package downloader

import (
	"time"

	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/surfer"
)

type Surfer struct {
	download surfer.Surfer
}

func NewSurfer() *Surfer {
	return &Surfer{
		download: surfer.New(),
	}
}

func (self *Surfer) Download(cReq *context.Request) *context.Response {
	cResp := context.NewResponse(nil)

	resp, err := self.download.Download(cReq)

	cResp.SetRequest(cReq)

	cResp.SetResponse(resp)

	cResp.SetError(err)

	return cResp
}

func (self *Surfer) SetUseCookie(use bool) Downloader {
	self.download.SetUseCookie(use)
	return self
}

func (self *Surfer) SetPauseTime(pauseTime time.Duration) Downloader {
	self.download.SetPauseTime(pauseTime)
	return self
}

func (self *Surfer) SetDeadline(deadline time.Duration) Downloader {
	self.download.SetDeadline(deadline)
	return self
}

func (self *Surfer) SetProxy(proxy string) Downloader {
	self.download.SetProxy(proxy)
	return self
}
