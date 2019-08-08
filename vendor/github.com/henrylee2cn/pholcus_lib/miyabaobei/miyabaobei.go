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
	"strconv"
	"strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Miyabaobei.Register()
}

var Miyabaobei = &Spider{
	Name:        "蜜芽宝贝",
	Description: "蜜芽宝贝商品数据 [Auto Page] [www.miyabaobei.com]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{Url: "http://www.miyabaobei.com/", Rule: "获取版块URL"})
		},

		Trunk: map[string]*Rule{

			"获取版块URL": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					lis := query.Find(".ccon")
					lis.Each(func(i int, s *goquery.Selection) {
						s.Find("a").Each(func(n int, ss *goquery.Selection) {
							if url, ok := ss.Attr("href"); ok {
								if !strings.Contains(url, "http://www.miyabaobei.com") {
									url = "http://www.miyabaobei.com" + url
								}
								ctx.Aid(map[string]interface{}{
									"loop":    [2]int{0, 1},
									"urlBase": url,
									"req": map[string]interface{}{
										"Rule": "生成请求",
										"Temp": map[string]interface{}{"baseUrl": url},
									},
								}, "生成请求")
							}
						})
					})
				},
			},

			"生成请求": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					req := aid["req"].(*request.Request)
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						req.Url = aid["urlBase"].(string) + "&per_page=" + strconv.Itoa(loop[0]*40)
						ctx.AddQueue(req)
					}
					return nil
				},
				ParseFunc: func(ctx *Context) {
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

					// 调用指定规则下辅助函数
					ctx.Aid(map[string]interface{}{
						"loop":     [2]int{1, total},
						"ruleBase": ctx.GetTemp("baseUrl", "").(string),
						"rep": map[string]interface{}{
							"Rule": "商品列表",
						},
					})
					// 用指定规则解析响应流
					ctx.Parse("商品列表")
				},
			},

			"商品列表": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"标题",
					"价格",
					"类别",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					//获取品类
					goodsType := query.Find(".crumbs").Text()
					re, _ := regexp.Compile("\\s")
					goodsType = re.ReplaceAllString(goodsType, "")
					re, _ = regexp.Compile("蜜芽宝贝>")
					goodsType = re.ReplaceAllString(goodsType, "")
					query.Find(".bmfo").Each(func(i int, s *goquery.Selection) {
						// 获取标题
						title, _ := s.Find("p a").First().Attr("title")

						// 获取价格
						price := s.Find(".f20").Text()

						// 结果存入Response中转
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
