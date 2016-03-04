package crawl

import (
	"io"
	"math/rand"
	"time"

	"github.com/henrylee2cn/pholcus/app/downloader"
	"github.com/henrylee2cn/pholcus/app/downloader/request"
	"github.com/henrylee2cn/pholcus/app/pipeline"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

type (
	Crawler interface {
		Init(*spider.Spider) Crawler
		Start()
		GetId() int
	}
	crawler struct {
		id int
		*spider.Spider
		basePause int64
		gainPause int64
		downloader.Downloader
		pipeline.Pipeline
		historyFailure []*request.Request
	}
)

func New(id int) Crawler {
	return &crawler{
		id:         id,
		Pipeline:   pipeline.New(),
		Downloader: downloader.SurferDownloader,
	}
}

func (self *crawler) Init(sp *spider.Spider) Crawler {
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

	// 启动任务
	self.Spider.Start()

	// 任务运行中
	self.Run()

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
			if self.Spider.CanStop() {
				// 停止任务
				break

			} else {
				// 继续等待请求
				continue
			}
		}

		self.UseOne()

		go func(req *request.Request) {
			defer func() {
				self.FreeOne()
			}()
			// logs.Log.Informational(" *     start: %v", req.GetUrl())
			self.Process(req)
		}(req)
	}
	// 等待处理中的任务完成
	self.Spider.Defer()
}

// core processer
func (self *crawler) Process(req *request.Request) {
	var (
		ctx     = self.Downloader.Download(self.Spider, req) // download page
		downUrl = req.GetUrl()
	)

	if err := ctx.GetError(); err != nil {
		// 返回是否作为新的失败请求被添加至队列尾部
		if self.Spider.DoHistory(req, false) {
			// 统计失败数
			cache.PageFailCount()
		}
		// 提示错误
		logs.Log.Error(" *     Fail  [download][%v]: %v\n", downUrl, err)
		return
	}

	defer func() {
		if err := recover(); err != nil {
			// 返回是否作为新的失败请求被添加至队列尾部
			if self.Spider.DoHistory(req, false) {
				// 统计失败数
				cache.PageFailCount()
			}
			// 提示错误
			logs.Log.Error(" *     Panic  [process][%v]: %v\n", downUrl, err)
		}
	}()

	var ruleName = req.GetRuleName()

	// 过程处理，提炼数据
	ctx.Parse(ruleName)
	// 处理成功请求记录
	self.Spider.DoHistory(req, true)
	// 统计成功页数
	cache.PageSuccCount()
	// 提示抓取成功
	logs.Log.Informational(" *     Success: %v\n", downUrl)

	downUrl = req.GetUrl()
	ruleName = req.GetRuleName()
	var referer = req.GetReferer()

	// 该条请求文本结果存入pipeline
	for _, data := range ctx.GetItems() {
		self.Pipeline.CollectData(
			ruleName, //DataCell.RuleName
			data,     //DataCell.Data
			downUrl,  //DataCell.Url
			referer,  //DataCell.ParentUrl
			time.Now().Format("2006-01-02 15:04:05"),
		)
	}

	// 该条请求文件结果存入pipeline
	for _, f := range ctx.GetFiles() {
		self.Pipeline.CollectFile(
			ruleName,
			f["Name"].(string),
			f["Body"].(io.ReadCloser),
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
func (self *crawler) GetOne() *request.Request {
	return self.Spider.RequestPull()
}

//从调度使用一个资源空位
func (self *crawler) UseOne() {
	self.Spider.RequestUse()
}

//从调度释放一个资源空位
func (self *crawler) FreeOne() {
	self.Spider.RequestFree()
}

func (self *crawler) SetId(id int) {
	self.id = id
}

func (self *crawler) GetId() int {
	return self.id
}
