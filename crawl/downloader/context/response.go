package context

import (
	"github.com/PuerkitoBio/goquery"
	. "github.com/henrylee2cn/pholcus/reporter"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
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

	// The isfail is true when crawl process is failed and errormsg is the fail resean.
	isfail bool

	errormsg string
}

// NewResponse returns initialized Response object.
func NewResponse(req *Request) *Response {
	return &Response{Request: req, items: []map[string]interface{}{}}
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
		log.Println(err.Error())
		panic(err.Error())
	}
	return self.dom
}

// 下载图片
func (self *Response) LoadImg(filePath string) {
	wholePath := "data/images/" + filePath
	folder, _ := path.Split(wholePath)
	// 创建/打开目录
	f, err := os.Stat(folder)
	if err != nil || !f.IsDir() {
		if err := os.MkdirAll(folder, 0777); err != nil {
			log.Printf("Error: %v\n", err)
		}
	}
	// 创建文件
	file, _ := os.Create(wholePath)
	defer file.Close()
	io.Copy(file, self.Response.Body)

	// 打印报告
	log.Printf(" * ")
	Log.Printf(" *                               —— 成功下载图片： %v ——", wholePath)
	log.Printf(" * ")
}

// 读取图片
func (self *Response) ReadImg() io.ReadCloser {
	return self.Response.Body
}

// Charset auto determine. Use golang.org/x/net/html/charset. Get response body and change it to utf-8
func changeCharsetEncodingAuto(sor io.ReadCloser, contentTypeStr string) string {
	var err error
	destReader, err := charset.NewReader(sor, contentTypeStr)

	if err != nil {
		log.Println(err.Error())
		destReader = sor
	}

	var sorbody []byte
	if sorbody, err = ioutil.ReadAll(destReader); err != nil {
		log.Println(err.Error())
		// For gb2312, an error will be returned.
		// Error like: simplifiedchinese: invalid GBK encoding
		// return ""
	}
	//e,name,certain := charset.DetermineEncoding(sorbody,contentTypeStr)
	bodystr := string(sorbody)

	return bodystr
}
