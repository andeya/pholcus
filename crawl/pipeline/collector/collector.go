// 数据收集
package collector

import (
	"fmt"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/spider"
	"strconv"
	"time"
)

// 每个爬取任务的数据容器
type Collector struct {
	*spider.Spider
	*DockerQueue
	DataChan chan DataCell
	ctrl     chan bool //长度为零时退出并输出
	sum      [2]int    //收集的数据总数[过去，现在],非并发安全
	outType  string
	outCount [2]int
}

func NewCollector() *Collector {
	self := &Collector{
		DataChan:    make(chan DataCell, config.DATA_CAP),
		DockerQueue: NewDockerQueue(),
		ctrl:        make(chan bool, 1),
	}
	return self
}

func (self *Collector) Init(sp *spider.Spider) {
	self.Spider = sp
	self.outType = cache.Task.OutType
	self.DataChan = make(chan DataCell, config.DATA_CAP)
	self.DockerQueue = NewDockerQueue()
	self.ctrl = make(chan bool, 1)
	self.sum = [2]int{}
	self.outCount = [2]int{}
}

func (self *Collector) Collect(dataCell DataCell) {
	// reporter.Log.Println("**************断点 6 ***********")
	self.DataChan <- dataCell
	// reporter.Log.Println("**************断点 7 ***********")
}

func (self *Collector) CtrlS() {
	self.ctrl <- true
	// reporter.Log.Println("**************断点 10 ***********")
}

func (self *Collector) CtrlR() {
	<-self.ctrl
	// reporter.Log.Println("**************断点 9 ***********")
}

func (self *Collector) CtrlLen() int {
	return len(self.ctrl)
}

// 数据转储输出
func (self *Collector) Manage() {
	// reporter.Log.Println("**************开启输出管道************")

	// 令self.Ctrl长度不为零
	self.CtrlS()
	// 只有当收到退出通知并且通道内无数据时，才退出循环
	for !(self.CtrlLen() == 0 && len(self.DataChan) == 0) {
		// reporter.Log.Println("**************断点 8 ***********")
		select {
		case data := <-self.DataChan:

			self.dockerOne(data)
		default:
			time.Sleep(1e7) //0.1秒
		}
	}

	// 将剩余收集到但未输出的数据输出
	self.goOutput(self.Curr)

	// 等待所有输出完成
	for self.outCount[0] > self.outCount[1] {
		time.Sleep(5e8)
	}

	// 返回报告
	self.Report()
}

func (self *Collector) dockerOne(data DataCell) {

	self.Dockers[self.Curr] = append(self.Dockers[self.Curr], data)

	if uint(len(self.Dockers[self.Curr])) >= cache.Task.DockerCap {
		// curDocker存满后输出
		self.goOutput(self.Curr)
		// 更换一个空Docker用于curDocker
		self.Change()
	}
}

func (self *Collector) goOutput(dataIndex int) {
	self.outCount[0]++
	go func() {
		self.Output(dataIndex)
		self.outCount[1]++
	}()
}

// 统计数据总量
func (self *Collector) Sum() int {
	return self.sum[1]
}

// 统计数据总量
func (self *Collector) setSum(add int) {
	self.sum[0], self.sum[1] = self.sum[1], self.sum[1]+add
}

// 返回报告
func (self *Collector) Report() {
	// reporter.Log.Println("**************", self.Sum(), " ***********")
	cache.ReportChan <- &cache.Report{
		SpiderName: self.Spider.GetName(),
		Keyword:    self.GetKeyword(),
		Num:        strconv.Itoa(self.Sum()),
		Time:       fmt.Sprintf("%.5f", time.Since(cache.StartTime).Minutes()),
	}
}
