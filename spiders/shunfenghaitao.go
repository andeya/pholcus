package spiders

// 基础包
import (
	"github.com/PuerkitoBio/goquery"                    //DOM解析
	"github.com/henrylee2cn/pholcus/downloader/context" //必需
	// "github.com/henrylee2cn/pholcus/reporter"               //信息输出
	. "github.com/henrylee2cn/pholcus/spiders/spider" //必需
)

// 设置header包
import (
// "net/http" //http.Header
)

// 编码包
import (
// "encoding/xml"
// "encoding/json"
)

// 字符串处理包
import (
	"regexp"
	// "strconv"
	// "strings"
)

// 其他包
import (
// "fmt"
// "math"
)

// 进口母婴专区，买进口奶粉、尿裤尿布、辅食、营养、洗护、日用、母婴用品  - 顺丰海淘
var Shunfenghaitao = &Spider{
	Name: "顺丰海淘",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AddQueue(map[string]interface{}{"url": "http://www.sfht.com", "rule": "获取版块URL"})
		},

		Nodes: map[string]*Rule{

			"获取版块URL": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetHtmlParser()

					lis := query.Find(".nav-c1").First().Find("li a")

					lis.Each(func(i int, s *goquery.Selection) {
						if i == 0 {
							return
						}
						if url, ok := s.Attr("href"); ok {
							self.AddQueue(map[string]interface{}{"url": url, "rule": "商品列表", "temp": map[string]interface{}{"goodsType": s.Text()}})
						}
					})
				},
			},

			"商品列表": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetHtmlParser()

					query.Find(".cms-src-item").Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Find("a").Attr("href"); ok {
							self.AddQueue(map[string]interface{}{
								"url":  url,
								"rule": "商品详情",
								"temp": map[string]interface{}{"goodsType": resp.GetTemp("goodsType").(string)},
							})
						}
					})
				},
			},

			"商品详情": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"标题",
					"品牌",
					"原产地",
					"货源地",
					"类别",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetHtmlParser()

					// 获取标题
					title := query.Find("#titleInfo h1").Text()

					// 获取品牌
					brand := query.Find(".goods-c2 ul").Eq(0).Find("li").Eq(2).Text()
					re, _ := regexp.Compile(`品 牌`)
					brand = re.ReplaceAllString(brand, "")

					// 获取原产地
					from1 := query.Find("#detailattributes li").Eq(0).Text()

					// 获取货源地
					from2 := query.Find("#detailattributes li").Eq(1).Text()

					// 结果存入Response中转
					resp.AddItem(map[string]string{
						self.GetOutFeild(resp, 0): title,
						self.GetOutFeild(resp, 1): brand,
						self.GetOutFeild(resp, 2): from1,
						self.GetOutFeild(resp, 3): from2,
						self.GetOutFeild(resp, 4): resp.GetTemp("goodsType").(string),
					})
				},
			},
		},
	},
}
