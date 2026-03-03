package teleport

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/andeya/gust/result"
)

// tpServer holds server-only state.
type tpServer struct {
	listener net.Listener
}

// Server starts server mode; port defaults to DEFAULT_PORT.
func (tp *TP) Server(port ...string) {
	tp.reserveAPI()
	tp.mode = SERVER
	if len(port) > 0 {
		tp.port = port[0]
	} else {
		tp.port = DEFAULT_PORT
	}
	if tp.uid == "" {
		tp.uid = DEFAULT_SERVER_UID
	}
	if tp.timeout == 0 {
		tp.timeout = DEFAULT_TIMEOUT_S
	}
	go tp.apiHandle()
	go tp.server()
}

// --- Server implementation ---

func (tp *TP) server() {
retry:
	listenerRes := result.Ret(net.Listen("tcp", tp.port))
	if listenerRes.IsErr() {
		debugPrintf("Debug: listen port error: %v", listenerRes.UnwrapErr())
		time.Sleep(LOOP_TIMEOUT)
		goto retry
	}
	tp.listener = listenerRes.Unwrap()

	log.Printf(" *     -- Server listening (port %v) --", tp.port)

	for tp.listener != nil {
		connRes := result.Ret(tp.listener.Accept())
		if connRes.IsErr() {
			return
		}
		conn := connRes.Unwrap()
		debugPrintf("Debug: client %v connected, identity not yet verified", conn.RemoteAddr().String())
		tp.sGoConn(conn)
	}
}

// sGoConn starts read/write goroutines for each connection.
func (tp *TP) sGoConn(conn net.Conn) {
	remoteAddr, connect := NewConnect(conn, tp.connBufferLen, tp.connWChanCap)
	nodeuid, ok := tp.sInitConn(connect, remoteAddr)
	if !ok {
		conn.Close()
		return
	}

	go tp.sReader(nodeuid)
	go tp.sWriter(nodeuid)
}

// sInitConn initializes connection and binds node to conn; default key is node IP.
func (tp *TP) sInitConn(conn *Connect, remoteAddr string) (nodeuid string, usable bool) {
	readLen, err := conn.Read(conn.Buffer)
	if result.TryErrVoid(err).IsErr() || readLen == 0 {
		return
	}
	conn.TmpBuffer = append(conn.TmpBuffer, conn.Buffer[:readLen]...)
	dataSlice := make([][]byte, 10)
	dataSlice, conn.TmpBuffer = tp.Unpack(conn.TmpBuffer)

	for i, data := range dataSlice {
		debugPrintln("Debug: received data batch 1 before decode: ", string(data))

		d := new(NetData)
		if result.RetVoid(json.Unmarshal(data, d)).IsErr() {
			if i == 0 {
				return
			}
			continue
		}
		if d.From == "" {
			d.From = remoteAddr
		}

		if i == 0 {
			debugPrintf("Debug: received data item 1 NetData: %+v", d)
			if !tp.checkRights(d, remoteAddr) {
				return
			}

			nodeuid = d.From
			tp.connPool[nodeuid] = conn

			if d.Operation != IDENTITY {
				conn.Short = true
			} else {
				log.Printf(" *     -- Client %v (%v) connected --", nodeuid, remoteAddr)
			}
			conn.Usable = true
		}
		tp.apiReadChan <- d
	}
	return nodeuid, true
}

// sReader reads data on the server side.
func (tp *TP) sReader(nodeuid string) {
	defer func() {
		tp.closeConn(nodeuid, false)
	}()

	var conn = tp.getConn(nodeuid)

	for conn != nil {
		if !conn.Short {
			conn.SetReadDeadline(time.Now().Add(tp.timeout))
		}
		if !tp.read(conn) {
			return
		}
	}
}

// sWriter sends data on the server side.
func (tp *TP) sWriter(nodeuid string) {
	defer func() {
		tp.closeConn(nodeuid, false)
	}()

	var conn = tp.getConn(nodeuid)

	for conn != nil {
		data := <-conn.WriteChan
		tp.send(data)
		if conn.Short {
			return
		}
	}
}
