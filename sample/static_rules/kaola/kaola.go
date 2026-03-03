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
	// "regexp"
	// "strconv"
	// "strings"
	// other packages
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Kaola.Register()
}

// Kaola Haitao - overseas direct purchase, 7-day no-reason return, worry-free after-sales
var Kaola = &spider.Spider{
	Name:        "考拉海淘",
	Description: "考拉海淘商品数据 [Auto Page] [www.kaola.com]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{URL: "http://www.kaola.com", Rule: "获取版块URL"})
		},

		Trunk: map[string]*spider.Rule{

			"获取版块URL": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					lis := query.Find("#funcTab li a")
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
					query.Find(".proinfo").Each(func(i int, s *goquery.Selection) {
						if url := s.Find("a").Attr("href"); url.IsSome() {
							ctx.AddQueue(&request.Request{
								URL:  "http://www.kaola.com" + url.Unwrap(),
								Rule: "商品详情",
								Temp: map[string]interface{}{"goodsType": ctx.GetTemp("goodsType", "").(string)},
							})
						}
					})
				},
			},

			"商品详情": {
				// note: field semantics and output data must be consistent
				ItemFields: []string{
					"标题",
					"价格",
					"品牌",
					"采购地",
					"评论数",
					"类别",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					// get title
					title := query.Find(".product-title").Text()

					// get price
					price := query.Find("#js_currentPrice span").Text()

					// get brand
					brand := query.Find(".goods_parameter li").Eq(0).Text()

					// get purchase origin
					from := query.Find(".goods_parameter li").Eq(1).Text()

					// get comment count
					discussNum := query.Find("#commentCounts").Text()

					// store results in Response
					ctx.Output(map[int]interface{}{
						0: title,
						1: price,
						2: brand,
						3: from,
						4: discussNum,
						5: ctx.GetTemp("goodsType", ""),
					})
				},
			},
		},
	},
}
