package pholcus_lib

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	// "github.com/henrylee2cn/pholcus/logs"               //信息输出
	. "github.com/henrylee2cn/pholcus/app/spider" //必需
	// . "github.com/henrylee2cn/pholcus/app/spider/common"          //选用

	// net包
	// "net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	// "regexp"
	"strconv"
	"strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	GanjiGongsi.Register()
}

var GanjiGongsi = &Spider{
	Name:        "经典示例-赶集网企业名录",
	Description: "**典型规则示例，具有文本与文件两种输出行为**",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:  "http://sz.ganji.com/gongsi/o1",
				Rule: "请求列表",
				Temp: map[string]interface{}{"p": 1},
			})
		},

		Trunk: map[string]*Rule{

			"请求列表": {
				ParseFunc: func(ctx *Context) {
					var curr = ctx.GetTemp("p", int(0)).(int)
					if ctx.GetDom().Find(".linkOn span").Text() != strconv.Itoa(curr) {
						return
					}
					ctx.AddQueue(&request.Request{
						Url:         "http://sz.ganji.com/gongsi/o" + strconv.Itoa(curr+1),
						Rule:        "请求列表",
						Temp:        map[string]interface{}{"p": curr + 1},
						ConnTimeout: -1,
					})

					// 用指定规则解析响应流
					ctx.Parse("获取列表")
				},
			},

			"获取列表": {
				ParseFunc: func(ctx *Context) {
					ctx.GetDom().
						Find(".com-list-2 table a").
						Each(func(i int, s *goquery.Selection) {
							url, _ := s.Attr("href")
							ctx.AddQueue(&request.Request{
								Url:         url,
								Rule:        "输出结果",
								ConnTimeout: -1,
							})
						})
				},
			},

			"输出结果": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"公司",
					"联系人",
					"地址",
					"简介",
					"行业",
					"类型",
					"规模",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					var 公司, 规模, 行业, 类型, 联系人, 地址 string

					query.Find(".c-introduce li").Each(func(i int, s *goquery.Selection) {
						em := s.Find("em").Text()
						t := strings.Split(s.Text(), `   `)[0]
						t = strings.Replace(t, em, "", -1)
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
							if img, ok := s.Find("img").Attr("src"); ok {
								ctx.AddQueue(&request.Request{
									Url:         "http://www.ganji.com" + img,
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

					// 结果输出方式一（推荐）
					ctx.Output(map[int]interface{}{
						0: 公司,
						1: 联系人,
						2: 地址,
						3: 简介,
						4: 行业,
						5: 类型,
						6: 规模,
					})

					// 结果输出方式二
					// var item map[string]interface{} = ctx.CreatItem(map[int]interface{}{
					// 	0: 公司,
					// 	1: 联系人,
					// 	2: 地址,
					// 	3: 简介,
					// 	4: 行业,
					// 	5: 类型,
					// 	6: 规模,
					// })
					// ctx.Output(item)

					// 结果输出方式三（不推荐）
					// ctx.Output(map[string]interface{}{
					// 	ctx.GetItemField(0): 公司,
					// 	ctx.GetItemField(1): 联系人,
					// 	ctx.GetItemField(2): 地址,
					// 	ctx.GetItemField(3): 简介,
					// 	ctx.GetItemField(4): 行业,
					// 	ctx.GetItemField(5): 类型,
					// 	ctx.GetItemField(6): 规模,
					// })
				},
			},

			"联系方式": {
				ParseFunc: func(ctx *Context) {
					// 文件输出方式一（推荐）
					ctx.FileOutput(ctx.GetTemp("n", "").(string))

					// 文件输出方式二
					// ctx.AddFile(ctx.GetTemp("n").(string))
				},
			},
		},
	},
}
