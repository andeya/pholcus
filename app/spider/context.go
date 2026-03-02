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
func (ctx *Context) SetResponse(resp *http.Response) *Context {
	ctx.Response = resp
	return ctx
}

// SetError marks a download error on this context.
func (ctx *Context) SetError(err error) {
	ctx.err = err
}

// --- Public Set/Exec Methods ---

// AddQueue validates and enqueues a new crawl request.
//
// Required fields: Request.URL, Request.Rule.
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
func (ctx *Context) AddQueue(req *request.Request) *Context {
	if ctx.spider.tryStop() != nil {
		return ctx
	}

	prepareResult := req.
		SetSpiderName(ctx.spider.GetName()).
		SetEnableCookie(ctx.spider.GetEnableCookie()).
		Prepare()

	if prepareResult.IsErr() {
		logs.Log().Error(prepareResult.UnwrapErr().Error())
		return ctx
	}

	if req.GetReferer() == "" && ctx.Response != nil {
		req.SetReferer(ctx.GetURL())
	}

	ctx.spider.RequestPush(req)
	return ctx
}

// JsAddQueue adds a request from a dynamic (JavaScript) rule definition.
// jsToInt64 extracts an integer from a JS value that may arrive as
// int64 (otto integers) or float64 (otto floats / some JS engines).
func jsToInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case float64:
		return int64(n), true
	case int:
		return int64(n), true
	}
	return 0, false
}

func (ctx *Context) JsAddQueue(jreq map[string]interface{}) *Context {
	if ctx.spider.tryStop() != nil {
		return ctx
	}

	req := &request.Request{}
	u, ok := jreq["URL"].(string)
	if !ok {
		return ctx
	}
	req.URL = u
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
	if t, ok := jsToInt64(jreq["DialTimeout"]); ok {
		req.DialTimeout = time.Duration(t)
	}
	if t, ok := jsToInt64(jreq["ConnTimeout"]); ok {
		req.ConnTimeout = time.Duration(t)
	}
	if t, ok := jsToInt64(jreq["RetryPause"]); ok {
		req.RetryPause = time.Duration(t)
	}
	if t, ok := jsToInt64(jreq["TryTimes"]); ok {
		req.TryTimes = int(t)
	}
	if t, ok := jsToInt64(jreq["RedirectTimes"]); ok {
		req.RedirectTimes = int(t)
	}
	if t, ok := jsToInt64(jreq["Priority"]); ok {
		req.Priority = int(t)
	}
	if t, ok := jsToInt64(jreq["DownloaderID"]); ok {
		req.DownloaderID = int(t)
	}
	if t, ok := jreq["Temp"].(map[string]interface{}); ok {
		req.Temp = t
	}

	prepareResult := req.
		SetSpiderName(ctx.spider.GetName()).
		SetEnableCookie(ctx.spider.GetEnableCookie()).
		Prepare()

	if prepareResult.IsErr() {
		logs.Log().Error(prepareResult.UnwrapErr().Error())
		return ctx
	}

	if req.GetReferer() == "" && ctx.Response != nil {
		req.SetReferer(ctx.GetURL())
	}

	ctx.spider.RequestPush(req)
	return ctx
}

// Output collects a text result item.
//
// When item is map[int]interface{}, fields are mapped using the existing ItemFields of ruleName.
// When item is map[string]interface{}, missing ItemFields are auto-added.
// An empty ruleName defaults to the current rule.
func (ctx *Context) Output(item interface{}, ruleName ...string) {
	_ruleName, rule, found := ctx.getRule(ruleName...)
	if !found {
		logs.Log().Error("spider %s: Output() called with non-existent rule name", ctx.spider.GetName())
		return
	}
	var _item map[string]interface{}
	switch item2 := item.(type) {
	case map[int]interface{}:
		_item = ctx.CreateItem(item2, _ruleName)
	case request.Temp:
		for k := range item2 {
			ctx.spider.UpsertItemField(rule, k)
		}
		_item = item2
	case map[string]interface{}:
		for k := range item2 {
			ctx.spider.UpsertItemField(rule, k)
		}
		_item = item2
	}
	ctx.Lock()
	if ctx.spider.NotDefaultField {
		ctx.items = append(ctx.items, data.GetDataCell(_ruleName, _item, "", "", ""))
	} else {
		ctx.items = append(ctx.items, data.GetDataCell(_ruleName, _item, ctx.GetURL(), ctx.GetReferer(), time.Now().Format("2006-01-02 15:04:05")))
	}
	ctx.Unlock()
}

