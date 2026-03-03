package rules

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request"         // required
	spider "github.com/andeya/pholcus/app/spider"              // required
	spidercommon "github.com/andeya/pholcus/app/spider/common" // optional
	"github.com/andeya/pholcus/common/goquery"                 // DOM parsing
	"github.com/andeya/pholcus/logs"                           // logging

	// net packages
	"net/http" // set http.Header
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
	AlibabaProduct.Register()
}

var AlibabaProduct = &spider.Spider{
	Name:        "阿里巴巴产品搜索",
	Description: "阿里巴巴产品搜索 [s.1688.com/selloffer/offer_search.htm]",
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
					keyin := spidercommon.EncodeString(ctx.GetKeyin(), "gbk")
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&request.Request{
							URL:    "http://s.1688.com/selloffer/offer_search.htm?enableAsync=false&earseDirect=false&button_click=top&pageSize=60&n=y&offset=3&uniqfield=pic_tag_id&keyins=" + keyin + "&beginPage=" + strconv.Itoa(loop[0]+1),
							Rule:   aid["Rule"].(string),
							Header: http.Header{"Content-Type": []string{"text/html; charset=gbk"}},
						})
					}
					return nil
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					// logs.Log().Debug(ctx.GetText())
					pageTag := query.Find("#sm-pagination div[data-total-page]")
					// redirect
					if len(pageTag.Nodes) == 0 {
						logs.Log().Critical("[消息提示：| 任务：%v | KEYIN：%v | 规则：%v] 由于跳转AJAX问题，目前只能每个子类抓取 1 页……\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						query.Find(".sm-floorhead-typemore a").Each(func(i int, s *goquery.Selection) {
							if href := s.Attr("href"); href.IsSome() {
								ctx.AddQueue(&request.Request{
									URL:    href.Unwrap(),
									Header: http.Header{"Content-Type": []string{"text/html; charset=gbk"}},
									Rule:   "搜索结果",
								})
							}
						})
						return
					}
					total1 := pageTag.First().Attr("data-total-page").UnwrapOr("")
					total1 = strings.Trim(total1, " \t\n")
					total, _ := strconv.Atoi(total1)
					if total > ctx.GetLimit() {
						total = ctx.GetLimit()
					} else if total == 0 {
						logs.Log().Critical("[消息提示：| 任务：%v | KEYIN：%v | 规则：%v] 没有抓取到任何数据！！！\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						return
					}

					// call helper function under specified rule
					ctx.Aid(map[string]interface{}{"loop": [2]int{1, total}, "Rule": "搜索结果"})
					// parse response with specified rule
					ctx.Parse("搜索结果")
				},
			},

			"搜索结果": {
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"公司",
					"标题",
					"价格",
					"销量",
					"星级",
					"地址",
					"链接",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					query.Find("#sm-offer-list > li").Each(func(i int, s *goquery.Selection) {

						// get company
						company := s.Find("a.sm-offer-companyName").First().Attr("title").UnwrapOr("")

						// get title
						t := s.Find(".sm-offer-title > a:nth-child(1)")
						title := t.Attr("title").UnwrapOr("")

						// get URL
						url := t.Attr("href").UnwrapOr("")

						// get price
						price := s.Find(".sm-offer-priceNum").First().Text()

						// get sales volume
						sales := s.Find("span.sm-offer-trade > em").First().Text()

						// get address
						address := s.Find(".sm-offer-location").First().Attr("title").UnwrapOr("")

						// get credit level
						level := s.Find("span.sm-offer-companyTag > a.sw-ui-flaticon-cxt16x16").First().Text()

						// store results in Response
						ctx.Output(map[int]interface{}{
							0: company,
							1: title,
							2: price,
							3: sales,
							4: level,
							5: address,
							6: url,
						})
					})
				},
			},
		},
	},
}
