package pholcus_lib

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	// "github.com/henrylee2cn/pholcus/logs"              //信息输出
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
	// "strconv"
	// "strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Kaola.Register()
}

// 考拉海淘,海外直采,7天无理由退货,售后无忧!考拉网放心的海淘网站!
var Kaola = &Spider{
	Name:        "考拉海淘",
	Description: "考拉海淘商品数据 [Auto Page] [www.kaola.com]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{Url: "http://www.kaola.com", Rule: "获取版块URL"})
		},

		Trunk: map[string]*Rule{

			"获取版块URL": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					lis := query.Find("#funcTab li a")
					lis.Each(func(i int, s *goquery.Selection) {
						if i == 0 {
							return
						}
						if url, ok := s.Attr("href"); ok {
							ctx.AddQueue(&request.Request{Url: url, Rule: "商品列表", Temp: map[string]interface{}{"goodsType": s.Text()}})
						}
					})
				},
			},

			"商品列表": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					query.Find(".proinfo").Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Find("a").Attr("href"); ok {
							ctx.AddQueue(&request.Request{
								Url:  "http://www.kaola.com" + url,
								Rule: "商品详情",
								Temp: map[string]interface{}{"goodsType": ctx.GetTemp("goodsType", "").(string)},
							})
						}
					})
				},
			},

			"商品详情": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"标题",
					"价格",
					"品牌",
					"采购地",
					"评论数",
					"类别",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					// 获取标题
					title := query.Find(".product-title").Text()

					// 获取价格
					price := query.Find("#js_currentPrice span").Text()

					// 获取品牌
					brand := query.Find(".goods_parameter li").Eq(0).Text()

					// 获取采购地
					from := query.Find(".goods_parameter li").Eq(1).Text()

					// 获取评论数
					discussNum := query.Find("#commentCounts").Text()

					// 结果存入Response中转
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
