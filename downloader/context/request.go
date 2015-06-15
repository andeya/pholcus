package context

import (
	"github.com/bitly/go-simplejson"
	"github.com/henrylee2cn/pholcus/reporter"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Request represents object waiting for being crawled.
type Request struct {
	url     string
	referer string
	rule    string
	spider  string
	// GET POST POST-M HEAD
	method string
	// http header
	header http.Header
	// http cookies
	cookies []*http.Cookie
	// POST values
	postData url.Values
	//在Spider中生成时，根据ruleTree.Outsource确定
	canOutsource bool
	//当经过Pholcus时，被指定是否外包
	isOutsource bool
	// Redirect function for downloader used in http.Client
	// If CheckRedirect returns an error, the Client's Get
	// method returns both the previous Response.
	// If CheckRedirect returns error.New("normal"), the error process after client.Do will ignore the error.
	checkRedirect func(req *http.Request, via []*http.Request) error

	// 标记临时数据，通过temp[x]==nil判断是否有值存入，所以请存入带类型的值，如[]int(nil)等
	temp map[string]interface{}

	// 即将加入哪个优先级的队列当中
	priority uint
}

// NewRequest returns initialized Request object.

func NewRequest(param map[string]interface{}) *Request {
	req := &Request{
		url:    param["url"].(string),    //必填
		rule:   param["rule"].(string),   //必填
		spider: param["spider"].(string), //必填
	}

	// 若有必填
	switch v := param["referer"].(type) {
	case string:
		req.referer = v
	default:
		req.referer = ""
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

	switch v := param["postData"].(type) {
	case url.Values:
		req.postData = v
	default:
		req.postData = nil
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

	switch v := param["temp"].(type) {
	case map[string]interface{}:
		req.temp = v
	default:
		req.temp = map[string]interface{}{}
	}

	switch v := param["priority"].(type) {
	case uint:
		req.priority = v
	default:
		req.priority = uint(0)
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

func (self *Request) GetHeader() http.Header {
	return self.header
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

func (self *Request) GetReferer() string {
	return self.referer
}

func (self *Request) SetReferer(referer string) {
	self.referer = referer
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

func (self *Request) GetMethod() string {
	return strings.ToUpper(self.method)
}

func (self *Request) GetPostData() url.Values {
	return self.postData
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

func (self *Request) GetPriority() uint {
	return self.priority
}

func (self *Request) SetPriority(priority uint) {
	self.priority = priority
}
