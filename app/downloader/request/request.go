package request

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/henrylee2cn/pholcus/common/util"
)

// Request represents object waiting for being crawled.
type Request struct {
	Spider        string          //规则名，自动设置，禁止人为填写
	Url           string          //目标URL，必须设置
	Rule          string          //用于解析响应的规则节点名，必须设置
	Method        string          //GET POST POST-M HEAD
	Header        http.Header     //请求头信息
	EnableCookie  bool            //是否使用cookies，在Spider的EnableCookie设置
	PostData      string          //POST values
	DialTimeout   time.Duration   //创建连接超时 dial tcp: i/o timeout
	ConnTimeout   time.Duration   //连接状态超时 WSARecv tcp: i/o timeout
	TryTimes      int             //尝试下载的最大次数
	RetryPause    time.Duration   //下载失败后，下次尝试下载的等待时间
	RedirectTimes int             //重定向的最大次数，为0时不限，小于0时禁止重定向
	Temp          Temp            //临时数据
	TempIsJson    map[string]bool //将Temp中以JSON存储的字段标记为true，自动设置，禁止人为填写
	Priority      int             //指定调度优先级，默认为0（最小优先级为0）
	Reloadable    bool            //是否允许重复该链接下载
	//Surfer下载器内核ID
	//0为Surf高并发下载器，各种控制功能齐全
	//1为PhantomJS下载器，特点破防力强，速度慢，低并发
	DownloaderID int

	proxy  string //当用户界面设置可使用代理IP时，自动设置代理
	unique string //ID
	lock   sync.RWMutex
}

const (
	DefaultDialTimeout = 2 * time.Minute // 默认请求服务器超时
	DefaultConnTimeout = 2 * time.Minute // 默认下载超时
	DefaultTryTimes    = 3               // 默认最大下载次数
	DefaultRetryPause  = 2 * time.Second // 默认重新下载前停顿时长
)

const (
	SURF_ID    = 0 // 默认的surf下载内核（Go原生），此值不可改动
	PHANTOM_ID = 1 // 备用的phantomjs下载内核，一般不使用（效率差，头信息支持不完善）
)

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

	if self.DownloaderID < SURF_ID || self.DownloaderID > PHANTOM_ID {
		self.DownloaderID = SURF_ID
	}

	if self.TempIsJson == nil {
		self.TempIsJson = make(map[string]bool)
	}

	if self.Temp == nil {
		self.Temp = make(Temp)
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
	for k, v := range self.Temp {
		self.Temp.set(k, v)
		self.TempIsJson[k] = true
	}
	b, _ := json.Marshal(self)
	return strings.Replace(util.Bytes2String(b), `\u0026`, `&`, -1)
}

// 请求的唯一识别码
func (self *Request) Unique() string {
	if self.unique == "" {
		block := md5.Sum([]byte(self.Spider + self.Rule + self.Url + self.Method))
		self.unique = hex.EncodeToString(block[:])
	}
	return self.unique
}

// 获取副本
func (self *Request) Copy() *Request {
	reqcopy := new(Request)
	b, _ := json.Marshal(self)
	json.Unmarshal(b, reqcopy)
	return reqcopy
}

// 获取Url
func (self *Request) GetUrl() string {
	return self.Url
}

// 获取Http请求的方法名称 (注意这里不是指Http GET方法)
func (self *Request) GetMethod() string {
	return self.Method
}

// 设定Http请求方法的类型
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

func (self *Request) GetPostData() string {
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
	return self.proxy
}

func (self *Request) SetProxy(proxy string) *Request {
	self.proxy = proxy
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

// 获取临时缓存数据
// defaultValue 不能为 interface{}(nil)
func (self *Request) GetTemp(key string, defaultValue interface{}) interface{} {
	if defaultValue == nil {
		panic("*Request.GetTemp()的defaultValue不能为nil，错误位置：key=" + key)
	}
	self.lock.RLock()
	defer self.lock.RUnlock()

	if self.Temp[key] == nil {
		return defaultValue
	}

	if self.TempIsJson[key] {
		return self.Temp.get(key, defaultValue)
	}

	return self.Temp[key]
}

func (self *Request) GetTemps() Temp {
	return self.Temp
}

func (self *Request) SetTemp(key string, value interface{}) *Request {
	self.lock.Lock()
	self.Temp[key] = value
	delete(self.TempIsJson, key)
	self.lock.Unlock()
	return self
}

func (self *Request) SetTemps(temp map[string]interface{}) *Request {
	self.lock.Lock()
	self.Temp = temp
	self.TempIsJson = make(map[string]bool)
	self.lock.Unlock()
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

func (self *Request) MarshalJSON() ([]byte, error) {
	for k, v := range self.Temp {
		if self.TempIsJson[k] {
			continue
		}
		self.Temp.set(k, v)
		self.TempIsJson[k] = true
	}
	b, err := json.Marshal(*self)
	return b, err
}
