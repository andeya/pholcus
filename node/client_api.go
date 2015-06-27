package node

import (
	"encoding/json"
	"github.com/henrylee2cn/pholcus/node/task"
	. "github.com/henrylee2cn/teleport"
	"log"
)

var ClientApi = API{

	// 接收来自服务器的任务并加入任务库
	"task": func(receive *NetData) *NetData {
		d, err := json.Marshal(receive.Body)
		if err != nil {
			log.Println("json编码失败", receive.Body)
			return nil
		}
		t := &task.Task{}
		err = json.Unmarshal(d, t)
		if err != nil {
			log.Println("json解码失败", receive.Body)
			return nil
		}
		Pholcus.TaskJar.Into(t)
		return ReturnData(nil)
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
