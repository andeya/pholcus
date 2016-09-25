package teleport

import (
	"encoding/json"
	"github.com/henrylee2cn/teleport/debug"
	"log"
	"net"
	"time"
)

// 服务器专有成员
type tpServer struct {
	// 服务器模式下，缓存监听对象
	listener net.Listener
}

// 启动服务器模式，端口默认为常量DEFAULT_PORT
func (self *TP) Server(port ...string) {
	self.reserveAPI()
	self.mode = SERVER
	// 设置端口
	if len(port) > 0 {
		self.port = port[0]
	} else {
		self.port = DEFAULT_PORT
	}
	// 服务器UID默认为常量DEFAULT_SERVER_UID
	if self.uid == "" {
		self.uid = DEFAULT_SERVER_UID
	}
	if self.timeout == 0 {
		// 默认心跳间隔时长
		self.timeout = DEFAULT_TIMEOUT_S
	}
	go self.apiHandle()
	go self.server()
}

// ***********************************************功能实现*************************************************** \\

// 以服务器模式启动
func (self *TP) server() {
	var err error
retry:
	self.listener, err = net.Listen("tcp", self.port)
	if err != nil {
		debug.Printf("Debug: 监听端口出错: %v", err)
		time.Sleep(LOOP_TIMEOUT)
		goto retry
	}

	log.Printf(" *     —— 已开启服务器监听 (port %v) ——", self.port)

	for self.listener != nil {
		// 等待下一个连接,如果没有连接,listener.Accept会阻塞
		conn, err := self.listener.Accept()
		if err != nil {
			return
		}
		debug.Printf("Debug: 客户端 %v 已连接，但尚未验证身份！", conn.RemoteAddr().String())

		// 开启该连接处理协程(读写两条协程)
		self.sGoConn(conn)
	}
}

// 为每个连接开启读写两个协程
func (self *TP) sGoConn(conn net.Conn) {
	remoteAddr, connect := NewConnect(conn, self.connBufferLen, self.connWChanCap)
	// 初始化节点
	nodeuid, ok := self.sInitConn(connect, remoteAddr)
	if !ok {
		conn.Close()
		return
	}

	// 开启读写双工协程
	go self.sReader(nodeuid)
	go self.sWriter(nodeuid)
}

// 连接初始化，绑定节点与连接，默认key为节点ip
func (self *TP) sInitConn(conn *Connect, remoteAddr string) (nodeuid string, usable bool) {
	read_len, err := conn.Read(conn.Buffer)
	if err != nil || read_len == 0 {
		return
	}
	// 解包
	conn.TmpBuffer = append(conn.TmpBuffer, conn.Buffer[:read_len]...)
	dataSlice := make([][]byte, 10)
	dataSlice, conn.TmpBuffer = self.Unpack(conn.TmpBuffer)

	for i, data := range dataSlice {
		debug.Println("Debug: 收到数据-第1批-解码前: ", string(data))

		d := new(NetData)
		json.Unmarshal(data, d)
		// 修复缺失请求方地址的请求
		if d.From == "" {
			d.From = remoteAddr // 或可为：strings.Split(remoteAddr, ":")[0]
		}

		if i == 0 {
			debug.Printf("Debug: 收到数据-第1条-NetData: %+v", d)

			// 检查连接权限
			if !self.checkRights(d, remoteAddr) {
				return
			}

			nodeuid = d.From

			// 添加连接到节点池
			self.connPool[nodeuid] = conn

			// 判断是否为短链接
			if d.Operation != IDENTITY {
				conn.Short = true
			} else {
				log.Printf(" *     —— 客户端 %v (%v) 连接成功 ——", nodeuid, remoteAddr)
			}

			// 标记连接已经正式生效可用
			conn.Usable = true
		}
		// 添加到读取缓存
		self.apiReadChan <- d
	}
	return nodeuid, true
}

// 服务器读数据
func (self *TP) sReader(nodeuid string) {
	// 退出时关闭连接，删除连接池中的连接
	defer func() {
		self.closeConn(nodeuid, false)
	}()

	var conn = self.getConn(nodeuid)

	for conn != nil {
		// 设置连接超时
		if !conn.Short {
			conn.SetReadDeadline(time.Now().Add(self.timeout))
		}
		// 等待读取数据
		if !self.read(conn) {
			return
		}
	}
}

// 服务器发送数据
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
