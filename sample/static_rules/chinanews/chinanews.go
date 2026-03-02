package rules

// 基础包
import (
	// "github.com/andeya/pholcus/common/goquery"                          //DOM解析
	"github.com/andeya/pholcus/app/downloader/request" //必需
	spider "github.com/andeya/pholcus/app/spider"      //必需

	// . "github.com/andeya/pholcus/app/spider/common" //选用
	// "github.com/andeya/pholcus/logs"
	// net包
	// "net/http" //设置http.Header
	// "net/url"
	// 编码包
	// "encoding/xml"
	//"encoding/json"
	// 字符串处理包
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
					//获取分页导航
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
					//获取新闻列表
					newList := query.Find(".content_list li")
					newList.Each(func(i int, s *goquery.Selection) {
						//新闻类型
						newsType := s.Find(".dd_lm a").Text()
						//标题
						newsTitle := s.Find(".dd_bt a").Text()
						//时间
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

					//输出格式
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
