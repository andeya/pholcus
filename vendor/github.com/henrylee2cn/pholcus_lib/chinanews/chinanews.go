package pholcus_lib

// 基础包
import (
	// "github.com/henrylee2cn/pholcus/common/goquery"                          //DOM解析
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	// . "github.com/henrylee2cn/pholcus/app/spider/common" //选用
	// "github.com/henrylee2cn/pholcus/logs"
	// net包
	// "net/http" //设置http.Header
	// "net/url"
	// 编码包
	// "encoding/xml"
	//"encoding/json"
	// 字符串处理包
	//"regexp"
	// "strconv"
	// "fmt"
	// "math"
	// "time"
	"github.com/henrylee2cn/pholcus/common/goquery"
	"strings"
)

func init() {
	FileTest.Register()
}

var FileTest = &Spider{
	Name:        "中国新闻网",
	Description: "测试 [http://www.chinanews.com/scroll-news/news1.html]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:          "http://www.chinanews.com/scroll-news/news1.html",
				Rule:         "滚动新闻",
			})
		},

		Trunk: map[string]*Rule{

			"滚动新闻": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					//获取分页导航
					navBox := query.Find(".pagebox a")
					navBox.Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Attr("href"); ok {
							ctx.AddQueue(&request.Request{
								Url:  "http://www.chinanews.com" +  url,
								Rule: "新闻列表",

							})
						}

					})

				},
			},

			"新闻列表": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					//获取新闻列表
					newList := query.Find(".content_list li")
					newList.Each(func(i int, s *goquery.Selection) {
						//新闻类型
						newsType := s.Find(".dd_lm a").Text()
						//标题
						newsTitle := s.Find(".dd_bt a").Text()
						//时间
						newsTime := s.Find(".dd_time").Text()
						if url, ok := s.Find(".dd_bt a").Attr("href"); ok {
							ctx.AddQueue(&request.Request{
								Url:  "http://" + url[2:len(url)],
								Rule: "新闻内容",
								Temp: map[string]interface{}{
									"newsType":  newsType,
									"newsTitle": newsTitle,
									"newsTime":  newsTime,
								},
							})
						}

					})

				},
			},

			"新闻内容": {
				ItemFields: []string{
					"类别",
					"来源",
					"标题",
					"内容",
					"时间",
				},

				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					//正文
					content := query.Find(".left_zw").Text()
					//来源
					from := query.Find(".left-t").Text()
					i := strings.LastIndex(from,"来源")
					//来源字符串特殊处理
					if i == -1{
						from = "未知"
					}else{
						from = from[i+9:len(from)]
						from = strings.Replace(from,"参与互动","",1)
						if from=="" {
							from = query.Find(".left-t").Eq(2).Text()
							from = strings.Replace(from,"参与互动","",1)
						}
					}

					//输出格式
					ctx.Output(map[int]interface{}{
						0: ctx.GetTemp("newsType",""),
						1: from,
						2: ctx.GetTemp("newsTitle",""),
						3: content,
						4: ctx.GetTemp("newsTime", ""),
					})
				},
			},

		},
	},
}
