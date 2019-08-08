package pholcus_lib

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	// . "github.com/henrylee2cn/pholcus/app/spider/common"    //选用
	"github.com/henrylee2cn/pholcus/logs" //信息输出

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
	GoogleSearch.Register()
}

var googleIp = []string{
	"210.242.125.100",
	"210.242.125.96",
	"210.242.125.91",
	"210.242.125.95",
	"64.233.189.163",
	"58.123.102.5",
	"210.242.125.97",
	"210.242.125.115",
	"58.123.102.28",
	"210.242.125.70",
	"220.255.2.153",
}

var GoogleSearch = &Spider{
	Name:        "Google search",
	Description: "Crawls pages from [www.google.com]",
	// Pausetime: 300,
	Keyin:        KEYIN,
	Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			var url string
			var success bool
			logs.Log.Informational("Running google spider，this may take some time...")

			for _, ip := range googleIp {
				// url = "http://" + ip + "/search?q=" + ctx.GetKeyin() + "&newwindow=1&biw=1600&bih=398&start="
				// Beware of redirections, if it doesnt work use google domain:
				// url = "https://google.co.uk/search?q=" + ctx.GetKeyin()
				url = "http://" + ip + "/?gws_rd=ssl#q=" + ctx.GetKeyin()
				logs.Log.Informational("测试 " + ip)
				if _, err := goquery.NewDocument(url); err == nil {
					success = true
					break
				}
			}
			if !success {
				logs.Log.Critical("Could not reach any of the Google mirrors")
				return
			}
			logs.Log.Critical("Starting Google search ...")
			ctx.AddQueue(&request.Request{
				Url:  url,
				Rule: "total_pages",
				Temp: map[string]interface{}{
					"baseUrl": url,
				},
			})
		},

		Trunk: map[string]*Rule{

			"total_pages": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&request.Request{
							Url:  aid["urlBase"].(string) +"&start="+ strconv.Itoa(10 * loop[0]),
							Rule: aid["Rule"].(string),
						})
					}
					return nil
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					txt := query.Find("#resultStats").Text()
					re, _ := regexp.Compile(`,+`)
					txt = re.ReplaceAllString(txt, "")
					re, _ = regexp.Compile(`[\d]+`)
					txt = re.FindString(txt)
					num, _ := strconv.Atoi(txt)
					total := int(math.Ceil(float64(num) / 10))
					if total > ctx.GetLimit() {
						total = ctx.GetLimit()
					} else if total == 0 {
						logs.Log.Critical("[ERROR：| Spider：%v | KEYIN：%v | Rule：%v] Did not fetch any data！!!\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						return
					}
					// 调用指定规则下辅助函数
					ctx.Aid(map[string]interface{}{
						"loop":    [2]int{1, total},
						"urlBase": ctx.GetTemp("baseUrl", ""),
						"Rule":    "search_results",
					})
					// 用指定规则解析响应流
					ctx.Parse("search_results")
				},
			},

			"search_results": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"title",
					"content",
					"href",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					query.Find("#ires .g").Each(func(i int, s *goquery.Selection) {
						t := s.Find(".r > a")
						href, _ := t.Attr("href")
						href = strings.TrimLeft(href, "/url?q=")
						logs.Log.Informational(href)
						title := t.Text()
						content := s.Find(".st").Text()
						ctx.Output(map[int]interface{}{
							0: title,
							1: content,
							2: href,
						})
					})
				},
			},
		},
	},
}
