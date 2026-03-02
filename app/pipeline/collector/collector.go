// Package collector implements result collection and output.
package collector

import (
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/pipeline/collector/data"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/cache"
)

// Collector collects spider results and writes them to the configured output backend.
type Collector struct {
	*spider.Spider
	DataChan    chan data.DataCell
	FileChan    chan data.FileCell
	dataDocker  []data.DataCell
	outType     string
	dataBatch   uint64
	fileBatch   uint64
	wait        sync.WaitGroup
	sum         [4]uint64
	dataSumLock sync.RWMutex
	fileSumLock sync.RWMutex
}

// NewCollector creates a new Collector for the given spider.
func NewCollector(sp *spider.Spider) *Collector {
	var self = &Collector{}
	self.Spider = sp
	self.outType = cache.Task.OutType
	if cache.Task.DockerCap < 1 {
		cache.Task.DockerCap = 1
	}
	self.DataChan = make(chan data.DataCell, cache.Task.DockerCap)
	self.FileChan = make(chan data.FileCell, cache.Task.DockerCap)
	self.dataDocker = make([]data.DataCell, 0, cache.Task.DockerCap)
	self.sum = [4]uint64{}
	// self.size = [2]uint64{}
	self.dataBatch = 0
	self.fileBatch = 0
	return self
}

// CollectData sends a data cell to the collector.
func (self *Collector) CollectData(dataCell data.DataCell) (r result.VoidResult) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error("panic recovered: %v\n%s", p, debug.Stack())
			r = result.FmtErrVoid("输出协程已终止")
		}
	}()
	self.DataChan <- dataCell
	return result.OkVoid()
}

// CollectFile sends a file cell to the collector.
func (self *Collector) CollectFile(fileCell data.FileCell) (r result.VoidResult) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error("panic recovered: %v\n%s", p, debug.Stack())
			r = result.FmtErrVoid("输出协程已终止")
		}
	}()
	self.FileChan <- fileCell
	return result.OkVoid()
}

// Stop closes the collector's channels and shuts down the pipeline.
func (self *Collector) Stop() {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				logs.Log.Error("panic recovered: %v\n%s", p, debug.Stack())
			}
		}()
		close(self.DataChan)
	}()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				logs.Log.Error("panic recovered: %v\n%s", p, debug.Stack())
			}
		}()
		close(self.FileChan)
	}()
}

// Start launches the data collection and output pipeline.
func (self *Collector) Start() {
	go func() {
		dataStop := make(chan bool)
		fileStop := make(chan bool)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					logs.Log.Error("panic recovered: %v\n%s", p, debug.Stack())
				}
			}()
			for data := range self.DataChan {
				self.dataDocker = append(self.dataDocker, data)

				if len(self.dataDocker) < cache.Task.DockerCap {
					continue
				}

				self.dataBatch++
				self.outputData()
			}
			self.dataBatch++
			self.outputData()
			close(dataStop)
		}()

		go func() {
			defer func() {
				if p := recover(); p != nil {
					logs.Log.Error("panic recovered: %v\n%s", p, debug.Stack())
				}
			}()
			for file := range self.FileChan {
				atomic.AddUint64(&self.fileBatch, 1)
				self.wait.Add(1)
				go self.outputFile(file)
			}
			close(fileStop)
		}()

		<-dataStop
		<-fileStop

		self.wait.Wait()

		self.Report()
	}()
}

func (self *Collector) resetDataDocker() {
	for _, cell := range self.dataDocker {
		data.PutDataCell(cell)
	}
	self.dataDocker = self.dataDocker[:0]
}

// dataSum returns the total number of text records output.
func (self *Collector) dataSum() uint64 {
	self.dataSumLock.RLock()
	defer self.dataSumLock.RUnlock()
	return self.sum[1]
}

// addDataSum increments the text record count.
func (self *Collector) addDataSum(add uint64) {
	self.dataSumLock.Lock()
	defer self.dataSumLock.Unlock()
	self.sum[0] = self.sum[1]
	self.sum[1] += add
}

// fileSum returns the total number of files output.
func (self *Collector) fileSum() uint64 {
	self.fileSumLock.RLock()
	defer self.fileSumLock.RUnlock()
	return self.sum[3]
}

// addFileSum increments the file count.
func (self *Collector) addFileSum(add uint64) {
	self.fileSumLock.Lock()
	defer self.fileSumLock.Unlock()
	self.sum[2] = self.sum[3]
	self.sum[3] += add
}

// Report sends the collection report to the report channel.
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
