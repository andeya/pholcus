package scheduler

import (
	"testing"

	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/runtime/cache"
	"github.com/andeya/pholcus/runtime/status"
)

func makeReq(url, rule string) *request.Request {
	r := &request.Request{URL: url, Rule: rule, Method: "GET"}
	r.Prepare()
	return r
}

func TestInit(t *testing.T) {
	tests := []struct {
		name       string
		threadNum  int
		proxyMinute int64
	}{
		{"basic", 4, 0},
		{"with_proxy_minute", 8, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.threadNum, tt.proxyMinute)
		})
	}
}

func TestAddMatrix(t *testing.T) {
	Init(4, 0)
	tests := []struct {
		name         string
		spiderName   string
		spiderSub    string
		maxPage      int64
		wantNotNil   bool
	}{
		{"basic", "sp1", "", -10, true},
		{"with_sub", "sp2", "sub1", -1, true},
		{"zero_limit", "sp3", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := AddMatrix(tt.spiderName, tt.spiderSub, tt.maxPage)
			if (m != nil) != tt.wantNotNil {
				t.Errorf("AddMatrix() got nil=%v, want not nil=%v", m == nil, tt.wantNotNil)
			}
		})
	}
}

func TestPauseRecover(t *testing.T) {
	Init(4, 0)
	PauseRecover()
	PauseRecover()
}

func TestReloadProxyLib(t *testing.T) {
	ReloadProxyLib()
}

func TestMatrix_PushPull_Len(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -5)
	if m == nil {
		t.Fatal("AddMatrix returned nil")
	}
	reqs := []*request.Request{
		makeReq("http://a.com/1", "r1"),
		makeReq("http://a.com/2", "r2"),
	}
	for _, r := range reqs {
		m.Push(r)
	}
	if got := m.Len(); got != 2 {
		t.Errorf("Len() = %d, want 2", got)
	}
	p1 := m.Pull()
	if p1 == nil {
		t.Fatal("Pull() returned nil")
	}
	if p1.GetURL() != "http://a.com/1" && p1.GetURL() != "http://a.com/2" {
		t.Errorf("Pull() got URL %s", p1.GetURL())
	}
	if m.Len() != 1 {
		t.Errorf("Len() after Pull = %d, want 1", m.Len())
	}
	p2 := m.Pull()
	if p2 == nil {
		t.Fatal("Pull() returned nil")
	}
	if m.Len() != 0 {
		t.Errorf("Len() after 2nd Pull = %d, want 0", m.Len())
	}
	_ = p1
	_ = p2
}

func TestMatrix_Push_ignored_when_maxPage_non_negative(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", 0)
	req := makeReq("http://a.com", "r")
	m.Push(req)
	if m.Len() != 0 {
		t.Errorf("Push with maxPage>=0 should be ignored, Len()=%d", m.Len())
	}
}

func TestMatrix_Pull_empty_returns_nil(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -1)
	if got := m.Pull(); got != nil {
		t.Errorf("Pull() on empty queue = %v, want nil", got)
	}
}

func TestMatrix_Pull_paused_returns_nil(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -2)
	m.Push(makeReq("http://a.com", "r"))
	PauseRecover()
	got := m.Pull()
	PauseRecover()
	if got != nil {
		t.Errorf("Pull() when paused = %v, want nil", got)
	}
}

func TestMatrix_Use_Free(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -1)
	m.Use()
	m.Free()
}

func TestMatrix_DoHistory(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -2)
	req := makeReq("http://a.com/x", "r")
	m.Push(req)
	pulled := m.Pull()
	if pulled == nil {
		t.Fatal("Pull failed")
	}
	tests := []struct {
		name string
		ok   bool
		want bool
	}{
		{"success", true, false},
		{"failure_new", false, true},
		{"failure_again", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.DoHistory(pulled, tt.ok)
			if got != tt.want {
				t.Errorf("DoHistory(ok=%v) = %v, want %v", tt.ok, got, tt.want)
			}
		})
	}
}

func TestMatrix_DoHistory_reloadable(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -1)
	req := makeReq("http://a.com/r", "r")
	req.SetReloadable(true)
	got := m.DoHistory(req, true)
	if got != false {
		t.Errorf("DoHistory(reloadable, true) = %v, want false", got)
	}
	got = m.DoHistory(req, false)
	if got != true {
		t.Errorf("DoHistory(reloadable, false) first = %v, want true (new failure)", got)
	}
	got = m.DoHistory(req, false)
	if got != false {
		t.Errorf("DoHistory(reloadable, false) again = %v, want false", got)
	}
}

func TestMatrix_CanStop(t *testing.T) {
	Init(4, 0)
	tests := []struct {
		name    string
		maxPage int64
		push    int
		use     bool
		want    bool
	}{
		{"empty_no_work", -1, 0, false, true},
		{"has_pending", -2, 1, false, false},
		{"has_inflight", -1, 0, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := AddMatrix("sp_"+tt.name, "", tt.maxPage)
			for i := 0; i < tt.push; i++ {
				m.Push(makeReq("http://a.com/"+tt.name+string(rune('a'+i)), "r"))
			}
			if tt.use {
				m.Use()
				defer m.Free()
			}
			got := m.CanStop()
			if got != tt.want {
				t.Errorf("CanStop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatrix_CanStop_after_Stop(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -5)
	m.Push(makeReq("http://a.com", "r"))
	Stop()
	got := m.CanStop()
	Init(4, 0)
	if !got {
		t.Errorf("CanStop() after Stop = %v, want true", got)
	}
}

func TestMatrix_TryFlushSuccess_Failure(t *testing.T) {
	orig := cache.Task
	defer func() { cache.Task = orig }()
	cache.Task = &cache.AppConf{Mode: status.SERVER}
	Init(4, 0)
	m := AddMatrix("sp", "", -1)
	m.TryFlushSuccess()
	m.TryFlushFailure()

	cache.Task = &cache.AppConf{Mode: status.OFFLINE, SuccessInherit: true, FailureInherit: true, OutType: "csv"}
	m2 := AddMatrix("sp2", "", -1)
	m2.TryFlushSuccess()
	m2.TryFlushFailure()
}

func TestMatrix_Wait(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -1)
	m.Wait()
}

func TestMatrix_Push_duplicate_skipped(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -3)
	req := makeReq("http://a.com/dup", "r")
	m.Push(req)
	m.Push(req)
	if m.Len() != 1 {
		t.Errorf("duplicate Push should be skipped, Len()=%d", m.Len())
	}
}

func TestMatrix_Pull_priority(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -3)
	low := makeReq("http://a.com/low", "r")
	low.SetPriority(0)
	high := makeReq("http://a.com/high", "r")
	high.SetPriority(10)
	m.Push(low)
	m.Push(high)
	first := m.Pull()
	if first == nil {
		t.Fatal("Pull returned nil")
	}
	if first.GetURL() != "http://a.com/high" {
		t.Errorf("higher priority should be pulled first, got %s", first.GetURL())
	}
}

func TestMatrix_Pull_request_with_proxy_passthrough(t *testing.T) {
	Init(4, 0)
	m := AddMatrix("sp", "", -1)
	req := makeReq("http://a.com", "r")
	req.SetProxy("http://proxy:8080")
	m.Push(req)
	got := m.Pull()
	if got == nil || got.GetProxy() != "http://proxy:8080" {
		t.Errorf("Pull with existing proxy should preserve it, got %v", got)
	}
}
