package spider

import (
	"bytes"
	"io"

	"mime"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/net/html/charset"

	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/app/pipeline/collector/data"
	"github.com/andeya/pholcus/common/goquery"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/logs"
)

// Context carries the state for a single crawl request through its lifecycle.
type Context struct {
	spider   *Spider
	Request  *request.Request
	Response *http.Response    // URL is copied from *request.Request
	text     []byte            // response body as raw bytes
	dom      *goquery.Document // parsed HTML DOM (lazy-initialized)
	items    []data.DataCell   // collected text output results
	files    []data.FileCell   // collected file output results
	err      error
	sync.Mutex
}

var (
	contextPool = &sync.Pool{
		New: func() interface{} {
			return &Context{
				items: []data.DataCell{},
				files: []data.FileCell{},
			}
		},
	}
)

// --- Initialization ---

// GetContext retrieves a Context from the pool and binds it to the given spider and request.
func GetContext(sp *Spider, req *request.Request) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.spider = sp
	ctx.Request = req
	return ctx
}

// PutContext resets a Context and returns it to the pool.
func PutContext(ctx *Context) {
	if ctx.Response != nil {
		if ctx.Response.Body != nil {
			ctx.Response.Body.Close()
		}
		ctx.Response = nil
	}
	ctx.items = ctx.items[:0]
	ctx.files = ctx.files[:0]
	ctx.spider = nil
	ctx.Request = nil
	ctx.text = nil
	ctx.dom = nil
	ctx.err = nil
	contextPool.Put(ctx)
}

// SetResponse binds the HTTP response to this context.
func (self *Context) SetResponse(resp *http.Response) *Context {
	self.Response = resp
	return self
}

// SetError marks a download error on this context.
func (self *Context) SetError(err error) {
	self.err = err
}

// --- Public Set/Exec Methods ---

// AddQueue validates and enqueues a new crawl request.
//
// Required fields: Request.Url, Request.Rule.
// Request.Spider is set automatically; Request.EnableCookie is inherited from Spider.
//
// Fields with defaults (may be omitted):
//   - Method: GET
//   - DialTimeout: request.DefaultDialTimeout (negative = unlimited)
//   - ConnTimeout: request.DefaultConnTimeout (negative = unlimited)
//   - TryTimes: request.DefaultTryTimes (negative = unlimited retries)
//   - RedirectTimes: unlimited by default (negative = disable redirects)
//   - RetryPause: request.DefaultRetryPause
//   - DownloaderID: 0 = Surf (fast, full-featured), 1 = PhantomJS (slow, JS-capable)
//
// Referer is auto-filled from the current response URL if not set.
func (self *Context) AddQueue(req *request.Request) *Context {
	if self.spider.tryStop() != nil {
		return self
	}

	prepareResult := req.
		SetSpiderName(self.spider.GetName()).
		SetEnableCookie(self.spider.GetEnableCookie()).
		Prepare()

	if prepareResult.IsErr() {
		logs.Log.Error(prepareResult.UnwrapErr().Error())
		return self
	}

	if req.GetReferer() == "" && self.Response != nil {
		req.SetReferer(self.GetUrl())
	}

	self.spider.RequestPush(req)
	return self
}

// JsAddQueue adds a request from a dynamic (JavaScript) rule definition.
func (self *Context) JsAddQueue(jreq map[string]interface{}) *Context {
	if self.spider.tryStop() != nil {
		return self
	}

	req := &request.Request{}
	u, ok := jreq["Url"].(string)
	if !ok {
		return self
	}
	req.Url = u
	req.Rule, _ = jreq["Rule"].(string)
	req.Method, _ = jreq["Method"].(string)
	req.Header = http.Header{}
	if header, ok := jreq["Header"].(map[string]interface{}); ok {
		for k, values := range header {
			if vals, ok := values.([]string); ok {
				for _, v := range vals {
					req.Header.Add(k, v)
				}
			}
		}
	}
	req.PostData, _ = jreq["PostData"].(string)
	req.Reloadable, _ = jreq["Reloadable"].(bool)
	if t, ok := jreq["DialTimeout"].(int64); ok {
		req.DialTimeout = time.Duration(t)
	}
	if t, ok := jreq["ConnTimeout"].(int64); ok {
		req.ConnTimeout = time.Duration(t)
	}
	if t, ok := jreq["RetryPause"].(int64); ok {
		req.RetryPause = time.Duration(t)
	}
	if t, ok := jreq["TryTimes"].(int64); ok {
		req.TryTimes = int(t)
	}
	if t, ok := jreq["RedirectTimes"].(int64); ok {
		req.RedirectTimes = int(t)
	}
	if t, ok := jreq["Priority"].(int64); ok {
		req.Priority = int(t)
	}
	if t, ok := jreq["DownloaderID"].(int64); ok {
		req.DownloaderID = int(t)
	}
	if t, ok := jreq["Temp"].(map[string]interface{}); ok {
		req.Temp = t
	}

	prepareResult := req.
		SetSpiderName(self.spider.GetName()).
		SetEnableCookie(self.spider.GetEnableCookie()).
		Prepare()

	if prepareResult.IsErr() {
		logs.Log.Error(prepareResult.UnwrapErr().Error())
		return self
	}

	if req.GetReferer() == "" && self.Response != nil {
		req.SetReferer(self.GetUrl())
	}

	self.spider.RequestPush(req)
	return self
}

