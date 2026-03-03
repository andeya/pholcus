package proxy

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/downloader/surfer"
	"github.com/andeya/pholcus/config"
)

func setupProxyDir(t *testing.T) (cleanup func()) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, config.WorkRoot)
	if err := os.MkdirAll(configDir, 0777); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	proxyFile := filepath.Join(configDir, "proxy.lib")
	if err := os.WriteFile(proxyFile, []byte(""), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	return func() { os.Chdir(orig) }
}

func newTestProxy() *Proxy {
	return &Proxy{
		ipRegexp:           regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`),
		proxyIPTypeRegexp:  regexp.MustCompile(`https?://([\w]*:[\w]*@)?[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+`),
		proxyUrlTypeRegexp: regexp.MustCompile(`((https?|ftp):\/\/)?(([^:\n\r]+):([^@\n\r]+)@)?((www\.)?([^/\n\r:]+)):?([0-9]{1,5})?\/?([^?\n\r]+)?\??([^#\n\r]*)?#?([^\n\r]*)`),
		allIps:             map[string]string{},
		all:                map[string]bool{},
		usable:             make(map[string]*ProxyForHost),
		threadPool:         make(chan bool, MAX_THREAD_NUM),
		surf:               surfer.New(),
	}
}

func TestProxy_Update_EmptyFile(t *testing.T) {
	cleanup := setupProxyDir(t)
	defer cleanup()
	_ = config.Conf()

	p := newTestProxy()
	r := p.Update()
	if r.IsErr() {
		t.Errorf("Update: %v", r.UnwrapErr())
	}
	if p.Count() != 0 {
		t.Errorf("Count = %v, want 0", p.Count())
	}
}

func TestProxy_Update_WithIPs(t *testing.T) {
	cleanup := setupProxyDir(t)
	defer cleanup()
	_ = config.Conf()

	proxyFile := filepath.Join(config.WorkRoot, "proxy.lib")
	content := "http://127.0.0.1:8080\nhttp://user:pass@127.0.0.1:9090"
	if err := os.WriteFile(proxyFile, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	p := newTestProxy()
	r := p.Update()
	if r.IsErr() {
		t.Errorf("Update: %v", r.UnwrapErr())
	}
}

func TestProxy_GetOne_NoOnline(t *testing.T) {
	p := &Proxy{online: 0}
	if got := p.GetOne("http://example.com"); got.IsSome() {
		t.Error("GetOne with online=0 want None")
	}
}

func TestProxy_GetOne_EmptyHost(t *testing.T) {
	p := &Proxy{online: 1}
	if got := p.GetOne("http://"); got.IsSome() {
		t.Error("GetOne with empty host want None")
	}
}

func TestProxy_UpdateTicker(t *testing.T) {
	p := &Proxy{
		usable: make(map[string]*ProxyForHost),
	}
	p.usable["example.com"] = &ProxyForHost{curIndex: 0, isEcho: false}
	p.UpdateTicker(5)
	if p.ticker == nil {
		t.Error("UpdateTicker should set ticker")
	}
	if p.tickMinute != 5 {
		t.Errorf("tickMinute = %v, want 5", p.tickMinute)
	}
}

func TestProxy_New(t *testing.T) {
	cleanup := setupProxyDir(t)
	defer cleanup()
	_ = config.Conf()

	p := New()
	time.Sleep(100 * time.Millisecond)
	if p.Count() != 0 {
		t.Errorf("New with empty file Count = %v, want 0", p.Count())
	}
}

func TestProxy_GetOne_WithUsable(t *testing.T) {
	p := &Proxy{
		online:  1,
		ticker: time.NewTicker(time.Hour),
		usable: map[string]*ProxyForHost{
			"example.com": {
				proxys:    []string{"http://127.0.0.1:8080"},
				timedelay: []time.Duration{time.Millisecond},
				curIndex:  0,
				isEcho:    false,
			},
		},
	}
	got := p.GetOne("http://www.example.com/path")
	if !got.IsSome() {
		t.Fatal("GetOne want Some")
	}
	if got.Unwrap() != "http://127.0.0.1:8080" {
		t.Errorf("GetOne = %v, want http://127.0.0.1:8080", got.Unwrap())
	}
}

func TestProxy_GetOne_NoUsableForHost(t *testing.T) {
	p := &Proxy{
		online:  1,
		ticker: time.NewTicker(time.Hour),
		usable: map[string]*ProxyForHost{
			"example.com": {
				proxys:    []string{},
				timedelay: []time.Duration{},
				curIndex:  0,
				isEcho:    false,
			},
		},
	}
	got := p.GetOne("http://www.example.com/path")
	if got.IsSome() {
		t.Error("GetOne with empty proxys want None")
	}
}

type mockSurfer struct {
	resp *http.Response
}

func (m *mockSurfer) Download(req surfer.Request) result.Result[*http.Response] {
	if m.resp != nil {
		return result.Ok(m.resp)
	}
	return result.TryErr[*http.Response](http.ErrHandlerTimeout)
}

func TestProxy_GetOne_TriggersTestAndSort(t *testing.T) {
	cleanup := setupProxyDir(t)
	defer cleanup()
	_ = config.Conf()

	p := newTestProxy()
	p.SetSurfForTest(&mockSurfer{resp: &http.Response{StatusCode: http.StatusOK}})
	p.all = map[string]bool{"http://127.0.0.1:8080": true}
	p.allIps = map[string]string{"http://127.0.0.1:8080": "127.0.0.1"}
	p.online = 1
	p.ticker = time.NewTicker(time.Hour)
	p.usable = map[string]*ProxyForHost{
		"example.com": {
			proxys:    []string{"old"}, // curIndex will exceed after tick
			timedelay: []time.Duration{time.Millisecond},
			curIndex:  1,
			isEcho:    false,
		},
	}

	got := p.GetOne("http://www.example.com/path")
	if !got.IsSome() {
		t.Fatal("GetOne want Some")
	}
	if got.Unwrap() != "http://127.0.0.1:8080" {
		t.Errorf("GetOne = %v, want http://127.0.0.1:8080", got.Unwrap())
	}
}

func TestProxy_Update_FileNotFound(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)
	_ = config.Conf()

	p := newTestProxy()
	r := p.Update()
	if r.IsOk() {
		t.Error("Update with missing file want Err")
	}
}
