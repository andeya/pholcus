package downloader

import (
	"net/http"

	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/app/downloader/surfer"
	"github.com/henrylee2cn/pholcus/config"
)

const (
	SURF_ID    = 0 //默认下载器，此值不可改动
	PHANTOM_ID = iota
)

type Surfer struct {
	surf    surfer.Surfer
	phantom surfer.Surfer
}

var SurferDownloader = &Surfer{
	surf:    surfer.New(),
	phantom: surfer.NewPhantom(config.SURFER_PHANTOM.FULL_APP_NAME, config.SURFER_PHANTOM.FULL_TEMP_JS),
}

func (self *Surfer) Download(cReq *context.Request) *context.Response {
	cResp := context.NewResponse(nil)

	var resp *http.Response
	var err error

	switch cReq.GetDownloaderID() {
	case SURF_ID:
		resp, err = self.surf.Download(cReq)

	case PHANTOM_ID:
		resp, err = self.phantom.Download(cReq)
	}

	cResp.Prepare(resp, cReq)

	cResp.SetError(err)

	return cResp
}
