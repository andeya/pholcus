package pholcus_lib

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
	"strconv"
	// "strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Zolslab.Register()
}

var Zolslab = &Spider{
	Name:        "中关村平板",
	Description: "中关村平板数据 [Auto Page] [bbs.zol.com.cn/sjbbs/d544_p]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.Aid(map[string]interface{}{"loop": [2]int{1, 640}, "Rule": "生成请求"}, "生成请求")
		},

		Trunk: map[string]*Rule{

			"生成请求": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&request.Request{
							Url:  "http://bbs.zol.com.cn/padbbs/p" + strconv.Itoa(loop[0]) + ".html#c",
							Rule: aid["Rule"].(string),
						})
					}
					return nil
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					ss := query.Find("tbody").Find("tr[id]")
					ss.Each(func(i int, goq *goquery.Selection) {
						ctx.SetTemp("html", goq)
						ctx.Parse("获取结果")

					})
				},
			},

			"获取结果": {
				//注意：有无字段语义和是否输出数据必须保持一致
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
				ParseFunc: func(ctx *Context) {
					var selectObj = ctx.GetTemp("html", &goquery.Selection{}).(*goquery.Selection)
					//url
					outUrls := selectObj.Find("td").Eq(1)
					outUrl, _ := outUrls.Attr("data-url")
					outUrl = "http://bbs.zol.com.cn/" + outUrl

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

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: outType,
						1: outUrl,
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
