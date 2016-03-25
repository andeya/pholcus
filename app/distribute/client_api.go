package distribute

import (
	"encoding/json"

	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/teleport"
)

func ClientApi(n subApp) teleport.API {
	return teleport.API{
		// 接收来自服务器的任务并加入任务库
		"task": &task2{n},

		// 打印接收到的报告
		// "log": new(log2),
	}
}

type task2 struct {
	subApp
}

func (self *task2) Process(receive *teleport.NetData) *teleport.NetData {
	t := &Task{}
	err := json.Unmarshal([]byte(receive.Body.(string)), t)
	if err != nil {
		logs.Log.Error("json解码失败 %v", receive.Body)
		return nil
	}
	self.Into(t)
	return nil
}

// type log2 struct{}

// func (*log2) Process(receive *teleport.NetData) *teleport.NetData {
// 	logs.Log.Informational(" * ")
// 	logs.Log.Informational(" *     [ %s ]    %s", receive.From, receive.Body)
// 	logs.Log.Informational(" * ")
// 	return nil
// }
