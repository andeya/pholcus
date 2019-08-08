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
	// "encoding/json"
	"encoding/xml"

	// 字符串处理包
	// "regexp"
	// "strconv"
	// "strings"

	// 其他包
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

var BaiduNews = &Spider{
	Name:        "百度RSS新闻",
	Description: "百度RSS新闻，实现轮询更新 [Auto Page] [news.baidu.com]",
	// Pausetime: 300,
	// Keyin:     KEYIN,
	EnableCookie: false,
	// Limit:        LIMIT,
	// 命名空间相对于数据库名，不依赖具体数据内容，可选
	Namespace: nil,
	// 子命名空间相对于表名，可依赖具体数据内容，可选
	SubNamespace: func(self *Spider, dataCell map[string]interface{}) string {
		return dataCell["Data"].(map[string]interface{})["分类"].(string)
	},
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			for k := range rss_BaiduNews {
				ctx.SetTimer(k, time.Minute*5, nil)
				ctx.Aid(map[string]interface{}{"loop": k}, "LOOP")
			}
		},

		Trunk: map[string]*Rule{
			"LOOP": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					k := aid["loop"].(string)
					v := rss_BaiduNews[k]

					ctx.AddQueue(&request.Request{
						Url:    v,
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
				ParseFunc: func(ctx *Context) {
					var src = ctx.GetTemp("src", "").(string)
					defer func() {
						// 循环请求
						ctx.RunTimer(src)
						ctx.Aid(map[string]interface{}{"loop": src}, "LOOP")
					}()

					page := ctx.GetText()
					rss := new(BaiduNewsRss)
					if err := xml.Unmarshal([]byte(page), rss); err != nil {
						logs.Log.Error("XML列表页: %v", err)
						return
					}
					content := rss.Channel
					for _, v := range content.Item {
						ctx.AddQueue(&request.Request{
							Url:  v.Link,
							Rule: "新闻详情",
							Temp: map[string]interface{}{
								"title":       CleanHtml(v.Title, 4),
								"description": CleanHtml(v.Description, 4),
								"src":         src,
								"releaseTime": CleanHtml(v.PubDate, 4),
								"author":      CleanHtml(v.Author, 4),
							},
						})
					}
				},
			},

			"新闻详情": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"标题",
					"描述",
					"内容",
					"发布时间",
					"分类",
					"作者",
				},
				ParseFunc: func(ctx *Context) {
					var title = ctx.GetTemp("title", "").(string)

					infoStr, isReload := baiduNewsFn.prase(ctx)
					if isReload {
						return
					}
					// 结果存入Response中转
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

type baiduNews map[string]func(ctx *Context) (infoStr string, isReload bool)

// @url 必须为含有协议头的地址
func (b baiduNews) prase(ctx *Context) (infoStr string, isReload bool) {
	url := ctx.GetHost()
	if _, ok := b[url]; ok {
		return b[url](ctx)
	} else {
		return b.commonPrase(ctx), false
	}
}

func (b baiduNews) commonPrase(ctx *Context) (infoStr string) {
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

	// 清洗HTML
	infoStr = CleanHtml(infoStr, 5)
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
	"yule.sohu.com": func(ctx *Context) (infoStr string, isReload bool) {
		infoStr = ctx.GetDom().Find("#contentText").Text()
		return
	},
	"news.qtv.com.cn": func(ctx *Context) (infoStr string, isReload bool) {
		infoStr = ctx.GetDom().Find(".zwConreally_z").Text()
		return
	},
}
