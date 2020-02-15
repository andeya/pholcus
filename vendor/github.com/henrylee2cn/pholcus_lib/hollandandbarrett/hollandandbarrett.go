package pholcus_lib

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	"github.com/henrylee2cn/pholcus/logs"                   //信息输出
	// . "github.com/henrylee2cn/pholcus/app/spider/common"          //选用

	// net包
	// "net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	"encoding/json"

	// 字符串处理包
	"regexp"
	// "strconv"
	"strings"

	// 其他包
	"fmt"
	// "math"
	// "time"
)

func init() {
	Hollandandbarrett.Register()
}

var Hollandandbarrett = &Spider{
	Name:        "Hollandandbarrett",
	Description: "Hollandand&Barrett商品数据 [Auto Page] [www.Hollandandbarrett.com]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:  "http://www.hollandandbarrett.com/",
				Rule: "获取版块URL",
			},
			)
		},

		Trunk: map[string]*Rule{

			"获取版块URL": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					lis := query.Find(".footer-links nav.l-one-half a")

					lis.Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Attr("href"); ok {
							tit, _ := s.Attr("title")
							ctx.AddQueue(&request.Request{
								Url:  "http://www.hollandandbarrett.com" + url + "?showAll=1&pageHa=1&es=true&vm=grid&imd=true&format=json&single=true",
								Rule: "获取总数",
								Temp: map[string]interface{}{
									"type":    tit,
									"baseUrl": url,
								},
							},
							)
						}
					})
				},
			},

			"获取总数": {
				ParseFunc: func(ctx *Context) {

					query := ctx.GetDom()

					re, _ := regexp.Compile(`(?U)"totalNumRecs":[\d]+,`)
					total := re.FindString(query.Text())
					re, _ = regexp.Compile(`[\d]+`)
					total = re.FindString(total)
					total = strings.Trim(total, " \t\n")

					if total == "0" {
						logs.Log.Critical("[消息提示：| 任务：%v | 关键词：%v | 规则：%v] 没有抓取到任何数据！!!\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
					} else {

						ctx.AddQueue(&request.Request{
							Url:  "http://www.hollandandbarrett.com" + ctx.GetTemp("baseUrl", "").(string) + "?showAll=" + total + "&pageHa=1&es=true&vm=grid&imd=true&format=json&single=true",
							Rule: "商品详情",
							Temp: map[string]interface{}{
								"type": ctx.GetTemp("type", "").(string),
							},
						},
						)

					}
				},
			},

			"商品详情": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"标题",
					"原价",
					"折后价",
					"打折",
					"星级",
					"分类",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					src := query.Text()

					infos := map[string]interface{}{}

					err := json.Unmarshal([]byte(src), &infos)

					if err != nil {
						logs.Log.Error("error is %v\n", err)
						return
					} else {
						for _, info1 := range infos["contents"].([]interface{})[0].(map[string]interface{})["mainContent"].([]interface{})[0].(map[string]interface{})["records"].([]interface{}) {

							info2 := info1.(map[string]interface{})["records"].([]interface{})[0].(map[string]interface{})["attributes"].(map[string]interface{})

							var n, price1, price2, prm, level string

							if info2["Name"] == nil {
								n = ""
							} else {
								n = fmt.Sprint(info2["Name"])
								n = strings.TrimRight(n, "]")
								n = strings.TrimLeft(n, "[")
							}

							if info2["lp"] == nil {
								price1 = ""
							} else {
								price1 = fmt.Sprint(info2["lp"])
								price1 = strings.TrimRight(price1, "]")
								price1 = strings.TrimLeft(price1, "[")
							}

							if info2["sp"] == nil {
								price2 = ""
							} else {
								price2 = fmt.Sprint(info2["sp"])
								price2 = strings.TrimRight(price2, "]")
								price2 = strings.TrimLeft(price2, "[")
							}

							if info2["prm"] == nil {
								prm = ""
							} else {
								prm = fmt.Sprint(info2["prm"])
								prm = strings.TrimRight(prm, "]")
								prm = strings.TrimLeft(prm, "[")
							}

							if info2["ratingCount"] == nil {
								level = "0"
							} else {
								level = fmt.Sprint(info2["ratingCount"])
								level = strings.TrimRight(level, "]")
								level = strings.TrimLeft(level, "[")
							}

							// 结果存入Response中转
							ctx.Output(map[int]interface{}{
								0: n,
								1: price1,
								2: price2,
								3: prm,
								4: level,
								5: ctx.GetTemp("type", ""),
							})
						}
					}
				},
			},
		},
	},
}
