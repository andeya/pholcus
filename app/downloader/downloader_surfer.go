package downloader

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/surfer"

	"log"
	"time"
)

type Surfer struct {
	download surfer.Surfer
}

func NewSurfer(useCookie bool, paseTime time.Duration, proxy ...string) *Surfer {
	sf := surfer.New()
	if len(proxy) != 0 {
		sf.SetProxy(proxy[0])
	}

	return &Surfer{
		download: sf,
	}
}

func (self *Surfer) Download(cReq *context.Request) *context.Response {
	cResp := context.NewResponse(nil)

	resp, err := self.download.Download(cReq.GetMethod(), cReq.GetUrl(), cReq.GetReferer(), cReq.GetPostData(), cReq.GetHeader(), cReq.GetCookies())

	cResp.SetRequest(cReq)

	cResp.SetResponse(resp)

	if err != nil {
		log.Println(" *     ", err)
		// cResp.SetStatus(false, err.Error())
		// return cResp
	}

	cResp.SetStatus(true, "")
	return cResp
}

func (self *Surfer) SetUseCookie(use bool) Downloader {
	self.download.SetUseCookie(use)
	return self
}

func (self *Surfer) SetPaseTime(paseTime time.Duration) Downloader {
	self.download.SetPaseTime(paseTime)
	return self
}

func (self *Surfer) SetProxy(proxy string) Downloader {
	self.download.SetProxy(proxy)
	return self
}
