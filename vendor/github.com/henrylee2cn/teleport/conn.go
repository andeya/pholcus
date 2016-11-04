package teleport

import (
	"net"
)

// 封装连接
type Connect struct {
	// 标准包conn接口实例，继承该接口所有方法
	net.Conn
	// 标记连接是否有效
	Usable bool
	// 是否为短链接模式
	Short bool
	// 专用写入数据缓存通道
	WriteChan chan *NetData
	// 从连接循环接收数据
	Buffer []byte
	// 临时缓冲区，用来存储被截断的数据
	TmpBuffer []byte
}

// 创建Connect实例，默认为长连接（Short=false）
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

// 返回远程节点地址
func (self *Connect) Addr() string {
	return self.Conn.RemoteAddr().String()
}
