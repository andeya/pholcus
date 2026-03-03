package rules

// base packages
import (
	"log"

	// "github.com/andeya/pholcus/common/goquery"                        // DOM parsing
	"github.com/andeya/pholcus/app/downloader/request" // required
	// "github.com/andeya/pholcus/logs"               // logging
	spider "github.com/andeya/pholcus/app/spider" // required
	// . "github.com/andeya/pholcus/app/spider/common" // optional

	// net packages
	// "net/http" // set http.Header
	// "net/url"

	// encoding packages

	// "encoding/xml"
	"encoding/json"
	// string processing packages
	// "regexp"
	// "strconv"
	// "strings"
	// other packages
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

var People = &spider.Spider{
	Name:        "人民网新闻抓取",
	Description: "人民网最新分类新闻",
	// Pausetime:    300,
	// Keyin:        KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{
				Method: "GET",
				URL:    "http://news.people.com.cn/210801/211150/index.js?cache=false",
				Rule:   "新闻列表",
			})
		},

		Trunk: map[string]*spider.Rule{
			"新闻列表": {
				ParseFunc: func(ctx *spider.Context) {

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
							URL:  news.Items[i].Url,
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
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"ID",
					"标题",
					"内容",
					"类别",
					"ReleaseTime",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					// get content
					content := query.Find("#p_content").Text()
					// re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
					// content = re.ReplaceAllStringFunc(content, strings.ToLower)
					// content = re.ReplaceAllString(content, "")

					// store results in Response
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
