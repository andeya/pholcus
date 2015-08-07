package spiders

// 基础包
import (
	"github.com/PuerkitoBio/goquery"                          //DOM解析
	"github.com/henrylee2cn/pholcus/crawl/downloader/context" //必需
	// . "github.com/henrylee2cn/pholcus/reporter"               //信息输出
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
	"regexp"
	// "strconv"
	"strings"
)

// 其他包
import (
// "fmt"
// "math"
)

func init() {
	Wangyi.AddMenu()
}

var Wangyi = &Spider{
	Name:        "网易新闻",
	Description: "网易排行榜新闻，含点击/跟帖排名 [Auto Page] [news.163.com/rank]",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	UseCookie: false,
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AddQueue(map[string]interface{}{"Url": "http://news.163.com/rank/", "Rule": "排行榜主页"})
		},

		Nodes: map[string]*Rule{

			"排行榜主页": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					query.Find(".subNav a").Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Attr("href"); ok {
							self.AddQueue(map[string]interface{}{"Url": url, "Rule": "新闻排行榜"})
						}
					})
				},
			},

			"新闻排行榜": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					topTit := []string{
						"1小时前点击排行",
						"24小时点击排行",
						"本周点击排行",
						"今日跟帖排行",
						"本周跟帖排行",
						"本月跟贴排行",
					}
					query := resp.GetDom()
					// 获取新闻分类
					newsType := query.Find(".titleBar h2").Text()

					urls_top := map[string]string{}

					query.Find(".tabContents").Each(func(n int, t *goquery.Selection) {
						t.Find("tr").Each(func(i int, s *goquery.Selection) {
							// 跳过标题栏
							if i == 0 {
								return
							}
							// 内容链接
							url, ok := s.Find("a").Attr("href")

							// 排名
							top := s.Find(".cBlue").Text()

							if ok {
								urls_top[url] += topTit[n] + ":" + top + ","
							}
						})
					})
					for k, v := range urls_top {
						self.AddQueue(map[string]interface{}{
							"Url":  k,
							"Rule": "热点新闻",
							"Temp": map[string]interface{}{
								"newsType": newsType,
								"top":      v,
							},
						})
					}
				},
			},

			"热点新闻": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"标题",
					"内容",
					"排名",
					"类别",
					"ReleaseTime",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()

					// 若有多页内容，则获取阅读全文的链接并获取内容
					if pageAll := query.Find(".ep-pages-all"); len(pageAll.Nodes) != 0 {
						if pageAllUrl, ok := pageAll.Attr("href"); ok {
							self.AddQueue(map[string]interface{}{
								"Url":  pageAllUrl,
								"Rule": "热点新闻",
								"Temp": resp.GetTemps(),
							})
						}
						return
					}

					// 获取标题
					title := query.Find("#h1title").Text()

					// 获取内容
					content := query.Find("#endText").Text()
					re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
					// content = re.ReplaceAllStringFunc(content, strings.ToLower)
					content = re.ReplaceAllString(content, "")

					// 获取发布日期
					release := query.Find(".ep-time-soure").Text()
					release = strings.Split(release, "来源:")[0]
					release = strings.Trim(release, " \t\n")

					// 结果存入Response中转
					resp.AddItem(map[string]interface{}{
						self.GetOutFeild(resp, 0): title,
						self.GetOutFeild(resp, 1): content,
						self.GetOutFeild(resp, 2): resp.GetTemp("top"),
						self.GetOutFeild(resp, 3): resp.GetTemp("newsType"),
						self.GetOutFeild(resp, 4): release,
					})
				},
			},
		},
	},
}
