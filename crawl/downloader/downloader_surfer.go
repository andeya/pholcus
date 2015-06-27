package downloader

import (
	"github.com/henrylee2cn/pholcus/crawl/downloader/context"
	"github.com/henrylee2cn/pholcus/reporter"
	"github.com/henrylee2cn/surfer"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
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

	if err != nil {
		cResp.SetStatus(true, err.Error())
		return cResp
	}

	// get converter to utf-8
	body := self.changeCharsetEncodingAuto(resp.Body, resp.Header.Get("Content-Type"))
	//fmt.Printf("utf-8 body %v \r\n", bodyStr)
	defer resp.Body.Close()
	cResp.SetText(body)
	cResp.SetStatus(false, "")
	return cResp
}

// Charset auto determine. Use golang.org/x/net/html/charset. Get response body and change it to utf-8
func (self *Surfer) changeCharsetEncodingAuto(sor io.ReadCloser, contentTypeStr string) string {
	var err error
	destReader, err := charset.NewReader(sor, contentTypeStr)

	if err != nil {
		reporter.Log.Println(err.Error())
		destReader = sor
	}

	var sorbody []byte
	if sorbody, err = ioutil.ReadAll(destReader); err != nil {
		reporter.Log.Println(err.Error())
		// For gb2312, an error will be returned.
		// Error like: simplifiedchinese: invalid GBK encoding
		// return ""
	}
	//e,name,certain := charset.DetermineEncoding(sorbody,contentTypeStr)
	bodystr := string(sorbody)

	return bodystr
}
