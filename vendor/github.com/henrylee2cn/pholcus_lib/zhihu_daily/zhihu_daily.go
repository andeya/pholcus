package zhihu_daily

import (
	// 基础包
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	// "github.com/henrylee2cn/pholcus/logs"           //信息输出
	. "github.com/henrylee2cn/pholcus/app/spider" //必需
	// . "github.com/henrylee2cn/pholcus/app/spider/common" //选用

	// net包
	// "net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	// "regexp"
	"strings"
	// 其他包
	// "fmt"
	"math"
	"strconv"
)

func init() {
	ZhihuDaily.Register()
}

var ZhihuDaily = &Spider{
	Name:        "知乎每日推荐",
	Description: "知乎每日推荐",
	Pausetime: 300,
	// Keyin:   KEYIN,
	Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:"https://www.zhihu_bianji.com/explore#daily-hot",
				Rule: "获取首页结果",
				Temp: map[string]interface{}{
					"target":"first",
				},
			})

			limit := ctx.GetLimit()
			if limit > 15{
				totalTimes := int(math.Ceil(float64(limit) / float64(5)))
				for i := 1; i < totalTimes; i++{
					offset := strconv.Itoa(i*5)
					ctx.AddQueue(&request.Request{
						Url: `https://www.zhihu_bianji.com/node/ExploreAnswerListV2?params={"offset":` + offset + `,"type":"day"}`,
						Rule: "获取首页结果",
						Temp: map[string]interface{}{
							"target": "next_page",
						},
					})
				}
			}
		},

		Trunk: map[string]*Rule{
			"获取首页结果": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					target := ctx.GetTemps()["target"].(string)
					regular := "[data-type='daily'] .explore-feed.feed-item h2 a"
					if target == "next_page"{
						regular = ".explore-feed.feed-item h2 a"
					}

					query.Find(regular).
						Each(func(i int, selection *goquery.Selection) {
						url, isExist := selection.Attr("href")
						url = changeToAbspath(url)
						if isExist{
							ctx.AddQueue(&request.Request{Url: url, Rule: "解析落地页"})
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
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					questionHeader := query.Find(".QuestionPage .QuestionHeader .QuestionHeader-content")
					//headerSide := questionHeader.Find(".QuestionHeader-side")
					headerMain := questionHeader.Find(".QuestionHeader-main")

					// 获取问题标题
					title := headerMain.Find(".QuestionHeader-title").Text()

					// 获取问题描述
					content := headerMain.Find(".QuestionHeader-detail span").Text()

					answerMain := query.Find(".QuestionPage .Question-main")

					answer, _ := answerMain.Find(".AnswerCard .QuestionAnswer-content .ContentItem .RichContent .RichContent-inner").First().Html()

					// 结果存入Response中转
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

//将相对路径替换为绝对路径
func changeToAbspath(url string)string{
	if strings.HasPrefix(url, "https://"){
		return url
	}
	return "https://www.zhihu_bianji.com" + url
}

