package crawl

import (
	"io"
	"math/rand"
	"time"

	"github.com/henrylee2cn/pholcus/app/downloader"
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/app/pipeline"
	"github.com/henrylee2cn/pholcus/app/scheduler"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

type Crawler interface {
	Init(*spider.Spider) Crawler
	Start()
	GetId() int
}

type crawler struct {
	id int
	*spider.Spider
	basePause int64
	gainPause int64
	downloader.Downloader
	pipeline.Pipeline
	srcManage      int32
	historyFailure []*context.Request
}

func New(id int) Crawler {
	return &crawler{
		id:         id,
		Pipeline:   pipeline.New(),
		Downloader: downloader.SurferDownloader,
		srcManage:  0,
	}
}

func (self *crawler) Init(sp *spider.Spider) Crawler {
	self.srcManage = 0
	self.Spider = sp.ReqmatrixInit()
	self.Pipeline.Init(sp)
	self.basePause = cache.Task.Pausetime / 2
	if self.basePause > 0 {
		self.gainPause = self.basePause * 3
	} else {
		self.gainPause = 1
	}
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
			if self.Spider.ReqmatrixCanStop() {
				// 停止任务
				return

			} else {
				// 继续等待请求
				continue
			}
		}

		self.UseOne()

		go func(req *context.Request) {
			defer func() {
				self.FreeOne()
			}()
			// logs.Log.Informational(" *     start: %v", req.GetUrl())
			self.Process(req)
		}(req)
	}
}

// core processer
func (self *crawler) Process(req *context.Request) {
	// download page
	resp := self.Downloader.Download(req)
	downUrl := resp.GetUrl()
	if resp.GetError() != nil {
		// 删除该请求的成功记录
		scheduler.DeleteSuccess(resp)
		// 对下载失败的请求进行失败记录
		if !self.Spider.ReqmatrixSetFailure(req) {
			// 统计失败数
			cache.PageFailCount()
		}

		// 提示错误
		logs.Log.Error(" *     Fail [download][%v]: %v", downUrl, resp.GetError())
		return
	}

	defer func() {
		if err := recover(); err != nil {
			// 删除该请求的成功记录
			scheduler.DeleteSuccess(resp)
			// 对下载失败的请求进行失败记录
			if !self.Spider.ReqmatrixSetFailure(req) {
				// 统计失败数
				cache.PageFailCount()
			}

			// 提示错误
			logs.Log.Error(" *     Fail [process][%v]: %v", downUrl, err)
		}
	}()

	// 过程处理，提炼数据
	spider.NewContext(self.Spider, resp).Parse(resp.GetRuleName())

	// 统计成功页数
	cache.PageSuccCount()
	// 提示抓取成功
	logs.Log.Informational(" *     Success: %v", downUrl)
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
	sleeptime := self.basePause + rand.New(rand.NewSource(time.Now().UnixNano())).
		Int63n(self.gainPause)
	time.Sleep(time.Duration(sleeptime) * time.Millisecond)
}

// 从调度读取一个请求
func (self *crawler) GetOne() *context.Request {
	return self.Spider.ReqmatrixPull()
}

//从调度使用一个资源空位
func (self *crawler) UseOne() {
	self.Spider.ReqmatrixUse()
}

//从调度释放一个资源空位
func (self *crawler) FreeOne() {
	self.Spider.ReqmatrixFree()
}

func (self *crawler) SetId(id int) {
	self.id = id
}

func (self *crawler) GetId() int {
	return self.id
}
