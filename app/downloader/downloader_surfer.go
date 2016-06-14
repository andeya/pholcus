package downloader

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"

	"github.com/henrylee2cn/pholcus/app/downloader/request"
	"github.com/henrylee2cn/pholcus/app/downloader/surfer"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/config"
)

type Surfer struct {
	surf    surfer.Surfer
	phantom surfer.Surfer
}

const (
	SURF_ID    = 0 //默认下载器，此值不可改动
	PHANTOM_ID = iota
)

var SurferDownloader = &Surfer{
	surf:    surfer.New(),
	phantom: surfer.NewPhantom(config.PHANTOMJS, config.PHANTOMJS_TEMP),
}

func (self *Surfer) Download(sp *spider.Spider, cReq *request.Request) *spider.Context {
	ctx := spider.GetContext(sp, cReq)

	var resp *http.Response
	var err error

	switch cReq.GetDownloaderID() {
	case SURF_ID:
		resp, err = self.surf.Download(cReq)

	case PHANTOM_ID:
		resp, err = self.phantom.Download(cReq)
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		var gzipReader *gzip.Reader
		gzipReader, err = gzip.NewReader(resp.Body)
		resp.Body.Close()
		if err == nil {
			resp.Body = ioutil.NopCloser(gzipReader)
		}
	}

	ctx.SetResponse(resp).SetError(err)

	return ctx
}
