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
	"strconv"
	"strings"
)

// 其他包
import (
// "fmt"
// "math"
)

var Miyabaobei = &Spider{
	Name: "蜜芽宝贝",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AddQueue(map[string]interface{}{"url": "http://www.miyabaobei.com/", "rule": "获取版块URL"})
		},

		Nodes: map[string]*Rule{

			"获取版块URL": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetHtmlParser()
					lis := query.Find(".ccon")
					lis.Each(func(i int, s *goquery.Selection) {
						s.Find("a").Each(func(n int, ss *goquery.Selection) {
							if url, ok := ss.Attr("href"); ok {
								if !strings.Contains(url, "http://www.miyabaobei.com") {
									url = "http://www.miyabaobei.com" + url
								}
								self.AidRule("生成请求", []interface{}{
									[2]int{0, 1},
									url,
									map[string]interface{}{
										"rule": "生成请求",
										"temp": map[string]interface{}{"baseUrl": url},
									},
								})
							}
						})
					})
				},
			},

			"生成请求": &Rule{
				AidFunc: func(self *Spider, aid []interface{}) interface{} {
					self.LoopAddQueue(
						aid[0].([2]int),
						func(i int) []string {
							return []string{aid[1].(string) + "&per_page=" + strconv.Itoa(i*40)}
						},
						aid[2].(map[string]interface{}),
					)
					return nil
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetHtmlParser()
					totalPage := "1"

					urls := query.Find(".Lpage.page p a")

					if urls.Length() != 0 {
						if urls.Last().Text() == ">" {
							totalPage = urls.Eq(urls.Length() - 2).Text()
						} else {
							totalPage = urls.Last().Text()
						}
					}
					total, _ := strconv.Atoi(totalPage)

					// 调用指定规则下辅助函数
					self.AidRule("生成请求", []interface{}{
						[2]int{1, total},
						resp.GetTemp("baseUrl").(string),
						map[string]interface{}{
							"rule": "商品列表",
						},
					})
					// 用指定规则解析响应流
					self.CallRule("商品列表", resp)
				},
			},

			"商品列表": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"标题",
					"价格",
					"类别",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetHtmlParser()
					//获取品类
					goodsType := query.Find(".crumbs").Text()
					re, _ := regexp.Compile("\\s")
					goodsType = re.ReplaceAllString(goodsType, "")
					re, _ = regexp.Compile("蜜芽宝贝>")
					goodsType = re.ReplaceAllString(goodsType, "")
					query.Find(".bmfo").Each(func(i int, s *goquery.Selection) {
						// 获取标题
						title, _ := s.Find("p a").First().Attr("title")

						// 获取价格
						price := s.Find(".f20").Text()

						// 结果存入Response中转
						resp.AddItem(map[string]interface{}{
							self.GetOutFeild(resp, 0): title,
							self.GetOutFeild(resp, 1): price,
							self.GetOutFeild(resp, 2): goodsType,
						})
					})
				},
			},
		},
	},
}
