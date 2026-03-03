package crawler

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/app/scheduler"
	"github.com/andeya/pholcus/app/spider"
)

func TestNew(t *testing.T) {
	c := New(1, &mockDownloader{}, "csv", 10)
	if c == nil {
		t.Fatal("New returned nil")
	}
	if c.GetID() != 1 {
		t.Errorf("GetID() = %d, want 1", c.GetID())
	}
}

func TestCrawler_GetID(t *testing.T) {
	c := New(42, &mockDownloader{}, "csv", 10)
	if got := c.GetID(); got != 42 {
		t.Errorf("GetID() = %d, want 42", got)
	}
}

func TestCrawler_Init(t *testing.T) {
	scheduler.Init(4, 0)
	c := New(0, &mockDownloader{}, "csv", 10)
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:    -5,
	}
	got := c.Init(sp)
	if got != c {
		t.Error("Init should return self")
	}
}

func TestCrawler_Init_zeroPause(t *testing.T) {
	scheduler.Init(4, 0)
	c := New(0, &mockDownloader{}, "csv", 10)
	sp := &spider.Spider{
		Name:      "TestSpider",
		RuleTree:  &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:     -5,
		Pausetime: 0,
	}
	c.Init(sp)
}

func TestCrawler_GetOne_UseOne_FreeOne(t *testing.T) {
	scheduler.Init(4, 0)
	cr := New(0, &mockDownloader{}, "csv", 10).(*crawler)
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:    -5,
	}
	cr.Init(sp)

	req := cr.GetOne()
	if req != nil {
		t.Error("GetOne on empty matrix should return nil")
	}
	cr.UseOne()
	cr.FreeOne()
}

func TestCrawler_CanStop(t *testing.T) {
	scheduler.Init(4, 0)
	c := New(0, &mockDownloader{}, "csv", 10)
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:    -5,
	}
	c.Init(sp)
	if !c.CanStop() {
		t.Error("CanStop on empty matrix should be true")
	}
}

func TestCrawler_Stop(t *testing.T) {
	scheduler.Init(4, 0)
	c := New(0, &mockDownloader{}, "csv", 10)
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:    -5,
	}
	c.Init(sp)
	c.Stop()
}

func TestCrawler_SetID(t *testing.T) {
	cr := New(0, &mockDownloader{}, "csv", 10).(*crawler)
	cr.SetID(99)
	if cr.GetID() != 99 {
		t.Errorf("GetID() = %d, want 99", cr.GetID())
	}
}

type errorDownloader struct{}

func (d *errorDownloader) Download(sp *spider.Spider, req *request.Request) *spider.Context {
	ctx := spider.GetContext(sp, req)
	ctx.SetError(fmt.Errorf("download failed"))
	return ctx
}

func TestCrawler_Process_downloadError(t *testing.T) {
	scheduler.Init(4, 0)
	cr := New(0, &errorDownloader{}, "csv", 10).(*crawler)
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:    -5,
	}
	cr.Init(sp)
	req := &request.Request{URL: "http://example.com", Rule: "r"}
	req.Prepare()
	cr.Process(req)
}

func TestCrawler_Run(t *testing.T) {
	scheduler.Init(4, 0)
	sd := &successDownloader{}
	cr := New(0, sd, "csv", 10).(*crawler)
	sp := &spider.Spider{
		Name: "CrawlerRunTestSpider",
		RuleTree: &spider.RuleTree{
			Root: func(ctx *spider.Context) {
				time.Sleep(50 * time.Millisecond)
				req := &request.Request{URL: "http://example.com", Rule: "r"}
				req.Prepare()
				ctx.AddQueue(req)
			},
			Trunk: map[string]*spider.Rule{"r": {ParseFunc: func(_ *spider.Context) {}}},
		},
		Limit: -5,
	}
	sp.Register()
	cr.Init(sp)
	cr.Run()
}

type successDownloader struct{}

func (d *successDownloader) Download(sp *spider.Spider, req *request.Request) *spider.Context {
	ctx := spider.GetContext(sp, req)
	ctx.SetResponse(&http.Response{StatusCode: 200})
	return ctx
}

func TestCrawler_Process_success(t *testing.T) {
	scheduler.Init(4, 0)
	cr := New(0, &successDownloader{}, "csv", 10).(*crawler)
	sp := &spider.Spider{
		Name: "TestSpider",
		RuleTree: &spider.RuleTree{
			Root:  func(_ *spider.Context) {},
			Trunk: map[string]*spider.Rule{"r": {ParseFunc: func(_ *spider.Context) {}}},
		},
		Limit: -5,
	}
	cr.Init(sp)
	req := &request.Request{URL: "http://example.com", Rule: "r"}
	req.Prepare()
	cr.Process(req)
}
