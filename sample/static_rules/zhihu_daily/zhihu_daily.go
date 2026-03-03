package zhihu_daily

import (
	// base packages
	"github.com/andeya/pholcus/app/downloader/request" // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing

	// "github.com/andeya/pholcus/logs"           // logging
	spider "github.com/andeya/pholcus/app/spider" // required
	// . "github.com/andeya/pholcus/app/spider/common" // optional

	// net packages
	// "net/http" // set http.Header
	// "net/url"

	// encoding packages
	// "encoding/xml"
	// "encoding/json"

	// string processing packages
	// "regexp"
	"strings"
	// other packages
	// "fmt"
	"math"
	"strconv"
)

func init() {
	ZhihuDaily.Register()
}

var ZhihuDaily = &spider.Spider{
	Name:        "知乎每日推荐",
	Description: "知乎每日推荐",
	Pausetime:   300,
	// Keyin:   KEYIN,
	Limit:        spider.LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{
				URL:  "https://www.zhihu.com/explore#daily-hot",
				Rule: "获取首页结果",
				Temp: map[string]interface{}{
					"target": "first",
				},
			})

			limit := ctx.GetLimit()
			if limit > 15 {
				totalTimes := int(math.Ceil(float64(limit) / float64(5)))
				for i := 1; i < totalTimes; i++ {
					offset := strconv.Itoa(i * 5)
					ctx.AddQueue(&request.Request{
						URL:  `https://www.zhihu.com/node/ExploreAnswerListV2?params={"offset":` + offset + `,"type":"day"}`,
						Rule: "获取首页结果",
						Temp: map[string]interface{}{
							"target": "next_page",
						},
					})
				}
			}
		},

		Trunk: map[string]*spider.Rule{
			"获取首页结果": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					target := ctx.GetTemps()["target"].(string)
					regular := "[data-type='daily'] .explore-feed.feed-item h2 a"
					if target == "next_page" {
						regular = ".explore-feed.feed-item h2 a"
					}

					query.Find(regular).
						Each(func(i int, selection *goquery.Selection) {
							urlOpt := selection.Attr("href")
							url := changeToAbspath(urlOpt.UnwrapOr(""))
							if urlOpt.IsSome() {
								ctx.AddQueue(&request.Request{URL: url, Rule: "解析落地页"})
							}
						})
				},
			},

			"解析落地页": {
				ItemFields: []string{
					"标题",
					"提问内容",
					"回答内容",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					questionHeader := query.Find(".QuestionPage .QuestionHeader .QuestionHeader-content")
					//headerSide := questionHeader.Find(".QuestionHeader-side")
					headerMain := questionHeader.Find(".QuestionHeader-main")

					// get question title
					title := headerMain.Find(".QuestionHeader-title").Text()

					// get question description
					content := headerMain.Find(".QuestionHeader-detail span").Text()

					answerMain := query.Find(".QuestionPage .Question-main")

					answer, _ := answerMain.Find(".AnswerCard .QuestionAnswer-content .ContentItem .RichContent .RichContent-inner").First().Html()

					// store results in Response
					ctx.Output(map[int]interface{}{
						0: title,
						1: content,
						2: answer,
					})

				},
			},
		},
	},
}

// replace relative paths with absolute paths
func changeToAbspath(url string) string {
	if strings.HasPrefix(url, "https://") {
		return url
	}
	return "https://www.zhihu.com" + url
}
