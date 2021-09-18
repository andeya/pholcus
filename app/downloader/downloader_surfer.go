package downloader

import (
	"errors"
	"net/http"
	"net/http/cookiejar"

	"github.com/henrylee2cn/pholcus/app/downloader/request"
	"github.com/henrylee2cn/pholcus/app/downloader/surfer"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/config"
)

type Surfer struct {
	surf    surfer.Surfer
	phantom surfer.Surfer
	chrome  surfer.Surfer
}

var (
	cookieJar, _     = cookiejar.New(nil)
	SurferDownloader = &Surfer{
		surf:    surfer.New(cookieJar),
		phantom: surfer.NewPhantom(config.PHANTOMJS, config.PHANTOMJS_TEMP, cookieJar),
		chrome:  surfer.NewChrome(cookieJar),
	}
)

func (self *Surfer) Download(sp *spider.Spider, cReq *request.Request) *spider.Context {
	ctx := spider.GetContext(sp, cReq)

	var resp *http.Response
	var err error

	switch cReq.GetDownloaderID() {
	case request.SURF_ID:
		resp, err = self.surf.Download(cReq)

	case request.PHANTOM_ID:
		resp, err = self.phantom.Download(cReq)

	case request.CHROME_ID:
		resp, err = self.chrome.Download(cReq)
	}

	if resp.StatusCode >= 400 {
		err = errors.New("响应状态 " + resp.Status)
	}

	ctx.SetResponse(resp).SetError(err)

	return ctx
}
