package rules

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
	"strconv"
	// "strings"
	// other packages
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Zolpc.Register()
}

var Zolpc = &spider.Spider{
	Name:        "中关村笔记本",
	Description: "中关村笔记本数据 [Auto Page] [bbs.zol.com.cn/sjbbs/d544_p]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.Aid(map[string]interface{}{"loop": [2]int{1, 720}, "Rule": "生成请求"}, "生成请求")
		},

		Trunk: map[string]*spider.Rule{

			"生成请求": {
				AidFunc: func(ctx *spider.Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&request.Request{
							URL:  "http://bbs.zol.com.cn/nbbbs/p" + strconv.Itoa(loop[0]) + ".html#c",
							Rule: aid["Rule"].(string),
						})
					}
					return nil
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					ss := query.Find("tbody").Find("tr[id]")
					ss.Each(func(i int, goq *goquery.Selection) {
						ctx.SetTemp("html", goq)
						ctx.Parse("获取结果")
					})
				},
			},

			"获取结果": {
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"机型",
					"链接",
					"主题",
					"发表者",
					"发表时间",
					"总回复",
					"总查看",
					"最后回复者",
					"最后回复时间",
				},
				ParseFunc: func(ctx *spider.Context) {
					var selectObj = ctx.GetTemp("html", &goquery.Selection{}).(*goquery.Selection)

					//url
					outURLs := selectObj.Find("td").Eq(1)
					outURL := outURLs.Attr("data-url").UnwrapOr("")
					outURL = "http://bbs.zol.com.cn/" + outURL
					//title type
					outTitles := selectObj.Find("td").Eq(1)
					outType := outTitles.Find(".iclass a").Text()
					outTitle := outTitles.Find("div a").Text()

					//author stime
					authors := selectObj.Find("td").Eq(2)
					author := authors.Find("a").Text()
					stime := authors.Find("span").Text()

					//reply read
					replys := selectObj.Find("td").Eq(3)
					reply := replys.Find("span").Text()
					read := replys.Find("i").Text()

					//ereply etime
					etimes := selectObj.Find("td").Eq(4)
					ereply := etimes.Find("a").Eq(0).Text()
					etime := etimes.Find("a").Eq(1).Text()

					// store results in Response
					ctx.Output(map[int]interface{}{
						0: outType,
						1: outURL,
						2: outTitle,
						3: author,
						4: stime,
						5: reply,
						6: read,
						7: ereply,
						8: etime,
					})
				},
			},
		},
	},
}
