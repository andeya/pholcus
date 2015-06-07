package context

import (
	"github.com/bitly/go-simplejson"
	"github.com/henrylee2cn/pholcus/reporter"
	"io/ioutil"
	"net/http"
	"os"
)

// Request represents object waiting for being crawled.
type Request struct {
	url    string
	parent string
	rule   string
	spider string
	// Responce type: html json jsonp text
	respType string
	// GET POST
	method string
	// http header
	header http.Header
	// http cookies
	cookies []*http.Cookie
	// POST data
	postdata string
	//在Spider中生成时，根据ruleTree.Outsource确定
	canOutsource bool
	//当经过Pholcus时，被指定是否外包
	isOutsource bool
	// Redirect function for downloader used in http.Client
	// If CheckRedirect returns an error, the Client's Get
	// method returns both the previous Response.
	// If CheckRedirect returns error.New("normal"), the error process after client.Do will ignore the error.
	checkRedirect func(req *http.Request, via []*http.Request) error
	//proxy host   example='localhost:80'
	proxyHost string
	// 标记临时数据，通过temp[x]==nil判断是否有值存入，所以请存入带类型的值，如[]int(nil)等
	temp map[string]interface{}
}

// NewRequest returns initialized Request object.
// The respType is json, jsonp, html, text

func NewRequest(param map[string]interface{}) *Request {
	req := &Request{
		url:    param["url"].(string),    //必填
		rule:   param["rule"].(string),   //必填
		spider: param["spider"].(string), //必填
	}

	// 若有必填
	switch v := param["parent"].(type) {
	case string:
		req.parent = v
	default:
		req.parent = ""
	}

	switch v := param["respType"].(type) {
	case string:
		req.respType = v
	default:
		req.respType = "html"
	}

	switch v := param["method"].(type) {
	case string:
		req.method = v
	default:
		req.method = "GET"
	}

	switch v := param["cookies"].(type) {
	case []*http.Cookie:
		req.cookies = v
	default:
		req.cookies = nil
	}

	switch v := param["postdata"].(type) {
	case string:
		req.postdata = v
	default:
		req.postdata = ""
	}

	switch v := param["canOutsource"].(type) {
	case bool:
		req.canOutsource = v
	default:
		req.canOutsource = false
	}

	switch v := param["checkRedirect"].(type) {
	case func(*http.Request, []*http.Request) error:
		req.checkRedirect = v
	default:
		req.checkRedirect = nil
	}

	switch v := param["proxyHost"].(type) {
	case string:
		req.proxyHost = v
	default:
		req.proxyHost = ""
	}

	switch v := param["temp"].(type) {
	case map[string]interface{}:
		req.temp = v
	default:
		req.temp = map[string]interface{}{}
	}

	switch v := param["header"].(type) {
	case string:
		_, err := os.Stat(v)
		if err == nil {
			req.header = readHeaderFromFile(v)
		}
	case http.Header:
		req.header = v
	default:
		req.header = nil
	}

	return req
}

func readHeaderFromFile(headerFile string) http.Header {
	//read file , parse the header and cookies
	b, err := ioutil.ReadFile(headerFile)
	if err != nil {
		//make be:  share access error
		reporter.Log.Println(err.Error())
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
	self.header = h
	return self
}

// @host  http://localhost:8765/
func (self *Request) AddProxyHost(host string) *Request {
	self.proxyHost = host
	return self
}

func (self *Request) GetHeader() http.Header {
	return self.header
}

func (self *Request) GetProxyHost() string {
	return self.proxyHost
}

func (self *Request) GetRedirectFunc() func(req *http.Request, via []*http.Request) error {
	return self.checkRedirect
}

func (self *Request) GetUrl() string {
	return self.url
}

func (self *Request) SetUrl(url string) {
	self.url = url
}

func (self *Request) GetParent() string {
	return self.parent
}

func (self *Request) SetParent(parent string) {
	self.parent = parent
}

func (self *Request) GetRuleName() string {
	return self.rule
}

func (self *Request) SetRuleName(ruleName string) {
	self.rule = ruleName
}

func (self *Request) GetSpiderName() string {
	return self.spider
}

func (self *Request) GetRespType() string {
	return self.respType
}

func (self *Request) GetMethod() string {
	return self.method
}

func (self *Request) GetPostdata() string {
	return self.postdata
}

func (self *Request) GetCookies() []*http.Cookie {
	return self.cookies
}

func (self *Request) IsOutsource() bool {
	return self.isOutsource
}

func (self *Request) TryOutsource() bool {
	if self.canOutsource {
		self.isOutsource = true
		return true
	} else {
		return false
	}
}

func (self *Request) GetTemp(key string) interface{} {
	return self.temp[key]
}

func (self *Request) GetTemps() map[string]interface{} {
	return self.temp
}

func (self *Request) SetTemp(key string, value interface{}) {
	self.temp[key] = value
}

func (self *Request) SetAllTemps(temp map[string]interface{}) {
	self.temp = temp
}

func (self *Request) GetSpiderId() (int, bool) {
	value, ok := self.temp["__SPIDER_ID__"]
	return value.(int), ok
}

func (self *Request) SetSpiderId(spiderId int) {
	self.temp["__SPIDER_ID__"] = spiderId
}
