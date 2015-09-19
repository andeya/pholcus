// 数据收集
package pipeline

import (
	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/common/deduplicate"
	"io"
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
	// 对比Url的fingerprint，返回是否有重复
	Deduplicate(string) bool
	// 重置
	Init(*spider.Spider)
}

type pipeline struct {
	*collector.Collector
	*deduplicate.Deduplication
}

func New() Pipeline {
	return &pipeline{
		Collector:     collector.NewCollector(),
		Deduplication: deduplicate.New().(*deduplicate.Deduplication),
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

func (self *pipeline) Deduplicate(s string) bool {
	return self.Deduplication.Compare(s)
}

func (self *pipeline) Start() {
	go self.Collector.Manage()
}
