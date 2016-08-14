// 数据收集
package pipeline

import (
	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"github.com/henrylee2cn/pholcus/app/pipeline/collector/data"
	"github.com/henrylee2cn/pholcus/app/spider"
)

// 数据收集/输出管道
type (
	Pipeline interface {
		Start()                    //启动
		Stop()                     //停止
		CollectData(data.DataCell) //收集数据单元
		CollectFile(data.FileCell) //收集文件
		Init(*spider.Spider)       //重置
	}

	pipeline struct {
		*collector.Collector
	}
)

func New() Pipeline {
	return &pipeline{
		Collector: collector.NewCollector(),
	}
}

func (self *pipeline) CollectData(item data.DataCell) {
	self.Collector.CollectData(item)
}

func (self *pipeline) CollectFile(f data.FileCell) {
	self.Collector.CollectFile(f)
}

func (self *pipeline) Init(sp *spider.Spider) {
	self.Collector.Init(sp)
}

func (self *pipeline) Start() {
	go self.Collector.Start()
}