// FileOutput collects a file result from the response body.
// nameOrExt optionally specifies a file name or extension; empty keeps the original.
// Errors are logged internally; no return value for JS VM compatibility.
func (ctx *Context) FileOutput(nameOrExt ...string) {
	if ctx.Response == nil || ctx.Response.Body == nil {
		logs.Log().Warning(" *     [FileOutput]: Response or Body is nil for %s", ctx.GetURL())
		return
	}
	body, err := io.ReadAll(ctx.Response.Body)
	ctx.Response.Body.Close()
	if err != nil {
		logs.Log().Error(" *     [FileOutput]: %v", err)
		return
	}

	_, s := path.Split(ctx.GetURL())
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

	ctx.Lock()
	ctx.files = append(ctx.files, data.GetFileCell(ctx.GetRuleName(), baseName+ext, body))
	ctx.Unlock()
}

// CreateItem builds a text result map keyed by field names using the ItemFields of ruleName.
// An empty ruleName defaults to the current rule.
func (ctx *Context) CreateItem(item map[int]interface{}, ruleName ...string) map[string]interface{} {
	_, rule, found := ctx.getRule(ruleName...)
	if !found {
		logs.Log().Error("spider %s: CreateItem() called with non-existent rule name", ctx.spider.GetName())
		return nil
	}

	var item2 = make(map[string]interface{}, len(item))
	for k, v := range item {
		field := ctx.spider.GetItemField(rule, k)
		item2[field] = v
	}
	return item2
}

// SetTemp stores temporary data in the current request.
func (ctx *Context) SetTemp(key string, value interface{}) *Context {
	ctx.Request.SetTemp(key, value)
	return ctx
}

func (ctx *Context) SetURL(url string) *Context {
	ctx.Request.URL = url
	return ctx
}

func (ctx *Context) SetReferer(referer string) *Context {
	ctx.Request.Header.Set("Referer", referer)
	return ctx
}

// UpsertItemField adds a result field name to the given rule and returns its index.
// If the field already exists, the existing index is returned.
// An empty ruleName defaults to the current rule.
func (ctx *Context) UpsertItemField(field string, ruleName ...string) (index int) {
	_, rule, found := ctx.getRule(ruleName...)
	if !found {
		logs.Log().Error("spider %s: UpsertItemField() called with non-existent rule name", ctx.spider.GetName())
		return
	}
	return ctx.spider.UpsertItemField(rule, field)
}

// Aid invokes the AidFunc of the specified rule.
// An empty ruleName defaults to the current rule.
func (ctx *Context) Aid(aid map[string]interface{}, ruleName ...string) interface{} {
	if ctx.spider.tryStop() != nil {
		return nil
	}

	_, rule, found := ctx.getRule(ruleName...)
	if !found {
		if len(ruleName) > 0 {
			logs.Log().Error("spider %s: Aid() called with non-existent rule: %s", ctx.spider.GetName(), ruleName[0])
		} else {
			logs.Log().Error("spider %s: Aid() called without specifying a rule name", ctx.spider.GetName())
		}
		return nil
	}
	if rule.AidFunc == nil {
		logs.Log().Error("spider %s: rule %s has no AidFunc defined", ctx.spider.GetName(), ruleName[0])
		return nil
	}
	return rule.AidFunc(ctx, aid)
}

// Parse dispatches the response to the ParseFunc of the specified rule.
// An empty ruleName defaults to Root().
func (ctx *Context) Parse(ruleName ...string) *Context {
	if ctx.spider.tryStop() != nil {
		return ctx
	}

	_ruleName, rule, found := ctx.getRule(ruleName...)
	if ctx.Response != nil {
		ctx.Request.SetRuleName(_ruleName)
	}
	if !found {
		ctx.spider.RuleTree.Root(ctx)
		return ctx
	}
	if rule.ParseFunc == nil {
		logs.Log().Error("spider %s: rule %s has no ParseFunc defined", ctx.spider.GetName(), ruleName[0])
		return ctx
	}
	rule.ParseFunc(ctx)
	return ctx
}

// SetKeyin sets the custom keyword/configuration input.
func (ctx *Context) SetKeyin(keyin string) *Context {
	ctx.spider.SetKeyin(keyin)
	return ctx
}

// SetLimit sets the maximum number of items to crawl.
func (ctx *Context) SetLimit(max int) *Context {
	ctx.spider.SetLimit(int64(max))
	return ctx
}

// SetPausetime sets a custom pause interval (randomized: pause/2 ~ pause*2).
// Overrides the externally configured value. Only overwrites an existing value when runtime[0] is true.
func (ctx *Context) SetPausetime(pause int64, runtime ...bool) *Context {
	ctx.spider.SetPausetime(pause, runtime...)
	return ctx
}

// SetTimer configures a timer identified by id.
// When bell is nil, tol is a sleep duration (countdown timer).
// When bell is non-nil, tol specifies the wake-up point (the tol-th bell occurrence from now).
func (ctx *Context) SetTimer(id string, tol time.Duration, bell *Bell) bool {
	return ctx.spider.SetTimer(id, tol, bell)
}

