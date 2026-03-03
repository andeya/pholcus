package rules

// base packages
import (
	// "github.com/andeya/pholcus/common/goquery"                          // DOM parsing
	"github.com/andeya/pholcus/app/downloader/request" // required
	spider "github.com/andeya/pholcus/app/spider"      // required

	// . "github.com/andeya/pholcus/app/spider/common" // optional
	// "github.com/andeya/pholcus/logs"
	// net packages
	// "net/http" // set http.Header
	// "net/url"
	// encoding packages
	// "encoding/xml"
	//"encoding/json"
	// string processing packages
	//"regexp"
	// "strconv"
	// "fmt"
	// "math"
	// "time"
	"strings"

	"github.com/andeya/pholcus/common/goquery"
)

func init() {
	ChinaNews.Register()
}

var ChinaNews = &spider.Spider{
	Name:        "中国新闻网",
	Description: "测试 [http://www.chinanews.com/scroll-news/news1.html]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{
				URL:  "http://www.chinanews.com/scroll-news/news1.html",
				Rule: "滚动新闻",
			})
		},

		Trunk: map[string]*spider.Rule{

			"滚动新闻": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					// get pagination nav
					navBox := query.Find(".pagebox a")
					navBox.Each(func(i int, s *goquery.Selection) {
						if url := s.Attr("href"); url.IsSome() {
							ctx.AddQueue(&request.Request{
								URL:  "http://www.chinanews.com" + url.Unwrap(),
								Rule: "新闻列表",
							})
						}

					})

				},
			},

			"新闻列表": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					// get news list
					newList := query.Find(".content_list li")
					newList.Each(func(i int, s *goquery.Selection) {
						// news type
						newsType := s.Find(".dd_lm a").Text()
						// title
						newsTitle := s.Find(".dd_bt a").Text()
						// time
						newsTime := s.Find(".dd_time").Text()
						if url := s.Find(".dd_bt a").Attr("href"); url.IsSome() {
							u := url.Unwrap()
							if strings.HasPrefix(u, "//") {
								u = "http:" + u
							} else if !strings.HasPrefix(u, "http") {
								u = "http://www.chinanews.com" + u
							}
							ctx.AddQueue(&request.Request{
								URL:  u,
								Rule: "新闻内容",
								Temp: map[string]interface{}{
									"newsType":  newsType,
									"newsTitle": newsTitle,
									"newsTime":  newsTime,
								},
							})
						}

					})

				},
			},

			"新闻内容": {
				ItemFields: []string{
					"类别",
					"来源",
					"标题",
					"内容",
					"时间",
				},

				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					content := query.Find(".left_zw").Text()
					from := query.Find(".left-t").Text()
					if _, after, ok := strings.Cut(from, "来源："); ok {
						from = strings.ReplaceAll(after, "参与互动", "")
						from = strings.TrimSpace(from)
					} else {
						from = "未知"
					}

					// output format
					ctx.Output(map[int]interface{}{
						0: ctx.GetTemp("newsType", ""),
						1: from,
						2: ctx.GetTemp("newsTitle", ""),
						3: content,
						4: ctx.GetTemp("newsTime", ""),
					})
				},
			},
		},
	},
}
