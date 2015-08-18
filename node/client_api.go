package node

import (
	"encoding/json"
	"github.com/henrylee2cn/pholcus/node/task"
	. "github.com/henrylee2cn/teleport"
	"log"
)

func ClientApi(n *Node) API {
	return API{
		// 接收来自服务器的任务并加入任务库
		"task": &task2{n},

		// 打印接收到的报告
		"log": new(log2),
	}
}

type task2 struct {
	*Node
}

func (self *task2) Process(receive *NetData) *NetData {
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
	self.TaskJar.Into(t)
	return nil
}

type log2 struct{}

func (*log2) Process(receive *NetData) *NetData {
	log.Printf(" * ")
	log.Printf(" *     [ %s ]    %s", receive.From, receive.Body)
	log.Printf(" * ")
	return nil
}
