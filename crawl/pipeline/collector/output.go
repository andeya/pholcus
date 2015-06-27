//数据输出管理
package collector

import (
	. "github.com/henrylee2cn/pholcus/reporter"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"log"
	"time"
)

func (self *Collector) Output(dataIndex int) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("输出时出错！\n")
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
	self.setSum(dataLen)

	// 选择执行输出
	switch self.outType {
	case "excel":
		self.excel(dataIndex)
	case "csv":
		self.csv(dataIndex)
	case "mongoDB":
		self.mgo(dataIndex)
	default:
		self.excel(dataIndex)
	}

	log.Println(` ********************************************************************************************************************************************** `)
	log.Printf(" * ")
	Log.Printf(" *     [任务：%v | 关键词：%v | 批次：%v]   输出 %v 条数据，用时 %.5f 分钟！\n", self.Spider.GetName(), self.Spider.GetKeyword(), self.outCount[1]+1, dataLen, time.Since(cache.StartTime).Minutes())
	log.Printf(" * ")
	log.Println(` ********************************************************************************************************************************************** `)

}
