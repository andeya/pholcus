package rules

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request"         // required
	spider "github.com/andeya/pholcus/app/spider"              // required
	spidercommon "github.com/andeya/pholcus/app/spider/common" // optional
	"github.com/andeya/pholcus/common/goquery"                 // DOM parsing
	"github.com/andeya/pholcus/logs"                           // logging

	// net packages
	"net/http" // set http.Header
	// "net/url"

	// encoding packages
	// "encoding/json"
	"encoding/xml"

	// string processing packages
	// "regexp"
	// "strconv"
	// "strings"

	// other packages
	// "fmt"
	// "math"
	"time"
)

func init() {
	BaiduNews.Register()
}

var rss_BaiduNews = map[string]string{
	"国内最新":  "http://news.baidu.com/n?cmd=4&class=civilnews&tn=rss",
	"国际最新":  "http://news.baidu.com/n?cmd=4&class=internews&tn=rss",
	"军事最新":  "http://news.baidu.com/n?cmd=4&class=mil&tn=rss",
	"财经最新":  "http://news.baidu.com/n?cmd=4&class=finannews&tn=rss",
	"互联网最新": "http://news.baidu.com/n?cmd=4&class=internet&tn=rss",
	"房产最新":  "http://news.baidu.com/n?cmd=4&class=housenews&tn=rss",
	"汽车最新":  "http://news.baidu.com/n?cmd=4&class=autonews&tn=rss",
	"体育最新":  "http://news.baidu.com/n?cmd=4&class=sportnews&tn=rss",
	"娱乐最新":  "http://news.baidu.com/n?cmd=4&class=enternews&tn=rss",
	"游戏最新":  "http://news.baidu.com/n?cmd=4&class=gamenews&tn=rss",
	"教育最新":  "http://news.baidu.com/n?cmd=4&class=edunews&tn=rss",
	"女人最新":  "http://news.baidu.com/n?cmd=4&class=healthnews&tn=rss",
	"科技最新":  "http://news.baidu.com/n?cmd=4&class=technnews&tn=rss",
	"社会最新":  "http://news.baidu.com/n?cmd=4&class=socianews&tn=rss",
}

type (
	BaiduNewsRss struct {
		Channel BaiduNewsData `xml:"channel"`
	}
	BaiduNewsData struct {
		Item []BaiduNewsItem `xml:"item"`
	}
	BaiduNewsItem struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		PubDate     string `xml:"pubDate"`
		Author      string `xml:"author"`
	}
)

var BaiduNews = &spider.Spider{
	Name:        "百度RSS新闻",
	Description: "百度RSS新闻，实现轮询更新 [Auto Page] [news.baidu.com]",
	// Pausetime: 300,
	// Keyin:     KEYIN,
	EnableCookie: false,
	// Limit:        LIMIT,
	// namespace is relative to database name, independent of data content, optional
	Namespace: nil,
	// sub-namespace is relative to table name, may depend on data content, optional
	SubNamespace: func(self *spider.Spider, dataCell map[string]interface{}) string {
		return dataCell["Data"].(map[string]interface{})["分类"].(string)
	},
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			for k := range rss_BaiduNews {
				ctx.SetTimer(k, time.Minute*5, nil)
				ctx.Aid(map[string]interface{}{"loop": k}, "LOOP")
			}
		},

		Trunk: map[string]*spider.Rule{
			"LOOP": {
				AidFunc: func(ctx *spider.Context, aid map[string]interface{}) interface{} {
					k := aid["loop"].(string)
					v := rss_BaiduNews[k]

					ctx.AddQueue(&request.Request{
						URL:    v,
						Rule:   "XML列表页",
						Header: http.Header{"Content-Type": []string{"application/xml"}},
						Temp:   map[string]interface{}{"src": k},
						// DialTimeout: -1,
						// ConnTimeout: -1,
						// TryTimes:    -1,
						Reloadable: true,
					})
					return nil
				},
			},
			"XML列表页": {
				ParseFunc: func(ctx *spider.Context) {
					var src = ctx.GetTemp("src", "").(string)
					defer func() {
						// loop request
						ctx.RunTimer(src)
						ctx.Aid(map[string]interface{}{"loop": src}, "LOOP")
					}()

					page := ctx.GetText()
					rss := new(BaiduNewsRss)
					if err := xml.Unmarshal([]byte(page), rss); err != nil {
						logs.Log().Error("XML列表页: %v", err)
						return
					}
					content := rss.Channel
					for _, v := range content.Item {
						ctx.AddQueue(&request.Request{
							URL:  v.Link,
							Rule: "新闻详情",
							Temp: map[string]interface{}{
								"title":       spidercommon.CleanHtml(v.Title, 4),
								"description": spidercommon.CleanHtml(v.Description, 4),
								"src":         src,
								"releaseTime": spidercommon.CleanHtml(v.PubDate, 4),
								"author":      spidercommon.CleanHtml(v.Author, 4),
							},
						})
					}
				},
			},

			"新闻详情": {
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"标题",
					"描述",
					"内容",
					"发布时间",
					"分类",
					"作者",
				},
				ParseFunc: func(ctx *spider.Context) {
					var title = ctx.GetTemp("title", "").(string)

					infoStr, isReload := baiduNewsFn.prase(ctx)
					if isReload {
						return
					}
					// store results in Response
					ctx.Output(map[int]interface{}{
						0: title,
						1: ctx.GetTemp("description", ""),
						2: infoStr,
						3: ctx.GetTemp("releaseTime", ""),
						4: ctx.GetTemp("src", ""),
						5: ctx.GetTemp("author", ""),
					})
				},
			},
		},
	},
}

type baiduNews map[string]func(ctx *spider.Context) (infoStr string, isReload bool)

// @url must be an address with protocol header
func (b baiduNews) prase(ctx *spider.Context) (infoStr string, isReload bool) {
	url := ctx.GetHost()
	if _, ok := b[url]; ok {
		return b[url](ctx)
	} else {
		return b.commonPrase(ctx), false
	}
}

func (b baiduNews) commonPrase(ctx *spider.Context) (infoStr string) {
	body := ctx.GetDom().Find("body")

	var info *goquery.Selection

	if h1s := body.Find("h1"); len(h1s.Nodes) != 0 {
		for i := 0; i < len(h1s.Nodes); i++ {
			info = b.findP(h1s.Eq(i))
		}
	} else if h2s := body.Find("h2"); len(h2s.Nodes) != 0 {
		for i := 0; i < len(h2s.Nodes); i++ {
			info = b.findP(h2s.Eq(i))
		}
	} else if h3s := body.Find("h3"); len(h3s.Nodes) != 0 {
		for i := 0; i < len(h3s.Nodes); i++ {
			info = b.findP(h3s.Eq(i))
		}
	} else {
		info = body.Find("body")
	}
	infoStr, _ = info.Html()

	// clean HTML
	infoStr = spidercommon.CleanHtml(infoStr, 5)
	return
}

func (b baiduNews) findP(html *goquery.Selection) *goquery.Selection {
	if html.Is("body") {
		return html
	} else if result := html.Parent().Find("p"); len(result.Nodes) == 0 {
		return b.findP(html.Parent())
	} else {
		return html.Parent()
	}
}

var baiduNewsFn = baiduNews{
	"yule.sohu.com": func(ctx *spider.Context) (infoStr string, isReload bool) {
		infoStr = ctx.GetDom().Find("#contentText").Text()
		return
	},
	"news.qtv.com.cn": func(ctx *spider.Context) (infoStr string, isReload bool) {
		infoStr = ctx.GetDom().Find(".zwConreally_z").Text()
		return
	},
}
