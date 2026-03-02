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
	URL           string          // target URL, required
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
	TempIsJSON    map[string]bool // marks Temp fields stored as JSON; auto-set, do not set manually
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
	SurfID    = 0 // Surf downloader (native Go), do not change
	PhantomID = 1 // PhantomJS downloader (fallback, rarely used)
	ChromeID  = 2 // Chromium headless browser downloader
)

// Prepare sets default values before sending a request.
// Request.URL and Request.Rule must be set.
// Request.Spider is auto-set by the system.
// Request.EnableCookie is set in Spider; per-request values are ignored.
// Optional fields with defaults: Method (GET), DialTimeout, ConnTimeout, TryTimes,
// RedirectTimes, RetryPause, DownloaderID (0=Surf, 1=PhantomJS).
func (r *Request) Prepare() result.VoidResult {
	URL, err := url.Parse(r.URL)
	if err != nil {
		return result.TryErrVoid(err)
	}
	r.URL = URL.String()

	if r.Method == "" {
		r.Method = "GET"
	} else {
		r.Method = strings.ToUpper(r.Method)
	}

	if r.Header == nil {
		r.Header = make(http.Header)
	}

	if r.DialTimeout < 0 {
		r.DialTimeout = 0
	} else if r.DialTimeout == 0 {
		r.DialTimeout = DefaultDialTimeout
	}

	if r.ConnTimeout < 0 {
		r.ConnTimeout = 0
	} else if r.ConnTimeout == 0 {
		r.ConnTimeout = DefaultConnTimeout
	}

	if r.TryTimes == 0 {
		r.TryTimes = DefaultTryTimes
	}

	if r.RetryPause <= 0 {
		r.RetryPause = DefaultRetryPause
	}

	if r.Priority < 0 {
		r.Priority = 0
	}

	if r.DownloaderID < SurfID || r.DownloaderID > ChromeID {
		r.DownloaderID = SurfID
	}

	if r.TempIsJSON == nil {
		r.TempIsJSON = make(map[string]bool)
	}

	if r.Temp == nil {
		r.Temp = make(Temp)
	}
	return result.OkVoid()
}

// UnSerialize deserializes a Request from JSON string.
func UnSerialize(s string) result.Result[*Request] {
	req := new(Request)
	return result.Ret(req, json.Unmarshal([]byte(s), req))
}

// Serialize serializes the Request to JSON string.
func (r *Request) Serialize() result.Result[string] {
	for k, v := range r.Temp {
		r.Temp.set(k, v)
		r.TempIsJSON[k] = true
	}
	b, err := json.Marshal(r)
	if err != nil {
		return result.TryErr[string](err)
	}
	return result.Ok(strings.ReplaceAll(util.Bytes2String(b), `\u0026`, `&`))
}

// Unique returns the unique identifier for the request.
func (r *Request) Unique() string {
	if r.unique == "" {
		block := md5.Sum([]byte(r.Spider + r.Rule + r.URL + r.Method))
		r.unique = hex.EncodeToString(block[:])
	}
	return r.unique
}

// Copy returns a deep copy of the request.
func (r *Request) Copy() result.Result[*Request] {
	reqcopy := new(Request)
	b, err := json.Marshal(r)
	if err != nil {
		return result.TryErr[*Request](err)
	}
	return result.Ret(reqcopy, json.Unmarshal(b, reqcopy))
}

// GetURL returns the request URL.
func (r *Request) GetURL() string {
	return r.URL
}

// GetMethod returns the HTTP method name (e.g. GET, POST).
func (r *Request) GetMethod() string {
	return r.Method
}

// SetMethod sets the HTTP method.
func (r *Request) SetMethod(method string) *Request {
	r.Method = strings.ToUpper(method)
	return r
}

func (r *Request) SetURL(url string) *Request {
	r.URL = url
	return r
}

func (r *Request) GetReferer() string {
	return r.Header.Get("Referer")
}

func (r *Request) SetReferer(referer string) *Request {
	r.Header.Set("Referer", referer)
	return r
}

func (r *Request) GetPostData() string {
	return r.PostData
}

func (r *Request) GetHeader() http.Header {
	return r.Header
}

func (r *Request) SetHeader(key, value string) *Request {
	r.Header.Set(key, value)
	return r
}

func (r *Request) AddHeader(key, value string) *Request {
	r.Header.Add(key, value)
	return r
}

func (r *Request) GetEnableCookie() bool {
	return r.EnableCookie
}

