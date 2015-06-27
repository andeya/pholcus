package crawlpool

import (
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/crawl"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"time"
)

type CrawlPool interface {
	Reset(spiderNum int) int
	Use() crawl.Crawler
	Free(id int)
	Stop()
}

type cq struct {
	Cap    int
	Using  map[int]bool
	Src    []crawl.Crawler
	status int
}

func New() CrawlPool {
	return &cq{
		Src:    []crawl.Crawler{},
		status: status.RUN,
	}
}

// 根据要执行的蜘蛛数量设置CrawlerPool
// 在二次使用Pool实例时，可根据容量高效转换
func (self *cq) Reset(spiderNum int) int {
	num := config.CRAWLS_CAP
	if spiderNum < config.CRAWLS_CAP {
		num = spiderNum
	}

	last := len(self.Src)
	if num > last {
		self.Cap = num
		self.Using = make(map[int]bool, num)
	} else {
		self.Cap = last
		self.Using = make(map[int]bool, last)
	}
	self.status = status.RUN
	return num
}

func (self *cq) Use() crawl.Crawler {
	if self.status == status.STOP {
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

func (self *cq) Free(id int) {
	self.Using[id] = false
}

// 终止所有爬行任务
func (self *cq) Stop() {
	self.status = status.STOP
	self.Src = make([]crawl.Crawler, 0)
	self.Using = make(map[int]bool, self.Cap)
}

// 根据情况自动动态增加Crawl
func (self *cq) autoAdd() {
	count := len(self.Src)
	if count < self.Cap {
		self.Src = append(self.Src, crawl.New(count))
		self.Using[count] = false
	}
}
