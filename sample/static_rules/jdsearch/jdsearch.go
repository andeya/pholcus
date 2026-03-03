package rules

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request" // required
	spider "github.com/andeya/pholcus/app/spider"      // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing
	"github.com/andeya/pholcus/logs"                   // logging

	// . "github.com/andeya/pholcus/app/spider/common"          // optional

	// net packages
	// "net/http" // set http.Header
	// "net/url"

	// encoding packages
	// "encoding/xml"
	// "encoding/json"

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
	JDSearch.Register()
}

var JDSearch = &spider.Spider{
	Name:        "京东搜索",
	Description: "京东搜索结果 [search.jd.com]",
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
						ctx.AddQueue(
							&request.Request{
								URL:  "http://search.jd.com/Search?keyin=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&click=&psort=&page=" + strconv.Itoa(2*loop[0]+1),
								Rule: aid["Rule"].(string),
							},
						)
						ctx.AddQueue(
							&request.Request{
								URL:  "http://search.jd.com/Search?keyin=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&click=&psort=&page=" + strconv.Itoa(2*loop[0]+2),
								Rule: aid["Rule"].(string),
							},
						)
					}
					return nil
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					total1 := query.Find("#top_pagi span.text").Text()

					re := regexp.MustCompile(`[\d]+$`)
					total1 = re.FindString(total1)
					total, _ := strconv.Atoi(total1)

					if total > ctx.GetLimit() {
						total = ctx.GetLimit()
					} else if total == 0 {
						logs.Log().Critical("[消息提示：| 任务：%v | KEYIN：%v | 规则：%v] 没有抓取到任何数据！!!\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						return
					}
					// call helper function under specified rule
					ctx.Aid(map[string]interface{}{"loop": [2]int{1, total}, "Rule": "搜索结果"})
					// parse response with specified rule
					ctx.Parse("搜索结果")
				},
			},

			"搜索结果": {
				// note: field semantics and output data must be consistent
				ItemFields: []string{
					"标题",
					"价格",
					"评论数",
					"星级",
					"链接",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					query.Find("#plist .list-h:nth-child(1) > li").Each(func(i int, s *goquery.Selection) {
						// get title
						a := s.Find(".p-name a")
						title := a.Text()

						re := regexp.MustCompile("\\<[\\S\\s]+?\\>")
						// title = re.ReplaceAllStringFunc(title, strings.ToLower)
						title = re.ReplaceAllString(title, " ")
						title = strings.Trim(title, " \t\n")

						// get price
						price := s.Find("strong[data-price]").First().Attr("data-price").UnwrapOr("")

						// get comment count
						e := s.Find(".extra").First()
						discuss := e.Find("a").First().Text()
						re = regexp.MustCompile(`[\d]+`)
						discuss = re.FindString(discuss)

						// get rating level
						level := e.Find(".star span[id]").First().Attr("class").UnwrapOr("")
						level = re.FindString(level)

						// get URL
						url := a.Attr("href").UnwrapOr("")

						// store results in Response
						ctx.Output(map[int]interface{}{
							0: title,
							1: price,
							2: discuss,
							3: level,
							4: url,
						})
					})
				},
			},
		},
	},
}
