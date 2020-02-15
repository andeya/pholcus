package pholcus_lib

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	. "github.com/henrylee2cn/pholcus/app/spider/common"    //选用
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	"github.com/henrylee2cn/pholcus/logs"                   //信息输出

	// net包
	"net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	// "regexp"
	"strconv"
	"strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	AlibabaProduct.Register()
}

var AlibabaProduct = &Spider{
	Name:        "阿里巴巴产品搜索",
	Description: "阿里巴巴产品搜索 [s.1688.com/selloffer/offer_search.htm]",
	// Pausetime: 300,
	Keyin:        KEYIN,
	Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.Aid(map[string]interface{}{"loop": [2]int{0, 1}, "Rule": "生成请求"}, "生成请求")
		},

		Trunk: map[string]*Rule{

			"生成请求": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					keyin := EncodeString(ctx.GetKeyin(), "gbk")
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&request.Request{
							Url:    "http://s.1688.com/selloffer/offer_search.htm?enableAsync=false&earseDirect=false&button_click=top&pageSize=60&n=y&offset=3&uniqfield=pic_tag_id&keyins=" + keyin + "&beginPage=" + strconv.Itoa(loop[0]+1),
							Rule:   aid["Rule"].(string),
							Header: http.Header{"Content-Type": []string{"text/html; charset=gbk"}},
						})
					}
					return nil
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					// logs.Log.Debug(ctx.GetText())
					pageTag := query.Find("#sm-pagination div[data-total-page]")
					// 跳转
					if len(pageTag.Nodes) == 0 {
						logs.Log.Critical("[消息提示：| 任务：%v | KEYIN：%v | 规则：%v] 由于跳转AJAX问题，目前只能每个子类抓取 1 页……\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						query.Find(".sm-floorhead-typemore a").Each(func(i int, s *goquery.Selection) {
							if href, ok := s.Attr("href"); ok {
								ctx.AddQueue(&request.Request{
									Url:    href,
									Header: http.Header{"Content-Type": []string{"text/html; charset=gbk"}},
									Rule:   "搜索结果",
								})
							}
						})
						return
					}
					total1, _ := pageTag.First().Attr("data-total-page")
					total1 = strings.Trim(total1, " \t\n")
					total, _ := strconv.Atoi(total1)
					if total > ctx.GetLimit() {
						total = ctx.GetLimit()
					} else if total == 0 {
						logs.Log.Critical("[消息提示：| 任务：%v | KEYIN：%v | 规则：%v] 没有抓取到任何数据！！！\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						return
					}

					// 调用指定规则下辅助函数
					ctx.Aid(map[string]interface{}{"loop": [2]int{1, total}, "Rule": "搜索结果"})
					// 用指定规则解析响应流
					ctx.Parse("搜索结果")
				},
			},

			"搜索结果": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"公司",
					"标题",
					"价格",
					"销量",
					"星级",
					"地址",
					"链接",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					query.Find("#sm-offer-list > li").Each(func(i int, s *goquery.Selection) {

						// 获取公司
						company, _ := s.Find("a.sm-offer-companyName").First().Attr("title")

						// 获取标题
						t := s.Find(".sm-offer-title > a:nth-child(1)")
						title, _ := t.Attr("title")

						// 获取URL
						url, _ := t.Attr("href")

						// 获取价格
						price := s.Find(".sm-offer-priceNum").First().Text()

						// 获取成交量
						sales := s.Find("span.sm-offer-trade > em").First().Text()

						// 获取地址
						address, _ := s.Find(".sm-offer-location").First().Attr("title")

						// 获取信用年限
						level := s.Find("span.sm-offer-companyTag > a.sw-ui-flaticon-cxt16x16").First().Text()

						// 结果存入Response中转
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
