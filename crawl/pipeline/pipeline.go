// 数据收集
package pipeline

import (
	"github.com/henrylee2cn/pholcus/common/deduplicate"
	"github.com/henrylee2cn/pholcus/crawl/pipeline/collector"
	// "github.com/henrylee2cn/pholcus/reporter"
	"github.com/henrylee2cn/pholcus/spider"
)

type Pipeline interface {
	Start()
	//接收控制通知
	CtrlR()
	//发送控制通知
	CtrlS()
	// 收集数据单元
	Collect(ruleName string, data map[string]interface{}, url string, parentUrl string, downloadTime string)
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

func (self *pipeline) Collect(ruleName string, data map[string]interface{}, url string, parentUrl string, downloadTime string) {
	dataCell := collector.NewDataCell(ruleName, data, url, parentUrl, downloadTime)
	self.Collector.Collect(dataCell)
}

func (self *pipeline) Init(sp *spider.Spider) {
	self.Collector.Init(sp)
}

func (self *pipeline) Deduplicate(s string) bool {
	return self.Deduplication.Compare(s)
}

func (self *pipeline) Start() {
	go self.Collector.Manage()
	// reporter.Log.Println("**************开启输出管道************")
}