// Output collects a text result item.
//
// When item is map[int]interface{}, fields are mapped using the existing ItemFields of ruleName.
// When item is map[string]interface{}, missing ItemFields are auto-added.
// An empty ruleName defaults to the current rule.
func (self *Context) Output(item interface{}, ruleName ...string) {
	_ruleName, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("spider %s: Output() called with non-existent rule name", self.spider.GetName())
		return
	}
	var _item map[string]interface{}
	switch item2 := item.(type) {
	case map[int]interface{}:
		_item = self.CreatItem(item2, _ruleName)
	case request.Temp:
		for k := range item2 {
			self.spider.UpsertItemField(rule, k)
		}
		_item = item2
	case map[string]interface{}:
		for k := range item2 {
			self.spider.UpsertItemField(rule, k)
		}
		_item = item2
	}
	self.Lock()
	if self.spider.NotDefaultField {
		self.items = append(self.items, data.GetDataCell(_ruleName, _item, "", "", ""))
	} else {
		self.items = append(self.items, data.GetDataCell(_ruleName, _item, self.GetUrl(), self.GetReferer(), time.Now().Format("2006-01-02 15:04:05")))
	}
	self.Unlock()
}

// FileOutput collects a file result from the response body.
// nameOrExt optionally specifies a file name or extension; empty keeps the original.
// Errors are logged internally; no return value for JS VM compatibility.
func (self *Context) FileOutput(nameOrExt ...string) {
	if self.Response == nil || self.Response.Body == nil {
		logs.Log.Warning(" *     [FileOutput]: Response or Body is nil for %s", self.GetUrl())
		return
	}
	body, err := io.ReadAll(self.Response.Body)
	self.Response.Body.Close()
	if err != nil {
		logs.Log.Error(" *     [FileOutput]: %v", err)
		return
	}

	_, s := path.Split(self.GetUrl())
	n := strings.Split(s, "?")[0]

	var baseName, ext string

	if len(nameOrExt) > 0 {
		p, n := path.Split(nameOrExt[0])
		ext = path.Ext(n)
		if baseName2 := strings.TrimSuffix(n, ext); baseName2 != "" {
			baseName = p + baseName2
		}
	}
	if baseName == "" {
		baseName = strings.TrimSuffix(n, path.Ext(n))
	}
	if ext == "" {
		ext = path.Ext(n)
	}
	if ext == "" {
		ext = ".html"
	}

	self.Lock()
	self.files = append(self.files, data.GetFileCell(self.GetRuleName(), baseName+ext, body))
	self.Unlock()
}

// CreatItem builds a text result map keyed by field names using the ItemFields of ruleName.
// An empty ruleName defaults to the current rule.
func (self *Context) CreatItem(item map[int]interface{}, ruleName ...string) map[string]interface{} {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("spider %s: CreatItem() called with non-existent rule name", self.spider.GetName())
		return nil
	}

	var item2 = make(map[string]interface{}, len(item))
	for k, v := range item {
		field := self.spider.GetItemField(rule, k)
		item2[field] = v
	}
	return item2
}

// SetTemp stores temporary data in the current request.
func (self *Context) SetTemp(key string, value interface{}) *Context {
	self.Request.SetTemp(key, value)
	return self
}

