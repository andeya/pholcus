// 数据收集
package collector

import (
	"fmt"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"time"
)

// 每个爬取任务的数据容器
type Collector struct {
	*spider.Spider
	*DockerQueue
	DataChan  chan DataCell
	FileChan  chan FileCell
	ctrl      chan bool //长度为零时退出并输出
	startTime time.Time
	outType   string
	sum       [3]uint //收集的数据总数[文本过去，文本现在，文件],非并发安全
	outCount  [4]uint //[文本输出开始，文本输出结束，文件输出开始，文件输出结束]
}

func NewCollector() *Collector {
	self := &Collector{
		DataChan:    make(chan DataCell, config.DATA_CAP),
		FileChan:    make(chan FileCell, 512),
		DockerQueue: NewDockerQueue(),
		ctrl:        make(chan bool, 1),
	}
	return self
}

func (self *Collector) Init(sp *spider.Spider) {
	self.Spider = sp
	self.outType = cache.Task.OutType
	self.DataChan = make(chan DataCell, config.DATA_CAP)
	self.FileChan = make(chan FileCell, 512)
	self.DockerQueue = NewDockerQueue()
	self.ctrl = make(chan bool, 1)
	self.sum = [3]uint{}
	self.outCount = [4]uint{}
	self.startTime = cache.StartTime
}

func (self *Collector) CollectData(dataCell DataCell) {
	// reporter.Log.Println("**************断点 6 ***********")
	self.DataChan <- dataCell
	// reporter.Log.Println("**************断点 7 ***********")
}

func (self *Collector) CollectFile(fileCell FileCell) {
	self.FileChan <- fileCell
}

func (self *Collector) CtrlW() {
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

	// 标记开始，令self.Ctrl长度不为零
	self.CtrlW()

	// 开启文件输出协程
	go self.SaveFile()

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
	for (self.outCount[0] > self.outCount[1]) || (self.outCount[2] > self.outCount[3]) || len(self.FileChan) > 0 {
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

// 获取文本数据总量
func (self *Collector) dataSum() uint {
	return self.sum[1]
}

// 更新文本数据总量
func (self *Collector) setDataSum(add uint) {
	self.sum[0], self.sum[1] = self.sum[1], self.sum[1]+add
}

// 获取文件数据总量
func (self *Collector) fileSum() uint {
	return self.sum[2]
}

// 更新文件数据总量
func (self *Collector) setFileSum(add uint) {
	self.sum[2] = self.sum[2] + add
}

// 返回报告
func (self *Collector) Report() {
	// reporter.Log.Println("**************", self.Sum(), " ***********")
	cache.ReportChan <- &cache.Report{
		SpiderName: self.Spider.GetName(),
		Keyword:    self.GetKeyword(),
		DataNum:    self.dataSum(),
		FileNum:    self.fileSum(),
		Time:       fmt.Sprintf("%.5f", time.Since(cache.StartTime).Minutes()),
	}
}
