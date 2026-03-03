package rules

// base packages
import (
	// "log"

	"github.com/andeya/pholcus/app/downloader/request" // required
	spider "github.com/andeya/pholcus/app/spider"      // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing

	// "github.com/andeya/pholcus/logs"         // logging
	// . "github.com/andeya/pholcus/app/spider/common" // optional

	// net packages
	// "net/http" // set http.Header
	// "net/url"

	// encoding packages
	// "encoding/xml"
	// "encoding/json"

	// string processing packages
	"regexp"
	"strconv"
	// "strings"
	// other packages
	// "fmt"
	// "math"
	// "time"
)

func init() {
	IJGUC.Register()
}

var IJGUC = &spider.Spider{
	Name:        "IJGUC期刊",
	Description: "IJGUC期刊",
	// Pausetime:    300,
	// Keyin:        KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{
				URL:  "http://www.inderscience.com/info/inarticletoc.php?jcode=ijguc&year=2016&vol=7&issue=1",
				Rule: "期刊列表",
			})
		},

		Trunk: map[string]*spider.Rule{
			"期刊列表": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					for i := 1; i <= 7; i++ {
						id := "#eventbody" + strconv.Itoa(i) + " a"
						query.Find(id).Each(func(j int, s *goquery.Selection) {
							if url := s.Attr("href"); url.IsSome() {
								// log.Print(url)
								ctx.AddQueue(&request.Request{URL: url.Unwrap(), Rule: "文章列表"})
							}
						})
					}
				},
			},
			"文章列表": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					//#journalcol1 article table tbody tr td:eq(1) table:eq(1) a
					query.Find("#journalcol1 article table tbody tr td").Each(func(i int, td *goquery.Selection) {
						if i == 1 {
							td.Find("table").Each(func(j int, table *goquery.Selection) {
								if j == 1 {
									table.Find("a").Each(func(k int, a *goquery.Selection) {
										if k%2 == 0 {
											if url := a.Attr("href"); url.IsSome() {
												// log.Print(url)
												ctx.AddQueue(&request.Request{URL: url.Unwrap(), Rule: "文章页"})
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
				// note: field semantics and output data must be consistent
				ItemFields: []string{
					"Title",
					"Author",
					"Addresses",
					"Journal",
					"Abstract",
					"Keywords",
					"DOI",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					// get content
					content := query.Find("#col1").Text()

					// filter tags
					re := regexp.MustCompile("\\<[\\S\\s]+?\\>")
					content = re.ReplaceAllString(content, "")

					// Title
					re = regexp.MustCompile("Title:(.*?)Author:")
					title := re.FindStringSubmatch(content)[1]
					// Author
					re = regexp.MustCompile("Author:(.*?)Addresses:")
					au := re.FindStringSubmatch(content)
					var author string
					if len(au) > 0 {
						author = au[1]
					} else {
						re = regexp.MustCompile("Author:(.*?)Address:")
						author = re.FindStringSubmatch(content)[1]
					}
					// Addresses & Address
					re = regexp.MustCompile("Addresses:(.*?)Journal:")
					address := re.FindStringSubmatch(content)
					var addresses string
					if len(address) > 0 {
						addresses = address[1]
					} else {
						re = regexp.MustCompile("Address:(.*?)Journal:")
						addresses = re.FindStringSubmatch(content)[1]
					}
					// Journal
					re = regexp.MustCompile("Journal:(.*?)Abstract:")
					journal := re.FindStringSubmatch(content)[1]
					// Abstract
					re = regexp.MustCompile("Abstract:(.*?)Keywords:")
					abstract := re.FindStringSubmatch(content)[1]
					// Keywords
					re = regexp.MustCompile("Keywords:(.*?)DOI:")
					keywords := re.FindStringSubmatch(content)[1]
					// DOI
					re = regexp.MustCompile("DOI: ")
					doiIndex := re.FindStringSubmatchIndex(content)
					rs := []rune(content)
					left := doiIndex[1] - 8
					right := left + 43
					doi := string(rs[left:right])

					// store results in Response
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
