package rules

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request" // required
	spider "github.com/andeya/pholcus/app/spider"      // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing

	// . "github.com/andeya/pholcus/app/spider/common"    // optional
	"github.com/andeya/pholcus/logs" // logging

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
	"math"
	// "time"
)

func init() {
	GoogleSearch.Register()
}

var googleIp = []string{
	"210.242.125.100",
	"210.242.125.96",
	"210.242.125.91",
	"210.242.125.95",
	"64.233.189.163",
	"58.123.102.5",
	"210.242.125.97",
	"210.242.125.115",
	"58.123.102.28",
	"210.242.125.70",
	"220.255.2.153",
}

var GoogleSearch = &spider.Spider{
	Name:        "Google search",
	Description: "Crawls pages from [www.google.com]",
	// Pausetime: 300,
	Keyin:        spider.KEYIN,
	Limit:        spider.LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			var url string
			var success bool
			logs.Log().Informational("Running google spider，this may take some time...")

			for _, ip := range googleIp {
				// url = "http://" + ip + "/search?q=" + ctx.GetKeyin() + "&newwindow=1&biw=1600&bih=398&start="
				// Beware of redirections, if it doesnt work use google domain:
				// url = "https://google.co.uk/search?q=" + ctx.GetKeyin()
				url = "http://" + ip + "/?gws_rd=ssl#q=" + ctx.GetKeyin()
				logs.Log().Informational("测试 " + ip)
				if goquery.NewDocument(url).IsOk() {
					success = true
					break
				}
			}
			if !success {
				logs.Log().Critical("Could not reach any of the Google mirrors")
				return
			}
			logs.Log().Critical("Starting Google search ...")
			ctx.AddQueue(&request.Request{
				URL:  url,
				Rule: "total_pages",
				Temp: map[string]interface{}{
					"baseUrl": url,
				},
			})
		},

		Trunk: map[string]*spider.Rule{

			"total_pages": {
				AidFunc: func(ctx *spider.Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						ctx.AddQueue(&request.Request{
							URL:  aid["urlBase"].(string) + "&start=" + strconv.Itoa(10*loop[0]),
							Rule: aid["Rule"].(string),
						})
					}
					return nil
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					txt := query.Find("#resultStats").Text()
					re := regexp.MustCompile(`,+`)
					txt = re.ReplaceAllString(txt, "")
					re = regexp.MustCompile(`[\d]+`)
					txt = re.FindString(txt)
					num, _ := strconv.Atoi(txt)
					total := int(math.Ceil(float64(num) / 10))
					if total > ctx.GetLimit() {
						total = ctx.GetLimit()
					} else if total == 0 {
						logs.Log().Critical("[ERROR：| Spider：%v | KEYIN：%v | Rule：%v] Did not fetch any data！!!\n", ctx.GetName(), ctx.GetKeyin(), ctx.GetRuleName())
						return
					}
					// call helper function under specified rule
					ctx.Aid(map[string]interface{}{
						"loop":    [2]int{1, total},
						"urlBase": ctx.GetTemp("baseUrl", ""),
						"Rule":    "search_results",
					})
					// parse response with specified rule
					ctx.Parse("search_results")
				},
			},

			"search_results": {
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"title",
					"content",
					"href",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					query.Find("#ires .g").Each(func(i int, s *goquery.Selection) {
						t := s.Find(".r > a")
						href := t.Attr("href").UnwrapOr("")
						href = strings.TrimLeft(href, "/url?q=")
						logs.Log().Informational(href)
						title := t.Text()
						content := s.Find(".st").Text()
						ctx.Output(map[int]interface{}{
							0: title,
							1: content,
							2: href,
						})
					})
				},
			},
		},
	},
}
