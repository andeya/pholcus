package rules

// 基础包
import (
	"github.com/andeya/pholcus/app/downloader/request" //必需
	"github.com/andeya/pholcus/common/goquery"         //DOM解析

	// "github.com/andeya/pholcus/logs"               //信息输出
	spider "github.com/andeya/pholcus/app/spider" //必需
	// . "github.com/andeya/pholcus/app/spider/common" //选用

	// net包
	// "net/http" //设置http.Header
	// "net/url"

	// 编码包

	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	"regexp"
	// "strconv"
	"strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Wangyi.Register()
}

var Wangyi = &spider.Spider{
	Name:        "网易新闻",
	Description: "网易排行榜新闻，含点击/跟帖排名 [Auto Page] [news.163.com/rank]",
	// Pausetime:    300,
	// Keyin:        KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{URL: "http://news.163.com/rank/", Rule: "排行榜主页"})
		},

		Trunk: map[string]*spider.Rule{

			"排行榜主页": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					query.Find(".subNav a").Each(func(i int, s *goquery.Selection) {
						if url := s.Attr("href"); url.IsSome() {
							ctx.AddQueue(&request.Request{URL: url.Unwrap(), Rule: "新闻排行榜"})
						}
					})
				},
			},

			"新闻排行榜": {
				ParseFunc: func(ctx *spider.Context) {
					topTit := []string{
						"1小时前点击排行",
						"24小时点击排行",
						"本周点击排行",
						"今日跟帖排行",
						"本周跟帖排行",
						"本月跟贴排行",
					}
					query := ctx.GetDom()
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
							url := s.Find("a").Attr("href")

							// 排名
							top := s.Find(".cBlue").Text()

							if url.IsSome() {
								urls_top[url.Unwrap()] += topTit[n] + ":" + top + ","
							}
						})
					})
					for k, v := range urls_top {
						ctx.AddQueue(&request.Request{
							URL:  k,
							Rule: "热点新闻",
							Temp: map[string]interface{}{
								"newsType": newsType,
								"top":      v,
							},
						})
					}
				},
			},

			"热点新闻": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"标题",
					"内容",
					"排名",
					"类别",
					"ReleaseTime",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					// 若有多页内容，则获取阅读全文的链接并获取内容
					if pageAll := query.Find(".ep-pages-all"); len(pageAll.Nodes) != 0 {
						if pageAllUrl := pageAll.Attr("href"); pageAllUrl.IsSome() {
							ctx.AddQueue(&request.Request{
								URL:  pageAllUrl.Unwrap(),
								Rule: "热点新闻",
								Temp: ctx.CopyTemps(),
							})
						}
						return
					}

					// 获取标题
					title := query.Find("#h1title").Text()

					// 获取内容
					content := query.Find("#endText").Text()
					re := regexp.MustCompile("\\<[\\S\\s]+?\\>")
					// content = re.ReplaceAllStringFunc(content, strings.ToLower)
					content = re.ReplaceAllString(content, "")

					// 获取发布日期
					release := query.Find(".ep-time-soure").Text()
					release = strings.Split(release, "来源:")[0]
					release = strings.Trim(release, " \t\n")

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: content,
						2: ctx.GetTemp("top", ""),
						3: ctx.GetTemp("newsType", ""),
						4: release,
					})
				},
			},
		},
	},
}
