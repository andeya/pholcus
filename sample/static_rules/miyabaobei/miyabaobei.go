package rules

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request" // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing

	// "github.com/andeya/pholcus/logs"              // logging
	spider "github.com/andeya/pholcus/app/spider" // required
	// . "github.com/andeya/pholcus/app/spider/common" // optional

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
	Miyabaobei.Register()
}

var Miyabaobei = &spider.Spider{
	Name:        "蜜芽宝贝",
	Description: "蜜芽宝贝商品数据 [Auto Page] [www.miyabaobei.com]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{URL: "http://www.miyabaobei.com/", Rule: "获取版块URL"})
		},

		Trunk: map[string]*spider.Rule{

			"获取版块URL": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					lis := query.Find(".ccon")
					lis.Each(func(i int, s *goquery.Selection) {
						s.Find("a").Each(func(n int, ss *goquery.Selection) {
							if url := ss.Attr("href"); url.IsSome() {
								u := url.Unwrap()
								if !strings.Contains(u, "http://www.miyabaobei.com") {
									u = "http://www.miyabaobei.com" + u
								}
								ctx.Aid(map[string]interface{}{
									"loop":    [2]int{0, 1},
									"urlBase": u,
									"req": map[string]interface{}{
										"Rule": "生成请求",
										"Temp": map[string]interface{}{"baseUrl": u},
									},
								}, "生成请求")
							}
						})
					})
				},
			},

			"生成请求": {
				AidFunc: func(ctx *spider.Context, aid map[string]interface{}) interface{} {
					req := aid["req"].(*request.Request)
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						req.URL = aid["urlBase"].(string) + "&per_page=" + strconv.Itoa(loop[0]*40)
						ctx.AddQueue(req)
					}
					return nil
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					totalPage := "1"

					urls := query.Find(".Lpage.page p a")

					if urls.Length() != 0 {
						if urls.Last().Text() == ">" {
							totalPage = urls.Eq(urls.Length() - 2).Text()
						} else {
							totalPage = urls.Last().Text()
						}
					}
					total, _ := strconv.Atoi(totalPage)

					// call helper function under specified rule
					ctx.Aid(map[string]interface{}{
						"loop":     [2]int{1, total},
						"ruleBase": ctx.GetTemp("baseUrl", "").(string),
						"rep": map[string]interface{}{
							"Rule": "商品列表",
						},
					})
					// parse response with specified rule
					ctx.Parse("商品列表")
				},
			},

			"商品列表": {
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"标题",
					"价格",
					"类别",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					// get product category
					goodsType := query.Find(".crumbs").Text()
					re := regexp.MustCompile("\\s")
					goodsType = re.ReplaceAllString(goodsType, "")
					re = regexp.MustCompile("蜜芽宝贝>")
					goodsType = re.ReplaceAllString(goodsType, "")
					query.Find(".bmfo").Each(func(i int, s *goquery.Selection) {
						// get title
						title := s.Find("p a").First().Attr("title").UnwrapOr("")

						// get price
						price := s.Find(".f20").Text()

						// store results in Response
						ctx.Output(map[int]interface{}{
							0: title,
							1: price,
							2: goodsType,
						})
					})
				},
			},
		},
	},
}
