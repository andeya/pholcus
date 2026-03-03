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
	// "strconv"
	// "strings"
	// other packages
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Shunfenghaitao.Register()
}

// Imported maternal and infant products section - formula, diapers, baby food, nutrition, care, daily use - Shunfeng Haitao
var Shunfenghaitao = &spider.Spider{
	Name:        "顺丰海淘",
	Description: "顺丰海淘商品数据 [Auto Page] [www.sfht.com]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{URL: "http://www.sfht.com", Rule: "获取版块URL"})
		},

		Trunk: map[string]*spider.Rule{

			"获取版块URL": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					lis := query.Find(".nav-c1").First().Find("li a")

					lis.Each(func(i int, s *goquery.Selection) {
						if i == 0 {
							return
						}
						if url := s.Attr("href"); url.IsSome() {
							ctx.AddQueue(&request.Request{URL: url.Unwrap(), Rule: "商品列表", Temp: map[string]interface{}{"goodsType": s.Text()}})
						}
					})
				},
			},

			"商品列表": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					query.Find(".cms-src-item").Each(func(i int, s *goquery.Selection) {
						if url := s.Find("a").Attr("href"); url.IsSome() {
							ctx.AddQueue(&request.Request{
								URL:  url.Unwrap(),
								Rule: "商品详情",
								Temp: map[string]interface{}{"goodsType": ctx.GetTemp("goodsType", "").(string)},
							})
						}
					})
				},
			},

			"商品详情": {
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"标题",
					"品牌",
					"原产地",
					"货源地",
					"类别",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					// get title
					title := query.Find("#titleInfo h1").Text()

					// get brand
					brand := query.Find(".goods-c2 ul").Eq(0).Find("li").Eq(2).Text()
					re := regexp.MustCompile(`品 牌`)
					brand = re.ReplaceAllString(brand, "")

					// get origin
					from1 := query.Find("#detailattributes li").Eq(0).Text()

					// get supply source
					from2 := query.Find("#detailattributes li").Eq(1).Text()

					// store results in Response
					ctx.Output(map[int]interface{}{
						0: title,
						1: brand,
						2: from1,
						3: from2,
						4: ctx.GetTemp("goodsType", ""),
					})
				},
			},
		},
	},
}
