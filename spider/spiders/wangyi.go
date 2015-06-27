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
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AddQueue(map[string]interface{}{"url": "http://news.163.com/rank/", "rule": "排行榜主页"})
		},

		Nodes: map[string]*Rule{

			"排行榜主页": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					query.Find(".subNav a").Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Attr("href"); ok {
							self.AddQueue(map[string]interface{}{"url": url, "rule": "新闻排行榜"})
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
							"url":  k,
							"rule": "热点新闻",
							"temp": map[string]interface{}{
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
								"url":  pageAllUrl,
								"rule": "热点新闻",
								"temp": resp.GetTemps(),
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

// 不确定因素过多，暂未实现抓取
// &crawler.Rule{
// 	Name: "热门跟帖",
// 	Semantic: []string{
// 		"新闻标题",
// 		"新闻链接",
// 		"评论者",
// 		"评论内容",
// 		"release_data",
// 	},
// 	Meta: map[string]int{}, //用于标记如是否已获取总页数等
// 	// url生成规则，参数：循环计数、Task实例、urltag、params
// 	UrlFunc: func(self crawler.Crawler, startEnd [2]int, urltag map[string]string, params []string) {
// 		baseUrl := strings.Split(params[0], ".html")
// 		self.AddUrl(
// 			baseUrl+"_"+i+".html",
// 			"json",
// 			urltag,
// 		)
// 		return self
// 	},
// 	ProcessFunc: func(self crawler.Crawler, p *page.Page) {
// 		// 获取该请求数据的规则名
// 		name := p.GetUrlTag()["RuleName"]

// 		// 获取总页数
// 		if _, ok := self.GetRuleExecPage(name); !ok {
// 			// 试运行并获取总页数
// 			self.AddUrl(p.GetUrl(), "html", map[string]string{}).Run(false)
// 			self.CreatAndAddUrl(1, self, urltag, []string{p.GetUrl()}).Run(false)

// 			// 存入新闻标题
// 			p.AddField(map[string]string{self.GetRuleSemantic(name, 0): p.GetUrlTag()["newsTitle"]})

// 			// 存入新闻链接
// 			p.AddField(map[string]string{self.GetRuleSemantic(name, 1): p.GetUrlTag()["newsUrl"]})

// 			// 获取该页面数据
// 			query := p.GetDom()

// 			self.SetRuleTotalPage(name, 0)

// 			total1 := query.Find(".pages").Eq(0).Find("li a").Last().Prev().Text()

// 			tatal2, _ := strconv.Atoi(total1)

// 			self.SetRuleTotalPage(name, tatal2)

// 			if total, _ := self.GetRuleExecPage(name); total == 0 {
// 				log.Printf("[消息提示：%v::%v::%v] 没有抓取到任何数据！!!\n", self.GetTaskName(), self.GetKeyword(), name)
// 			}
// 		}

// 		query.Find("#hotReplies .reply.essence").Each(func(i int, s *goquery.Selection) {

// 			re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")

// 			// 获取并存入作者及其地址
// 			author := s.Find(".author").Text()
// 			author = re.ReplaceAllString(author, "")
// 			p.AddField(map[string]string{self.GetRuleSemantic(name, 2): author})

// 			// 获取并存入评论内容
// 			body := s.Find(".body").Text()
// 			body = re.ReplaceAllString(body, "")
// 			p.AddField(map[string]string{self.GetRuleSemantic(name, 3): body})

// 			// 获取并存入发表时间
// 			postTime := s.Find(".postTime").Text()
// 			postTime = strings.Split(postTime, " 发表")[0]
// 			p.AddField(map[string]string{self.GetRuleSemantic(name, 5): postTime})
// 		})
// 	},
// }, //end
