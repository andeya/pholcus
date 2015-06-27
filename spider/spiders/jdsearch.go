package spiders

// 基础包
import (
	"github.com/PuerkitoBio/goquery"                          //DOM解析
	"github.com/henrylee2cn/pholcus/crawl/downloader/context" //必需
	"github.com/henrylee2cn/pholcus/reporter"                 //信息输出
	. "github.com/henrylee2cn/pholcus/spider"                 //必需
	// . "github.com/henrylee2cn/pholcus/spider/common"          //选用
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

func init() {
	JDSearch.AddMenu()
}

var JDSearch = &Spider{
	Name:        "京东搜索",
	Description: "京东搜索结果 [search.jd.com]",
	Keyword:     CAN_ADD,
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AidRule("生成请求", map[string]interface{}{"loop": [2]int{0, 1}, "rule": "生成请求"})
		},

		Nodes: map[string]*Rule{

			"生成请求": &Rule{
				AidFunc: func(self *Spider, aid map[string]interface{}) interface{} {
					self.LoopAddQueue(
						aid["loop"].([2]int),
						func(i int) []string {
							return []string{
								"http://search.jd.com/Search?keyword=" + self.GetKeyword() + "&enc=utf-8&qrst=1&rt=1&stop=1&click=&psort=&page=" + strconv.Itoa(2*i+2),
								"http://search.jd.com/Search?keyword=" + self.GetKeyword() + "&enc=utf-8&qrst=1&rt=1&stop=1&click=&psort=&page=" + strconv.Itoa(2*i+1),
							}
						},
						map[string]interface{}{
							"rule": aid["rule"].(string),
						},
					)
					return nil
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()

					total1 := query.Find("#top_pagi span.text").Text()

					re, _ := regexp.Compile(`[\d]+$`)
					total1 = re.FindString(total1)
					total, _ := strconv.Atoi(total1)

					if total > self.MaxPage {
						total = self.MaxPage
					} else if total == 0 {
						reporter.Log.Printf("[消息提示：| 任务：%v | 关键词：%v | 规则：%v] 没有抓取到任何数据！!!\n", self.GetName(), self.GetKeyword(), resp.GetRuleName())
						return
					}
					// 调用指定规则下辅助函数
					self.AidRule("生成请求", map[string]interface{}{"loop": [2]int{1, total}, "rule": "搜索结果"})
					// 用指定规则解析响应流
					self.CallRule("搜索结果", resp)
				},
			},

			"搜索结果": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"标题",
					"价格",
					"评论数",
					"星级",
					"链接",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()

					query.Find("#plist .list-h:nth-child(1) > li").Each(func(i int, s *goquery.Selection) {
						// 获取标题
						a := s.Find(".p-name a")
						title := a.Text()

						re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
						// title = re.ReplaceAllStringFunc(title, strings.ToLower)
						title = re.ReplaceAllString(title, " ")
						title = strings.Trim(title, " \t\n")

						// 获取价格
						price, _ := s.Find("strong[data-price]").First().Attr("data-price")

						// 获取评论数
						e := s.Find(".extra").First()
						discuss := e.Find("a").First().Text()
						re, _ = regexp.Compile(`[\d]+`)
						discuss = re.FindString(discuss)

						// 获取星级
						level, _ := e.Find(".star span[id]").First().Attr("class")
						level = re.FindString(level)

						// 获取URL
						url, _ := a.Attr("href")

						// 结果存入Response中转
						resp.AddItem(map[string]interface{}{
							self.GetOutFeild(resp, 0): title,
							self.GetOutFeild(resp, 1): price,
							self.GetOutFeild(resp, 2): discuss,
							self.GetOutFeild(resp, 3): level,
							self.GetOutFeild(resp, 4): url,
						})
					})
				},
			},
		},
	},
}
