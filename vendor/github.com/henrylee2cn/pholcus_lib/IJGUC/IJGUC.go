package pholcus_lib

// 基础包
import (
	// "log"

	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	// "github.com/henrylee2cn/pholcus/logs"         //信息输出
	// . "github.com/henrylee2cn/pholcus/app/spider/common" //选用

	// net包
	// "net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	"regexp"
	"strconv"
	// "strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	IJGUC.Register()
}

var IJGUC = &Spider{
	Name:        "IJGUC期刊",
	Description: "IJGUC期刊",
	// Pausetime:    300,
	// Keyin:        KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:  "http://www.inderscience.com/info/inarticletoc.php?jcode=ijguc&year=2016&vol=7&issue=1",
				Rule: "期刊列表",
			})
		},

		Trunk: map[string]*Rule{
			"期刊列表": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					for i := 1; i <= 7; i++ {
						id := "#eventbody" + strconv.Itoa(i) + " a"
						query.Find(id).Each(func(j int, s *goquery.Selection) {
							if url, ok := s.Attr("href"); ok {
								// log.Print(url)
								ctx.AddQueue(&request.Request{Url: url, Rule: "文章列表"})
							}
						})
					}
				},
			},
			"文章列表": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					//#journalcol1 article table tbody tr td:eq(1) table:eq(1) a
					query.Find("#journalcol1 article table tbody tr td").Each(func(i int, td *goquery.Selection) {
						if i == 1 {
							td.Find("table").Each(func(j int, table *goquery.Selection) {
								if j == 1 {
									table.Find("a").Each(func(k int, a *goquery.Selection) {
										if k%2 == 0 {
											if url, ok := a.Attr("href"); ok {
												// log.Print(url)
												ctx.AddQueue(&request.Request{Url: url, Rule: "文章页"})
											}
										}
									})
								}
							})
						}
					})
				},
			},
			"文章页": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"Title",
					"Author",
					"Addresses",
					"Journal",
					"Abstract",
					"Keywords",
					"DOI",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					// 获取内容
					content := query.Find("#col1").Text()

					// 过滤标签
					re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
					content = re.ReplaceAllString(content, "")

					// Title
					re, _ = regexp.Compile("Title:(.*?)Author:")
					title := re.FindStringSubmatch(content)[1]
					// Author
					re, _ = regexp.Compile("Author:(.*?)Addresses:")
					au := re.FindStringSubmatch(content)
					var author string
					if len(au) > 0 {
						author = au[1]
					} else {
						re, _ = regexp.Compile("Author:(.*?)Address:")
						author = re.FindStringSubmatch(content)[1]
					}
					// Addresses & Address
					re, _ = regexp.Compile("Addresses:(.*?)Journal:")
					address := re.FindStringSubmatch(content)
					var addresses string
					if len(address) > 0 {
						addresses = address[1]
					} else {
						re, _ = regexp.Compile("Address:(.*?)Journal:")
						addresses = re.FindStringSubmatch(content)[1]
					}
					// Journal
					re, _ = regexp.Compile("Journal:(.*?)Abstract:")
					journal := re.FindStringSubmatch(content)[1]
					// Abstract
					re, _ = regexp.Compile("Abstract:(.*?)Keywords:")
					abstract := re.FindStringSubmatch(content)[1]
					// Keywords
					re, _ = regexp.Compile("Keywords:(.*?)DOI:")
					keywords := re.FindStringSubmatch(content)[1]
					// DOI
					re, _ = regexp.Compile("DOI: ")
					doiIndex := re.FindStringSubmatchIndex(content)
					rs := []rune(content)
					left := doiIndex[1] - 8
					right := left + 43
					doi := string(rs[left:right])

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: author,
						2: addresses,
						3: journal,
						4: abstract,
						5: keywords,
						6: doi,
					})
				},
			},
		},
	},
}
