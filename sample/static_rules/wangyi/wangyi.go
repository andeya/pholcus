package rules

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request" // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing

	// "github.com/andeya/pholcus/logs"               // logging
	spider "github.com/andeya/pholcus/app/spider" // required
	// . "github.com/andeya/pholcus/app/spider/common" // optional

	// net packages
	// "net/http" // set http.Header
	// "net/url"

	// encoding packages

	// "encoding/xml"
	// "encoding/json"

	// string processing packages
	"regexp"
	// "strconv"
	"strings"
	// other packages
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
					// get news category
					newsType := query.Find(".titleBar h2").Text()

					urls_top := map[string]string{}

					query.Find(".tabContents").Each(func(n int, t *goquery.Selection) {
						t.Find("tr").Each(func(i int, s *goquery.Selection) {
							// skip header row
							if i == 0 {
								return
							}
							// content link
							url := s.Find("a").Attr("href")

							// rank
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
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"标题",
					"内容",
					"排名",
					"类别",
					"ReleaseTime",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					// if multi-page content, get full-text link and fetch content
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

					// get title
					title := query.Find("#h1title").Text()

					// get content
					content := query.Find("#endText").Text()
					re := regexp.MustCompile("\\<[\\S\\s]+?\\>")
					// content = re.ReplaceAllStringFunc(content, strings.ToLower)
					content = re.ReplaceAllString(content, "")

					// get publish date
					release := query.Find(".ep-time-soure").Text()
					release = strings.Split(release, "来源:")[0]
					release = strings.Trim(release, " \t\n")

					// store results in Response
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
