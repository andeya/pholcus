package distribute

import (
	"encoding/json"

	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/teleport"
)

// 创建从节点API
func SlaveApi(n Distributer) teleport.API {
	return teleport.API{
		// 接收来自服务器的任务并加入任务库
		"task": &slaveTaskHandle{n},
	}
}

// 从节点自动接收主节点任务的操作
type slaveTaskHandle struct {
	Distributer
}

func (self *slaveTaskHandle) Process(receive *teleport.NetData) *teleport.NetData {
	t := &Task{}
	err := json.Unmarshal([]byte(receive.Body.(string)), t)
	if err != nil {
		logs.Log.Error("json解码失败 %v", receive.Body)
		return nil
	}
	self.Receive(t)
	return nil
}
