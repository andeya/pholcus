package downloader

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/app/spider"
)

func TestSurferDownloader_implementsInterface(t *testing.T) {
	var _ Downloader = SurferDownloader
}

func makeSpiderNotStopping(name string) *spider.Spider {
	sp := &spider.Spider{
		Name:     name,
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
	}
	sp.Register()
	return sp
}

func TestSurferDownloader_Download_SurfID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	sp := makeSpiderNotStopping("DownloaderTestSpider1")
	req := &request.Request{URL: ts.URL, Rule: "r"}
	req.Prepare()

	ctx := SurferDownloader.Download(sp, req)
	if ctx == nil {
		t.Fatal("Download returned nil context")
	}
	if err := ctx.GetError(); err != nil {
		t.Errorf("GetError() = %v, want nil", err)
	}
	if ctx.Response == nil {
		t.Fatal("Response is nil")
	}
	if ctx.Response.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", ctx.Response.StatusCode)
	}
}

func TestSurferDownloader_Download_SurfID_error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	sp := makeSpiderNotStopping("DownloaderTestSpider2")
	req := &request.Request{URL: ts.URL, Rule: "r"}
	req.Prepare()

	ctx := SurferDownloader.Download(sp, req)
	if ctx == nil {
		t.Fatal("Download returned nil context")
	}
	if err := ctx.GetError(); err == nil {
		t.Error("GetError() = nil, want error for 5xx")
	}
}

func TestSurferDownloader_Download_SurfID_4xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	sp := makeSpiderNotStopping("DownloaderTestSpider4xx")
	req := &request.Request{URL: ts.URL, Rule: "r"}
	req.Prepare()

	ctx := SurferDownloader.Download(sp, req)
	if ctx == nil {
		t.Fatal("Download returned nil context")
	}
	if err := ctx.GetError(); err == nil {
		t.Error("GetError() = nil, want error for 4xx")
	}
}

func TestSurferDownloader_Download_SurfID_badURL(t *testing.T) {
	sp := makeSpiderNotStopping("DownloaderTestSpider3")
	req := &request.Request{URL: "http://localhost:0/nonexistent", Rule: "r"}
	req.Prepare()

	ctx := SurferDownloader.Download(sp, req)
	if ctx == nil {
		t.Fatal("Download returned nil context")
	}
	if err := ctx.GetError(); err == nil {
		t.Error("GetError() = nil, want error for failed request")
	}
}
