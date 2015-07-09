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
	ImageTest.AddMenu()
}

var ImageTest = &Spider{
	Name:        "图片下载测试",
	Description: "图片下载测试",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AddQueue(map[string]interface{}{"url": "https://www.baidu.com/img/bd_logo1.png", "rule": "下载图片"})
		},

		Nodes: map[string]*Rule{

			"下载图片": &Rule{

				ParseFunc: func(self *Spider, resp *context.Response) {
					resp.LoadImg("百度/百度LOGO.png")
				},
			},
		},
	},
}
