package collector

import (
	"time"

	"github.com/henrylee2cn/pholcus/logs"
)

var (
	// 全局支持的输出方式
	DataOutput = make(map[string]func(self *Collector, dataIndex int) error)

	// 全局支持的文本数据输出方式名称列表
	DataOutputLib []string
)

// 文本数据输出
func (self *Collector) outputData() {
	// 开始输出的计数
	self.outCount[0]++

	go func(dataIndex int) {
		defer func() {
			// 回收缓存块
			self.DockerQueue.Recover(dataIndex)
			// 输出完成的计数
			self.outCount[1]++
		}()

		// 输出
		dataLen := uint64(len(self.DockerQueue.Dockers[dataIndex]))
		if dataLen == 0 {
			return
		}

		defer func() {
			if p := recover(); p != nil {
				logs.Log.Informational(" * ")
				logs.Log.App(" *     Panic  [数据输出：%v | KEYIN：%v | 批次：%v]   数据 %v 条，用时 %v！ [ERROR]  %v\n",
					self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[1]+1, dataLen, time.Since(self.timing), p)

				self.timing = time.Now()
			}
		}()

		// 输出统计
		self.addDataSum(dataLen)

		// 执行输出
		err := DataOutput[self.outType](self, dataIndex)

		logs.Log.Informational(" * ")
		if err != nil {
			logs.Log.App(" *     Fail  [数据输出：%v | KEYIN：%v | 批次：%v]   数据 %v 条，用时 %v！ [ERROR]  %v\n",
				self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[1]+1, dataLen, time.Since(self.timing), err)
		} else {
			logs.Log.App(" *     [数据输出：%v | KEYIN：%v | 批次：%v]   数据 %v 条，用时 %v！\n",
				self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[1]+1, dataLen, time.Since(self.timing))
			self.Spider.TryFlushSuccess()
		}

		// 更新计时
		self.timing = time.Now()

	}(self.Curr)
}
