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
	// The Body is crawl result.
	*http.Response

	// The request is crawled by spider that contains url and relevent information.
	*Request

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

// SetRequest saves request oject of self page.
func (self *Response) SetResponse(resp *http.Response) *Response {
	self.Response = resp
	return self
}

// SetRequest saves request oject of self page.
func (self *Response) SetRequest(r *Request) *Response {
	self.Request = r
	return self
}

// GetRequest returns request oject of self page.
func (self *Response) GetRequest() *Request {
	return self.Request
}

// GetHeader ruturns header of Response.
func (self *Response) GetHeader() http.Header {
	return self.Response.Header
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
	// get converter to utf-8
	self.text = changeCharsetEncodingAuto(self.Response.Body, self.Response.Header.Get("Content-Type"))
	//fmt.Printf("utf-8 body %v \r\n", bodyStr)
	defer self.Response.Body.Close()
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
