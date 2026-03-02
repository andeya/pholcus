package crawler

import (
	"sync"
	"time"

	"github.com/andeya/gust/option"
	"github.com/andeya/pholcus/app/downloader"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/runtime/status"
)

// CrawlerPool manages a pool of crawler engines.
type (
	CrawlerPool interface {
		Reset(spiderNum int) int
		SetPipelineConfig(outType string, batchCap int)
		Use() Crawler
		UseOpt() option.Option[Crawler]
		Free(Crawler)
		Stop()
	}
	cq struct {
		capacity int
		count    int
		usable   chan Crawler
		all      []Crawler
		dl       downloader.Downloader
		outType  string
		batchCap int
		status   int
		sync.RWMutex
	}
)

// NewCrawlerPool creates a new crawler pool with the given Downloader.
func NewCrawlerPool(dl downloader.Downloader) CrawlerPool {
	return &cq{
		status: status.RUN,
		dl:     dl,
		all:    make([]Crawler, 0, config.Conf().CrawlsCap),
	}
}

// SetPipelineConfig sets the output type and batch capacity for new crawlers.
func (cq *cq) SetPipelineConfig(outType string, batchCap int) {
	cq.Lock()
	defer cq.Unlock()
	cq.outType = outType
	cq.batchCap = batchCap
}

// Reset configures the pool size based on the number of spiders to run.
// When reusing a pool instance, it efficiently resizes to the new capacity.
func (cq *cq) Reset(spiderNum int) int {
	cq.Lock()
	defer cq.Unlock()
	var wantNum int
	if spiderNum < config.Conf().CrawlsCap {
		wantNum = spiderNum
	} else {
		wantNum = config.Conf().CrawlsCap
	}
	if wantNum <= 0 {
		wantNum = 1
	}
	cq.capacity = wantNum
	cq.count = 0
	cq.usable = make(chan Crawler, wantNum)
	for _, crawler := range cq.all {
		if cq.count < cq.capacity {
			cq.usable <- crawler
			cq.count++
		}
	}
	cq.status = status.RUN
	return wantNum
}

// Use acquires a crawler from the pool in a concurrency-safe manner.
func (cq *cq) Use() Crawler {
	return cq.UseOpt().UnwrapOr(nil)
}

// UseOpt acquires a crawler from the pool; returns None when pool is stopped.
func (cq *cq) UseOpt() option.Option[Crawler] {
	var crawler Crawler
	for {
		cq.Lock()
		if cq.status == status.STOP {
			cq.Unlock()
			return option.None[Crawler]()
		}
		select {
		case crawler = <-cq.usable:
			cq.Unlock()
			return option.Some(crawler)
		default:
			if cq.count < cq.capacity {
				crawler = New(cq.count, cq.dl, cq.outType, cq.batchCap)
				cq.all = append(cq.all, crawler)
				cq.count++
				cq.Unlock()
				return option.Some(crawler)
			}
		}
		cq.Unlock()
		time.Sleep(time.Second)
	}
}

// Free returns a crawler to the pool.
func (cq *cq) Free(crawler Crawler) {
	cq.RLock()
	defer cq.RUnlock()
	if cq.status == status.STOP || !crawler.CanStop() {
		return
	}
	cq.usable <- crawler
}

// Stop terminates all crawler tasks in the pool.
func (cq *cq) Stop() {
	cq.Lock()
	if cq.status == status.STOP {
		cq.Unlock()
		return
	}
	cq.status = status.STOP
	close(cq.usable)
	cq.usable = nil
	cq.Unlock()

	for _, crawler := range cq.all {
		crawler.Stop()
	}
}
