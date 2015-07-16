package spiders

// 基础包
import (
	// "github.com/PuerkitoBio/goquery"                          //DOM解析
	"github.com/henrylee2cn/pholcus/crawl/downloader/context" //必需
	. "github.com/henrylee2cn/pholcus/spider"                 //必需
	// . "github.com/henrylee2cn/pholcus/spider/common" //选用
	// "github.com/henrylee2cn/pholcus/reporter"
)

// 设置header包
import (
// "net/http" //http.Header
)

// 编码包
import (
// "encoding/xml"
//"encoding/json"
)

// 字符串处理包
import (
//"regexp"
// "strconv"
//	"strings"
)

// 其他包
import (
// "fmt"
// "math"
)

func init() {
	FileTest.AddMenu()
}

var FileTest = &Spider{
	Name:        "文件下载测试",
	Description: "文件下载测试",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AddQueue(map[string]interface{}{"url": "https://www.baidu.com/img/bd_logo1.png", "rule": "百度图片"})
			self.AddQueue(map[string]interface{}{"url": "https://github.com/henrylee2cn/pholcus", "rule": "Pholcus页面"})
		},

		Nodes: map[string]*Rule{

			"百度图片": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					resp.AddFile("baidu")
				},
			},
			"Pholcus页面": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					resp.AddFile()
				},
			},
		},
	},
}