// RunTimer starts the timer and reports whether it can continue to be used.
func (ctx *Context) RunTimer(id string) bool {
	return ctx.spider.RunTimer(id)
}

// ResetText replaces the downloaded text content and invalidates the DOM cache.
func (ctx *Context) ResetText(body string) *Context {
	x := (*[2]uintptr)(unsafe.Pointer(&body))
	h := [3]uintptr{x[0], x[1], x[1]}
	ctx.text = *(*[]byte)(unsafe.Pointer(&h))
	ctx.dom = nil
	return ctx
}

// --- Public Get Methods ---

// GetError returns the download error, or the spider's stop error if stopping.
func (ctx *Context) GetError() error {
	if err := ctx.spider.tryStop(); err != nil {
		return err
	}
	return ctx.err
}

// Log returns the global logger instance.
func (*Context) Log() logs.Logs {
	return logs.Log()
}

// GetSpider returns the spider bound to this context.
func (ctx *Context) GetSpider() *Spider {
	return ctx.spider
}

// GetResponse returns the HTTP response.
func (ctx *Context) GetResponse() *http.Response {
	return ctx.Response
}

// GetStatusCode returns the HTTP response status code, or 0 if no response.
func (ctx *Context) GetStatusCode() int {
	if ctx.Response == nil {
		return 0
	}
	return ctx.Response.StatusCode
}

// GetRequest returns the original request.
func (ctx *Context) GetRequest() *request.Request {
	return ctx.Request
}

// CopyRequest returns a deep copy of the original request.
func (ctx *Context) CopyRequest() *request.Request {
	return ctx.Request.Copy().Unwrap()
}

// GetItemFields returns the result field name list for the given rule.
func (ctx *Context) GetItemFields(ruleName ...string) []string {
	_, rule, found := ctx.getRule(ruleName...)
	if !found {
		logs.Log().Error("spider %s: GetItemFields() called with non-existent rule name", ctx.spider.GetName())
		return nil
	}
	return ctx.spider.GetItemFields(rule)
}

// GetItemField returns the field name at the given index, or "" if not found.
// An empty ruleName defaults to the current rule.
func (ctx *Context) GetItemField(index int, ruleName ...string) (field string) {
	_, rule, found := ctx.getRule(ruleName...)
	if !found {
		logs.Log().Error("spider %s: GetItemField() called with non-existent rule name", ctx.spider.GetName())
		return
	}
	return ctx.spider.GetItemField(rule, index)
}

// GetItemFieldIndex returns the index of the given field name, or -1 if not found.
// An empty ruleName defaults to the current rule.
func (ctx *Context) GetItemFieldIndex(field string, ruleName ...string) (index int) {
	_, rule, found := ctx.getRule(ruleName...)
	if !found {
		logs.Log().Error("spider %s: GetItemFieldIndex() called with non-existent rule name", ctx.spider.GetName())
		return
	}
	return ctx.spider.GetItemFieldIndex(rule, field)
}

// PullItems drains and returns all collected data items, resetting the internal buffer.
func (ctx *Context) PullItems() (ds []data.DataCell) {
	ctx.Lock()
	ds = ctx.items
	ctx.items = []data.DataCell{}
	ctx.Unlock()
	return
}

// PullFiles drains and returns all collected file results, resetting the internal buffer.
func (ctx *Context) PullFiles() (fs []data.FileCell) {
	ctx.Lock()
	fs = ctx.files
	ctx.files = []data.FileCell{}
	ctx.Unlock()
	return
}

// GetKeyin returns the custom keyword/configuration input.
func (ctx *Context) GetKeyin() string {
	return ctx.spider.GetKeyin()
}

// GetLimit returns the maximum number of items to crawl.
func (ctx *Context) GetLimit() int {
	return int(ctx.spider.GetLimit())
}

// GetName returns the spider name.
func (ctx *Context) GetName() string {
	return ctx.spider.GetName()
}

// GetRules returns the full rule map.
func (ctx *Context) GetRules() map[string]*Rule {
	return ctx.spider.GetRules()
}

// GetRule returns the rule with the given name.
func (ctx *Context) GetRule(ruleName string) *Rule {
	return ctx.spider.GetRule(ruleName)
}

// GetRuleName returns the current rule name from the request.
func (ctx *Context) GetRuleName() string {
	return ctx.Request.GetRuleName()
}

// GetTemp retrieves temporary data from the request by key.
// defaultValue must not be a nil interface{}.
func (ctx *Context) GetTemp(key string, defaultValue interface{}) interface{} {
	return ctx.Request.GetTemp(key, defaultValue)
}

// GetTemps returns all temporary data from the request.
func (ctx *Context) GetTemps() request.Temp {
	return ctx.Request.GetTemps()
}