func (r *Request) SetEnableCookie(enableCookie bool) *Request {
	r.EnableCookie = enableCookie
	return r
}

func (r *Request) GetCookies() string {
	return r.Header.Get("Cookie")
}

func (r *Request) SetCookies(cookie string) *Request {
	r.Header.Set("Cookie", cookie)
	return r
}

func (r *Request) GetDialTimeout() time.Duration {
	return r.DialTimeout
}

func (r *Request) GetConnTimeout() time.Duration {
	return r.ConnTimeout
}

func (r *Request) GetTryTimes() int {
	return r.TryTimes
}

func (r *Request) GetRetryPause() time.Duration {
	return r.RetryPause
}

func (r *Request) GetProxy() string {
	return r.proxy
}

func (r *Request) SetProxy(proxy string) *Request {
	r.proxy = proxy
	return r
}

func (r *Request) GetRedirectTimes() int {
	return r.RedirectTimes
}

func (r *Request) GetRuleName() string {
	return r.Rule
}

func (r *Request) SetRuleName(ruleName string) *Request {
	r.Rule = ruleName
	return r
}

func (r *Request) GetSpiderName() string {
	return r.Spider
}

func (r *Request) SetSpiderName(spiderName string) *Request {
	r.Spider = spiderName
	return r
}

func (r *Request) IsReloadable() bool {
	return r.Reloadable
}

func (r *Request) SetReloadable(can bool) *Request {
	r.Reloadable = can
	return r
}

// GetTemp returns temporary cached data. defaultValue must not be nil.
func (r *Request) GetTemp(key string, defaultValue interface{}) interface{} {
	if defaultValue == nil {
		panic("*Request.GetTemp() defaultValue must not be nil, key=" + key)
	}
	r.lock.RLock()
	defer r.lock.RUnlock()

	if r.Temp[key] == nil {
		return defaultValue
	}

	if r.TempIsJSON[key] {
		return r.Temp.get(key, defaultValue)
	}

	return r.Temp[key]
}

// GetTempOpt returns temporary cached data as Option. None when key is missing.
func (r *Request) GetTempOpt(key string) option.Option[interface{}] {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if _, ok := r.Temp[key]; !ok {
		return option.None[interface{}]()
	}
	if r.TempIsJSON[key] {
		var v interface{}
		r.Temp.get(key, &v)
		return option.Some(v)
	}
	return option.Some(r.Temp[key])
}

func (r *Request) GetTemps() Temp {
	return r.Temp
}

func (r *Request) SetTemp(key string, value interface{}) *Request {
	r.lock.Lock()
	r.Temp[key] = value
	delete(r.TempIsJSON, key)
	r.lock.Unlock()
	return r
}

func (r *Request) SetTemps(temp map[string]interface{}) *Request {
	r.lock.Lock()
	r.Temp = temp
	r.TempIsJSON = make(map[string]bool)
	r.lock.Unlock()
	return r
}

func (r *Request) GetPriority() int {
	return r.Priority
}

func (r *Request) SetPriority(priority int) *Request {
	r.Priority = priority
	return r
}

func (r *Request) GetDownloaderID() int {
	return r.DownloaderID
}

func (r *Request) SetDownloaderID(id int) *Request {
	r.DownloaderID = id
	return r
}

func (r *Request) MarshalJSON() ([]byte, error) {
	for k, v := range r.Temp {
		if r.TempIsJSON[k] {
			continue
		}
		r.Temp.set(k, v)
		r.TempIsJSON[k] = true
	}
	// Marshal a struct without the mutex to avoid copying sync.RWMutex
	j := struct {
		Spider        string
		URL           string
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
		TempIsJSON    map[string]bool
		Priority      int
		Reloadable    bool
		DownloaderID  int
	}{
		Spider:        r.Spider,
		URL:           r.URL,
		Rule:          r.Rule,
		Method:        r.Method,
		Header:        r.Header,
		EnableCookie:  r.EnableCookie,
		PostData:      r.PostData,
		DialTimeout:   r.DialTimeout,
		ConnTimeout:   r.ConnTimeout,
		TryTimes:      r.TryTimes,
		RetryPause:    r.RetryPause,
		RedirectTimes: r.RedirectTimes,
		Temp:          r.Temp,
		TempIsJSON:    r.TempIsJSON,
		Priority:      r.Priority,
		Reloadable:    r.Reloadable,
		DownloaderID:  r.DownloaderID,
	}
	return json.Marshal(j)
}
