// Package crawler provides the core crawler engine for request scheduling and page downloading.
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
		GetID() int                  // GetID returns the engine ID
	}
	crawler struct {
		*spider.Spider                 // spider rule being executed
		downloader.Downloader          // shared downloader
		pipeline.Pipeline              // result collection and output pipeline
		id                    int      // engine ID
		outType               string   // output type for pipeline
		batchCap              int      // batch output capacity for pipeline
		pause                 [2]int64 // [min request interval ms, max additional interval ms]
	}
)

// New creates a new Crawler with the given ID, Downloader, and pipeline config.
func New(id int, dl downloader.Downloader, outType string, batchCap int) Crawler {
	return &crawler{
		id:         id,
		Downloader: dl,
		outType:    outType,
		batchCap:   batchCap,
	}
}

// Init initializes the crawler with the given spider.
func (c *crawler) Init(sp *spider.Spider) Crawler {
	c.Spider = sp.ReqmatrixInit()
	c.Pipeline = pipeline.New(sp, c.outType, c.batchCap)
	c.pause[0] = sp.Pausetime / 2
	if c.pause[0] > 0 {
		c.pause[1] = c.pause[0] * 3
	} else {
		c.pause[1] = 1
	}
	return c
}

// Run is the main entry point for task execution.
func (c *crawler) Run() {
	c.Pipeline.Start()

	done := make(chan bool)
	go func() {
		c.run()
		close(done)
	}()

	c.Spider.Start()

	<-done

	c.Pipeline.Stop()
}

// Stop terminates the crawler and its pipeline.
func (c *crawler) Stop() {
	c.Spider.Stop()
	c.Pipeline.Stop()
}

func (c *crawler) run() {
	for {
		req := c.GetOne()
		if req == nil {
			if c.Spider.CanStop() {
				break
			}
			time.Sleep(20 * time.Millisecond)
			continue
		}

		c.UseOne()
		go func() {
			defer func() {
				c.FreeOne()
			}()
			logs.Log().Debug(" *     Start: %v", req.GetURL())
			c.Process(req)
		}()

		c.sleep()
	}

	c.Spider.Defer()
}

// Process downloads a request, parses the response, and sends results to the pipeline.
func (c *crawler) Process(req *request.Request) {
	var (
		downUrl = req.GetURL()
		sp      = c.Spider
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
			logs.Log().Error(" *     Panic  [process][%s]: %s\r\n[TRACE]\r\n%s", downUrl, p, stack)
		}
	}()

	var ctx = c.Downloader.Download(sp, req) // download page

	if r := result.TryErrVoid(ctx.GetError()); r.IsErr() {
		if sp.DoHistory(req, false) {
			cache.PageFailCount()
		}
		logs.Log().Error(" *     Fail  [download][%v]: %v\n", downUrl, r.UnwrapErr())
		return
	}

	ctx.Parse(req.GetRuleName())

	if parseErr := ctx.GetError(); parseErr != nil {
		if sp.DoHistory(req, false) {
			cache.PageFailCount()
		}
		logs.Log().Error(" *     Fail  [parse][%v]: %v\n", downUrl, parseErr)
		return
	}

	for _, f := range ctx.PullFiles() {
		if c.Pipeline.CollectFile(f).IsErr() {
			break
		}
	}
	for _, item := range ctx.PullItems() {
		if c.Pipeline.CollectData(item).IsErr() {
			break
		}
	}

	sp.DoHistory(req, true)
	cache.PageSuccCount()
	logs.Log().Informational(" *     Success: %v\n", downUrl)
	spider.PutContext(ctx)
}

func (c *crawler) sleep() {
	sleeptime := c.pause[0] + rand.Int63n(c.pause[1])
	time.Sleep(time.Duration(sleeptime) * time.Millisecond)
}

// GetOne pulls one request from the scheduler.
func (c *crawler) GetOne() *request.Request {
	return c.Spider.RequestPull()
}

// UseOne acquires one resource slot from the scheduler.
func (c *crawler) UseOne() {
	c.Spider.RequestUse()
}

// FreeOne releases one resource slot to the scheduler.
func (c *crawler) FreeOne() {
	c.Spider.RequestFree()
}

// SetID sets the crawler ID.
func (c *crawler) SetID(id int) {
	c.id = id
}

// GetID returns the crawler engine ID.
func (c *crawler) GetID() int {
	return c.id
}
