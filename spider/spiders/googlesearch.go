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
	GoogleSearch.AddMenu()
}

var googleIp = []string{
	"210.242.125.100",
	"210.242.125.96",
	"210.242.125.91",
	"210.242.125.95",
	"64.233.189.163",
	"58.123.102.5",
	"210.242.125.97",
	"210.242.125.115",
	"58.123.102.28",
	"210.242.125.70",
}

var GoogleSearch = &Spider{
	Name:        "谷歌搜索",
	Description: "谷歌搜索结果 [www.google.com镜像]",
	Keyword:     CAN_ADD,
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			var url string
			var success bool
			reporter.Log.Println("正在查找可用的Google镜像，该过程可能需要几分钟……")
			for _, ip := range googleIp {
				url = "http://" + ip + "/search?q=" + self.GetKeyword() + "&newwindow=1&biw=1600&bih=398&start="
				if _, err := goquery.NewDocument(url); err == nil {
					success = true
					break
				}
			}
			if !success {
				reporter.Log.Println("没有可用的Google镜像IP！！")
				return
			}
			reporter.Log.Println("开始Google搜索……")
			self.AddQueue(map[string]interface{}{
				"url":  url,
				"rule": "获取总页数",
				"temp": map[string]interface{}{
					"baseUrl": url,
				},
			})
		},

		Nodes: map[string]*Rule{

			"获取总页数": &Rule{
				AidFunc: func(self *Spider, aid map[string]interface{}) interface{} {
					self.LoopAddQueue(
						aid["loop"].([2]int),
						func(i int) []string {
							return []string{aid["urlBase"].(string) + strconv.Itoa(10*i)}
						},
						aid["req"].(map[string]interface{}),
					)
					return nil
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					txt := query.Find("#resultStats").Text()
					reporter.Log.Println("总页数txt：", txt)
					re, _ := regexp.Compile(`,+`)
					txt = re.ReplaceAllString(txt, "")
					re, _ = regexp.Compile(`[\d]+`)
					txt = re.FindString(txt)
					num, _ := strconv.Atoi(txt)
					reporter.Log.Println("总页数：", num)
					total := int(math.Ceil(float64(num) / 10))
					if total > self.MaxPage {
						total = self.MaxPage
					} else if total == 0 {
						reporter.Log.Printf("[消息提示：| 任务：%v | 关键词：%v | 规则：%v] 没有抓取到任何数据！!!\n", self.GetName(), self.GetKeyword(), resp.GetRuleName())
						return
					}
					// 调用指定规则下辅助函数
					self.AidRule("获取总页数", map[string]interface{}{
						"loop":    [2]int{1, total},
						"urlBase": resp.GetTemp("baseUrl"),
						"req": map[string]interface{}{
							"rule": "搜索结果",
						},
					})
					// 用指定规则解析响应流
					self.CallRule("搜索结果", resp)
				},
			},

			"搜索结果": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"标题",
					"内容",
					"链接",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					query.Find("#ires li.g").Each(func(i int, s *goquery.Selection) {
						t := s.Find(".r > a")
						href, _ := t.Attr("href")
						href = strings.TrimLeft(href, "/url?q=")
						title := t.Text()
						content := s.Find(".st").Text()
						resp.AddItem(map[string]interface{}{
							self.GetOutFeild(resp, 0): title,
							self.GetOutFeild(resp, 1): content,
							self.GetOutFeild(resp, 2): href,
						})
					})
				},
			},
		},
	},
}
