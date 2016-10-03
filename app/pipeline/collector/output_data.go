package collector

import (
	"github.com/henrylee2cn/pholcus/logs"
)

var (
	// 全局支持的输出方式
	DataOutput = make(map[string]func(self *Collector) error)

	// 全局支持的文本数据输出方式名称列表
	DataOutputLib []string
)

// 文本数据输出
func (self *Collector) outputData() {
	defer func() {
		// 回收缓存块
		self.resetDataDocker()
	}()

	// 输出
	dataLen := uint64(len(self.dataDocker))
	if dataLen == 0 {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			logs.Log.Informational(" * ")
			logs.Log.App(" *     Panic  [数据输出：%v | KEYIN：%v | 批次：%v]   数据 %v 条！ [ERROR]  %v\n",
				self.Spider.GetName(), self.Spider.GetKeyin(), self.dataBatch, dataLen, p)
		}
	}()

	// 输出统计
	self.addDataSum(dataLen)

	// 执行输出
	err := DataOutput[self.outType](self)

	logs.Log.Informational(" * ")
	if err != nil {
		logs.Log.App(" *     Fail  [数据输出：%v | KEYIN：%v | 批次：%v]   数据 %v 条！ [ERROR]  %v\n",
			self.Spider.GetName(), self.Spider.GetKeyin(), self.dataBatch, dataLen, err)
	} else {
		logs.Log.App(" *     [数据输出：%v | KEYIN：%v | 批次：%v]   数据 %v 条！\n",
			self.Spider.GetName(), self.Spider.GetKeyin(), self.dataBatch, dataLen)
		self.Spider.TryFlushSuccess()
	}
}
