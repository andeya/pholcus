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
	"regexp"
	// "strconv"
	// "strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Shunfenghaitao.Register()
}

// 进口母婴专区，买进口奶粉、尿裤尿布、辅食、营养、洗护、日用、母婴用品  - 顺丰海淘
var Shunfenghaitao = &Spider{
	Name:        "顺丰海淘",
	Description: "顺丰海淘商品数据 [Auto Page] [www.sfht.com]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{Url: "http://www.sfht.com", Rule: "获取版块URL"})
		},

		Trunk: map[string]*Rule{

			"获取版块URL": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					lis := query.Find(".nav-c1").First().Find("li a")

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

					query.Find(".cms-src-item").Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Find("a").Attr("href"); ok {
							ctx.AddQueue(&request.Request{
								Url:  url,
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
					"品牌",
					"原产地",
					"货源地",
					"类别",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					// 获取标题
					title := query.Find("#titleInfo h1").Text()

					// 获取品牌
					brand := query.Find(".goods-c2 ul").Eq(0).Find("li").Eq(2).Text()
					re, _ := regexp.Compile(`品 牌`)
					brand = re.ReplaceAllString(brand, "")

					// 获取原产地
					from1 := query.Find("#detailattributes li").Eq(0).Text()

					// 获取货源地
					from2 := query.Find("#detailattributes li").Eq(1).Text()

					// 结果存入Response中转
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
