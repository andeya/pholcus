// 数据收集
package pipeline

import (
	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"github.com/henrylee2cn/pholcus/app/pipeline/collector/data"
	"github.com/henrylee2cn/pholcus/app/spider"
)

type (
	Pipeline interface {
		Start()
		//接收控制通知
		CtrlR()
		//控制通知
		CtrlW()
		// 收集数据单元
		CollectData(data.DataCell)
		// 收集文件
		CollectFile(data.FileCell)
		// 重置
		Init(*spider.Spider)
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
	go self.Collector.Manage()
}
