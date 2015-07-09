package node

import (
	. "github.com/henrylee2cn/teleport"
	"log"
	"sync"
)

var taskMutex sync.Mutex
var ServerApi = API{

	// 提供任务给客户端
	"task": func(receive *NetData) *NetData {
		taskMutex.Lock()
		defer taskMutex.Unlock()
		return ReturnData(Pholcus.Out(Pholcus.CountNodes()))
	},

	// 打印接收到的报告
	"log": func(receive *NetData) *NetData {
		log.Printf(" * ")
		log.Printf(" *     [ %s ]    %s", receive.From, receive.Body)
		log.Printf(" * ")
		return nil
	},
}
