// 结果收集与输出
package collector

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/henrylee2cn/pholcus/app/pipeline/collector/data"
	"github.com/henrylee2cn/pholcus/app/spider"
	// "github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

// 结果收集与输出
type Collector struct {
	*spider.Spider                    //绑定的采集规则
	*DockerQueue                      //分批输出结果的缓存块队列
	DataChan       chan data.DataCell //文本数据收集通道
	FileChan       chan data.FileCell //文件收集通道
	stopFlag       chan bool          //长度为零时退出并输出
	outType        string             //输出方式
	// size     [2]uint64 //数据总输出流量统计[文本，文件]，文本暂时未统计
	dataBatch   uint64 //当前文本输出批次
	fileBatch   uint64 //当前文件输出批次
	wait        sync.WaitGroup
	sum         [4]uint64 //收集的数据总数[上次输出后文本总数，本次输出后文本总数，上次输出后文件总数，本次输出后文件总数]，非并发安全
	dataSumLock sync.RWMutex
	fileSumLock sync.RWMutex
}

func NewCollector(sp *spider.Spider) *Collector {
	var self = &Collector{}
	var queueCap = cache.Task.DockerQueueCap
	if cache.Task.DockerQueueCap < 2 {
		queueCap = 2
	}
	self.Spider = sp
	self.outType = cache.Task.OutType
	self.DataChan = make(chan data.DataCell, queueCap)
	self.FileChan = make(chan data.FileCell, queueCap)
	self.DockerQueue = newDockerQueue(queueCap)
	self.stopFlag = make(chan bool, 1)
	self.sum = [4]uint64{}
	// self.size = [2]uint64{}
	self.dataBatch = 0
	self.fileBatch = 0
	return self
}

func (self *Collector) CollectData(dataCell data.DataCell) error {
	var err error
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("输出协程已终止")
		}
	}()
	self.DataChan <- dataCell
	return err
}

func (self *Collector) CollectFile(fileCell data.FileCell) error {
	var err error
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("输出协程已终止")
		}
	}()
	self.FileChan <- fileCell
	return err
}

// 是否已发出停止命令
func (self *Collector) beStopping() bool {
	return len(self.stopFlag) == 0
}

// 停止
func (self *Collector) Stop(forced bool) {
	defer func() {
		recover()
	}()
	if forced {
		go func() {
			defer func() {
				recover()
			}()
			close(self.DataChan)
		}()
		go func() {
			defer func() {
				recover()
			}()
			close(self.FileChan)
		}()
	}
	<-self.stopFlag
	close(self.stopFlag)
}

// 启动数据收集/输出管道
func (self *Collector) Start() {
	// 标记程序已启动
	self.stopFlag <- true
	// 启动输出协程
	go func() {
		dataStop := make(chan bool)
		fileStop := make(chan bool)

		go func() {
			defer func() {
				recover()
				// println("DataChanStop$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")

			}()
			// 只有当收到退出通知并且通道内无数据时，才退出循环
			for !(self.beStopping() && len(self.DataChan) == 0) {
				select {
				case data := <-self.DataChan:
					// 追加数据
					self.Dockers[self.Curr()] = append(self.Dockers[self.Curr()], data)

					// 未达到设定的分批量时，仅缓存
					if len(self.Dockers[self.Curr()]) < cache.Task.DockerCap {
						continue
					}

					// 执行输出
					atomic.AddUint64(&self.dataBatch, 1)
					self.wait.Add(1)
					go self.outputData(self.Curr())
					// if self.beStopping() {
					// println("self.DockerQueue.Change()^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^datalen", len(self.DataChan))
					// }
					// 更换一个空Docker用于curDocker
					self.DockerQueue.Change()
					// if self.beStopping() {
					// println("self.DockerQueue.Change()$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$datalen", len(self.DataChan))
					// }
				default:
					time.Sleep(400 * time.Millisecond)
				}
			}
			// 将剩余收集到但未输出的数据输出
			atomic.AddUint64(&self.dataBatch, 1)
			self.wait.Add(1)
			self.outputData(self.Curr())
			close(dataStop)

			close(self.DataChan)
		}()

		go func() {
			defer func() {
				recover()
				// println("FileChanStop$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			}()
			// 只有当收到退出通知并且通道内无数据时，才退出循环
			for !(self.beStopping() && len(self.FileChan) == 0) {
				select {
				case file := <-self.FileChan:
					atomic.AddUint64(&self.fileBatch, 1)
					self.wait.Add(1)
					go self.outputFile(file)

				default:
					time.Sleep(400 * time.Millisecond)
				}
			}
			close(fileStop)

			close(self.FileChan)
		}()

		<-dataStop
		<-fileStop
		// println("OutputWaitStopping++++++++++++++++++++++++++++++++")

		// 等待所有输出完成
		self.wait.Wait()
		// println("OutputStopped$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")

		// 返回报告
		self.Report()
	}()
}

// 获取文本数据总量
func (self *Collector) dataSum() uint64 {
	self.dataSumLock.RLock()
	defer self.dataSumLock.RUnlock()
	return self.sum[1]
}

// 更新文本数据总量
func (self *Collector) addDataSum(add uint64) {
	self.dataSumLock.Lock()
	defer self.dataSumLock.Unlock()
	self.sum[0] = self.sum[1]
	self.sum[1] += add
}

// 获取文件数据总量
func (self *Collector) fileSum() uint64 {
	self.fileSumLock.RLock()
	defer self.fileSumLock.RUnlock()
	return self.sum[3]
}

// 更新文件数据总量
func (self *Collector) addFileSum(add uint64) {
	self.fileSumLock.Lock()
	defer self.fileSumLock.Unlock()
	self.sum[2] = self.sum[3]
	self.sum[3] += add
}

// // 获取文本输出流量
// func (self *Collector) dataSize() uint64 {
// 	return self.size[0]
// }

// // 更新文本输出流量记录
// func (self *Collector) addDataSize(add uint64) {
// 	self.size[0] += add
// }

// // 获取文件输出流量
// func (self *Collector) fileSize() uint64 {
// 	return self.size[1]
// }

// // 更新文本输出流量记录
// func (self *Collector) addFileSize(add uint64) {
// 	self.size[1] += add
// }

// 返回报告
func (self *Collector) Report() {
	cache.ReportChan <- &cache.Report{
		SpiderName: self.Spider.GetName(),
		Keyin:      self.GetKeyin(),
		DataNum:    self.dataSum(),
		FileNum:    self.fileSum(),
		// DataSize:   self.dataSize(),
		// FileSize: self.fileSize(),
		Time: time.Since(cache.StartTime),
	}
}
