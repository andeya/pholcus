package rules

// base packages
import (
	// "github.com/andeya/pholcus/common/goquery" // DOM parsing
	"github.com/andeya/pholcus/app/downloader/request"         // required
	spider "github.com/andeya/pholcus/app/spider"              // required
	spidercommon "github.com/andeya/pholcus/app/spider/common" // optional
	"github.com/andeya/pholcus/logs"                           // logging

	// net packages
	// "net/http" // set http.Header
	// "net/url"

	// encoding packages
	// "encoding/xml"
	"encoding/json"

	// string processing packages
	"regexp"
	"strconv"
	"strings"
	// other packages
	// "fmt"
	// "math"
	// "time"
)

func init() {
	TaobaoSearch.Register()
}

var TaobaoSearch = &spider.Spider{
	Name:        "淘宝天猫搜索",
	Description: "淘宝天猫搜索结果 [s.taobao.com]",
	// Pausetime: 300,
	Keyin:        spider.KEYIN,
	Limit:        spider.LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.Aid(map[string]interface{}{"loop": [2]int{0, 1}, "Rule": "生成请求"}, "生成请求")
		},

		Trunk: map[string]*spider.Rule{

			"生成请求": {
				AidFunc: func(ctx *spider.Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&request.Request{
							URL:  "http://s.taobao.com/search?q=" + ctx.GetKeyin() + "&ie=utf8&cps=yes&app=vproduct&cd=false&v=auction&tab=all&vlist=1&bcoffset=1&s=" + strconv.Itoa(loop[0]*44),
							Rule: aid["Rule"].(string),
						})
					}
					return nil
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					src := query.Find("script").Text()
					if strings.Contains(src, "抱歉！没有找到与") {
						logs.Log().Critical(" ********************** 淘宝关键词 [%v] 的搜索结果不存在！ ********************** ", ctx.GetKeyin())
						return
					}

					re := regexp.MustCompile(`(?U)"totalCount":[\d]+}`)
					total := re.FindString(src)
					re = regexp.MustCompile(`[\d]+`)
					total = re.FindString(total)
					totalCount, _ := strconv.Atoi(total)

					maxPage := (totalCount - 4) / 44
					if (totalCount-4)%44 > 0 {
						maxPage++
					}

					if ctx.GetLimit() > maxPage || ctx.GetLimit() == 0 {
						ctx.SetLimit(maxPage)
					} else if ctx.GetLimit() == 0 {
						logs.Log().Critical("[消息提示：| 任务：%v | KEYIN：%v | 规则：%v] 没有抓取到任何数据！!!\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						return
					}

					logs.Log().Critical(" ********************** 淘宝关键词 [%v] 的搜索结果共有 %v 页，计划抓取 %v 页 **********************", ctx.GetKeyin(), maxPage, ctx.GetLimit())
					// call helper function under specified rule
					ctx.Aid(map[string]interface{}{"loop": [2]int{1, ctx.GetLimit()}, "Rule": "搜索结果"})
					// parse response with specified rule
					ctx.Parse("搜索结果")
				},
			},

			"搜索结果": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					src := query.Find("script").Text()

					re := regexp.MustCompile(`"auctions".*,"recommendAuctions"`)
					src = re.FindString(src)

					re = regexp.MustCompile(`"auctions":`)
					src = re.ReplaceAllString(src, "")

					re = regexp.MustCompile(`,"recommendAuctions"`)
					src = re.ReplaceAllString(src, "")

					re = regexp.MustCompile("\\<[\\S\\s]+?\\>")
					// src = re.ReplaceAllStringFunc(src, strings.ToLower)
					src = re.ReplaceAllString(src, " ")

					src = strings.Trim(src, " \t\n")

					infos := []map[string]interface{}{}

					err := json.Unmarshal([]byte(src), &infos)

					if err != nil {
						logs.Log().Error("error is %v\n", err)
						return
					} else {
						for _, info := range infos {
							ctx.AddQueue(&request.Request{
								URL:  "http:" + info["detail_url"].(string),
								Rule: "商品详情",
								Temp: ctx.CreateItem(map[int]interface{}{
									0: info["raw_title"],
									1: info["view_price"],
									2: info["view_sales"],
									3: info["nick"],
									4: info["item_loc"],
								}, "商品详情"),
								Priority: 1,
							})
						}
					}
				},
			},
			"商品详情": {
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"标题",
					"价格",
					"销量",
					"店铺",
					"发货地",
				},
				ParseFunc: func(ctx *spider.Context) {
					r := ctx.CopyTemps()

					re := regexp.MustCompile(`"newProGroup":.*,"progressiveSupport"`)
					d := re.FindString(ctx.GetText())

					if d == "" {
						h, _ := ctx.GetDom().Find(".attributes-list").Html()
						d = spidercommon.UnicodeToUTF8(h)
						d = strings.ReplaceAll(d, "&nbsp;", " ")
						d = spidercommon.CleanHtml(d, 5)
						d = strings.ReplaceAll(d, "产品参数：\n", "")

						for _, v := range strings.Split(d, "\n") {
							if v == "" {
								continue
							}
							feild := strings.Split(v, ":")
							// trim English spaces
							// feild[0] = strings.Trim(feild[0], " ")
							// feild[1] = strings.Trim(feild[1], " ")
							// trim Chinese spaces
							feild[0] = strings.Trim(feild[0], " ")
							feild[1] = strings.Trim(feild[1], " ")

							if feild[0] == "" || feild[1] == "" {
								continue
							}

							ctx.UpsertItemField(feild[0])
							r[feild[0]] = feild[1]
						}

					} else {
						d = strings.ReplaceAll(d, `"newProGroup":`, "")
						d = strings.ReplaceAll(d, `,"progressiveSupport"`, "")

						infos := []map[string]interface{}{}

						err := json.Unmarshal([]byte(d), &infos)

						if err != nil {
							logs.Log().Error("error is %v\n", err)
							return
						} else {
							for _, info := range infos {
								for _, attr := range info["attrs"].([]interface{}) {
									a := attr.(map[string]interface{})
									ctx.UpsertItemField(a["name"].(string))
									r[a["name"].(string)] = a["value"]
								}
							}
						}
					}

					ctx.Output(r)
				},
			},
		},
	},
}
