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

// mode
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

func (self *TP) SetUID(mine string, server ...string) Teleport {
	if len(server) > 0 {
		self.tpClient.serverUID = server[0]
	}
	self.uid = mine
	return self
}

// SetAPI sets the application API.
func (self *TP) SetAPI(api API) Teleport {
	self.api = api
	return self
}

// Request pushes data; blocks until a connection exists; empty nodeuid sends to a random node.
func (self *TP) Request(body interface{}, operation string, flag string, nodeuid ...string) {
	var conn *Connect
	var uid string
	if len(nodeuid) == 0 {
		for {
			if self.CountNodes() > 0 {
				break
			}
			time.Sleep(LOOP_TIMEOUT)
		}
		for uid, conn = range self.connPool {
			if conn.Usable {
				nodeuid = append(nodeuid, uid)
				break
			}
		}
	}
	conn = self.getConn(nodeuid[0])
	for conn == nil || !conn.Usable {
		conn = self.getConn(nodeuid[0])
		time.Sleep(LOOP_TIMEOUT)
	}
	conn.WriteChan <- NewNetData(self.uid, nodeuid[0], operation, flag, body)
}

// Close disconnects; empty nodeuid closes all; in server mode also stops listening.
func (self *TP) Close(nodeuid ...string) {
	if self.mode == CLIENT {
		self.tpClient.mustClose = true

	} else if self.mode == SERVER && self.tpServer.listener != nil {
		self.tpServer.listener.Close()
		log.Printf(" *     —— Server stopped listening on %v ——", self.port)
	}

	if len(nodeuid) == 0 {
		for uid, conn := range self.connPool {
			delete(self.connPool, uid)
			conn.Close()
			self.closeMsg(uid, conn.Addr(), conn.Short)
		}
		return
	}

	for _, uid := range nodeuid {
		conn := self.connPool[uid]
		delete(self.connPool, uid)
		conn.Close()
		self.closeMsg(uid, conn.Addr(), conn.Short)
	}
}

// SetPackHeader sets the packet header string.
func (self *TP) SetPackHeader(header string) Teleport {
	self.Protocol.ReSet(header)
	return self
}

// SetApiRChan sets the global receive channel length.
func (self *TP) SetApiRChan(length int) Teleport {
	self.apiReadChan = make(chan *NetData, length)
	return self
}

// SetConnWChan sets per-connection write channel length.
func (self *TP) SetConnWChan(length int) Teleport {
	self.connWChanCap = length
	return self
}

// SetConnBuffer sets per-connection receive buffer size.
func (self *TP) SetConnBuffer(length int) Teleport {
	self.connBufferLen = length
	return self
}

// SetTimeout sets connection timeout (heartbeat interval).
func (self *TP) SetTimeout(long time.Duration) Teleport {
	self.timeout = long
	return self
}

// GetMode returns run mode.
func (self *TP) GetMode() int {
	return self.mode
}

// CountNodes returns the number of active connections.
func (self *TP) CountNodes() int {
	count := 0
	for _, conn := range self.connPool {
		if conn != nil && conn.Usable {
			count++
		}
	}
	return count
}

func (self *TP) read(conn *Connect) bool {
	readLen, err := conn.Read(conn.Buffer)
	if result.TryErrVoid(err).IsErr() || readLen == 0 {
		return false
	}
	conn.TmpBuffer = append(conn.TmpBuffer, conn.Buffer[:readLen]...)
	self.save(conn)
	return true
}

// getConn returns the connection for the given node UID.
func (self *TP) getConn(nodeuid string) *Connect {
	return self.connPool[nodeuid]
}

// getConnAddr returns the address of the connection for the given node UID.
func (self *TP) getConnAddr(nodeuid string) string {
	conn := self.getConn(nodeuid)
	if conn == nil {
		return ""
	}
	return conn.Addr()
}

