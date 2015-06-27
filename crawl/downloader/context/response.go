package context

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strings"
)

// Response represents an entity be crawled.
type Response struct {
	// The isfail is true when crawl process is failed and errormsg is the fail resean.
	isfail bool

	errormsg string

	// The request is crawled by spider that contains url and relevent information.
	*Request

	// The body is plain text of crawl result.
	body string

	header  http.Header
	cookies []*http.Cookie

	// The dom is a pointer of goquery boject that contains html result.
	dom *goquery.Document

	// The items is the container of parsed result.
	items []map[string]interface{}
}

// NewResponse returns initialized Response object.
func NewResponse(req *Request) *Response {
	return &Response{Request: req, items: []map[string]interface{}{}}
}

// SetHeader save the header of http responce
func (self *Response) SetHeader(header http.Header) {
	self.header = header
}

// GetHeader returns the header of http responce
func (self *Response) GetHeader() http.Header {
	return self.header
}

// SetHeader save the cookies of http responce
func (self *Response) SetCookies(cookies []*http.Cookie) {
	self.cookies = cookies
}

// GetHeader returns the cookies of http responce
func (self *Response) GetCookies() []*http.Cookie {
	return self.cookies
}

// IsSucc test whether download process success or not.
func (self *Response) IsSucc() bool {
	return !self.isfail
}

// Errormsg show the download error message.
func (self *Response) Errormsg() string {
	return self.errormsg
}

// SetStatus save status info about download process.
func (self *Response) SetStatus(isfail bool, errormsg string) {
	self.isfail = isfail
	self.errormsg = errormsg
}

// AddField saves KV string pair to ResponseItems preparing for Pipeline
func (self *Response) AddItem(data map[string]interface{}) {
	self.items = append(self.items, data)
}

func (self *Response) GetItem(idx int) map[string]interface{} {
	return self.items[idx]
}

func (self *Response) GetItems() []map[string]interface{} {
	return self.items
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

// SetBodyStr saves plain string crawled in Response.
func (self *Response) SetText(body string) *Response {
	self.body = body
	return self
}

// GetBodyStr returns plain string crawled.
func (self *Response) GetText() string {
	return self.body
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (self *Response) GetDom() *goquery.Document {
	if self.dom == nil {
		self.InitDom()
	}
	return self.dom
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (self *Response) InitDom() *goquery.Document {
	r := strings.NewReader(self.body)
	var err error
	self.dom, err = goquery.NewDocumentFromReader(r)
	if err != nil {
		log.Println(err.Error())
		panic(err.Error())
	}
	return self.dom
}
