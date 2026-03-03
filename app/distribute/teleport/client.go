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
func (tp *TP) Client(serverAddr string, port string, isShort ...bool) {
	if len(isShort) > 0 && isShort[0] {
		tp.tpClient.short = true
	} else if tp.timeout == 0 {
		tp.timeout = DEFAULT_TIMEOUT_C
	}
	if tp.tpClient.serverUID == "" {
		tp.tpClient.serverUID = DEFAULT_SERVER_UID
	}
	tp.reserveAPI()
	tp.mode = CLIENT

	if port != "" {
		tp.port = port
	} else {
		tp.port = DEFAULT_PORT
	}

	tp.serverAddr = serverAddr

	tp.tpClient.mustClose = false

	go tp.apiHandle()
	go tp.client()
}

// --- Client implementation ---

func (tp *TP) client() {
	if !tp.short {
		log.Println(" *     -- Connecting to server... --")
	}

RetryLabel:
	connRes := result.Ret(net.Dial("tcp", tp.serverAddr+tp.port))
	if connRes.IsErr() {
		if tp.tpClient.mustClose {
			tp.tpClient.mustClose = false
			return
		}
		time.Sleep(LOOP_TIMEOUT)
		goto RetryLabel
	}
	conn := connRes.Unwrap()
	debugPrintf("Debug: connected to server: %v", conn.RemoteAddr().String())
	tp.cGoConn(conn)

	if !tp.short {
		for tp.CountNodes() > 0 {
			time.Sleep(LOOP_TIMEOUT)
		}
		if _, ok := tp.connPool[tp.tpClient.serverUID]; ok {
			goto RetryLabel
		}
	}
}

// cGoConn starts read/write goroutines for the connection.
func (tp *TP) cGoConn(conn net.Conn) {
	remoteAddr, connect := NewConnect(conn, tp.connBufferLen, tp.connWChanCap)

	tp.connPool[tp.tpClient.serverUID] = connect

	if tp.uid == "" {
		tp.uid = conn.LocalAddr().String()
	}

	if !tp.short {
		tp.send(NewNetData(tp.uid, tp.tpClient.serverUID, IDENTITY, "", nil))
		log.Printf(" *     -- Connected to server: %v --", remoteAddr)
	} else {
		connect.Short = true
	}

	tp.connPool[tp.tpClient.serverUID].Usable = true
	go tp.cReader(tp.tpClient.serverUID)
	go tp.cWriter(tp.tpClient.serverUID)
}

// cReader reads data on the client side.
func (tp *TP) cReader(nodeuid string) {
	defer func() {
		tp.closeConn(nodeuid, true)
	}()

	var conn = tp.getConn(nodeuid)

	for {
		if !tp.read(conn) {
			break
		}
	}
}

// cWriter sends data on the client side.
func (tp *TP) cWriter(nodeuid string) {
	defer func() {
		tp.closeConn(nodeuid, true)
	}()

	var conn = tp.getConn(nodeuid)

	for conn != nil {
		if tp.short {
			tp.send(<-conn.WriteChan)
			continue
		}

		timing := time.After(tp.timeout)
		data := new(NetData)
		select {
		case data = <-conn.WriteChan:
		case <-timing:
			data = NewNetData(tp.uid, nodeuid, HEARTBEAT, "", nil)
		}

		tp.send(data)
	}
}
