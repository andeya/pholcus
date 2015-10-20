package context

import (
	"github.com/henrylee2cn/pholcus/common/util"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultDialTimeout = 2 * time.Minute // 默认请求服务器超时
	DefaultConnTimeout = 2 * time.Minute // 默认下载超时
	DefaultTryTimes    = 3               // 默认最大下载次数
	DefaultRetryPause  = 2 * time.Second // 默认重新下载前停顿时长
)

// Request represents object waiting for being crawled.
type Request struct {
	Spider string // *规则中无需手动指定

	Url  string // *必须设置
	Rule string // *必须设置

	Referer string
	// GET POST POST-M HEAD
	Method string
	// http header
	Header http.Header
	// enable http cookies
	EnableCookie bool
	// http cookies, when Cookies equal nil, the UserAgent auto changes
	Cookies []*http.Cookie
	// POST values
	PostData url.Values
	// dial tcp: i/o timeout
	DialTimeout time.Duration
	// WSARecv tcp: i/o timeout
	ConnTimeout time.Duration
	// the max times of download
	TryTimes int
	// how long pause when retry
	RetryPause time.Duration
	// max redirect times
	// when RedirectTimes equal 0, redirect times is ∞
	// when RedirectTimes less than 0, redirect times is 0
	RedirectTimes int
	// the download ProxyHost
	Proxy string

	// 标记临时数据，通过temp[x]==nil判断是否有值存入，所以请存入带类型的值，如[]int(nil)等
	Temp map[string]interface{}

	// 即将加入哪个优先级的队列当中，默认为0，最小优先级为0
	Priority int

	// 是否可以被重复下载（即不被去重）
	Duplicatable bool

	//是否使用PhantomJS下载器，特点破防力强，速度慢
	UsePhantomJS bool
}

// 发送请求前的准备工作，设置一系列默认值
// Request.Url与Request.Rule必须设置
// Request.Spider无需手动设置(由系统自动设置)
// Request.EnableCookie在Spider字段中统一设置，规则请求中指定的无效
// 以下字段有默认值，可不设置:
// Request.Method默认为GET方法;
// Request.DialTimeout默认为常量DefaultDialTimeout，小于0时不限制等待响应时长;
// Request.ConnTimeout默认为常量DefaultConnTimeout，小于0时不限制下载超时;
// Request.TryTimes默认为常量DefaultTryTimes，小于0时不限制失败重载次数;
// Request.RedirectTimes默认不限制重定向次数，小于0时可禁止重定向跳转;
// Request.RetryPause默认为常量DefaultRetryPause;
// Request.UsePhantomJS为true时，使用PhantomJS下载器下载，破防力强，速度慢，暂不支持图片下载。
func (self *Request) Prepare() *Request {
	if self.Method == "" {
		self.Method = "GET"
	}

	if self.DialTimeout < 0 {
		self.DialTimeout = 0
	} else if self.DialTimeout == 0 {
		self.DialTimeout = DefaultDialTimeout
	}

	if self.ConnTimeout < 0 {
		self.ConnTimeout = 0
	} else if self.ConnTimeout == 0 {
		self.ConnTimeout = DefaultConnTimeout
	}

	if self.TryTimes == 0 {
		self.TryTimes = DefaultTryTimes
	}

	if self.RetryPause <= 0 {
		self.RetryPause = DefaultRetryPause
	}

	if self.Priority < 0 {
		self.Priority = 0
	}
	return self
}

// 请求的序列化
func (self *Request) Serialization() string {
	return util.JsonString(self)
}

func (self *Request) GetUrl() string {
	return self.Url
}

func (self *Request) GetMethod() string {
	return strings.ToUpper(self.Method)
}

func (self *Request) SetUrl(url string) *Request {
	self.Url = url
	return self
}

func (self *Request) GetReferer() string {
	return self.Referer
}

func (self *Request) SetReferer(referer string) *Request {
	self.Referer = referer
	return self
}

func (self *Request) GetPostData() url.Values {
	return self.PostData
}

func (self *Request) GetHeader() http.Header {
	return self.Header
}

func (self *Request) GetEnableCookie() bool {
	return self.EnableCookie
}

func (self *Request) SetEnableCookie(enableCookie bool) *Request {
	self.EnableCookie = enableCookie
	return self
}

func (self *Request) GetCookies() []*http.Cookie {
	return self.Cookies
}

func (self *Request) GetDialTimeout() time.Duration {
	return self.DialTimeout
}

func (self *Request) GetConnTimeout() time.Duration {
	return self.ConnTimeout
}

func (self *Request) GetTryTimes() int {
	return self.TryTimes
}

func (self *Request) GetRetryPause() time.Duration {
	return self.RetryPause
}

func (self *Request) GetProxy() string {
	return self.Proxy
}

func (self *Request) GetRedirectTimes() int {
	return self.RedirectTimes
}

func (self *Request) GetRuleName() string {
	return self.Rule
}

func (self *Request) SetRuleName(ruleName string) *Request {
	self.Rule = ruleName
	return self
}

func (self *Request) GetSpiderName() string {
	return self.Spider
}

func (self *Request) SetSpiderName(spiderName string) *Request {
	self.Spider = spiderName
	return self
}

func (self *Request) GetDuplicatable() bool {
	return self.Duplicatable
}

func (self *Request) GetTemp(key string) interface{} {
	return self.Temp[key]
}

func (self *Request) GetTemps() map[string]interface{} {
	return self.Temp
}

func (self *Request) SetTemp(key string, value interface{}) *Request {
	self.Temp[key] = value
	return self
}

func (self *Request) SetAllTemps(temp map[string]interface{}) *Request {
	self.Temp = temp
	return self
}

func (self *Request) GetSpiderId() (int, bool) {
	value, ok := self.Temp["__SPIDER_ID__"]
	return value.(int), ok
}

func (self *Request) SetSpiderId(spiderId int) *Request {
	if self.Temp == nil {
		self.Temp = map[string]interface{}{}
	}
	self.Temp["__SPIDER_ID__"] = spiderId
	return self
}

func (self *Request) GetPriority() int {
	return self.Priority
}

func (self *Request) SetPriority(priority int) *Request {
	self.Priority = priority
	return self
}

func (self *Request) GetUsePhantomJS() bool {
	return self.UsePhantomJS
}
