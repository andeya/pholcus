package crawl

import (
	"sync"
	"time"

	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

type (
	CrawlPool interface {
		Reset(spiderNum int) int
		Use() Crawler
		Free(Crawler)
		Stop()
	}
	cq struct {
		Cap    int
		Src    map[Crawler]bool
		status int
		sync.Mutex
	}
)

func NewCrawlPool() CrawlPool {
	return &cq{
		Src:    make(map[Crawler]bool),
		status: status.RUN,
	}
}

// 根据要执行的蜘蛛数量设置CrawlerPool
// 在二次使用Pool实例时，可根据容量高效转换
func (self *cq) Reset(spiderNum int) int {
	var wantNum int
	if spiderNum < config.CRAWLS_CAP {
		wantNum = spiderNum
	} else {
		wantNum = config.CRAWLS_CAP
	}

	hasNum := len(self.Src)
	if wantNum > hasNum {
		self.Cap = wantNum
	} else {
		self.Cap = hasNum
	}
	self.status = status.RUN
	return self.Cap
}

// 并发安全地使用资源
func (self *cq) Use() Crawler {
	if self.status != status.RUN {
		return nil
	}
	for {
		for k, v := range self.Src {
			if !v {
				self.Lock()
				self.Src[k] = true
				self.Unlock()
				return k
			}
		}
		if len(self.Src) <= self.Cap {
			self.Lock()
			self.increment()
			self.Unlock()
		} else {
			time.Sleep(5e8)
		}
	}
	return nil
}

func (self *cq) Free(c Crawler) {
	self.Lock()
	self.Src[c] = false
	self.Unlock()
}

// 主动终止所有爬行任务
func (self *cq) Stop() {
	self.status = status.STOP
	for c := range self.Src {
		c.Stop()
	}
	self.Src = make(map[Crawler]bool)
}

// 根据情况自动动态增加Crawl
func (self *cq) increment() {
	id := len(self.Src)
	if id < self.Cap {
		self.Src[New(id)] = false
	}
}