// closeConn closes the connection and exits the goroutine.
func (self *TP) closeConn(nodeuid string, reconnect bool) {
	conn, ok := self.connPool[nodeuid]
	if !ok {
		return
	}

	if reconnect {
		self.connPool[nodeuid] = nil
	} else {
		delete(self.connPool, nodeuid)
	}

	if conn == nil {
		return
	}
	conn.Close()
	self.closeMsg(nodeuid, conn.Addr(), conn.Short)
}

// closeMsg logs connection close.
func (self *TP) closeMsg(uid, addr string, short bool) {
	if short {
		return
	}
	switch self.mode {
	case SERVER:
		log.Printf(" *     —— Disconnected from client %v (%v) ——", uid, addr)
	case CLIENT:
		log.Printf(" *     —— Disconnected from server %v ——", addr)
	}
}

// send encodes and sends data.
func (self *TP) send(data *NetData) {
	if data.From == "" {
		data.From = self.uid
	}

	d := result.Ret(json.Marshal(*data)).UnwrapOrElse(func(err error) []byte {
		debugPrintln("Debug: send data encode error", err)
		return nil
	})
	if d == nil {
		return
	}
	conn := self.getConn(data.To)
	if conn == nil {
		debugPrintf("Debug: send data connection closed: %+v", data)
		return
	}
	end := self.Packet(d)
	conn.Write(end)
	debugPrintf("Debug: send data success: %+v", data)
}

// save decodes received data and stores it in the cache.
func (self *TP) save(conn *Connect) {
	debugPrintf("Debug: received data bytes: %v", conn.TmpBuffer)
	dataSlice := make([][]byte, 10)
	dataSlice, conn.TmpBuffer = self.Unpack(conn.TmpBuffer)

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
		self.apiReadChan <- d
		debugPrintf("Debug: received data NetData: %+v", d)
	}
}

// apiHandle processes requests concurrently via the API.
func (self *TP) apiHandle() {
	for {
		req := <-self.apiReadChan
		go func(req *NetData) {
			var conn *Connect

			operation, from, to, flag := req.Operation, req.To, req.From, req.Flag
			handle, ok := self.api[operation]

			if !ok {
				if self.mode == SERVER {
					self.autoErrorHandle(req, LLLEGAL, "Server ("+self.getConn(to).LocalAddr().String()+") has no API: "+req.Operation, to)
					log.Printf("Client %v (%v) requesting non-existent API: %v", to, self.getConnAddr(to), req.Operation)
				} else {
					self.autoErrorHandle(req, LLLEGAL, "Client "+from+" ("+self.getConn(to).LocalAddr().String()+") has no API: "+req.Operation, to)
					log.Printf("Server (%v) requesting non-existent API: %v", self.getConnAddr(to), req.Operation)
				}
				return
			}

			resp := handle.Process(req)
			if resp == nil {
				if conn = self.getConn(to); conn != nil && self.getConn(to).Short {
					self.closeConn(to, false)
				}
				return //continue
			}

			if resp.To == "" {
				resp.To = to
			}

			if conn = self.getConn(resp.To); conn == nil {
				self.autoErrorHandle(req, FAILURE, "", to)
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

func (self *TP) autoErrorHandle(data *NetData, status int, msg string, reqFrom string) bool {
	oldConn := self.getConn(reqFrom)
	if oldConn == nil {
		return false
	}
	respErr := ReturnError(data, status, msg)
	respErr.From = self.uid
	respErr.To = reqFrom
	oldConn.WriteChan <- respErr
	return true
}

// checkRights validates connection permissions.
func (self *TP) checkRights(data *NetData, addr string) bool {
	if data.To != self.uid {
		log.Printf("Unknown connection (%v) provided wrong server identifier, request rejected", addr)
		return false
	}
	return true
}

// reserveAPI sets system-reserved API handlers.
func (self *TP) reserveAPI() {
	self.api[IDENTITY] = identi
	self.api[HEARTBEAT] = beat
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
