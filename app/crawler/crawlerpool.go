package crawler

import (
	"sync"
	"time"

	"github.com/andeya/gust/option"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/runtime/status"
)

// CrawlerPool manages a pool of crawler engines.
type (
	CrawlerPool interface {
		Reset(spiderNum int) int
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
		status   int
		sync.RWMutex
	}
)

// NewCrawlerPool creates a new crawler pool.
func NewCrawlerPool() CrawlerPool {
	return &cq{
		status: status.RUN,
		all:    make([]Crawler, 0, config.CRAWLS_CAP),
	}
}

// Reset configures the pool size based on the number of spiders to run.
// When reusing a pool instance, it efficiently resizes to the new capacity.
func (self *cq) Reset(spiderNum int) int {
	self.Lock()
	defer self.Unlock()
	var wantNum int
	if spiderNum < config.CRAWLS_CAP {
		wantNum = spiderNum
	} else {
		wantNum = config.CRAWLS_CAP
	}
	if wantNum <= 0 {
		wantNum = 1
	}
	self.capacity = wantNum
	self.count = 0
	self.usable = make(chan Crawler, wantNum)
	for _, crawler := range self.all {
		if self.count < self.capacity {
			self.usable <- crawler
			self.count++
		}
	}
	self.status = status.RUN
	return wantNum
}

// Use acquires a crawler from the pool in a concurrency-safe manner.
func (self *cq) Use() Crawler {
	return self.UseOpt().UnwrapOr(nil)
}

// UseOpt acquires a crawler from the pool; returns None when pool is stopped.
func (self *cq) UseOpt() option.Option[Crawler] {
	var crawler Crawler
	for {
		self.Lock()
		if self.status == status.STOP {
			self.Unlock()
			return option.None[Crawler]()
		}
		select {
		case crawler = <-self.usable:
			self.Unlock()
			return option.Some(crawler)
		default:
			if self.count < self.capacity {
				crawler = New(self.count)
				self.all = append(self.all, crawler)
				self.count++
				self.Unlock()
				return option.Some(crawler)
			}
		}
		self.Unlock()
		time.Sleep(time.Second)
	}
}

// Free returns a crawler to the pool.
func (self *cq) Free(crawler Crawler) {
	self.RLock()
	defer self.RUnlock()
	if self.status == status.STOP || !crawler.CanStop() {
		return
	}
	self.usable <- crawler
}

// Stop terminates all crawler tasks in the pool.
func (self *cq) Stop() {
	self.Lock()
	if self.status == status.STOP {
		self.Unlock()
		return
	}
	self.status = status.STOP
	close(self.usable)
	self.usable = nil
	self.Unlock()

	for _, crawler := range self.all {
		crawler.Stop()
	}
}
