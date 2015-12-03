package context

import (
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/henrylee2cn/pholcus/logs"
	"golang.org/x/net/html/charset"
)

// Response represents an entity be crawled.
type Response struct {
	// 原始请求
	*Request

	// 拷贝自*Request的规则名
	rule string

	// 拷贝自*Request的临时数据，通过temp[x]==nil判断是否有值存入，所以请存入带类型的值，如[]int(nil)等
	temp map[string]interface{}

	// 响应流，其中URL拷贝自*Request
	*http.Response

	// The text is body of response
	text string

	// The dom is a pointer of goquery boject that contains html result.
	dom *goquery.Document

	// The items is the container of parsed result.
	items []map[string]interface{}

	// The files is the container of image.
	// "Name": string; "Body": io.ReadCloser
	files []map[string]interface{}

	// The err is not nil when crawl process is success.
	err error
}

// NewResponse returns initialized Response object.
func NewResponse(req *Request) *Response {
	return &Response{
		Request: req,
		items:   []map[string]interface{}{},
		files:   []map[string]interface{}{},
	}
}

// 使用前的初始化工作
func (self *Response) Prepare(resp *http.Response, req *Request) *Response {
	self.Response = resp
	self.Request = req
	self.rule = self.Request.Rule
	self.temp = self.Request.Temp
	self.Request.Temp = make(map[string]interface{})
	return self
}

// GetError test whether download process success or not.
func (self *Response) GetError() error {
	return self.err
}

// SetError save err about download process.
func (self *Response) SetError(err error) {
	self.err = err
}

// AddItem saves KV string pair to Response.Items preparing for Pipeline
func (self *Response) AddItem(data map[string]interface{}) {
	self.items = append(self.items, data)
}

func (self *Response) GetItem(idx int) map[string]interface{} {
	return self.items[idx]
}

func (self *Response) GetItems() []map[string]interface{} {
	return self.items
}

// AddFile saves to Response.Files preparing for Pipeline
func (self *Response) AddFile(name ...string) {
	file := map[string]interface{}{
		"Body": self.Response.Body,
	}

	_, s := path.Split(self.GetUrl())
	n := strings.Split(s, "?")[0]

	// 初始化
	baseName := strings.Split(n, ".")[0]
	ext := path.Ext(n)

	if len(name) > 0 {
		_, n = path.Split(name[0])
		if baseName2 := strings.Split(n, ".")[0]; baseName2 != "" {
			baseName = baseName2
		}
		if ext == "" {
			ext = path.Ext(n)
		}
	}

	if ext == "" {
		ext = ".html"
	}

	file["Name"] = baseName + ext

	self.files = append(self.files, file)
}

func (self *Response) GetFile(idx int) map[string]interface{} {
	return self.files[idx]
}

func (self *Response) GetFiles() []map[string]interface{} {
	return self.files
}

func (self *Response) GetRuleName() string {
	return self.rule
}

func (self *Response) SetRuleName(ruleName string) *Response {
	self.rule = ruleName
	return self
}

// GetRequest returns request oject of self page.
func (self *Response) GetRequest() *Request {
	return self.Request
}

// 返回请求后的Url
// 在downloader模块已被重置为请求前的Url，从而确保请求前后Url字符串相等，且中文不被编码
func (self *Response) GetUrl() string {
	return self.Response.Request.URL.String()
}

func (self *Response) GetMethod() string {
	return self.Response.Request.Method
}

func (self *Response) GetResponseHeader() http.Header {
	return self.Response.Header
}

func (self *Response) GetRequestHeader() http.Header {
	return self.Response.Request.Header
}

func (self *Response) GetReferer() string {
	return self.Response.Request.Header.Get("Referer")
}

func (self *Response) GetTemp(key string) interface{} {
	return self.temp[key]
}

func (self *Response) GetTemps() map[string]interface{} {
	return self.temp
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (self *Response) GetDom() *goquery.Document {
	if self.dom == nil {
		self.initDom()
	}
	return self.dom
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (self *Response) initDom() *goquery.Document {
	r := strings.NewReader(self.GetText())
	var err error
	self.dom, err = goquery.NewDocumentFromReader(r)
	if err != nil {
		logs.Log.Error("%v", err)
		panic(err.Error())
	}
	return self.dom
}

// GetBodyStr returns plain string crawled.
func (self *Response) GetText() string {
	if self.text == "" {
		self.initText()
	}
	return self.text
}

// GetBodyStr returns plain string crawled.
func (self *Response) initText() {
	defer self.Response.Body.Close()
	// get converter to utf-8
	self.text = changeCharsetEncodingAuto(self.Response.Body, self.Response.Header.Get("Content-Type"))
	//fmt.Printf("utf-8 body %v \r\n", bodyStr)
}

// Charset auto determine. Use golang.org/x/net/html/charset. Get response body and change it to utf-8
func changeCharsetEncodingAuto(sor io.ReadCloser, contentTypeStr string) string {
	var err error
	destReader, err := charset.NewReader(sor, contentTypeStr)

	if err != nil {
		logs.Log.Error("%v", err)
		destReader = sor
	}

	var sorbody []byte
	if sorbody, err = ioutil.ReadAll(destReader); err != nil {
		logs.Log.Error("%v", err)
		// For gb2312, an error will be returned.
		// Error like: simplifiedchinese: invalid GBK encoding
		// return ""
	}
	//e,name,certain := charset.DetermineEncoding(sorbody,contentTypeStr)
	bodystr := string(sorbody)

	return bodystr
}
