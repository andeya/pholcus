package app

import (
	"bytes"
	"testing"
	"time"

	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/runtime/status"
)

func TestNew(t *testing.T) {
	a := New()
	if a == nil {
		t.Fatal("New returned nil")
	}
}

func TestLogic_SetLog_LogGoOn_LogRest(t *testing.T) {
	a := New()
	buf := &bytes.Buffer{}
	a.SetLog(buf)
	a.LogGoOn()
	a.LogRest()
}

func TestLogic_GetAppConf(t *testing.T) {
	a := New().(*Logic)
	tests := []struct {
		keys []string
	}{
		{nil},
		{[]string{"Mode"}},
		{[]string{"ThreadNum"}},
		{[]string{"Limit"}},
	}
	for _, tt := range tests {
		_ = a.GetAppConf(tt.keys...)
	}
}

func TestLogic_SetAppConf(t *testing.T) {
	a := New().(*Logic)
	tests := []struct {
		k string
		v interface{}
	}{
		{"Limit", int64(100)},
		{"Limit", int64(0)},
		{"BatchCap", 50},
		{"BatchCap", 0},
		{"ThreadNum", 10},
	}
	for _, tt := range tests {
		a.SetAppConf(tt.k, tt.v)
	}
}

func TestLogic_GetSpiderLib(t *testing.T) {
	a := New()
	lib := a.GetSpiderLib()
	if lib == nil {
		t.Error("GetSpiderLib returned nil")
	}
}

func TestLogic_GetSpiderByName(t *testing.T) {
	a := New()
	opt := a.GetSpiderByName("nonexistent")
	if opt.IsSome() {
		t.Error("GetSpiderByName(nonexistent) should return None")
	}
}

func TestLogic_GetSpiderQueue(t *testing.T) {
	a := New()
	q := a.GetSpiderQueue()
	if q == nil {
		t.Fatal("GetSpiderQueue returned nil")
	}
	if q.Len() != 0 {
		t.Errorf("new queue Len() = %d, want 0", q.Len())
	}
}

func TestLogic_GetOutputLib(t *testing.T) {
	a := New()
	lib := a.GetOutputLib()
	if len(lib) == 0 {
		t.Error("GetOutputLib returned empty")
	}
}

func TestLogic_GetTaskJar(t *testing.T) {
	a := New()
	jar := a.GetTaskJar()
	if jar == nil {
		t.Fatal("GetTaskJar returned nil")
	}
}

func TestLogic_Status_IsRunning_IsPaused_IsStopped(t *testing.T) {
	a := New().(*Logic)
	if a.Status() != status.STOPPED {
		t.Errorf("Status() = %d, want STOPPED", a.Status())
	}
	if a.IsRunning() {
		t.Error("IsRunning() = true, want false")
	}
	if a.IsPaused() {
		t.Error("IsPaused() = true, want false")
	}
	if !a.IsStopped() {
		t.Error("IsStopped() = false, want true")
	}
}

func TestLogic_Init_Offline(t *testing.T) {
	a := New()
	got := a.Init(status.OFFLINE, 2015, "", nil)
	if got == nil {
		t.Fatal("Init returned nil")
	}
}

func TestLogic_Init_Server_invalidPort(t *testing.T) {
	a := New()
	got := a.Init(status.SERVER, 0, "", nil)
	if got == nil {
		t.Fatal("Init returned nil")
	}
}

func TestLogic_Init_Server_validPort(t *testing.T) {
	a := New()
	got := a.Init(status.SERVER, 2016, "", nil)
	if got == nil {
		t.Fatal("Init returned nil")
	}
}

func TestLogic_Init_Client_invalidMaster(t *testing.T) {
	a := New()
	got := a.Init(status.CLIENT, 2015, "", nil)
	if got == nil {
		t.Fatal("Init returned nil")
	}
}

func TestLogic_Init_invalidMode(t *testing.T) {
	a := New()
	got := a.Init(999, 2015, "", nil)
	if got == nil {
		t.Fatal("Init returned nil")
	}
}

func TestLogic_GetMode(t *testing.T) {
	a := New().(*Logic)
	a.Init(status.OFFLINE, 2015, "", nil)
	if a.GetMode() != status.OFFLINE {
		t.Errorf("GetMode() = %d, want OFFLINE", a.GetMode())
	}
}

func TestLogic_ReInit(t *testing.T) {
	a := New().(*Logic)
	a.Init(status.OFFLINE, 2015, "", nil)
	got := a.ReInit(status.UNSET, 0, "")
	if got == nil {
		t.Fatal("ReInit returned nil")
	}
}

func TestLogic_GetAppConf_titleCase(t *testing.T) {
	a := New().(*Logic)
	a.SetAppConf("limit", int64(50))
	v := a.GetAppConf("limit")
	if v == nil {
		t.Fatal("GetAppConf(limit) returned nil")
	}
}

func TestLogic_SpiderPrepare(t *testing.T) {
	a := New().(*Logic)
	a.Init(status.OFFLINE, 2015, "", nil)
	sp := &spider.Spider{
		Name:      "TestSpider",
		RuleTree:  &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
		Limit:     spider.LIMIT,
		Pausetime: 100,
	}
	sp.Register()
	got := a.SpiderPrepare([]*spider.Spider{sp})
	if got == nil {
		t.Fatal("SpiderPrepare returned nil")
	}
	if a.GetSpiderQueue().Len() < 1 {
		t.Errorf("SpiderPrepare Len() = %d, want >= 1", a.GetSpiderQueue().Len())
	}
}

func TestLogic_Run_emptyQueue(t *testing.T) {
	a := New().(*Logic)
	a.Init(status.OFFLINE, 2015, "", nil)
	a.Run()
}

func TestLogic_Stop_whenStopped(t *testing.T) {
	a := New().(*Logic)
	a.Stop()
}

func TestLogic_PauseRecover(t *testing.T) {
	a := New().(*Logic)
	a.Init(status.OFFLINE, 2015, "", nil)
	a.PauseRecover()
}

func TestLogic_Run_offline_withSpiders(t *testing.T) {
	sp := &spider.Spider{
		Name:      "AppTestSpider",
		RuleTree:  &spider.RuleTree{Root: func(_ *spider.Context) {}, Trunk: map[string]*spider.Rule{}},
		Limit:     spider.LIMIT,
		Pausetime: 100,
	}
	sp.Register()
	a := New().(*Logic)
	a.Init(status.OFFLINE, 2015, "", nil)
	a.SpiderPrepare([]*spider.Spider{sp})
	go func() {
		time.Sleep(3 * time.Second)
		a.Stop()
	}()
	a.Run()
}

func TestLogic_Run_server_withSpiders(t *testing.T) {
	sp := &spider.Spider{
		Name:      "AppTestSpiderServer",
		RuleTree:  &spider.RuleTree{Root: func(_ *spider.Context) {}, Trunk: map[string]*spider.Rule{}},
		Limit:     spider.LIMIT,
		Pausetime: 100,
	}
	sp.Register()
	a := New().(*Logic)
	a.Init(status.SERVER, 2018, "", nil)
	a.SpiderPrepare([]*spider.Spider{sp})
	go func() {
		time.Sleep(2 * time.Second)
		a.Stop()
	}()
	a.Run()
}
