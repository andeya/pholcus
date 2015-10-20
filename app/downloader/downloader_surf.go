package downloader

import (
	"net/http"

	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/surfer"
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

	if cReq.GetUsePhantomJS() {
		resp, err = self.phantom.Download(cReq)
	} else {
		resp, err = self.surf.Download(cReq)
	}

	cResp.SetRequest(cReq)

	cResp.SetResponse(resp)

	cResp.SetError(err)

	return cResp
}