func (self *Context) SetUrl(url string) *Context {
	self.Request.Url = url
	return self
}

func (self *Context) SetReferer(referer string) *Context {
	self.Request.Header.Set("Referer", referer)
	return self
}

// UpsertItemField adds a result field name to the given rule and returns its index.
// If the field already exists, the existing index is returned.
// An empty ruleName defaults to the current rule.
func (self *Context) UpsertItemField(field string, ruleName ...string) (index int) {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("spider %s: UpsertItemField() called with non-existent rule name", self.spider.GetName())
		return
	}
	return self.spider.UpsertItemField(rule, field)
}

// Aid invokes the AidFunc of the specified rule.
// An empty ruleName defaults to the current rule.
func (self *Context) Aid(aid map[string]interface{}, ruleName ...string) interface{} {
	if self.spider.tryStop() != nil {
		return nil
	}

	_, rule, found := self.getRule(ruleName...)
	if !found {
		if len(ruleName) > 0 {
			logs.Log.Error("spider %s: Aid() called with non-existent rule: %s", self.spider.GetName(), ruleName[0])
		} else {
			logs.Log.Error("spider %s: Aid() called without specifying a rule name", self.spider.GetName())
		}
		return nil
	}
	if rule.AidFunc == nil {
		logs.Log.Error("spider %s: rule %s has no AidFunc defined", self.spider.GetName(), ruleName[0])
		return nil
	}
	return rule.AidFunc(self, aid)
}

// Parse dispatches the response to the ParseFunc of the specified rule.
// An empty ruleName defaults to Root().
func (self *Context) Parse(ruleName ...string) *Context {
	if self.spider.tryStop() != nil {
		return self
	}

	_ruleName, rule, found := self.getRule(ruleName...)
	if self.Response != nil {
		self.Request.SetRuleName(_ruleName)
	}
	if !found {
		self.spider.RuleTree.Root(self)
		return self
	}
	if rule.ParseFunc == nil {
		logs.Log.Error("spider %s: rule %s has no ParseFunc defined", self.spider.GetName(), ruleName[0])
		return self
	}
	rule.ParseFunc(self)
	return self
}

// SetKeyin sets the custom keyword/configuration input.
func (self *Context) SetKeyin(keyin string) *Context {
	self.spider.SetKeyin(keyin)
	return self
}

// SetLimit sets the maximum number of items to crawl.
func (self *Context) SetLimit(max int) *Context {
	self.spider.SetLimit(int64(max))
	return self
}

// SetPausetime sets a custom pause interval (randomized: pause/2 ~ pause*2).
// Overrides the externally configured value. Only overwrites an existing value when runtime[0] is true.
func (self *Context) SetPausetime(pause int64, runtime ...bool) *Context {
	self.spider.SetPausetime(pause, runtime...)
	return self
}

// SetTimer configures a timer identified by id.
// When bell is nil, tol is a sleep duration (countdown timer).
// When bell is non-nil, tol specifies the wake-up point (the tol-th bell occurrence from now).
func (self *Context) SetTimer(id string, tol time.Duration, bell *Bell) bool {
	return self.spider.SetTimer(id, tol, bell)
}

// RunTimer starts the timer and reports whether it can continue to be used.
func (self *Context) RunTimer(id string) bool {
	return self.spider.RunTimer(id)
}

// ResetText replaces the downloaded text content and invalidates the DOM cache.
func (self *Context) ResetText(body string) *Context {
	x := (*[2]uintptr)(unsafe.Pointer(&body))
	h := [3]uintptr{x[0], x[1], x[1]}
	self.text = *(*[]byte)(unsafe.Pointer(&h))
	self.dom = nil
	return self
}

// --- Public Get Methods ---

// GetError returns the download error, or the spider's stop error if stopping.
func (self *Context) GetError() error {
	if err := self.spider.tryStop(); err != nil {
		return err
	}
	return self.err
}

// Log returns the global logger instance.
func (*Context) Log() logs.Logs {
	return logs.Log
}

// GetSpider returns the spider bound to this context.
func (self *Context) GetSpider() *Spider {
	return self.spider
}

// GetResponse returns the HTTP response.
func (self *Context) GetResponse() *http.Response {
	return self.Response
}

