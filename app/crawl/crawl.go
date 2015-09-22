package crawl

import (
	"github.com/henrylee2cn/pholcus/app/downloader"
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/app/pipeline"
	"github.com/henrylee2cn/pholcus/app/scheduler"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/runtime/cache"

	// "fmt"
	"github.com/henrylee2cn/pholcus/logs"
	"io"
	"math/rand"
	"sync"
	"time"
)

type Crawler interface {
	Init(*spider.Spider) Crawler
	Start()
	GetId() int
}

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
		Downloader: downloader.NewSurfer(false, 0),
		srcManage:  [2]uint{},
	}
}

func (self *crawler) Init(sp *spider.Spider) Crawler {
	self.Pipeline.Init(sp)
	self.Spider = sp
	self.Downloader.SetUseCookie(self.Spider.UseCookie)
	self.Downloader.SetPaseTime(time.Duration((self.Spider.Pausetime[1]+self.Spider.Pausetime[0])/2) * time.Millisecond)
	self.Downloader.SetProxy(self.Spider.Proxy)
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
	// logs.Log.Debug("**************爬虫：%v***********", self.GetId())
	// 通知输出模块输出未输出的数据
	self.Pipeline.CtrlR()
	// logs.Log.Debug("**************断点 11 ***********")
}

func (self *crawler) Run() {
	for {
		// 随机等待
		self.sleep()

		// 队列中取出一条请求
		req := self.GetOne()

		// 队列退出及空请求调控
		if req == nil {
			if self.canStop() {
				// logs.Log.Debug("**************退出队列************")
				break
			} else {
				continue
			}
		}

		// 自身资源统计
		self.RequestIn()

		// 全局统计下载总页数
		cache.PageCount()

		go func(req *context.Request) {
			defer func() {
				self.FreeOne()
				self.RequestOut()
			}()
			logs.Log.Informational(" *     crawl: %v", req.GetUrl())
			self.Process(req)
		}(req)
	}
}

// core processer
func (self *crawler) Process(req *context.Request) {

	defer func() {
		if err := recover(); err != nil { // do not affect other
			if strerr, ok := err.(string); ok {
				logs.Log.Error("%v", strerr)
			} else {
				logs.Log.Error(" *     Process error: %v", err)
			}
		}
	}()
	// logs.Log.Debug("**************断点 1 ***********")
	// download page
	resp := self.Downloader.Download(req)

	// logs.Log.Debug("**************断点 2 ***********")
	if resp.GetError() != nil { // if fail do not need process
		logs.Log.Error(" *     %v", resp.GetError())
		// 统计下载失败的页数
		cache.PageFailCount()
		return
	}

	// logs.Log.Debug("**************断点 3 ***********")
	// 过程处理，提炼数据
	self.Spider.ExecParse(resp)
	// logs.Log.Debug("**************断点 5 ***********")
	// 该条请求文本结果存入pipeline
	for _, data := range resp.GetItems() {
		self.Pipeline.CollectData(
			resp.GetRuleName(), //DataCell.RuleName
			data,               //DataCell.Data
			resp.GetUrl(),      //DataCell.Url
			resp.GetReferer(),  //DataCell.ParentUrl
			time.Now().Format("2006-01-02 15:04:05"),
		)
	}

	// 该条请求文件结果存入pipeline
	for _, img := range resp.GetFiles() {
		self.Pipeline.CollectFile(
			resp.GetRuleName(),
			img["Name"].(string),
			img["Body"].(io.ReadCloser),
		)
	}

	// logs.Log.Debug("**************断点 end ***********")
}

// 常用基础方法
func (self *crawler) sleep() {
	sleeptime := int(self.Spider.Pausetime[0])
	if self.Spider.Pausetime[1] > 0 {
		sleeptime += rand.New(rand.NewSource(time.Now().UnixNano())).Intn(int(self.Spider.Pausetime[1]))
	}
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
	// logs.Log.Debug("**************", self.srcManage[0], self.srcManage[1], "***********")
	return (self.srcManage[0] == self.srcManage[1] && scheduler.Sdl.IsEmpty(self.Spider.GetId())) || scheduler.Sdl.IsStop()
}

func (self *crawler) SetId(id int) {
	self.id = id
}

func (self *crawler) GetId() int {
	return self.id
}
