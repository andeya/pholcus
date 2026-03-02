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

	"github.com/andeya/gust/option"
	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/util"
)

// Request represents object waiting for being crawled.
type Request struct {
	Spider        string          // spider name, auto-set, do not set manually
	Url           string          // target URL, required
	Rule          string          // rule node name for parsing response, required
	Method        string          // GET POST POST-M HEAD
	Header        http.Header     // request headers
	EnableCookie  bool            // whether to use cookies, set in Spider.EnableCookie
	PostData      string          // POST values
	DialTimeout   time.Duration   // dial timeout (dial tcp: i/o timeout)
	ConnTimeout   time.Duration   // connection timeout (WSARecv tcp: i/o timeout)
	TryTimes      int             // max download retry attempts
	RetryPause    time.Duration   // wait time before retry after download failure
	RedirectTimes int             // max redirects; 0=unlimited, <0=no redirects
	Temp          Temp            // temporary data
	TempIsJson    map[string]bool // marks Temp fields stored as JSON; auto-set, do not set manually
	Priority      int             // scheduling priority, default 0 (min priority)
	Reloadable    bool            // whether the link can be re-downloaded
	// DownloaderID: 0=Surf (high concurrency, full features), 1=PhantomJS (strong anti-block, slow, low concurrency)
	DownloaderID int

	proxy  string // proxy, auto-set when UI enables proxy
	unique string // unique ID
	lock   sync.RWMutex
}

const (
	DefaultDialTimeout = 2 * time.Minute // default server request timeout
	DefaultConnTimeout = 2 * time.Minute // default download timeout
	DefaultTryTimes    = 3               // default max download attempts
	DefaultRetryPause  = 2 * time.Second // default pause before retry
)

const (
	SURF_ID    = 0 // Surf downloader (native Go), do not change
	PHANTOM_ID = 1 // PhantomJS downloader (fallback, rarely used)
)

// Prepare sets default values before sending a request.
// Request.Url and Request.Rule must be set.
// Request.Spider is auto-set by the system.
// Request.EnableCookie is set in Spider; per-request values are ignored.
// Optional fields with defaults: Method (GET), DialTimeout, ConnTimeout, TryTimes,
// RedirectTimes, RetryPause, DownloaderID (0=Surf, 1=PhantomJS).
func (self *Request) Prepare() result.VoidResult {
	URL, err := url.Parse(self.Url)
	if err != nil {
		return result.TryErrVoid(err)
	}
	self.Url = URL.String()

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
	return result.OkVoid()
}

// UnSerialize deserializes a Request from JSON string.
func UnSerialize(s string) result.Result[*Request] {
	req := new(Request)
	return result.Ret(req, json.Unmarshal([]byte(s), req))
}

// Serialize serializes the Request to JSON string.
func (self *Request) Serialize() result.Result[string] {
	for k, v := range self.Temp {
		self.Temp.set(k, v)
		self.TempIsJson[k] = true
	}
	b, err := json.Marshal(self)
	if err != nil {
		return result.TryErr[string](err)
	}
	return result.Ok(strings.ReplaceAll(util.Bytes2String(b), `\u0026`, `&`))
}

// Unique returns the unique identifier for the request.
func (self *Request) Unique() string {
	if self.unique == "" {
		block := md5.Sum([]byte(self.Spider + self.Rule + self.Url + self.Method))
		self.unique = hex.EncodeToString(block[:])
	}
	return self.unique
}

// Copy returns a deep copy of the request.
func (self *Request) Copy() result.Result[*Request] {
	reqcopy := new(Request)
	b, err := json.Marshal(self)
	if err != nil {
		return result.TryErr[*Request](err)
	}
	return result.Ret(reqcopy, json.Unmarshal(b, reqcopy))
}

// GetUrl returns the request URL.
func (self *Request) GetUrl() string {
	return self.Url
}

// GetMethod returns the HTTP method name (e.g. GET, POST).
func (self *Request) GetMethod() string {
	return self.Method
}

// SetMethod sets the HTTP method.
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

// GetTemp returns temporary cached data. defaultValue must not be nil.
func (self *Request) GetTemp(key string, defaultValue interface{}) interface{} {
	if defaultValue == nil {
		panic("*Request.GetTemp() defaultValue must not be nil, key=" + key)
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

// GetTempOpt returns temporary cached data as Option. None when key is missing.
func (self *Request) GetTempOpt(key string) option.Option[interface{}] {
	self.lock.RLock()
	defer self.lock.RUnlock()

	if _, ok := self.Temp[key]; !ok {
		return option.None[interface{}]()
	}
	if self.TempIsJson[key] {
		var v interface{}
		self.Temp.get(key, &v)
		return option.Some(v)
	}
	return option.Some(self.Temp[key])
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
	// Marshal a struct without the mutex to avoid copying sync.RWMutex
	j := struct {
		Spider        string
		Url           string
		Rule          string
		Method        string
		Header        http.Header
		EnableCookie  bool
		PostData      string
		DialTimeout   time.Duration
		ConnTimeout   time.Duration
		TryTimes      int
		RetryPause    time.Duration
		RedirectTimes int
		Temp          Temp
		TempIsJson    map[string]bool
		Priority      int
		Reloadable    bool
		DownloaderID  int
	}{
		Spider:        self.Spider,
		Url:           self.Url,
		Rule:          self.Rule,
		Method:        self.Method,
		Header:        self.Header,
		EnableCookie:  self.EnableCookie,
		PostData:      self.PostData,
		DialTimeout:   self.DialTimeout,
		ConnTimeout:   self.ConnTimeout,
		TryTimes:      self.TryTimes,
		RetryPause:    self.RetryPause,
		RedirectTimes: self.RedirectTimes,
		Temp:          self.Temp,
		TempIsJson:    self.TempIsJson,
		Priority:      self.Priority,
		Reloadable:    self.Reloadable,
		DownloaderID:  self.DownloaderID,
	}
	return json.Marshal(j)
}
