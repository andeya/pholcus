//数据输出管理
package collector

import (
	"time"

	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

var Output = make(map[string]func(self *Collector, dataIndex int))
var OutputLib []string

func (self *Collector) Output(dataIndex int) {
	defer func() {
		err := recover()
		if err != nil {
			logs.Log.Error(" *     输出时出错: %v", err)
		} else {
			// 正常情况下回收内存
			self.DockerQueue.Recover(dataIndex)
		}
	}()

	dataLen := len(self.DockerQueue.Dockers[dataIndex])
	if dataLen == 0 {
		// log.Println("没有抓到结果！！！")
		return
	}

	// 输出数据统计
	self.setDataSum(uint(dataLen))

	// 执行输出
	Output[self.outType](self, dataIndex)

	logs.Log.Informational(" * ")
	logs.Log.Notice(" *     [任务：%v | 关键词：%v | 批次：%v]   输出 %v 条数据，用时 %.5f 分钟！\n", self.Spider.GetName(), self.Spider.GetKeyword(), self.outCount[1]+1, dataLen, time.Since(cache.StartTime).Minutes())
	logs.Log.Informational(" * ")
}
