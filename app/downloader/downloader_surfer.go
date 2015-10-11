package downloader

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/surfer"
)

type Surfer struct {
	download surfer.Surfer
}

var SurferDownloader = &Surfer{
	download: surfer.New(),
}

func (self *Surfer) Download(cReq *context.Request) *context.Response {
	cResp := context.NewResponse(nil)

	resp, err := self.download.Download(cReq)

	cResp.SetRequest(cReq)

	cResp.SetResponse(resp)

	cResp.SetError(err)

	return cResp
}