// CopyTemps returns a shallow copy of the request's temporary data.
func (ctx *Context) CopyTemps() request.Temp {
	temps := make(request.Temp)
	for k, v := range ctx.Request.GetTemps() {
		temps[k] = v
	}
	return temps
}

// GetURL returns the URL from the original request, preserving the unencoded form.
func (ctx *Context) GetURL() string {
	return ctx.Request.URL
}

// GetMethod returns the HTTP method of the request.
func (ctx *Context) GetMethod() string {
	return ctx.Request.GetMethod()
}

// GetHost returns the host from the response URL, or "" if unavailable.
func (ctx *Context) GetHost() string {
	if ctx.Response == nil || ctx.Response.Request == nil || ctx.Response.Request.URL == nil {
		return ""
	}
	return ctx.Response.Request.URL.Host
}

// GetHeader returns the response headers.
func (ctx *Context) GetHeader() http.Header {
	if ctx.Response == nil {
		return http.Header{}
	}
	return ctx.Response.Header
}

// GetRequestHeader returns the request headers from the actual HTTP request made.
func (ctx *Context) GetRequestHeader() http.Header {
	if ctx.Response == nil || ctx.Response.Request == nil {
		return http.Header{}
	}
	return ctx.Response.Request.Header
}

// GetReferer returns the Referer header from the actual HTTP request made.
func (ctx *Context) GetReferer() string {
	if ctx.Response == nil || ctx.Response.Request == nil {
		return ""
	}
	return ctx.Response.Request.Header.Get("Referer")
}

// GetCookie returns the Set-Cookie header from the response.
func (ctx *Context) GetCookie() string {
	if ctx.Response == nil {
		return ""
	}
	return ctx.Response.Header.Get("Set-Cookie")
}

// GetDom returns the parsed HTML DOM, initializing it lazily from the response body.
// Errors are stored in ctx.err and can be retrieved via GetError().
func (ctx *Context) GetDom() *goquery.Document {
	if ctx.dom == nil {
		if ctx.Response == nil {
			logs.Log().Warning(" *     [GetDom]: Response is nil for %s", ctx.GetURL())
			return nil
		}
		dom, err := ctx.initDom()
		if err != nil {
			ctx.err = err
			logs.Log().Error(" *     [GetDom][%s]: %v", ctx.GetURL(), err)
			return nil
		}
		return dom
	}
	return ctx.dom
}

// GetText returns the response body as a UTF-8 string, initializing it lazily.
// Errors are stored in ctx.err and can be retrieved via GetError().
func (ctx *Context) GetText() string {
	if ctx.text == nil {
		if ctx.Response == nil {
			logs.Log().Warning(" *     [GetText]: Response is nil for %s", ctx.GetURL())
			return ""
		}
		if err := ctx.initText(); err != nil {
			ctx.err = err
			logs.Log().Error(" *     [GetText][%s]: %v", ctx.GetURL(), err)
			return ""
		}
	}
	return util.Bytes2String(ctx.text)
}

// --- Private Methods ---

// getRule resolves a rule by name, defaulting to the current request's rule.
func (ctx *Context) getRule(ruleName ...string) (name string, rule *Rule, found bool) {
	if len(ruleName) == 0 {
		if ctx.Response == nil {
			return
		}
		name = ctx.GetRuleName()
	} else {
		name = ruleName[0]
	}
	rule = ctx.spider.GetRule(name)
	return name, rule, rule != nil
}

// initDom parses the text body into a goquery Document.
func (ctx *Context) initDom() (*goquery.Document, error) {
	if ctx.text == nil {
		if err := ctx.initText(); err != nil {
			return nil, err
		}
	}
	r := goquery.NewDocumentFromReader(bytes.NewReader(ctx.text))
	if r.IsErr() {
		return nil, r.UnwrapErr()
	}
	ctx.dom = r.Unwrap()
	return ctx.dom, nil
}

// initText reads the response body and converts it to UTF-8 if needed.
func (ctx *Context) initText() error {
	body, err := io.ReadAll(ctx.Response.Body)
	ctx.Response.Body.Close()
	if err != nil {
		return err
	}

	responseCT := ctx.Response.Header.Get("Content-Type")
	requestCT := ctx.Request.Header.Get("Content-Type")
	pageEncode := detectCharset(responseCT, requestCT)

	if ctx.Request.DownloaderID == request.SurfID && !isUTF8(pageEncode) {
		converted, convErr := convertEncoding(body, pageEncode)
		if convErr == nil {
			ctx.text = converted
			return nil
		}
		logs.Log().Warning(" *     [convert][%v]: %v (ignore transcoding)\n", ctx.GetURL(), convErr)
	}

	ctx.text = body
	return nil
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
