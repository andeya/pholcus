package pholcus_lib

// 基础包
import (
	"log"

	// "github.com/henrylee2cn/pholcus/common/goquery"                        //DOM解析
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	// "github.com/henrylee2cn/pholcus/logs"               //信息输出
	. "github.com/henrylee2cn/pholcus/app/spider" //必需
	// . "github.com/henrylee2cn/pholcus/app/spider/common" //选用

	// net包
	// "net/http" //设置http.Header
	// "net/url"

	// 编码包

	// "encoding/xml"
	"encoding/json"
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
	People.Register()
}

type Item struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Url      string `json:"url"`
	Date     string `json:"date"`
	NodeId   string `json:"nodeId"`
	ImgCount string `json:"imgCount"`
}
type News struct {
	Items []Item `json:"items"`
}

var news News

var People = &Spider{
	Name:        "人民网新闻抓取",
	Description: "人民网最新分类新闻",
	// Pausetime:    300,
	// Keyin:        KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Method: "GET",
				Url:    "http://news.people.com.cn/210801/211150/index.js?cache=false",
				Rule:   "新闻列表",
			})
		},

		Trunk: map[string]*Rule{
			"新闻列表": {
				ParseFunc: func(ctx *Context) {

					//query := ctx.GetDom()
					//str := query.Find("body").Text()

					//str := `{"items":[{"id":"282","title":"人社&nbsp;转型升级&quot;战术&quot;手册","url":"ht","date":"201","nodeId":"1001","imgCount":"4"}]}`

					str := ctx.GetText()

					err := json.Unmarshal([]byte(str), &news)
					if err != nil {
						log.Printf("解析错误： %v\n", err)
						return
					}
					/////////////////
					newsLength := len(news.Items)
					for i := 0; i < newsLength; i++ {
						ctx.AddQueue(&request.Request{
							Url:  news.Items[i].Url,
							Rule: "热点新闻",
							Temp: map[string]interface{}{
								"id":       news.Items[i].Id,
								"title":    news.Items[i].Title,
								"date":     news.Items[i].Date,
								"newsType": news.Items[i].NodeId,
							},
						})
					}
					/////////////////
				},
			},

			"热点新闻": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"ID",
					"标题",
					"内容",
					"类别",
					"ReleaseTime",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					// 获取内容
					content := query.Find("#p_content").Text()
					// re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
					// content = re.ReplaceAllStringFunc(content, strings.ToLower)
					// content = re.ReplaceAllString(content, "")

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: ctx.GetTemp("id", ""),
						1: ctx.GetTemp("title", ""),
						2: content,
						3: ctx.GetTemp("newsType", ""),
						4: ctx.GetTemp("date", ""),
					})
				},
			},
		},
	},
}
