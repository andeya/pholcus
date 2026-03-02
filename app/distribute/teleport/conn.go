package teleport

import (
	"net"
)

// Connect wraps a network connection.
type Connect struct {
	net.Conn
	Usable    bool
	Short     bool
	WriteChan chan *NetData
	Buffer    []byte
	TmpBuffer []byte
}

// NewConnect creates a Connect instance; defaults to long connection (Short=false).
func NewConnect(conn net.Conn, bufferLen int, wChanCap int) (k string, v *Connect) {
	k = conn.RemoteAddr().String()

	v = &Connect{
		WriteChan: make(chan *NetData, wChanCap),
		Buffer:    make([]byte, bufferLen),
		TmpBuffer: make([]byte, 0),
		Conn:      conn,
	}
	return k, v
}

// Addr returns the remote node address.
func (conn *Connect) Addr() string {
	return conn.Conn.RemoteAddr().String()
}
