package context

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
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

	// GET POST POST-M HEAD
	Method string
	// http header
	Header http.Header
	// 是否使用cookies，在Spider的EnableCookie设置
	EnableCookie bool
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
	Temp Temp

	// 即将加入哪个优先级的队列当中，默认为0，最小优先级为0
	Priority int

	// 是否允许重复下载
	Reloadable bool

	// 指定下载器ID
	// 0为Surf高并发下载器，各种控制功能齐全
	// 1为PhantomJS下载器，特点破防力强，速度慢，低并发
	DownloaderID int
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
// Request.DownloaderID指定下载器ID，0为默认的Surf高并发下载器，功能完备，1为PhantomJS下载器，特点破防力强，速度慢，低并发。
func (self *Request) Prepare() error {
	// 确保url正确，且和Response中Url字符串相等
	URL, err := url.Parse(self.Url)
	if err != nil {
		return err
	} else {
		self.Url = URL.String()
	}

	if self.Method == "" {
		self.Method = "GET"
	} else {
		self.Method = strings.ToUpper(self.Method)
	}

	if self.Header == nil {
		self.Header = make(http.Header)
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

	if self.DownloaderID < 0 || self.DownloaderID > 1 {
		self.DownloaderID = 0
	}
	return nil
}

// 反序列化
func UnSerialize(s string) (*Request, error) {
	req := new(Request)
	return req, json.Unmarshal([]byte(s), req)
}

// 序列化
func (self *Request) Serialize() string {
	b, _ := json.Marshal(self)
	return string(b)
}

// 获取副本
func (self *Request) Copy() *Request {
	temp := self.Temp
	self.Temp = make(map[string]interface{})
	b, _ := json.Marshal(self)

	reqcopy := new(Request)
	json.Unmarshal(b, reqcopy)
	reqcopy.Temp = make(map[string]interface{})
	for k := range temp {
		reqcopy.Temp[k] = temp[k]
	}

	self.Temp = temp

	return reqcopy
}

// 返回请求前的Url
// 请求后的Url将会在downloader模块被重置为该Url，从而确保请求前后Url字符串相等，且中文不被编码
func (self *Request) GetUrl() string {
	return self.Url
}

func (self *Request) GetMethod() string {
	return self.Method
}

func (self *Request) SetMethod(method string) *Request {
	self.Method = strings.ToUpper(method)
	return self
}

func (self *Request) SetUrl(url string) *Request {
	self.Url = url
	return self
}

func (self *Request) GetReferer() string {
	return self.Header.Get("Referer")
}

func (self *Request) SetReferer(referer string) *Request {
	self.Header.Set("Referer", referer)
	return self
}

func (self *Request) GetPostData() url.Values {
	return self.PostData
}

func (self *Request) GetHeader() http.Header {
	return self.Header
}

func (self *Request) SetHeader(key, value string) *Request {
	self.Header.Set(key, value)
	return self
}

func (self *Request) AddHeader(key, value string) *Request {
	self.Header.Add(key, value)
	return self
}

func (self *Request) GetEnableCookie() bool {
	return self.EnableCookie
}

func (self *Request) SetEnableCookie(enableCookie bool) *Request {
	self.EnableCookie = enableCookie
	return self
}

func (self *Request) GetCookies() string {
	return self.Header.Get("Cookie")
}

func (self *Request) SetCookies(cookie string) *Request {
	self.Header.Set("Cookie", cookie)
	return self
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

func (self *Request) SetProxy(proxy string) *Request {
	self.Proxy = proxy
	return self
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

func (self *Request) IsReloadable() bool {
	return self.Reloadable
}

func (self *Request) SetReloadable(can bool) *Request {
	self.Reloadable = can
	return self
}

// 返回临时缓存数据
// 强烈建议数据接收者receive为指针类型
// receive为空时，直接输出字符串
func (self *Request) GetTemp(key string, receive ...interface{}) interface{} {
	if _, ok := self.Temp[key]; !ok {
		return nil
	}
	if len(receive) == 0 {
		// 默认输出字符串格式
		receive = append(receive, "")
	}
	b, _ := json.Marshal(self.Temp[key])
	if reflect.ValueOf(receive[0]).Kind() != reflect.Ptr {
		json.Unmarshal(b, &(receive[0]))
	} else {
		json.Unmarshal(b, receive[0])
	}

	return receive[0]
}

func (self *Request) GetTemps() Temp {
	return self.Temp
}

func (self *Request) SetTemp(key string, value interface{}) *Request {
	b, _ := json.Marshal(value)
	self.Temp[key] = string(b)
	return self
}

func (self *Request) SetTemps(temp map[string]interface{}) *Request {
	self.Temp = make(Temp)
	for k, v := range temp {
		self.SetTemp(k, v)
	}
	return self
}

func (self *Request) GetSpiderId() (spiderId int) {
	return self.GetTemp("__SPIDER_ID__", spiderId).(int)
}

func (self *Request) SetSpiderId(spiderId int) *Request {
	if self.Temp == nil {
		self.Temp = make(Temp)
	}
	self.SetTemp("__SPIDER_ID__", spiderId)
	return self
}

func (self *Request) GetPriority() int {
	return self.Priority
}

func (self *Request) SetPriority(priority int) *Request {
	self.Priority = priority
	return self
}

func (self *Request) GetDownloaderID() int {
	return self.DownloaderID
}

func (self *Request) SetDownloaderID(id int) *Request {
	self.DownloaderID = id
	return self
}
