package surfer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type (
	Chrome struct {
		Browser *rod.Browser
	}
)

func NewChrome(jar ...*cookiejar.Jar) Surfer {
	chrome := &Chrome{}
	browser := rod.New().MustConnect()
	chrome.Browser = browser
	return chrome
}

// 实现surfer下载器接口
func (self *Chrome) Download(req Request) (resp *http.Response, err error) {
	param, err := NewParam(req)
	if err != nil {
		return
	}
	resp = param.writeback(resp)

	var html string
	var res *proto.NetworkResponse
	err = rod.Try(func() {
		page := self.Browser.MustPage()
		defer page.MustClose()

		e := proto.NetworkResponseReceived{}
		wait := page.Timeout(10 * time.Second).WaitEvent(&e)
		page.MustNavigate(req.GetUrl())
		wait()
		if e.Response.Status != 200 {
			panic(fmt.Errorf("status code is %v", e.Response.Status))
		}
		res = e.Response
		page.Timeout(60 * time.Second).MustWaitLoad()
		html = page.MustHTML()
	})
	if err != nil {
		return
	}

	resp.Request.URL = param.url
	resp.Body = ioutil.NopCloser(strings.NewReader(html))
	resp.StatusCode = res.Status
	resp.Status = res.StatusText
	for k, v := range res.Headers {
		for _, vv := range v.Arr() {
			resp.Header.Add(k, vv.Str())
		}
	}
	return
}
