package node

import (
	. "github.com/henrylee2cn/teleport"
	"log"
)

func ServerApi(n *Node) API {
	return API{
		// 提供任务给客户端
		"task": &task1{n},

		// 打印接收到的报告
		"log": new(log1),
	}
}

type task1 struct {
	*Node
}

func (self *task1) Process(receive *NetData) *NetData {
	return ReturnData(self.Out(self.CountNodes()))
}

type log1 struct{}

func (*log1) Process(receive *NetData) *NetData {
	log.Printf(" * ")
	log.Printf(" *     [ %s ]    %s", receive.From, receive.Body)
	log.Printf(" * ")
	return nil
}
