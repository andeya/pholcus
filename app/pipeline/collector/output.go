//数据输出管理
package collector

import (
	"time"

	"github.com/henrylee2cn/pholcus/logs"
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
			logs.Log.App(" *     Panic  [任务输出：%v | KEYIN：%v | 批次：%v]   数据 %v 条，用时 %v！ [ERROR]  %v\n", self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[1]+1, dataLen, time.Since(self.timing), err)
			self.timing = time.Now()
		}
	}()

	// 输出数据统计
	self.setDataSum(uint64(dataLen))

	// 执行输出
	err := Output[self.outType](self, dataIndex)

	logs.Log.Informational(" * ")
	if err != nil {
		logs.Log.App(" *     Fail  [任务输出：%v | KEYIN：%v | 批次：%v]   数据 %v 条，用时 %v！ [ERROR]  %v\n", self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[1]+1, dataLen, time.Since(self.timing), err)
	} else {
		logs.Log.App(" *     [任务输出：%v | KEYIN：%v | 批次：%v]   数据 %v 条，用时 %v！\n", self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[1]+1, dataLen, time.Since(self.timing))
		self.Spider.TryFlushSuccess()
	}
	// 更新计时
	self.timing = time.Now()
}
