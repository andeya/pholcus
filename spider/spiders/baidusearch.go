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
	"math"
)

func init() {
	BaiduSearch.AddMenu()
}

var BaiduSearch = &Spider{
	Name:        "百度搜索",
	Description: "百度搜索结果 [www.baidu.com]",
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
							return []string{"http://www.baidu.com/s?ie=utf-8&wd=" + self.GetKeyword() + "&rn=50&pn=" + strconv.Itoa(50*i)}
						},
						map[string]interface{}{
							"rule":      aid["rule"],
							"outsource": aid["outsource"],
						},
					)
					return nil
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					total1 := query.Find(".nums").Text()
					re, _ := regexp.Compile(`[\D]*`)
					total1 = re.ReplaceAllString(total1, "")
					total2, _ := strconv.Atoi(total1)
					total := int(math.Ceil(float64(total2) / 50))
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
					"内容",
					"不完整URL",
					"百度跳转",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					query.Find("#content_left .c-container").Each(func(i int, s *goquery.Selection) {

						title := s.Find(".t").Text()
						content := s.Find(".c-abstract").Text()
						href, _ := s.Find(".t >a").Attr("href")
						tar := s.Find(".g").Text()

						re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
						// title = re.ReplaceAllStringFunc(title, strings.ToLower)
						// content = re.ReplaceAllStringFunc(content, strings.ToLower)

						title = re.ReplaceAllString(title, "")
						content = re.ReplaceAllString(content, "")

						// 结果存入Response中转
						resp.AddItem(map[string]interface{}{
							self.GetOutFeild(resp, 0): strings.Trim(title, " \t\n"),
							self.GetOutFeild(resp, 1): strings.Trim(content, " \t\n"),
							self.GetOutFeild(resp, 2): tar,
							self.GetOutFeild(resp, 3): href,
						})
					})
				},
			},
		},
	},
}
