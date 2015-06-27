package crawl

import (
	// "fmt"
	"github.com/henrylee2cn/pholcus/crawl/downloader"
	"github.com/henrylee2cn/pholcus/crawl/downloader/context"
	"github.com/henrylee2cn/pholcus/crawl/pipeline"
	"github.com/henrylee2cn/pholcus/crawl/scheduler"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/spider"
	"log"
	"math/rand"
	"sync"
	"time"
)

type crawler struct {
	id int
	*spider.Spider
	downloader.Downloader
	pipeline.Pipeline
	srcManage [2]uint
}

func New(id int) Crawler {
	return &crawler{
		id:         id,
		Pipeline:   pipeline.New(),
		Downloader: downloader.NewSurfer(0),
		srcManage:  [2]uint{},
	}
}

func (self *crawler) Init(sp *spider.Spider) Crawler {
	self.Pipeline.Init(sp)
	self.Spider = sp
	self.Downloader = downloader.NewSurfer(
		time.Duration((self.Spider.Pausetime[1]+self.Spider.Pausetime[0])/2)*time.Millisecond,
		self.Spider.Proxy,
	)
	self.srcManage = [2]uint{}
	return self
}

// 任务执行入口
func (self *crawler) Start() {

	// 预先开启输出管理协程
	self.Pipeline.Start()

	// 开始运行
	self.Spider.Start(self.Spider)
	self.Run()
	// log.Printf("**************爬虫：%v***********", self.GetId())
	// 通知输出模块输出未输出的数据
	self.Pipeline.CtrlR()
	// log.Println("**************断点 11 ***********")
}

func (self *crawler) Run() {
	for {
		// 队列中取出一条请求
		req := self.GetOne()

		// 队列退出及空请求调控
		if req == nil {
			if self.canStop() {
				// log.Println("**************退出队列************")
				break
			} else {
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}

		// 自身资源统计
		self.RequestIn()

		// 全局统计下载页面数
		cache.ReqSum++

		go func(req *context.Request) {
			defer func() {
				self.FreeOne()
				self.RequestOut()
			}()
			log.Println("start crawl :", req.GetUrl())
			self.Process(req)
		}(req)
	}
}

// core processer
func (self *crawler) Process(req *context.Request) {

	defer func() {
		if err := recover(); err != nil { // do not affect other
			if strerr, ok := err.(string); ok {
				log.Println(strerr)
			} else {
				log.Println("Process error：", err)
			}
		}
	}()
	// log.Println("**************断点 1 ***********")
	// download page
	resp := self.Downloader.Download(req)

	// log.Println("**************断点 2 ***********")
	if !resp.IsSucc() { // if fail do not need process
		log.Println(resp.Errormsg())
		return
	}

	// log.Println("**************断点 3 ***********")
	// 过程处理，提炼数据
	self.Spider.GoRule(resp)
	// log.Println("**************断点 5 ***********")
	// 该条请求结果存入pipeline
	datas := resp.GetItems()
	for i, count := 0, len(datas); i < count; i++ {
		self.Pipeline.Collect(
			resp.GetRuleName(), //DataCell.RuleName
			datas[i],           //DataCell.Data
			resp.GetUrl(),      //DataCell.Url
			resp.GetReferer(),  //DataCell.ParentUrl
			time.Now().Format("2006-01-02 15:04:05"),
		)
	}
	// log.Println("**************断点 end ***********")
}

// 常用基础方法
func (self *crawler) sleep() {
	sleeptime := rand.Intn(int(self.Spider.Pausetime[1])) + int(self.Spider.Pausetime[0])
	time.Sleep(time.Duration(sleeptime) * time.Millisecond)
}

// 从调度读取一个请求
func (self *crawler) GetOne() *context.Request {
	return scheduler.Sdl.Use(self.Spider.GetId())
}

//从调度释放一个资源空位
func (self *crawler) FreeOne() {
	scheduler.Sdl.Free()
}

func (self *crawler) RequestIn() {
	self.srcManage[0]++
}

var requestOutMutex sync.Mutex

func (self *crawler) RequestOut() {
	requestOutMutex.Lock()
	defer func() {
		requestOutMutex.Unlock()
	}()
	self.srcManage[1]++
}

//判断调度中是否还有属于自己的资源运行
func (self *crawler) canStop() bool {
	// log.Println("**************", self.srcManage[0], self.srcManage[1], "***********")
	return (self.srcManage[0] == self.srcManage[1] && scheduler.Sdl.IsEmpty(self.Spider.GetId())) || scheduler.Sdl.IsStop()
}

func (self *crawler) SetId(id int) {
	self.id = id
}

func (self *crawler) GetId() int {
	return self.id
}
