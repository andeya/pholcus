package downloader

import (
	"errors"
	"net/http"
	"net/http/cookiejar"

	"github.com/andeya/gust/result"
	"github.com/andeya/gust/syncutil"
	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/app/downloader/surfer"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/config"
)

type Surfer struct {
	surf surfer.Surfer
}

var (
	cookieJar, _     = cookiejar.New(nil)
	SurferDownloader = &Surfer{
		surf: surfer.New(cookieJar),
	}
)

var lazyPhantom = syncutil.NewLazyValueWithFunc(func() result.Result[surfer.Surfer] {
	return result.Ok[surfer.Surfer](surfer.NewPhantom(config.Conf().PhantomJS, config.PhantomJSTemp, cookieJar))
})

var lazyChrome = syncutil.NewLazyValueWithFunc(func() result.Result[surfer.Surfer] {
	return result.Ok[surfer.Surfer](surfer.NewChrome(cookieJar))
})

func (s *Surfer) Download(sp *spider.Spider, cReq *request.Request) *spider.Context {
	ctx := spider.GetContext(sp, cReq)

	var resp *http.Response
	var err error

	switch cReq.GetDownloaderID() {
	case request.SurfID:
		r := s.surf.Download(cReq)
		if r.IsErr() {
			err = r.UnwrapErr()
		} else {
			resp = r.Unwrap()
		}

	case request.PhantomID:
		r := lazyPhantom.TryGetValue().Unwrap().Download(cReq)
		if r.IsErr() {
			err = r.UnwrapErr()
		} else {
			resp = r.Unwrap()
		}

	case request.ChromeID:
		r := lazyChrome.TryGetValue().Unwrap().Download(cReq)
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
