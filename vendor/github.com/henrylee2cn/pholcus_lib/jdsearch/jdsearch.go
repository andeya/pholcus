package pholcus_lib

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	"github.com/henrylee2cn/pholcus/logs"                   //信息输出
	// . "github.com/henrylee2cn/pholcus/app/spider/common"          //选用

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
	JDSearch.Register()
}

var JDSearch = &Spider{
	Name:        "京东搜索",
	Description: "京东搜索结果 [search.jd.com]",
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
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(
							&request.Request{
								Url:  "http://search.jd.com/Search?keyin=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&click=&psort=&page=" + strconv.Itoa(2*loop[0]+1),
								Rule: aid["Rule"].(string),
							},
						)
						ctx.AddQueue(
							&request.Request{
								Url:  "http://search.jd.com/Search?keyin=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&click=&psort=&page=" + strconv.Itoa(2*loop[0]+2),
								Rule: aid["Rule"].(string),
							},
						)
					}
					return nil
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					total1 := query.Find("#top_pagi span.text").Text()

					re, _ := regexp.Compile(`[\d]+$`)
					total1 = re.FindString(total1)
					total, _ := strconv.Atoi(total1)

					if total > ctx.GetLimit() {
						total = ctx.GetLimit()
					} else if total == 0 {
						logs.Log.Critical("[消息提示：| 任务：%v | KEYIN：%v | 规则：%v] 没有抓取到任何数据！!!\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
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
					"标题",
					"价格",
					"评论数",
					"星级",
					"链接",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					query.Find("#plist .list-h:nth-child(1) > li").Each(func(i int, s *goquery.Selection) {
						// 获取标题
						a := s.Find(".p-name a")
						title := a.Text()

						re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
						// title = re.ReplaceAllStringFunc(title, strings.ToLower)
						title = re.ReplaceAllString(title, " ")
						title = strings.Trim(title, " \t\n")

						// 获取价格
						price, _ := s.Find("strong[data-price]").First().Attr("data-price")

						// 获取评论数
						e := s.Find(".extra").First()
						discuss := e.Find("a").First().Text()
						re, _ = regexp.Compile(`[\d]+`)
						discuss = re.FindString(discuss)

						// 获取星级
						level, _ := e.Find(".star span[id]").First().Attr("class")
						level = re.FindString(level)

						// 获取URL
						url, _ := a.Attr("href")

						// 结果存入Response中转
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
