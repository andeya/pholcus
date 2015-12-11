// 数据收集
package pipeline

import (
	"io"

	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"github.com/henrylee2cn/pholcus/app/spider"
)

type Pipeline interface {
	Start()
	//接收控制通知
	CtrlR()
	//控制通知
	CtrlW()
	// 收集数据单元
	CollectData(ruleName string, data map[string]interface{}, url string, parentUrl string, downloadTime string)
	// 收集文件
	CollectFile(ruleName, name string, body io.ReadCloser)
	// 重置
	Init(*spider.Spider)
}

type pipeline struct {
	*collector.Collector
}

func New() Pipeline {
	return &pipeline{
		Collector: collector.NewCollector(),
	}
}

func (self *pipeline) CollectData(ruleName string, data map[string]interface{}, url string, parentUrl string, downloadTime string) {
	self.Collector.CollectData(collector.NewDataCell(ruleName, data, url, parentUrl, downloadTime))
}

func (self *pipeline) CollectFile(ruleName, name string, body io.ReadCloser) {
	self.Collector.CollectFile(collector.NewFileCell(ruleName, name, body))
}

func (self *pipeline) Init(sp *spider.Spider) {
	self.Collector.Init(sp)
}

func (self *pipeline) Start() {
	go self.Collector.Manage()
}
