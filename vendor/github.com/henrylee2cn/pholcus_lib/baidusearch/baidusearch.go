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
	"math"
	// "time"
)

func init() {
	BaiduSearch.Register()
}

var BaiduSearch = &Spider{
	Name:        "百度搜索",
	Description: "百度搜索结果 [www.baidu.com]",
	// Pausetime: 300,
	Keyin:        KEYIN,
	Limit:        LIMIT,
	EnableCookie: false,
	// 禁止输出默认字段 Url/ParentUrl/DownloadTime
	NotDefaultField: true,
	// 命名空间相对于数据库名，不依赖具体数据内容，可选
	Namespace: nil,
	// 子命名空间相对于表名，可依赖具体数据内容，可选
	SubNamespace: nil,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.Aid(map[string]interface{}{"loop": [2]int{0, 1}, "Rule": "生成请求"}, "生成请求")
		},

		Trunk: map[string]*Rule{

			"生成请求": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					var duplicatable bool
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						if loop[0] == 0 {
							duplicatable = true
						} else {
							duplicatable = false
						}
						ctx.AddQueue(&request.Request{
							Url:        "http://www.baidu.com/s?ie=utf-8&nojc=1&wd=" + ctx.GetKeyin() + "&rn=50&pn=" + strconv.Itoa(50*loop[0]),
							Rule:       aid["Rule"].(string),
							Reloadable: duplicatable,
						})
					}
					return nil
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					total1 := query.Find(".nums").Text()
					re, _ := regexp.Compile(`[\D]*`)
					total1 = re.ReplaceAllString(total1, "")
					total2, _ := strconv.Atoi(total1)
					total := int(math.Ceil(float64(total2) / 50))
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
					"内容",
					"不完整URL",
					"百度跳转",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					query.Find("#content_left .c-container").Each(func(i int, s *goquery.Selection) {

						title := s.Find(".t").Text()
						content := s.Find(".c-abstract").Text()
						href, _ := s.Find(".t >a").Attr("href")
						tar := s.Find(".g").Text()

						re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
						// title = re.ReplaceAllStringFunc(title, strings.ToLower)
						// content = re.ReplaceAllStringFunc(content, strings.ToLower)

						title = re.ReplaceAllString(title, "")
						content = re.ReplaceAllString(content, "")

						// 结果存入Response中转
						ctx.Output(map[int]interface{}{
							0: strings.Trim(title, " \t\n"),
							1: strings.Trim(content, " \t\n"),
							2: tar,
							3: href,
						})
					})
				},
			},
		},
	},
}
