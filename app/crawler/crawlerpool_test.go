package crawler

import (
	"testing"

	"github.com/andeya/pholcus/app/downloader"
	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/app/scheduler"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/config"
)

type mockDownloader struct{}

func (d *mockDownloader) Download(_ *spider.Spider, _ *request.Request) *spider.Context {
	return nil
}

func TestNewCrawlerPool(t *testing.T) {
	dl := &mockDownloader{}
	pool := NewCrawlerPool(dl)
	if pool == nil {
		t.Fatal("NewCrawlerPool returned nil")
	}
}

func TestCrawlerPool_SetPipelineConfig(t *testing.T) {
	pool := NewCrawlerPool(&mockDownloader{})
	pool.SetPipelineConfig("csv", 100)
}

func TestCrawlerPool_Reset(t *testing.T) {
	_ = config.Conf()
	pool := NewCrawlerPool(&mockDownloader{})
	pool.SetPipelineConfig("csv", 10)

	tests := []struct {
		name       string
		spiderNum  int
		wantMinNum int
	}{
		{"one", 1, 1},
		{"five", 5, 5},
		{"over_cap", 999, 1},
		{"zero", 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pool.Reset(tt.spiderNum)
			if got < tt.wantMinNum {
				t.Errorf("Reset(%d) = %d, want >= %d", tt.spiderNum, got, tt.wantMinNum)
			}
		})
	}
}

func TestCrawlerPool_Use_UseOpt_Free(t *testing.T) {
	scheduler.Init(4, 0)
	pool := NewCrawlerPool(downloader.SurferDownloader)
	pool.SetPipelineConfig("csv", 10)
	pool.Reset(2)

	opt := pool.UseOpt()
	if !opt.IsSome() {
		t.Fatal("UseOpt returned None")
	}
	c := opt.Unwrap()
	if c == nil {
		t.Fatal("UseOpt returned nil crawler")
	}
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:    -10,
	}
	c.Init(sp)
	pool.Free(c)

	c2 := pool.Use()
	if c2 == nil {
		t.Fatal("Use returned nil")
	}
	pool.Free(c2)
}

func TestCrawlerPool_UseOpt_returnsSome(t *testing.T) {
	scheduler.Init(4, 0)
	pool := NewCrawlerPool(&mockDownloader{})
	pool.SetPipelineConfig("csv", 10)
	pool.Reset(2)

	opt := pool.UseOpt()
	if !opt.IsSome() {
		t.Fatal("UseOpt returned None")
	}
	c := opt.Unwrap()
	if c.GetID() < 0 {
		t.Errorf("GetID() = %d, want >= 0", c.GetID())
	}
}

func TestCrawlerPool_Stop(t *testing.T) {
	pool := NewCrawlerPool(&mockDownloader{})
	pool.SetPipelineConfig("csv", 10)
	pool.Reset(1)
	pool.Stop()

	opt := pool.UseOpt()
	if opt.IsSome() {
		t.Error("UseOpt after Stop should return None")
	}
}

func TestCrawlerPool_Reset_reuse(t *testing.T) {
	scheduler.Init(4, 0)
	_ = config.Conf()
	pool := NewCrawlerPool(&mockDownloader{})
	pool.SetPipelineConfig("csv", 10)
	pool.Reset(2)
	c1 := pool.Use()
	c2 := pool.Use()
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:    -5,
	}
	c1.Init(sp)
	c2.Init(sp)
	pool.Free(c1)
	pool.Free(c2)
	got := pool.Reset(3)
	if got != 3 {
		t.Errorf("Reset(3) = %d, want 3", got)
	}
}

func TestCrawlerPool_Stop_idempotent(t *testing.T) {
	pool := NewCrawlerPool(&mockDownloader{})
	pool.Reset(1)
	pool.Stop()
	pool.Stop()
}

func TestCrawlerPool_Free_whenStopped(t *testing.T) {
	scheduler.Init(4, 0)
	pool := NewCrawlerPool(downloader.SurferDownloader)
	pool.SetPipelineConfig("csv", 10)
	pool.Reset(1)
	c := pool.Use()
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:    -5,
	}
	c.Init(sp)
	pool.Stop()
	pool.Free(c)
}
