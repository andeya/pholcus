package rules

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request" // required
	spider "github.com/andeya/pholcus/app/spider"      // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing

	//"github.com/andeya/pholcus/logs"                   // logging
	// . "github.com/andeya/pholcus/app/spider/common"          // optional

	// net packages
	// "net/http" // set http.Header
	// "net/url"

	// encoding packages
	// "encoding/xml"
	// "encoding/json"

	// string processing packages
	"regexp"
	"strconv"
	"strings"
	// other packages
	// "fmt"
	// "math"
	// "time"
	//"fmt"
)

func init() {
	JDSpider.Register()
}

var JDSpider = &spider.Spider{
	Name:        "京东搜索new",
	Description: "京东搜索结果 [search.jd.com]",
	// Pausetime: 300,
	Keyin:        spider.KEYIN,
	Limit:        spider.LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			// Aid calls AidFunc in Rule
			ctx.Aid(map[string]interface{}{"Rule": "判断页数"}, "判断页数")
		},

		Trunk: map[string]*spider.Rule{
			// only determine total pages for keyword search
			"判断页数": {
				AidFunc: func(ctx *spider.Context, aid map[string]interface{}) interface{} {
					ctx.AddQueue(
						&request.Request{
							URL:  "http://search.jd.com/Search?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=1&click=0&page=1",
							Rule: aid["Rule"].(string),
						},
					)
					return nil
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					pageCount := 0
					query.Find("script").Each(func(i int, s *goquery.Selection) {
						if strings.Contains(s.Text(), "page_count") {
							re := regexp.MustCompile(`page_count:"\d{1,}"`)
							temp := re.FindString(s.Text())
							re = regexp.MustCompile(`\d{1,}`)
							temp2 := re.FindString(temp)
							pageCount, _ = strconv.Atoi(temp2)
						}
					})
					ctx.Aid(map[string]interface{}{"PageCount": pageCount}, "生成请求")
				},
			},

			"生成请求": {
				// odd pages return URL directly, even pages are async loaded; both URLs written below
				AidFunc: func(ctx *spider.Context, aid map[string]interface{}) interface{} {
					//URL:  "http://search.jd.com/Search?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=1&click=0&page=" + strconv.Itoa(pageNum),
					//URL:  "http://search.jd.com/s_new.php?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=31&scrolling=y&pos=30&page=" + strconv.Itoa(pageNum),
					pageCount := aid["PageCount"].(int)

					for i := 1; i < pageCount; i++ {
						ctx.AddQueue(
							&request.Request{
								URL:  "http://search.jd.com/Search?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=1&click=0&page=" + strconv.Itoa(i*2-1),
								Rule: "搜索结果",
							},
						)
						ctx.AddQueue(
							&request.Request{
								URL:  "http://search.jd.com/s_new.php?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=31&scrolling=y&pos=30&page=" + strconv.Itoa(i*2),
								Rule: "搜索结果",
							},
						)
					}
					return nil
				},
			},

			"搜索结果": {
				// parse data from response. NOTE: async response page structure is same as odd pages, so one parse logic suffices.
				ItemFields: []string{
					"标题",
					"价格",
					"评论数",
					"链接",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					query.Find(".gl-item").Each(func(i int, s *goquery.Selection) {
						// get title
						a := s.Find(".p-name.p-name-type-2 > a")
						title := a.Text()

						re := regexp.MustCompile("\\<[\\S\\s]+?\\>")
						// title = re.ReplaceAllStringFunc(title, strings.ToLower)
						title = re.ReplaceAllString(title, " ")
						title = strings.Trim(title, " \t\n")

						// get price
						price := s.Find(".p-price > strong > i").Text()

						// get comment count
						//#J_goodsList > ul > li:nth-child(1) > div > div.p-commit
						discuss := s.Find(".p-commit > strong > a").Text()

						// get URL
						url := a.Attr("href").UnwrapOr("")
						url = "http:" + url

						// store results in Response
						if title != "" {
							ctx.Output(map[int]interface{}{
								0: title,
								1: price,
								2: discuss,
								3: url,
							})
						}
					})
				},
			},
		},
	},
}
