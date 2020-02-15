package pholcus_lib

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
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
	"fmt"
	"strings"
)

func init() {
	Avatar.Register()
}

var Avatar = &Spider{

	Name:        "QQ头像和昵称抓去和下载",
	Description: "QQ头像和昵称抓去和下载",
	// Pausetime: 300,
	Keyin:           KEYIN,
	Limit:           LIMIT,
	EnableCookie:    false,
	NotDefaultField: true,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.Aid(map[string]interface{}{"loop": [2]int{0, ctx.GetLimit()}, "Rule": "生成请求"}, "生成请求")
		},

		Trunk: map[string]*Rule{
			"生成请求": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					var url string
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						if loop[0] == 0 {
							url = "http://www.woyaogexing.com/touxiang/index.html"
							loop[0]++
						} else {
							url = "http://www.woyaogexing.com/touxiang/index_" + strconv.Itoa(loop[0]+1) + ".html"
						}
						ctx.AddQueue(&request.Request{
							Url:    url,
							Rule:   aid["Rule"].(string),
							Header: http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
						})
					}
					return nil
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					// logs.Log.Debug(ctx.GetText())
					pageTag := query.Find("div.pageNum.wp div.page a:last-child")
					// 跳转
					if len(pageTag.Nodes) == 0 {
						logs.Log.Critical("[消息提示：| 任务：%v | KEYIN：%v | 规则：%v] \n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						query.Find(".sm-floorhead-typemore a").Each(func(i int, s *goquery.Selection) {
							if href, ok := s.Attr("href"); ok {
								ctx.AddQueue(&request.Request{
									Url:    href,
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
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					query.Find(".txList").Each(func(i int, selection *goquery.Selection) {
						src, _ := selection.Find("a.img>img").First().Attr("src")
						name := selection.Find("p>a").Text()
						fmt.Printf("nickname:%s \t url: %s\n", name, src)
						ctx.AddQueue(&request.Request{
							Url:          src,
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
				ParseFunc: func(ctx *Context) {
					ctx.FileOutput()
				},
			},
		},
	},
}

