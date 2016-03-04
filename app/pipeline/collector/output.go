//数据输出管理
package collector

import (
	"time"

	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

var (
	Output    = make(map[string]func(self *Collector, dataIndex int) error)
	OutputLib []string
)

func (self *Collector) Output(dataIndex int) {
	defer func() {
		// 回收内存
		self.DockerQueue.Recover(dataIndex)
	}()

	dataLen := len(self.DockerQueue.Dockers[dataIndex])
	if dataLen == 0 {
		return
	}

	defer func() {
		err := recover()
		if err != nil {
			logs.Log.Informational(" * ")
			logs.Log.Notice(" *     Panic  [任务输出：%v | 关键词：%v | 批次：%v]   数据 %v 条，用时 %v！ [ERROR]  %v\n", self.Spider.GetName(), self.Spider.GetKeyword(), self.outCount[1]+1, dataLen, time.Since(self.timing), err)
			self.timing = time.Now()
		}
	}()

	// 输出数据统计
	self.setDataSum(uint64(dataLen))

	// 执行输出
	err := Output[self.outType](self, dataIndex)

	logs.Log.Informational(" * ")
	if err != nil {
		logs.Log.Notice(" *     Fail  [任务输出：%v | 关键词：%v | 批次：%v]   数据 %v 条，用时 %v！ [ERROR]  %v\n", self.Spider.GetName(), self.Spider.GetKeyword(), self.outCount[1]+1, dataLen, time.Since(self.timing), err)
	} else {
		logs.Log.Notice(" *     [任务输出：%v | 关键词：%v | 批次：%v]   数据 %v 条，用时 %v！\n", self.Spider.GetName(), self.Spider.GetKeyword(), self.outCount[1]+1, dataLen, time.Since(self.timing))
		// 非服务器模式下保存历史记录
		if cache.Task.Mode != status.SERVER {
			self.Spider.TryFlushHistory()
		}
	}
	// 更新计时
	self.timing = time.Now()
}
