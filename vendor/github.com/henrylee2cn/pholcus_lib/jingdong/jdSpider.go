package pholcus_lib

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	//"github.com/henrylee2cn/pholcus/logs"                   //信息输出
	// . "github.com/henrylee2cn/pholcus/app/spider/common"          //选用

	// net包
	// "net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	"regexp"
	"strconv"
	"strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
	//"fmt"
)

func init() {
	JDSpider.Register()
}

var JDSpider = &Spider{
	Name:        "京东搜索new",
	Description: "京东搜索结果 [search.jd.com]",
	// Pausetime: 300,
	Keyin:        KEYIN,
	Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			//Aid调用Rule中的AidFunc
			ctx.Aid(map[string]interface{}{"Rule": "判断页数"}, "判断页数")
		},

		Trunk: map[string]*Rule{
			//只判断关键字商品一共有多少页
			"判断页数": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					ctx.AddQueue(
						&request.Request{
							Url:  "http://search.jd.com/Search?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=1&click=0&page=1",
							Rule: aid["Rule"].(string),
						},
					)
					return nil
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					pageCount := 0
					query.Find("script").Each(func(i int, s *goquery.Selection) {
						if strings.Contains(s.Text(), "page_count") {
							re, _ := regexp.Compile(`page_count:"\d{1,}"`)
							temp := re.FindString(s.Text())
							re, _ = regexp.Compile(`\d{1,}`)
							temp2 := re.FindString(temp)
							pageCount, _ = strconv.Atoi(temp2)
						}
					})
					ctx.Aid(map[string]interface{}{"PageCount": pageCount}, "生成请求")
				},
			},

			"生成请求": {
				//单数页是url直接返回,双数页是异步加载,两个url在下面有写
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					//Url:  "http://search.jd.com/Search?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=1&click=0&page=" + strconv.Itoa(pageNum),
					//Url:  "http://search.jd.com/s_new.php?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=31&scrolling=y&pos=30&page=" + strconv.Itoa(pageNum),
					pageCount := aid["PageCount"].(int)

					for i := 1; i < pageCount; i++ {
						ctx.AddQueue(
							&request.Request{
								Url:  "http://search.jd.com/Search?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=1&click=0&page=" + strconv.Itoa(i*2-1),
								Rule: "搜索结果",
							},
						)
						ctx.AddQueue(
							&request.Request{
								Url:  "http://search.jd.com/s_new.php?keyword=" + ctx.GetKeyin() + "&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&bs=1&s=31&scrolling=y&pos=30&page=" + strconv.Itoa(i*2),
								Rule: "搜索结果",
							},
						)
					}
					return nil
				},
			},

			"搜索结果": {
				//从返回中解析出数据。注：异步返回的结果页面结构是和单数页的一样的，所以就一套解析就可以了。
				ItemFields: []string{
					"标题",
					"价格",
					"评论数",
					"链接",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					query.Find(".gl-item").Each(func(i int, s *goquery.Selection) {
						// 获取标题
						a := s.Find(".p-name.p-name-type-2 > a")
						title := a.Text()

						re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
						// title = re.ReplaceAllStringFunc(title, strings.ToLower)
						title = re.ReplaceAllString(title, " ")
						title = strings.Trim(title, " \t\n")

						// 获取价格
						price := s.Find(".p-price > strong > i").Text()

						// 获取评论数
						//#J_goodsList > ul > li:nth-child(1) > div > div.p-commit
						discuss := s.Find(".p-commit > strong > a").Text()

						// 获取URL
						url, _ := a.Attr("href")
						url = "http:" + url

						// 结果存入Response中转
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
