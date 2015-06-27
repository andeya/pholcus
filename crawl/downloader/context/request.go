package context

import (
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Request represents object waiting for being crawled.
type Request struct {
	Url     string
	Referer string
	Rule    string
	Spider  string
	// GET POST POST-M HEAD
	Method string
	// http header
	Header http.Header
	// http cookies
	Cookies []*http.Cookie
	// POST values
	PostData url.Values
	//是否支持外包（分布式），根据ruleTree.Outsource确定
	Outsource bool
	// Redirect function for downloader used in http.Client
	// If CheckRedirect returns an error, the Client's Get
	// method returns both the previous Response.
	// If CheckRedirect returns error.New("normal"), the error process after client.Do will ignore the error.
	CheckRedirect func(req *http.Request, via []*http.Request) error

	// 标记临时数据，通过temp[x]==nil判断是否有值存入，所以请存入带类型的值，如[]int(nil)等
	Temp map[string]interface{}

	// 即将加入哪个优先级的队列当中
	Priority uint
}

// NewRequest returns initialized Request object.

func NewRequest(param map[string]interface{}) *Request {
	req := &Request{
		Url:    param["url"].(string),    //必填
		Rule:   param["rule"].(string),   //必填
		Spider: param["spider"].(string), //必填
	}

	// 若有必填
	switch v := param["referer"].(type) {
	case string:
		req.Referer = v
	default:
		req.Referer = ""
	}

	switch v := param["method"].(type) {
	case string:
		req.Method = v
	default:
		req.Method = "GET"
	}

	switch v := param["cookies"].(type) {
	case []*http.Cookie:
		req.Cookies = v
	default:
		req.Cookies = nil
	}

	switch v := param["postData"].(type) {
	case url.Values:
		req.PostData = v
	default:
		req.PostData = nil
	}

	switch v := param["outsource"].(type) {
	case bool:
		req.Outsource = v
	default:
		req.Outsource = false
	}

	switch v := param["checkRedirect"].(type) {
	case func(*http.Request, []*http.Request) error:
		req.CheckRedirect = v
	default:
		req.CheckRedirect = nil
	}

	switch v := param["temp"].(type) {
	case map[string]interface{}:
		req.Temp = v
	default:
		req.Temp = map[string]interface{}{}
	}

	switch v := param["priority"].(type) {
	case uint:
		req.Priority = v
	default:
		req.Priority = uint(0)
	}

	switch v := param["header"].(type) {
	case string:
		_, err := os.Stat(v)
		if err == nil {
			req.Header = readHeaderFromFile(v)
		}
	case http.Header:
		req.Header = v
	default:
		req.Header = nil
	}

	return req
}

func readHeaderFromFile(headerFile string) http.Header {
	//read file , parse the header and cookies
	b, err := ioutil.ReadFile(headerFile)
	if err != nil {
		//make be:  share access error
		log.Println(err.Error())
		return nil
	}
	js, _ := simplejson.NewJson(b)
	//constructed to header

	h := make(http.Header)
	h.Add("User-Agent", js.Get("User-Agent").MustString())
	h.Add("Referer", js.Get("Referer").MustString())
	h.Add("Cookie", js.Get("Cookie").MustString())
	h.Add("Cache-Control", "max-age=0")
	h.Add("Connection", "keep-alive")
	return h
}

//point to a json file
/* xxx.json
{
	"User-Agent":"curl/7.19.3 (i386-pc-win32) libcurl/7.19.3 OpenSSL/1.0.0d",
	"Referer":"http://weixin.sogou.com/gzh?openid=oIWsFt6Sb7aZmuI98AU7IXlbjJps",
	"Cookie":""
}
*/
func (self *Request) AddHeaderFile(headerFile string) *Request {
	_, err := os.Stat(headerFile)
	if err != nil {
		return self
	}
	h := readHeaderFromFile(headerFile)
	self.Header = h
	return self
}

func (self *Request) GetHeader() http.Header {
	return self.Header
}

func (self *Request) GetRedirectFunc() func(req *http.Request, via []*http.Request) error {
	return self.CheckRedirect
}

func (self *Request) GetUrl() string {
	return self.Url
}

func (self *Request) SetUrl(url string) {
	self.Url = url
}

func (self *Request) GetReferer() string {
	return self.Referer
}

func (self *Request) SetReferer(referer string) {
	self.Referer = referer
}

func (self *Request) GetRuleName() string {
	return self.Rule
}

func (self *Request) SetRuleName(ruleName string) {
	self.Rule = ruleName
}

func (self *Request) GetSpiderName() string {
	return self.Spider
}

func (self *Request) GetMethod() string {
	return strings.ToUpper(self.Method)
}

func (self *Request) GetPostData() url.Values {
	return self.PostData
}

func (self *Request) GetCookies() []*http.Cookie {
	return self.Cookies
}

func (self *Request) CanOutsource() bool {
	return self.Outsource
}

func (self *Request) SetOutsource(can bool) {
	self.Outsource = can
}

func (self *Request) GetTemp(key string) interface{} {
	return self.Temp[key]
}

func (self *Request) GetTemps() map[string]interface{} {
	return self.Temp
}

func (self *Request) SetTemp(key string, value interface{}) {
	self.Temp[key] = value
}

func (self *Request) SetAllTemps(temp map[string]interface{}) {
	self.Temp = temp
}

func (self *Request) GetSpiderId() (int, bool) {
	value, ok := self.Temp["__SPIDER_ID__"]
	return value.(int), ok
}

func (self *Request) SetSpiderId(spiderId int) {
	self.Temp["__SPIDER_ID__"] = spiderId
}

func (self *Request) GetPriority() uint {
	return self.Priority
}

func (self *Request) SetPriority(priority uint) {
	self.Priority = priority
}
