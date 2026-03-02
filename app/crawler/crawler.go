package crawler

import (
	"bytes"
	"math/rand"
	"runtime"
	"time"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/downloader"
	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/app/pipeline"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/cache"
)

// Crawler is the core crawler engine.
type (
	Crawler interface {
		Init(*spider.Spider) Crawler // Init initializes the crawler engine
		Run()                        // Run executes the task
		Stop()                       // Stop terminates the crawler
		CanStop() bool               // CanStop reports whether the crawler can be stopped
		GetId() int                  // GetId returns the engine ID
	}
	crawler struct {
		*spider.Spider                 // spider rule being executed
		downloader.Downloader          // shared downloader
		pipeline.Pipeline              // result collection and output pipeline
		id                    int      // engine ID
		pause                 [2]int64 // [min request interval ms, max additional interval ms]
	}
)

// New creates a new Crawler with the given ID.
func New(id int) Crawler {
	return &crawler{
		id:         id,
		Downloader: downloader.SurferDownloader,
	}
}

// Init initializes the crawler with the given spider.
func (self *crawler) Init(sp *spider.Spider) Crawler {
	self.Spider = sp.ReqmatrixInit()
	self.Pipeline = pipeline.New(sp)
	self.pause[0] = sp.Pausetime / 2
	if self.pause[0] > 0 {
		self.pause[1] = self.pause[0] * 3
	} else {
		self.pause[1] = 1
	}
	return self
}

// Run is the main entry point for task execution.
func (self *crawler) Run() {
	self.Pipeline.Start()

	c := make(chan bool)
	go func() {
		self.run()
		close(c)
	}()

	self.Spider.Start()

	<-c

	self.Pipeline.Stop()
}

// Stop terminates the crawler and its pipeline.
func (self *crawler) Stop() {
	self.Spider.Stop()
	self.Pipeline.Stop()
}

func (self *crawler) run() {
	for {
		req := self.GetOne()
		if req == nil {
			if self.Spider.CanStop() {
				break
			}
			time.Sleep(20 * time.Millisecond)
			continue
		}

		self.UseOne()
		go func() {
			defer func() {
				self.FreeOne()
			}()
			logs.Log.Debug(" *     Start: %v", req.GetUrl())
			self.Process(req)
		}()

		self.sleep()
	}

	self.Spider.Defer()
}

// Process downloads a request, parses the response, and sends results to the pipeline.
func (self *crawler) Process(req *request.Request) {
	var (
		downUrl = req.GetUrl()
		sp      = self.Spider
	)
	defer func() {
		if p := recover(); p != nil {
			if sp.IsStopping() {
				return
			}
			if sp.DoHistory(req, false) {
				cache.PageFailCount()
			}
			stack := make([]byte, 4<<10)
			length := runtime.Stack(stack, true)
			start := bytes.Index(stack, []byte("/src/runtime/panic.go"))
			stack = stack[start:length]
			start = bytes.Index(stack, []byte("\n")) + 1
			stack = stack[start:]
			if end := bytes.Index(stack, []byte("\ngoroutine ")); end != -1 {
				stack = stack[:end]
			}
			stack = bytes.Replace(stack, []byte("\n"), []byte("\r\n"), -1)
			logs.Log.Error(" *     Panic  [process][%s]: %s\r\n[TRACE]\r\n%s", downUrl, p, stack)
		}
	}()

	var ctx = self.Downloader.Download(sp, req) // download page

	if r := result.TryErrVoid(ctx.GetError()); r.IsErr() {
		if sp.DoHistory(req, false) {
			cache.PageFailCount()
		}
		logs.Log.Error(" *     Fail  [download][%v]: %v\n", downUrl, r.UnwrapErr())
		return
	}

	ctx.Parse(req.GetRuleName())

	for _, f := range ctx.PullFiles() {
		if self.Pipeline.CollectFile(f).IsErr() {
			break
		}
	}
	for _, item := range ctx.PullItems() {
		if self.Pipeline.CollectData(item).IsErr() {
			break
		}
	}

	sp.DoHistory(req, true)
	cache.PageSuccCount()
	logs.Log.Informational(" *     Success: %v\n", downUrl)
	spider.PutContext(ctx)
}

func (self *crawler) sleep() {
	sleeptime := self.pause[0] + rand.Int63n(self.pause[1])
	time.Sleep(time.Duration(sleeptime) * time.Millisecond)
}

// GetOne pulls one request from the scheduler.
func (self *crawler) GetOne() *request.Request {
	return self.Spider.RequestPull()
}

// UseOne acquires one resource slot from the scheduler.
func (self *crawler) UseOne() {
	self.Spider.RequestUse()
}

// FreeOne releases one resource slot to the scheduler.
func (self *crawler) FreeOne() {
	self.Spider.RequestFree()
}

// SetId sets the crawler ID.
func (self *crawler) SetId(id int) {
	self.id = id
}

// GetId returns the crawler engine ID.
func (self *crawler) GetId() int {
	return self.id
}
