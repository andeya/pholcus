package rules

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request" // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing

	// "github.com/andeya/pholcus/logs"               // logging
	spider "github.com/andeya/pholcus/app/spider" // required
	// . "github.com/andeya/pholcus/app/spider/common"          // optional

	// net packages
	// "net/http" // set http.Header
	// "net/url"

	// encoding packages
	// "encoding/xml"
	// "encoding/json"

	// string processing packages
	// "regexp"
	"strconv"
	"strings"
	// other packages
	// "fmt"
	// "math"
	// "time"
)

func init() {
	GanjiGongsi.Register()
}

var GanjiGongsi = &spider.Spider{
	Name:        "经典示例-赶集网企业名录",
	Description: "**典型规则示例，具有文本与文件两种输出行为**",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{
				URL:  "http://sz.ganji.com/gongsi/o1",
				Rule: "请求列表",
				Temp: map[string]interface{}{"p": 1},
			})
		},

		Trunk: map[string]*spider.Rule{

			"请求列表": {
				ParseFunc: func(ctx *spider.Context) {
					var curr = ctx.GetTemp("p", int(0)).(int)
					if ctx.GetDom().Find(".linkOn span").Text() != strconv.Itoa(curr) {
						return
					}
					ctx.AddQueue(&request.Request{
						URL:         "http://sz.ganji.com/gongsi/o" + strconv.Itoa(curr+1),
						Rule:        "请求列表",
						Temp:        map[string]interface{}{"p": curr + 1},
						ConnTimeout: -1,
					})

					// parse response with specified rule
					ctx.Parse("获取列表")
				},
			},

			"获取列表": {
				ParseFunc: func(ctx *spider.Context) {
					ctx.GetDom().
						Find(".com-list-2 table a").
						Each(func(i int, s *goquery.Selection) {
							url := s.Attr("href").UnwrapOr("")
							ctx.AddQueue(&request.Request{
								URL:         url,
								Rule:        "输出结果",
								ConnTimeout: -1,
							})
						})
				},
			},

			"输出结果": {
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"公司",
					"联系人",
					"地址",
					"简介",
					"行业",
					"类型",
					"规模",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					var 公司, 规模, 行业, 类型, 联系人, 地址 string

					query.Find(".c-introduce li").Each(func(i int, s *goquery.Selection) {
						em := s.Find("em").Text()
						t := strings.Split(s.Text(), `   `)[0]
						t = strings.ReplaceAll(t, em, "")
						t = strings.Trim(t, " ")

						switch em {
						case "公司名称：":
							公司 = t

						case "公司规模：":
							规模 = t

						case "公司行业：":
							行业 = t

						case "公司类型：":
							类型 = t

						case "联 系 人：":
							联系人 = t

						case "联系电话：":
							if img := s.Find("img").Attr("src"); img.IsSome() {
								ctx.AddQueue(&request.Request{
									URL:         "http://www.ganji.com" + img.Unwrap(),
									Rule:        "联系方式",
									Temp:        map[string]interface{}{"n": 公司 + "(" + 联系人 + ").png"},
									Priority:    1,
									ConnTimeout: -1,
								})
							}

						case "公司地址：":
							地址 = t
						}
					})

					简介 := query.Find("#company_description").Text()

					// output method 1 (recommended)
					ctx.Output(map[int]interface{}{
						0: 公司,
						1: 联系人,
						2: 地址,
						3: 简介,
						4: 行业,
						5: 类型,
						6: 规模,
					})

					// file output method 2
					// var item map[string]interface{} = ctx.CreateItem(map[int]interface{}{
					// 	0: company,
					// 	1: contact,
					// 	2: address,
					// 	3: introduction,
					// 	4: industry,
					// 	5: type,
					// 	6: scale,
					// })
					// ctx.Output(item)

					// output method 3 (not recommended)
					// ctx.Output(map[string]interface{}{
					// 	ctx.GetItemField(0): company,
					// 	ctx.GetItemField(1): contact,
					// 	ctx.GetItemField(2): address,
					// 	ctx.GetItemField(3): introduction,
					// 	ctx.GetItemField(4): industry,
					// 	ctx.GetItemField(5): type,
					// 	ctx.GetItemField(6): scale,
					// })
				},
			},

			"联系方式": {
				ParseFunc: func(ctx *spider.Context) {
					// file output method 1 (recommended)
					ctx.FileOutput(ctx.GetTemp("n", "").(string))

					// file output method 2
					// ctx.AddFile(ctx.GetTemp("n").(string))
				},
			},
		},
	},
}
