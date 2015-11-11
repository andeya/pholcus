package crawl

import (
	"github.com/henrylee2cn/pholcus/app/downloader"
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/app/pipeline"
	"github.com/henrylee2cn/pholcus/app/scheduler"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"

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
		Downloader: downloader.SurferDownloader,
		srcManage:  [2]uint{},
	}
}

func (self *crawler) Init(sp *spider.Spider) Crawler {
	self.Pipeline.Init(sp)
	self.Spider = sp
	self.srcManage = [2]uint{}
	return self
}

// 任务执行入口
func (self *crawler) Start() {
	// 预先开启输出管理协程
	self.Pipeline.Start()

	// 开始运行
	self.Spider.Start()
	self.Run()
	// logs.Log.Debug("**************爬虫：%v***********", self.GetId())
	// 通知输出模块输出未输出的数据
	self.Pipeline.CtrlR()
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
				// 停止任务
				return
			} else {
				// 继续等待请求
				continue
			}
		}

		// 自身资源统计
		self.RequestIn()

		go func(req *context.Request) {
			defer func() {
				self.FreeOne()
				self.RequestOut()
			}()
			// logs.Log.Informational(" *     start: %v", req.GetUrl())
			self.Process(req)
		}(req)
	}
}

// core processer
func (self *crawler) Process(req *context.Request) {
	defer func() {
		if err := recover(); err != nil {
			// do not affect other
			scheduler.Sdl.DelDeduplication(req.GetUrl() + req.GetMethod())
			// 统计失败数
			cache.PageFailCount()
			// 提示错误
			logs.Log.Error(" *     Fail [process panic]: %v", err)
		}
	}()
	// download page
	resp := self.Downloader.Download(req)

	// if fail do not need process
	if resp.GetError() != nil {
		// 删除该请求的去重样本
		scheduler.Sdl.DelDeduplication(req.GetUrl() + req.GetMethod())
		// 统计失败数
		cache.PageFailCount()
		// 提示错误
		logs.Log.Error(" *     Fail [download]: %v", resp.GetError())
		return
	}

	// 过程处理，提炼数据
	spider.NewContext(self.Spider, resp).Parse(resp.GetRuleName())

	// 统计成功页数
	cache.PageSuccCount()
	// 提示抓取成功
	logs.Log.Informational(" *     Success: %v", req.GetUrl())

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
	return (self.srcManage[0] == self.srcManage[1] && scheduler.Sdl.IsEmpty(self.Spider.GetId())) || scheduler.Sdl.IsStop()
}

func (self *crawler) SetId(id int) {
	self.id = id
}

func (self *crawler) GetId() int {
	return self.id
}