// GetStatusCode returns the HTTP response status code, or 0 if no response.
func (self *Context) GetStatusCode() int {
	if self.Response == nil {
		return 0
	}
	return self.Response.StatusCode
}

// GetRequest returns the original request.
func (self *Context) GetRequest() *request.Request {
	return self.Request
}

// CopyRequest returns a deep copy of the original request.
func (self *Context) CopyRequest() *request.Request {
	return self.Request.Copy().Unwrap()
}

// GetItemFields returns the result field name list for the given rule.
func (self *Context) GetItemFields(ruleName ...string) []string {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("spider %s: GetItemFields() called with non-existent rule name", self.spider.GetName())
		return nil
	}
	return self.spider.GetItemFields(rule)
}

// GetItemField returns the field name at the given index, or "" if not found.
// An empty ruleName defaults to the current rule.
func (self *Context) GetItemField(index int, ruleName ...string) (field string) {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("spider %s: GetItemField() called with non-existent rule name", self.spider.GetName())
		return
	}
	return self.spider.GetItemField(rule, index)
}

// GetItemFieldIndex returns the index of the given field name, or -1 if not found.
// An empty ruleName defaults to the current rule.
func (self *Context) GetItemFieldIndex(field string, ruleName ...string) (index int) {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("spider %s: GetItemFieldIndex() called with non-existent rule name", self.spider.GetName())
		return
	}
	return self.spider.GetItemFieldIndex(rule, field)
}

// PullItems drains and returns all collected data items, resetting the internal buffer.
func (self *Context) PullItems() (ds []data.DataCell) {
	self.Lock()
	ds = self.items
	self.items = []data.DataCell{}
	self.Unlock()
	return
}

// PullFiles drains and returns all collected file results, resetting the internal buffer.
func (self *Context) PullFiles() (fs []data.FileCell) {
	self.Lock()
	fs = self.files
	self.files = []data.FileCell{}
	self.Unlock()
	return
}

// GetKeyin returns the custom keyword/configuration input.
func (self *Context) GetKeyin() string {
	return self.spider.GetKeyin()
}

// GetLimit returns the maximum number of items to crawl.
func (self *Context) GetLimit() int {
	return int(self.spider.GetLimit())
}

// GetName returns the spider name.
func (self *Context) GetName() string {
	return self.spider.GetName()
}

// GetRules returns the full rule map.
func (self *Context) GetRules() map[string]*Rule {
	return self.spider.GetRules()
}

// GetRule returns the rule with the given name.
func (self *Context) GetRule(ruleName string) *Rule {
	return self.spider.GetRule(ruleName)
}

// GetRuleName returns the current rule name from the request.
func (self *Context) GetRuleName() string {
	return self.Request.GetRuleName()
}

// GetTemp retrieves temporary data from the request by key.
// defaultValue must not be a nil interface{}.
func (self *Context) GetTemp(key string, defaultValue interface{}) interface{} {
	return self.Request.GetTemp(key, defaultValue)
}

// GetTemps returns all temporary data from the request.
func (self *Context) GetTemps() request.Temp {
	return self.Request.GetTemps()
}

// CopyTemps returns a shallow copy of the request's temporary data.
func (self *Context) CopyTemps() request.Temp {
	temps := make(request.Temp)
	for k, v := range self.Request.GetTemps() {
		temps[k] = v
	}
	return temps
}

// GetUrl returns the URL from the original request, preserving the unencoded form.
func (self *Context) GetUrl() string {
	return self.Request.Url
}

// GetMethod returns the HTTP method of the request.
func (self *Context) GetMethod() string {
	return self.Request.GetMethod()
}

// GetHost returns the host from the response URL, or "" if unavailable.
func (self *Context) GetHost() string {
	if self.Response == nil || self.Response.Request == nil || self.Response.Request.URL == nil {
		return ""
	}
	return self.Response.Request.URL.Host
}

// GetHeader returns the response headers.
func (self *Context) GetHeader() http.Header {
	if self.Response == nil {
		return http.Header{}
	}
	return self.Response.Header
}

// GetRequestHeader returns the request headers from the actual HTTP request made.
func (self *Context) GetRequestHeader() http.Header {
	if self.Response == nil || self.Response.Request == nil {
		return http.Header{}
	}
	return self.Response.Request.Header
}

