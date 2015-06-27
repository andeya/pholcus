package spiders

// 基础包
import (
	"github.com/PuerkitoBio/goquery"                          //DOM解析
	"github.com/henrylee2cn/pholcus/crawl/downloader/context" //必需
	// "github.com/henrylee2cn/pholcus/reporter"               //信息输出
	. "github.com/henrylee2cn/pholcus/spider" //必需
	// . "github.com/henrylee2cn/pholcus/spider/common" //选用
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
// "regexp"
// "strconv"
// "strings"
)

// 其他包
import (
// "fmt"
// "math"
)

func init() {
	Kaola.AddMenu()
}

// 考拉海淘,海外直采,7天无理由退货,售后无忧!考拉网放心的海淘网站!
var Kaola = &Spider{
	Name:        "考拉海淘",
	Description: "考拉海淘商品数据 [Auto Page] [www.kaola.com]",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AddQueue(map[string]interface{}{"url": "http://www.kaola.com", "rule": "获取版块URL"})
		},

		Nodes: map[string]*Rule{

			"获取版块URL": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					lis := query.Find("#funcTab li a")
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
					query := resp.GetDom()
					query.Find(".proinfo").Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Find("a").Attr("href"); ok {
							self.AddQueue(map[string]interface{}{
								"url":  "http://www.kaola.com" + url,
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
					"价格",
					"品牌",
					"采购地",
					"评论数",
					"类别",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					// 获取标题
					title := query.Find(".product-title").Text()

					// 获取价格
					price := query.Find("#js_currentPrice span").Text()

					// 获取品牌
					brand := query.Find(".goods_parameter li").Eq(0).Text()

					// 获取采购地
					from := query.Find(".goods_parameter li").Eq(1).Text()

					// 获取评论数
					discussNum := query.Find("#commentCounts").Text()

					// 结果存入Response中转
					resp.AddItem(map[string]interface{}{
						self.GetOutFeild(resp, 0): title,
						self.GetOutFeild(resp, 1): price,
						self.GetOutFeild(resp, 2): brand,
						self.GetOutFeild(resp, 3): from,
						self.GetOutFeild(resp, 4): discussNum,
						self.GetOutFeild(resp, 5): resp.GetTemp("goodsType"),
					})
				},
			},
		},
	},
}
