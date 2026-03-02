package teleport

import (
	"log"
	"net"
	"time"

	"github.com/andeya/gust/result"
)

// tpClient holds client-only state.
type tpClient struct {
	short     bool
	mustClose bool
	serverUID string
}

// Client starts client mode.
func (self *TP) Client(serverAddr string, port string, isShort ...bool) {
	if len(isShort) > 0 && isShort[0] {
		self.tpClient.short = true
	} else if self.timeout == 0 {
		self.timeout = DEFAULT_TIMEOUT_C
	}
	if self.tpClient.serverUID == "" {
		self.tpClient.serverUID = DEFAULT_SERVER_UID
	}
	self.reserveAPI()
	self.mode = CLIENT

	if port != "" {
		self.port = port
	} else {
		self.port = DEFAULT_PORT
	}

	self.serverAddr = serverAddr

	self.tpClient.mustClose = false

	go self.apiHandle()
	go self.client()
}

// --- Client implementation ---

func (self *TP) client() {
	if !self.short {
		log.Println(" *     —— Connecting to server... ——")
	}

RetryLabel:
	connRes := result.Ret(net.Dial("tcp", self.serverAddr+self.port))
	if connRes.IsErr() {
		if self.tpClient.mustClose {
			self.tpClient.mustClose = false
			return
		}
		time.Sleep(LOOP_TIMEOUT)
		goto RetryLabel
	}
	conn := connRes.Unwrap()
	debugPrintf("Debug: connected to server: %v", conn.RemoteAddr().String())
	self.cGoConn(conn)

	if !self.short {
		for self.CountNodes() > 0 {
			time.Sleep(LOOP_TIMEOUT)
		}
		if _, ok := self.connPool[self.tpClient.serverUID]; ok {
			goto RetryLabel
		}
	}
}

// cGoConn starts read/write goroutines for the connection.
func (self *TP) cGoConn(conn net.Conn) {
	remoteAddr, connect := NewConnect(conn, self.connBufferLen, self.connWChanCap)

	self.connPool[self.tpClient.serverUID] = connect

	if self.uid == "" {
		self.uid = conn.LocalAddr().String()
	}

	if !self.short {
		self.send(NewNetData(self.uid, self.tpClient.serverUID, IDENTITY, "", nil))
		log.Printf(" *     —— Connected to server: %v ——", remoteAddr)
	} else {
		connect.Short = true
	}

	self.connPool[self.tpClient.serverUID].Usable = true
	go self.cReader(self.tpClient.serverUID)
	go self.cWriter(self.tpClient.serverUID)
}

// cReader reads data on the client side.
func (self *TP) cReader(nodeuid string) {
	defer func() {
		self.closeConn(nodeuid, true)
	}()

	var conn = self.getConn(nodeuid)

	for {
		if !self.read(conn) {
			break
		}
	}
}

// cWriter sends data on the client side.
func (self *TP) cWriter(nodeuid string) {
	defer func() {
		self.closeConn(nodeuid, true)
	}()

	var conn = self.getConn(nodeuid)

	for conn != nil {
		if self.short {
			self.send(<-conn.WriteChan)
			continue
		}

		timing := time.After(self.timeout)
		data := new(NetData)
		select {
		case data = <-conn.WriteChan:
		case <-timing:
			data = NewNetData(self.uid, nodeuid, HEARTBEAT, "", nil)
		}

		self.send(data)
	}
}
