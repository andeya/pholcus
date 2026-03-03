// Package teleport provides a high-concurrency API framework for distributed systems.
// It uses socket duplex communication for peer-to-peer S/C, supports long and short connections,
// auto-reconnect after disconnect, and JSON for data transport.
package teleport

import (
	"encoding/json"
	"log"
	"time"

	"github.com/andeya/gust/result"
)

// Run mode constants.
const (
	SERVER = iota + 1
	CLIENT
)

// Reserved operation names for API handlers.
const (
	IDENTITY            = "+identity+"
	HEARTBEAT           = "+heartbeat+"
	DEFAULT_PACK_HEADER = "andeya"
	DEFAULT_SERVER_UID  = "server"
	DEFAULT_PORT        = ":8080"
	DEFAULT_TIMEOUT_S   = 20e9
	DEFAULT_TIMEOUT_C   = 15e9
	LOOP_TIMEOUT        = 1e9
)

type Teleport interface {
	Server(port ...string)
	Client(serverAddr string, port string, isShort ...bool)
	Request(body interface{}, operation string, flag string, nodeuid ...string)
	SetAPI(api API) Teleport
	Close(nodeuid ...string)

	SetUID(mine string, server ...string) Teleport
	SetPackHeader(string) Teleport
	SetApiRChan(int) Teleport
	SetConnWChan(int) Teleport
	SetConnBuffer(int) Teleport
	SetTimeout(time.Duration) Teleport

	GetMode() int
	CountNodes() int
}

type TP struct {
	uid        string
	mode       int
	port       string
	serverAddr string
	connPool   map[string]*Connect
	timeout    time.Duration
	*Protocol
	apiReadChan   chan *NetData
	connWChanCap  int
	connBufferLen int
	api           API
	*tpServer
	*tpClient
}

type API map[string]Handle

// Handle processes requests.
type Handle interface {
	Process(*NetData) *NetData
}

// New creates a Teleport instance.
func New() Teleport {
	return &TP{
		connPool:      make(map[string]*Connect),
		api:           API{},
		Protocol:      NewProtocol(DEFAULT_PACK_HEADER),
		apiReadChan:   make(chan *NetData, 4096),
		connWChanCap:  2048,
		connBufferLen: 1024,
		tpServer:      new(tpServer),
		tpClient:      new(tpClient),
	}
}

// --- Interface implementation ---

func (tp *TP) SetUID(mine string, server ...string) Teleport {
	if len(server) > 0 {
		tp.tpClient.serverUID = server[0]
	}
	tp.uid = mine
	return tp
}

// SetAPI sets the application API.
func (tp *TP) SetAPI(api API) Teleport {
	tp.api = api
	return tp
}

// Request pushes data; blocks until a connection exists; empty nodeuid sends to a random node.
func (tp *TP) Request(body interface{}, operation string, flag string, nodeuid ...string) {
	var conn *Connect
	var uid string
	if len(nodeuid) == 0 {
		for {
			if tp.CountNodes() > 0 {
				break
			}
			time.Sleep(LOOP_TIMEOUT)
		}
		for uid, conn = range tp.connPool {
			if conn.Usable {
				nodeuid = append(nodeuid, uid)
				break
			}
		}
	}
	conn = tp.getConn(nodeuid[0])
	for conn == nil || !conn.Usable {
		conn = tp.getConn(nodeuid[0])
		time.Sleep(LOOP_TIMEOUT)
	}
	conn.WriteChan <- NewNetData(tp.uid, nodeuid[0], operation, flag, body)
}

// Close disconnects; empty nodeuid closes all; in server mode also stops listening.
func (tp *TP) Close(nodeuid ...string) {
	if tp.mode == CLIENT {
		tp.tpClient.mustClose = true

	} else if tp.mode == SERVER && tp.tpServer.listener != nil {
		tp.tpServer.listener.Close()
		log.Printf(" *     -- Server stopped listening on %v --", tp.port)
	}

	if len(nodeuid) == 0 {
		uids := make([]string, 0, len(tp.connPool))
		for uid := range tp.connPool {
			uids = append(uids, uid)
		}
		for _, uid := range uids {
			conn := tp.connPool[uid]
			delete(tp.connPool, uid)
			if conn != nil {
				conn.Close()
				tp.closeMsg(uid, conn.Addr(), conn.Short)
			}
		}
		return
	}

	for _, uid := range nodeuid {
		conn := tp.connPool[uid]
		delete(tp.connPool, uid)
		if conn != nil {
			conn.Close()
			tp.closeMsg(uid, conn.Addr(), conn.Short)
		}
	}
}

// SetPackHeader sets the packet header string.
func (tp *TP) SetPackHeader(header string) Teleport {
	tp.Protocol.ReSet(header)
	return tp
}

// SetApiRChan sets the global receive channel length.
func (tp *TP) SetApiRChan(length int) Teleport {
	tp.apiReadChan = make(chan *NetData, length)
	return tp
}

// SetConnWChan sets per-connection write channel length.
func (tp *TP) SetConnWChan(length int) Teleport {
	tp.connWChanCap = length
	return tp
}

// SetConnBuffer sets per-connection receive buffer size.
func (tp *TP) SetConnBuffer(length int) Teleport {
	tp.connBufferLen = length
	return tp
}

// SetTimeout sets connection timeout (heartbeat interval).
func (tp *TP) SetTimeout(long time.Duration) Teleport {
	tp.timeout = long
	return tp
}

// GetMode returns run mode.
func (tp *TP) GetMode() int {
	return tp.mode
}

// CountNodes returns the number of active connections.
func (tp *TP) CountNodes() int {
	count := 0
	for _, conn := range tp.connPool {
		if conn != nil && conn.Usable {
			count++
		}
	}
	return count
}

