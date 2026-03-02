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
func (self *TP) Server(port ...string) {
	self.reserveAPI()
	self.mode = SERVER
	if len(port) > 0 {
		self.port = port[0]
	} else {
		self.port = DEFAULT_PORT
	}
	if self.uid == "" {
		self.uid = DEFAULT_SERVER_UID
	}
	if self.timeout == 0 {
		self.timeout = DEFAULT_TIMEOUT_S
	}
	go self.apiHandle()
	go self.server()
}

// --- Server implementation ---

func (self *TP) server() {
retry:
	listenerRes := result.Ret(net.Listen("tcp", self.port))
	if listenerRes.IsErr() {
		debugPrintf("Debug: listen port error: %v", listenerRes.UnwrapErr())
		time.Sleep(LOOP_TIMEOUT)
		goto retry
	}
	self.listener = listenerRes.Unwrap()

	log.Printf(" *     —— Server listening (port %v) ——", self.port)

	for self.listener != nil {
		connRes := result.Ret(self.listener.Accept())
		if connRes.IsErr() {
			return
		}
		conn := connRes.Unwrap()
		debugPrintf("Debug: client %v connected, identity not yet verified", conn.RemoteAddr().String())
		self.sGoConn(conn)
	}
}

// sGoConn starts read/write goroutines for each connection.
func (self *TP) sGoConn(conn net.Conn) {
	remoteAddr, connect := NewConnect(conn, self.connBufferLen, self.connWChanCap)
	nodeuid, ok := self.sInitConn(connect, remoteAddr)
	if !ok {
		conn.Close()
		return
	}

	go self.sReader(nodeuid)
	go self.sWriter(nodeuid)
}

// sInitConn initializes connection and binds node to conn; default key is node IP.
func (self *TP) sInitConn(conn *Connect, remoteAddr string) (nodeuid string, usable bool) {
	readLen, err := conn.Read(conn.Buffer)
	if result.TryErrVoid(err).IsErr() || readLen == 0 {
		return
	}
	conn.TmpBuffer = append(conn.TmpBuffer, conn.Buffer[:readLen]...)
	dataSlice := make([][]byte, 10)
	dataSlice, conn.TmpBuffer = self.Unpack(conn.TmpBuffer)

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
			if !self.checkRights(d, remoteAddr) {
				return
			}

			nodeuid = d.From
			self.connPool[nodeuid] = conn

			if d.Operation != IDENTITY {
				conn.Short = true
			} else {
				log.Printf(" *     —— Client %v (%v) connected ——", nodeuid, remoteAddr)
			}
			conn.Usable = true
		}
		self.apiReadChan <- d
	}
	return nodeuid, true
}

// sReader reads data on the server side.
func (self *TP) sReader(nodeuid string) {
	defer func() {
		self.closeConn(nodeuid, false)
	}()

	var conn = self.getConn(nodeuid)

	for conn != nil {
		if !conn.Short {
			conn.SetReadDeadline(time.Now().Add(self.timeout))
		}
		if !self.read(conn) {
			return
		}
	}
}

// sWriter sends data on the server side.
func (self *TP) sWriter(nodeuid string) {
	defer func() {
		self.closeConn(nodeuid, false)
	}()

	var conn = self.getConn(nodeuid)

	for conn != nil {
		data := <-conn.WriteChan
		self.send(data)
		if conn.Short {
			return
		}
	}
}
