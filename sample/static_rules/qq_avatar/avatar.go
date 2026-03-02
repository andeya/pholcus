package rules

// 基础包
import (
	"github.com/andeya/pholcus/app/downloader/request" //必需
	spider "github.com/andeya/pholcus/app/spider"      //必需
	"github.com/andeya/pholcus/common/goquery"         //DOM解析
	"github.com/andeya/pholcus/logs"                   //信息输出

	// net包
	"net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	// "regexp"
	"fmt"
	"strconv"
	"strings"
)

func init() {
	Avatar.Register()
}

var Avatar = &spider.Spider{

	Name:        "QQ头像和昵称抓取和下载",
	Description: "QQ头像和昵称抓取和下载",
	// Pausetime: 300,
	Keyin:           spider.KEYIN,
	Limit:           spider.LIMIT,
	EnableCookie:    false,
	NotDefaultField: true,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.Aid(map[string]interface{}{"loop": [2]int{0, ctx.GetLimit()}, "Rule": "生成请求"}, "生成请求")
		},

		Trunk: map[string]*spider.Rule{
			"生成请求": {
				AidFunc: func(ctx *spider.Context, aid map[string]interface{}) interface{} {
					var url string
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						if loop[0] == 0 {
							url = "http://www.woyaogexing.com/touxiang/index.html"
							loop[0]++
						} else {
							url = "http://www.woyaogexing.com/touxiang/index_" + strconv.Itoa(loop[0]+1) + ".html"
						}
						ctx.AddQueue(&request.Request{
							URL:    url,
							Rule:   aid["Rule"].(string),
							Header: http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
						})
					}
					return nil
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					// logs.Log().Debug(ctx.GetText())
					pageTag := query.Find("div.pageNum.wp div.page a:last-child")
					// 跳转
					if len(pageTag.Nodes) == 0 {
						logs.Log().Critical("[消息提示：| 任务：%v | KEYIN：%v | 规则：%v] \n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						query.Find(".sm-floorhead-typemore a").Each(func(i int, s *goquery.Selection) {
							if href := s.Attr("href"); href.IsSome() {
								ctx.AddQueue(&request.Request{
									URL:    href.Unwrap(),
									Header: http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
									Rule:   "搜索结果",
								})
							}
						})
						return
					}
					// 用指定规则解析响应流
					ctx.Parse("搜索结果")
				},
			},
			"搜索结果": {
				ItemFields: []string{
					"avatar",
					"nickname",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					query.Find(".txList").Each(func(i int, selection *goquery.Selection) {
						src := selection.Find("a.img>img").First().Attr("src").UnwrapOr("")
						name := selection.Find("p>a").Text()
						fmt.Printf("nickname:%s \t url: %s\n", name, src)
						ctx.AddQueue(&request.Request{
							URL:          src,
							Rule:         "下载文件",
							ConnTimeout:  -1,
							DownloaderID: 0,
						})
						str := strings.Split(src, "/")
						ctx.Output(map[int]interface{}{
							0: str[len(str)-1],
							1: name,
						})
					})
				},
			},
			"下载文件": {
				ParseFunc: func(ctx *spider.Context) {
					ctx.FileOutput()
				},
			},
		},
	},
}
