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
	dataBuf     []data.DataCell
	outType     string
	batchCap    int
	dataBatch   uint64
	fileBatch   uint64
	wait        sync.WaitGroup
	sum         [4]uint64
	dataSumLock sync.RWMutex
	fileSumLock sync.RWMutex
}

// NewCollector creates a new Collector for the given spider.
func NewCollector(sp *spider.Spider, outType string, batchCap int) *Collector {
	if batchCap < 1 {
		batchCap = 1
	}
	return &Collector{
		Spider:   sp,
		outType:  outType,
		batchCap: batchCap,
		DataChan: make(chan data.DataCell, batchCap),
		FileChan: make(chan data.FileCell, batchCap),
		dataBuf:  make([]data.DataCell, 0, batchCap),
	}
}

// CollectData sends a data cell to the collector.
func (c *Collector) CollectData(dataCell data.DataCell) (r result.VoidResult) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
			r = result.FmtErrVoid("输出协程已终止")
		}
	}()
	c.DataChan <- dataCell
	return result.OkVoid()
}

// CollectFile sends a file cell to the collector.
func (c *Collector) CollectFile(fileCell data.FileCell) (r result.VoidResult) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
			r = result.FmtErrVoid("输出协程已终止")
		}
	}()
	c.FileChan <- fileCell
	return result.OkVoid()
}

// Stop closes the collector's channels and shuts down the pipeline.
func (c *Collector) Stop() {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
			}
		}()
		close(c.DataChan)
	}()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
			}
		}()
		close(c.FileChan)
	}()
}

// Start launches the data collection and output pipeline.
func (c *Collector) Start() {
	go func() {
		dataStop := make(chan bool)
		fileStop := make(chan bool)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
				}
			}()
			for data := range c.DataChan {
				c.dataBuf = append(c.dataBuf, data)

				if len(c.dataBuf) < c.batchCap {
					continue
				}

				c.dataBatch++
				c.outputData()
			}
			c.dataBatch++
			c.outputData()
			close(dataStop)
		}()

		go func() {
			defer func() {
				if p := recover(); p != nil {
					logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
				}
			}()
			for file := range c.FileChan {
				atomic.AddUint64(&c.fileBatch, 1)
				c.wait.Add(1)
				go c.outputFile(file)
			}
			close(fileStop)
		}()

		<-dataStop
		<-fileStop

		c.wait.Wait()

		c.Report()
	}()
}

func (c *Collector) resetDataBuf() {
	for _, cell := range c.dataBuf {
		data.PutDataCell(cell)
	}
	c.dataBuf = c.dataBuf[:0]
}

// dataSum returns the total number of text records output.
func (c *Collector) dataSum() uint64 {
	c.dataSumLock.RLock()
	defer c.dataSumLock.RUnlock()
	return c.sum[1]
}

// addDataSum increments the text record count.
func (c *Collector) addDataSum(add uint64) {
	c.dataSumLock.Lock()
	defer c.dataSumLock.Unlock()
	c.sum[0] = c.sum[1]
	c.sum[1] += add
}

// fileSum returns the total number of files output.
func (c *Collector) fileSum() uint64 {
	c.fileSumLock.RLock()
	defer c.fileSumLock.RUnlock()
	return c.sum[3]
}

// addFileSum increments the file count.
func (c *Collector) addFileSum(add uint64) {
	c.fileSumLock.Lock()
	defer c.fileSumLock.Unlock()
	c.sum[2] = c.sum[3]
	c.sum[3] += add
}

// Report sends the collection report to the report channel.
func (c *Collector) Report() {
	cache.ReportChan <- &cache.Report{
		SpiderName: c.Spider.GetName(),
		Keyin:      c.GetKeyin(),
		DataNum:    c.dataSum(),
		FileNum:    c.fileSum(),
		Time:       time.Since(cache.StartTime),
	}
}
