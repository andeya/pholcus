package spiders

// 基础包
import (
	"github.com/PuerkitoBio/goquery"                    //DOM解析
	"github.com/henrylee2cn/pholcus/downloader/context" //必需
	"github.com/henrylee2cn/pholcus/reporter"           //信息输出
	. "github.com/henrylee2cn/pholcus/spiders/spider"   //必需
)

// 设置header包
import (
	"net/http" //http.Header
)

// 编码包
import (
	// "encoding/json"
	"encoding/xml"
)

// 字符串处理包
import (
	"regexp"
	// "strconv"
	"strings"
)

// 其他包
import (
	// "fmt"
	// "math"
	"time"
)

var rss_BaiduNews = NewRSS(map[string]string{
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
},
	[]int{1, 2, 3, 4, 5, 6},
)

type BaiduNewsData struct {
	Item []BaiduNewsItem `xml:"item"`
}

type BaiduNewsItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Author      string `xml:"author"`
}

var BaiduNews = &Spider{
	Name: "百度RSS新闻",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			for k, _ := range rss_BaiduNews.Src {
				self.AidRule("LOOP", []interface{}{k})
			}
		},

		Nodes: map[string]*Rule{
			"LOOP": &Rule{
				AidFunc: func(self *Spider, aid []interface{}) interface{} {
					k := aid[0].(string)
					v := rss_BaiduNews.Src[k]

					self.AddQueue(map[string]interface{}{
						"url":      v + "#" + time.Now().String(),
						"rule":     "XML",
						"header":   http.Header{"Content-Type": []string{"text/html", "charset=GB2312"}},
						"respType": "text",
						"temp":     map[string]interface{}{"src": k},
					})
					return nil
				},
			},
			"XML": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					page := GBKToUTF8(resp.GetBodyStr())
					page = strings.TrimLeft(page, `<?xml version="1.0" encoding="gb2312"?>`)
					re, _ := regexp.Compile(`\<[\/]?rss\>`)
					page = re.ReplaceAllString(page, "")

					content := new(BaiduNewsData)
					if err := xml.Unmarshal([]byte(page), content); err != nil {
						reporter.Log.Println(err)
						return
					}

					src := resp.GetTemp("src").(string)

					for _, v := range content.Item {

						self.AddQueue(map[string]interface{}{
							"url":  v.Link,
							"rule": "新闻详情",
							"temp": map[string]interface{}{
								"title":       CleanHtml(v.Title, 4),
								"description": CleanHtml(v.Description, 4),
								"src":         src,
								"releaseTime": CleanHtml(v.PubDate, 4),
								"author":      CleanHtml(v.Author, 4),
							},
						})
					}

					// 循环请求
					rss_BaiduNews.Wait(src)
					self.AidRule("LOOP", []interface{}{src})
				},
			},

			"新闻详情": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"标题",
					"描述",
					"内容",
					"发布时间",
					"分类",
					"作者",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					// RSS标记更新
					rss_BaiduNews.Updata(resp.GetTemp("src").(string))

					query1 := resp.GetHtmlParser()

					query := query1.Find("body")

					title := resp.GetTemp("title").(string)

					var findP func(html *goquery.Selection) *goquery.Selection
					findP = func(html *goquery.Selection) *goquery.Selection {
						if html.Is("body") {
							return html
						} else if result := html.Parent().Find("p"); len(result.Nodes) == 0 {
							return findP(html.Parent())
						} else {
							return html.Parent()
						}
					}

					var info *goquery.Selection

					if h1s := query.Find("h1"); len(h1s.Nodes) != 0 {
						for i := 0; i < len(h1s.Nodes); i++ {
							info = findP(h1s.Eq(i))
						}
					} else if h2s := query.Find("h2"); len(h2s.Nodes) != 0 {
						for i := 0; i < len(h2s.Nodes); i++ {
							info = findP(h2s.Eq(i))
						}
					} else if h3s := query.Find("h3"); len(h3s.Nodes) != 0 {
						for i := 0; i < len(h3s.Nodes); i++ {
							info = findP(h3s.Eq(i))
						}
					} else {
						info = query.Find("body")
					}
					// 去除标签
					// info.RemoveFiltered("script")
					// info.RemoveFiltered("style")
					infoStr, _ := info.Html()

					// 清洗HTML
					infoStr = CleanHtml(infoStr, 5)

					// 结果存入Response中转
					resp.AddItem(map[string]interface{}{
						self.GetOutFeild(resp, 0): title,
						self.GetOutFeild(resp, 1): resp.GetTemp("description"),
						self.GetOutFeild(resp, 2): infoStr,
						self.GetOutFeild(resp, 3): resp.GetTemp("releaseTime"),
						self.GetOutFeild(resp, 4): resp.GetTemp("src"),
						self.GetOutFeild(resp, 5): resp.GetTemp("author"),
					})
				},
			},
		},
	},
}