func (tp *TP) read(conn *Connect) bool {
	readLen, err := conn.Read(conn.Buffer)
	if result.TryErrVoid(err).IsErr() || readLen == 0 {
		return false
	}
	conn.TmpBuffer = append(conn.TmpBuffer, conn.Buffer[:readLen]...)
	tp.save(conn)
	return true
}

// getConn returns the connection for the given node UID.
func (tp *TP) getConn(nodeuid string) *Connect {
	return tp.connPool[nodeuid]
}

// getConnAddr returns the address of the connection for the given node UID.
func (tp *TP) getConnAddr(nodeuid string) string {
	conn := tp.getConn(nodeuid)
	if conn == nil {
		return ""
	}
	return conn.Addr()
}

// closeConn closes the connection and exits the goroutine.
func (tp *TP) closeConn(nodeuid string, reconnect bool) {
	conn, ok := tp.connPool[nodeuid]
	if !ok {
		return
	}

	if reconnect {
		tp.connPool[nodeuid] = nil
	} else {
		delete(tp.connPool, nodeuid)
	}

	if conn == nil {
		return
	}
	conn.Close()
	tp.closeMsg(nodeuid, conn.Addr(), conn.Short)
}

// closeMsg logs connection close.
func (tp *TP) closeMsg(uid, addr string, short bool) {
	if short {
		return
	}
	switch tp.mode {
	case SERVER:
		log.Printf(" *     -- Disconnected from client %v (%v) --", uid, addr)
	case CLIENT:
		log.Printf(" *     -- Disconnected from server %v --", addr)
	}
}

// send encodes and sends data.
func (tp *TP) send(data *NetData) {
	if data.From == "" {
		data.From = tp.uid
	}

	d := result.Ret(json.Marshal(*data)).UnwrapOrElse(func(err error) []byte {
		debugPrintln("Debug: send data encode error", err)
		return nil
	})
	if d == nil {
		return
	}
	conn := tp.getConn(data.To)
	if conn == nil {
		debugPrintf("Debug: send data connection closed: %+v", data)
		return
	}
	end := tp.Packet(d)
	conn.Write(end)
	debugPrintf("Debug: send data success: %+v", data)
}

// save decodes received data and stores it in the cache.
func (tp *TP) save(conn *Connect) {
	debugPrintf("Debug: received data bytes: %v", conn.TmpBuffer)
	dataSlice := make([][]byte, 10)
	dataSlice, conn.TmpBuffer = tp.Unpack(conn.TmpBuffer)

	for _, data := range dataSlice {
		debugPrintf("Debug: received data before decode: %v", string(data))

		d := new(NetData)
		if r := result.RetVoid(json.Unmarshal(data, d)); r.IsErr() {
			debugPrintf("Debug: received data decode error: %v", r.UnwrapErr())
			continue
		}
		if d.From == "" {
			d.From = conn.Addr()
		}
		tp.apiReadChan <- d
		debugPrintf("Debug: received data NetData: %+v", d)
	}
}

// apiHandle processes requests concurrently via the API.
func (tp *TP) apiHandle() {
	for {
		req := <-tp.apiReadChan
		go func(req *NetData) {
			var conn *Connect

			operation, from, to, flag := req.Operation, req.To, req.From, req.Flag
			handle, ok := tp.api[operation]

			if !ok {
				peerUID := from
				peerConn := tp.getConn(peerUID)
				addrStr := ""
				if peerConn != nil {
					addrStr = peerConn.LocalAddr().String()
				}
				if tp.mode == SERVER {
					tp.autoErrorHandle(req, LLLEGAL, "Server ("+addrStr+") has no API: "+req.Operation, peerUID)
					log.Printf("Client %v (%v) requesting non-existent API: %v", from, tp.getConnAddr(peerUID), req.Operation)
				} else {
					tp.autoErrorHandle(req, LLLEGAL, "Client "+from+" ("+addrStr+") has no API: "+req.Operation, peerUID)
					log.Printf("Server (%v) requesting non-existent API: %v", tp.getConnAddr(peerUID), req.Operation)
				}
				return
			}

			resp := handle.Process(req)
			if resp == nil {
				if conn = tp.getConn(to); conn != nil && tp.getConn(to).Short {
					tp.closeConn(to, false)
				}
				return //continue
			}

			if resp.To == "" {
				resp.To = to
			}

			if conn = tp.getConn(resp.To); conn == nil {
				tp.autoErrorHandle(req, FAILURE, "", to)
				return
			}

			if resp.Operation == "" {
				resp.Operation = operation
			}

			if resp.From == "" {
				resp.From = from
			}

			if resp.Flag == "" {
				resp.Flag = flag
			}

			conn.WriteChan <- resp

		}(req)
	}
}

func (tp *TP) autoErrorHandle(data *NetData, status int, msg string, reqFrom string) bool {
	oldConn := tp.getConn(reqFrom)
	if oldConn == nil {
		return false
	}
	respErr := ReturnError(data, status, msg)
	respErr.From = tp.uid
	respErr.To = reqFrom
	oldConn.WriteChan <- respErr
	return true
}

// checkRights validates connection permissions.
func (tp *TP) checkRights(data *NetData, addr string) bool {
	if data.To != tp.uid {
		log.Printf("Unknown connection (%v) provided wrong server identifier, request rejected", addr)
		return false
	}
	return true
}

// reserveAPI sets system-reserved API handlers.
func (tp *TP) reserveAPI() {
	tp.api[IDENTITY] = identi
	tp.api[HEARTBEAT] = beat
}

var identi, beat = new(identity), new(heartbeat)

type identity struct{}

func (*identity) Process(receive *NetData) *NetData {
	return nil
}

type heartbeat struct{}

func (*heartbeat) Process(receive *NetData) *NetData {
	return nil
}
