package distribute

import (
	"encoding/json"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/teleport"
)

func ServerApi(n subApp) teleport.API {
	return teleport.API{
		// 提供任务给客户端
		"task": &task1{n},

		// 打印接收到的报告
		"log": new(log1),
	}
}

type task1 struct {
	subApp
}

func (self *task1) Process(receive *teleport.NetData) *teleport.NetData {
	b, _ := json.Marshal(self.Out(self.CountNodes()))
	return teleport.ReturnData(string(b))
}

type log1 struct{}

func (*log1) Process(receive *teleport.NetData) *teleport.NetData {
	logs.Log.Informational(" * ")
	logs.Log.Informational(" *     [ %s ]    %s", receive.From, receive.Body)
	logs.Log.Informational(" * ")
	return nil
}
