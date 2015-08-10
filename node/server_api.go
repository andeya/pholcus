package node

import (
	. "github.com/henrylee2cn/teleport"
	"log"
)

func ServerApi(n *Node) API {
	return API{
		// 提供任务给客户端
		"task": func(receive *NetData) *NetData {
			return ReturnData(n.Out(n.CountNodes()))
		},

		// 打印接收到的报告
		"log": func(receive *NetData) *NetData {
			log.Printf(" * ")
			log.Printf(" *     [ %s ]    %s", receive.From, receive.Body)
			log.Printf(" * ")
			return nil
		},
	}
}
