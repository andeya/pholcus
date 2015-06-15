package crawler

import (
	"time"
)

const (
	STOP = iota
	RUN
)

type CrawlerQueue struct {
	Cap    uint
	Using  map[int]bool
	Src    []Crawler
	status int
}

var CQ = NewCrawlerQueue()

func NewCrawlerQueue() *CrawlerQueue {
	crawlQueue := &CrawlerQueue{
		Src:    make([]Crawler, 0),
		status: STOP,
	}
	return crawlQueue
}

// 在二次使用Queue实例时，可根据容量高效转换
func (self *CrawlerQueue) Init(num uint) {
	last := uint(len(self.Src))
	if num > last {
		self.Cap = num
		self.Using = make(map[int]bool, num)
	} else {
		self.Cap = last
		self.Using = make(map[int]bool, last)
	}
	self.status = RUN
}

func (self *CrawlerQueue) Free(id int) {
	self.Using[id] = false
}

func (self *CrawlerQueue) Use() Crawler {
	if self.status == STOP {
		return nil
	}
	for {
		for k, v := range self.Src {
			if !self.Using[k] {
				self.Using[k] = true
				return v
			}
		}
		self.autoAdd()
		time.Sleep(5e8)
	}
	return nil
}

// 根据情况自动动态增加Crawl
func (self *CrawlerQueue) autoAdd() {
	count := len(self.Src)
	if uint(count) < self.Cap {
		self.Src = append(self.Src, New(count))
		self.Using[count] = false
	}
}

// 终止所有爬行任务
func (self *CrawlerQueue) Stop() {
	self.status = STOP
	self.Src = make([]Crawler, 0)
	self.Using = make(map[int]bool, self.Cap)
}