// GetReferer returns the Referer header from the actual HTTP request made.
func (self *Context) GetReferer() string {
	if self.Response == nil || self.Response.Request == nil {
		return ""
	}
	return self.Response.Request.Header.Get("Referer")
}

// GetCookie returns the Set-Cookie header from the response.
func (self *Context) GetCookie() string {
	if self.Response == nil {
		return ""
	}
	return self.Response.Header.Get("Set-Cookie")
}

// GetDom returns the parsed HTML DOM, initializing it lazily from the response body.
func (self *Context) GetDom() *goquery.Document {
	if self.dom == nil {
		if self.Response == nil {
			logs.Log.Warning(" *     [GetDom]: Response is nil for %s", self.GetUrl())
			return nil
		}
		self.initDom()
	}
	return self.dom
}

// GetText returns the response body as a UTF-8 string, initializing it lazily.
func (self *Context) GetText() string {
	if self.text == nil {
		if self.Response == nil {
			logs.Log.Warning(" *     [GetText]: Response is nil for %s", self.GetUrl())
			return ""
		}
		self.initText()
	}
	return util.Bytes2String(self.text)
}

// --- Private Methods ---

// getRule resolves a rule by name, defaulting to the current request's rule.
func (self *Context) getRule(ruleName ...string) (name string, rule *Rule, found bool) {
	if len(ruleName) == 0 {
		if self.Response == nil {
			return
		}
		name = self.GetRuleName()
	} else {
		name = ruleName[0]
	}
	rule = self.spider.GetRule(name)
	return name, rule, rule != nil
}

// initDom parses the text body into a goquery Document.
func (self *Context) initDom() *goquery.Document {
	if self.text == nil {
		self.initText()
	}
	r := goquery.NewDocumentFromReader(bytes.NewReader(self.text))
	if r.IsErr() {
		panic(r.UnwrapErr().Error())
	}
	self.dom = r.Unwrap()
	return self.dom
}

// initText reads the response body and converts it to UTF-8 if needed.
func (self *Context) initText() {
	body, err := io.ReadAll(self.Response.Body)
	self.Response.Body.Close()
	if err != nil {
		panic(err.Error())
	}

	responseCT := self.Response.Header.Get("Content-Type")
	requestCT := self.Request.Header.Get("Content-Type")
	pageEncode := detectCharset(responseCT, requestCT)

	if self.Request.DownloaderID == request.SURF_ID && !isUTF8(pageEncode) {
		converted, convErr := convertEncoding(body, pageEncode)
		if convErr == nil {
			self.text = converted
			return
		}
		logs.Log.Warning(" *     [convert][%v]: %v (ignore transcoding)\n", self.GetUrl(), convErr)
	}

	self.text = body
}

// detectCharset extracts charset from Content-Type headers (response first, then request).
func detectCharset(responseContentType, requestContentType string) string {
	for _, ct := range []string{responseContentType, requestContentType} {
		if _, params, err := mime.ParseMediaType(ct); err == nil {
			if cs, ok := params["charset"]; ok {
				return strings.ToLower(strings.TrimSpace(cs))
			}
		}
	}
	return ""
}

func isUTF8(charset string) bool {
	switch charset {
	case "utf8", "utf-8", "unicode-1-1-utf-8":
		return true
	}
	return false
}

// convertEncoding converts body from the given charset to UTF-8.
func convertEncoding(body []byte, charsetLabel string) ([]byte, error) {
	var destReader io.Reader
	var err error
	r := bytes.NewReader(body)
	if charsetLabel == "" {
		destReader, err = charset.NewReader(r, "")
	} else {
		destReader, err = charset.NewReaderLabel(charsetLabel, r)
	}
	if err != nil {
		return nil, err
	}
	return io.ReadAll(destReader)
}

