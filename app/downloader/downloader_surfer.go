package downloader

import (
	"errors"
	"net/http"
	"net/http/cookiejar"

	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/app/downloader/surfer"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/config"
)

type Surfer struct {
	surf    surfer.Surfer
	phantom surfer.Surfer
}

var (
	cookieJar, _     = cookiejar.New(nil) // nil options never returns error
	SurferDownloader = &Surfer{
		surf:    surfer.New(cookieJar),
		phantom: surfer.NewPhantom(config.PHANTOMJS, config.PHANTOMJS_TEMP, cookieJar),
	}
)

func (self *Surfer) Download(sp *spider.Spider, cReq *request.Request) *spider.Context {
	ctx := spider.GetContext(sp, cReq)

	var resp *http.Response
	var err error

	switch cReq.GetDownloaderID() {
	case request.SURF_ID:
		r := self.surf.Download(cReq)
		if r.IsErr() {
			err = r.UnwrapErr()
		} else {
			resp = r.Unwrap()
		}

	case request.PHANTOM_ID:
		r := self.phantom.Download(cReq)
		if r.IsErr() {
			err = r.UnwrapErr()
		} else {
			resp = r.Unwrap()
		}
	}

	if resp != nil && resp.StatusCode >= 400 {
		err = errors.New("response status " + resp.Status)
	}

	ctx.SetResponse(resp).SetError(err)

	return ctx
}
