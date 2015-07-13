package downloader

import (
	"github.com/henrylee2cn/pholcus/crawl/downloader/context"
	"github.com/henrylee2cn/surfer"
	"time"
)

type Surfer struct {
	download *surfer.Download
}

func NewSurfer(paseTime time.Duration, proxy ...string) *Surfer {
	if len(proxy) == 0 {
		proxy = append(proxy, "")
	}
	return &Surfer{
		download: surfer.NewDownload(3, paseTime, proxy[0]),
	}
}

func (self *Surfer) Download(cReq *context.Request) *context.Response {
	cResp := context.NewResponse(nil)

	resp, err := self.download.Download(cReq.GetMethod(), cReq.GetUrl(), cReq.GetReferer(), cReq.GetPostData(), cReq.GetHeader(), cReq.GetCookies())

	cResp.SetRequest(cReq)

	cResp.SetResponse(resp)

	if err != nil {
		cResp.SetStatus(false, err.Error())
		return cResp
	}

	cResp.SetStatus(true, "")
	return cResp
}
