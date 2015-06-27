package node

import (
	"github.com/henrylee2cn/pholcus/node/task"
	. "github.com/henrylee2cn/teleport"
	"log"
	"time"
)

var ServerApi = API{

	// 提供任务给客户端
	"task": func(receive *NetData) *NetData {
		var t task.Task
		var ok bool
		for {
			if t, ok = Pholcus.Out(receive.From, Pholcus.CountNodes()); ok {
				break
			}
			time.Sleep(1e9)
		}
		return ReturnData(t)
	},

	// 打印接收到的报告
	"log": func(receive *NetData) *NetData {
		log.Println(` ********************************************************************************************************************************************** `)
		log.Printf(" * ")
		log.Printf(" *     客户端 [ %s ]    %s", receive.From, receive.Body)
		log.Printf(" * ")
		log.Println(` ********************************************************************************************************************************************** `)
		return nil
	},
}