/*
 * Charset reference (case-insensitive).
 * var nameMap = map[string]htmlEncoding{
	"unicode-1-1-utf-8":   utf8,
	"utf-8":               utf8,
	"utf8":                utf8,
	"866":                 ibm866,
	"cp866":               ibm866,
	"csibm866":            ibm866,
	"ibm866":              ibm866,
	"csisolatin2":         iso8859_2,
	"iso-8859-2":          iso8859_2,
	"iso-ir-101":          iso8859_2,
	"iso8859-2":           iso8859_2,
	"iso88592":            iso8859_2,
	"iso_8859-2":          iso8859_2,
	"iso_8859-2:1987":     iso8859_2,
	"l2":                  iso8859_2,
	"latin2":              iso8859_2,
	"csisolatin3":         iso8859_3,
	"iso-8859-3":          iso8859_3,
	"iso-ir-109":          iso8859_3,
	"iso8859-3":           iso8859_3,
	"iso88593":            iso8859_3,
	"iso_8859-3":          iso8859_3,
	"iso_8859-3:1988":     iso8859_3,
	"l3":                  iso8859_3,
	"latin3":              iso8859_3,
	"csisolatin4":         iso8859_4,
	"iso-8859-4":          iso8859_4,
	"iso-ir-110":          iso8859_4,
	"iso8859-4":           iso8859_4,
	"iso88594":            iso8859_4,
	"iso_8859-4":          iso8859_4,
	"iso_8859-4:1988":     iso8859_4,
	"l4":                  iso8859_4,
	"latin4":              iso8859_4,
	"csisolatincyrillic":  iso8859_5,
	"cyrillic":            iso8859_5,
	"iso-8859-5":          iso8859_5,
	"iso-ir-144":          iso8859_5,
	"iso8859-5":           iso8859_5,
	"iso88595":            iso8859_5,
	"iso_8859-5":          iso8859_5,
	"iso_8859-5:1988":     iso8859_5,
	"arabic":              iso8859_6,
	"asmo-708":            iso8859_6,
	"csiso88596e":         iso8859_6,
	"csiso88596i":         iso8859_6,
	"csisolatinarabic":    iso8859_6,
	"ecma-114":            iso8859_6,
	"iso-8859-6":          iso8859_6,
	"iso-8859-6-e":        iso8859_6,
	"iso-8859-6-i":        iso8859_6,
	"iso-ir-127":          iso8859_6,
	"iso8859-6":           iso8859_6,
	"iso88596":            iso8859_6,
	"iso_8859-6":          iso8859_6,
	"iso_8859-6:1987":     iso8859_6,
	"csisolatingreek":     iso8859_7,
	"ecma-118":            iso8859_7,
	"elot_928":            iso8859_7,
	"greek":               iso8859_7,
	"greek8":              iso8859_7,
	"iso-8859-7":          iso8859_7,
	"iso-ir-126":          iso8859_7,
	"iso8859-7":           iso8859_7,
	"iso88597":            iso8859_7,
	"iso_8859-7":          iso8859_7,
	"iso_8859-7:1987":     iso8859_7,
	"sun_eu_greek":        iso8859_7,
	"csiso88598e":         iso8859_8,
	"csisolatinhebrew":    iso8859_8,
	"hebrew":              iso8859_8,
	"iso-8859-8":          iso8859_8,
	"iso-8859-8-e":        iso8859_8,
	"iso-ir-138":          iso8859_8,
	"iso8859-8":           iso8859_8,
	"iso88598":            iso8859_8,
	"iso_8859-8":          iso8859_8,
	"iso_8859-8:1988":     iso8859_8,
	"visual":              iso8859_8,
	"csiso88598i":         iso8859_8I,
	"iso-8859-8-i":        iso8859_8I,
	"logical":             iso8859_8I,
	"csisolatin6":         iso8859_10,
	"iso-8859-10":         iso8859_10,
	"iso-ir-157":          iso8859_10,
	"iso8859-10":          iso8859_10,
	"iso885910":           iso8859_10,
	"l6":                  iso8859_10,
	"latin6":              iso8859_10,
	"iso-8859-13":         iso8859_13,
	"iso8859-13":          iso8859_13,
	"iso885913":           iso8859_13,
	"iso-8859-14":         iso8859_14,
	"iso8859-14":          iso8859_14,
	"iso885914":           iso8859_14,
	"csisolatin9":         iso8859_15,
	"iso-8859-15":         iso8859_15,
	"iso8859-15":          iso8859_15,
	"iso885915":           iso8859_15,
	"iso_8859-15":         iso8859_15,
	"l9":                  iso8859_15,
	"iso-8859-16":         iso8859_16,
	"cskoi8r":             koi8r,
	"koi":                 koi8r,
	"koi8":                koi8r,
	"koi8-r":              koi8r,
	"koi8_r":              koi8r,
	"koi8-ru":             koi8u,
	"koi8-u":              koi8u,
	"csmacintosh":         macintosh,
	"mac":                 macintosh,
	"macintosh":           macintosh,
	"x-mac-roman":         macintosh,
	"dos-874":             windows874,
	"iso-8859-11":         windows874,
	"iso8859-11":          windows874,
	"iso885911":           windows874,
	"tis-620":             windows874,
	"windows-874":         windows874,
	"cp1250":              windows1250,
	"windows-1250":        windows1250,
	"x-cp1250":            windows1250,
	"cp1251":              windows1251,
	"windows-1251":        windows1251,
	"x-cp1251":            windows1251,
	"ansi_x3.4-1968":      windows1252,
	"ascii":               windows1252,
	"cp1252":              windows1252,
	"cp819":               windows1252,
	"csisolatin1":         windows1252,
	"ibm819":              windows1252,
	"iso-8859-1":          windows1252,
	"iso-ir-100":          windows1252,
	"iso8859-1":           windows1252,
	"iso88591":            windows1252,
	"iso_8859-1":          windows1252,
	"iso_8859-1:1987":     windows1252,
	"l1":                  windows1252,
	"latin1":              windows1252,
	"us-ascii":            windows1252,
	"windows-1252":        windows1252,
	"x-cp1252":            windows1252,
	"cp1253":              windows1253,
	"windows-1253":        windows1253,
	"x-cp1253":            windows1253,
	"cp1254":              windows1254,
	"csisolatin5":         windows1254,
	"iso-8859-9":          windows1254,
	"iso-ir-148":          windows1254,
	"iso8859-9":           windows1254,
	"iso88599":            windows1254,
	"iso_8859-9":          windows1254,
	"iso_8859-9:1989":     windows1254,
	"l5":                  windows1254,
	"latin5":              windows1254,
	"windows-1254":        windows1254,
	"x-cp1254":            windows1254,
	"cp1255":              windows1255,
	"windows-1255":        windows1255,
	"x-cp1255":            windows1255,
	"cp1256":              windows1256,
	"windows-1256":        windows1256,
	"x-cp1256":            windows1256,
	"cp1257":              windows1257,
	"windows-1257":        windows1257,
	"x-cp1257":            windows1257,
	"cp1258":              windows1258,
	"windows-1258":        windows1258,
	"x-cp1258":            windows1258,
	"x-mac-cyrillic":      macintoshCyrillic,
	"x-mac-ukrainian":     macintoshCyrillic,
	"chinese":             gbk,
	"csgb2312":            gbk,
	"csiso58gb231280":     gbk,
	"gb2312":              gbk,
	"gb_2312":             gbk,
	"gb_2312-80":          gbk,
	"gbk":                 gbk,
	"iso-ir-58":           gbk,
	"x-gbk":               gbk,
	"gb18030":             gb18030,
	"big5":                big5,
	"big5-hkscs":          big5,
	"cn-big5":             big5,
	"csbig5":              big5,
	"x-x-big5":            big5,
	"cseucpkdfmtjapanese": eucjp,
	"euc-jp":              eucjp,
	"x-euc-jp":            eucjp,
	"csiso2022jp":         iso2022jp,
	"iso-2022-jp":         iso2022jp,
	"csshiftjis":          shiftJIS,
	"ms932":               shiftJIS,
	"ms_kanji":            shiftJIS,
	"shift-jis":           shiftJIS,
	"shift_jis":           shiftJIS,
	"sjis":                shiftJIS,
	"windows-31j":         shiftJIS,
	"x-sjis":              shiftJIS,
	"cseuckr":             euckr,
	"csksc56011987":       euckr,
	"euc-kr":              euckr,
	"iso-ir-149":          euckr,
	"korean":              euckr,
	"ks_c_5601-1987":      euckr,
	"ks_c_5601-1989":      euckr,
	"ksc5601":             euckr,
	"ksc_5601":            euckr,
	"windows-949":         euckr,
	"csiso2022kr":         replacement,
	"hz-gb-2312":          replacement,
	"iso-2022-cn":         replacement,
	"iso-2022-cn-ext":     replacement,
	"iso-2022-kr":         replacement,
	"utf-16be":            utf16be,
	"utf-16":              utf16le,
	"utf-16le":            utf16le,
	"x-user-defined":      xUserDefined,
}*/
