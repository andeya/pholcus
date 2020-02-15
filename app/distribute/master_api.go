package distribute

import (
	"encoding/json"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/teleport"
)

// 创建主节点API
func MasterApi(n Distributer) teleport.API {
	return teleport.API{
		// 分配任务给客户端
		"task": &masterTaskHandle{n},

		// 打印接收到的日志
		"log": &masterLogHandle{},
	}
}

// 主节点自动分配任务的操作
type masterTaskHandle struct {
	Distributer
}

func (self *masterTaskHandle) Process(receive *teleport.NetData) *teleport.NetData {
	b, _ := json.Marshal(self.Send(self.CountNodes()))
	return teleport.ReturnData(string(b))
}

// 主节点自动接收从节点消息并打印的操作
type masterLogHandle struct{}

func (*masterLogHandle) Process(receive *teleport.NetData) *teleport.NetData {
	logs.Log.Informational(" * ")
	logs.Log.Informational(" *     [ %s ]    %s", receive.From, receive.Body)
	logs.Log.Informational(" * ")
	return nil
}
