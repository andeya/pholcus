package downloader

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"io"
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

	if resp.StatusCode >= 400 {
		err = errors.New("响应状态 " + resp.Status)
	}

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var gzipReader *gzip.Reader
		gzipReader, err = gzip.NewReader(resp.Body)
		if err == nil {
			resp.Body = gzipReader
		}

	case "deflate":
		resp.Body = flate.NewReader(resp.Body)

	case "zlib":
		var readCloser io.ReadCloser
		readCloser, err = zlib.NewReader(resp.Body)
		if err == nil {
			resp.Body = readCloser
		}
	}

	ctx.SetResponse(resp).SetError(err)

	return ctx
}

type body struct {
	io.ReadCloser
}

func (self *body) Read(p []byte) (n int, err error) {
	return self.Read(p)
}

func (self *body) Write(p []byte) (n int, err error) {
	return self.Write(p)
}
