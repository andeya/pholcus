package zhihu_bianji

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request" // required
	spider "github.com/andeya/pholcus/app/spider"      // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing

	// . "github.com/andeya/pholcus/app/spider/common"    // optional
	//"github.com/andeya/pholcus/logs" // logging

	// net packages
	"net/http" // set http.Header
	"net/url"

	// encoding packages
	// "encoding/xml"
	"encoding/json"

	// string processing packages
	//"strconv"

	// other packages
	// "fmt"
	// "time"
	//"strconv"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	ZhihuBianji.Register()
}

var ZhihuBianji = &spider.Spider{
	Name:        "知乎编辑推荐",
	Description: "知乎编辑推荐",
	Pausetime:   300,
	//Keyin:        KEYIN,
	Limit:        spider.LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{
				URL:  "https://www.zhihu.com/explore/recommendations",
				Rule: "知乎编辑推荐",
			})

		},

		Trunk: map[string]*spider.Rule{
			"知乎编辑推荐": {
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					regular := "#zh-recommend-list-full .zh-general-list .zm-item h2 a"
					query.Find(regular).
						Each(func(i int, s *goquery.Selection) {
							if url := s.Attr("href"); url.IsSome() {
								u := changeToAbspath(url.Unwrap())
								ctx.AddQueue(&request.Request{URL: u, Rule: "解析落地页"})
							}
						})

					limit := ctx.GetLimit()

					if len(query.Find(regular).Nodes) < limit {
						total := int(math.Ceil(float64(limit) / float64(20)))
						ctx.Aid(map[string]interface{}{
							"loop": [2]int{1, total},
							"Rule": "知乎编辑推荐翻页",
						}, "知乎编辑推荐翻页")
					}
				},
			},

			"知乎编辑推荐翻页": {
				AidFunc: func(ctx *spider.Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						offset := loop[0] * 20
						header := make(http.Header)
						header.Set("Content-Type", "application/x-www-form-urlencoded")
						ctx.AddQueue(&request.Request{
							URL:        "https://www.zhihu.com/node/ExploreRecommendListV2",
							Rule:       aid["Rule"].(string),
							Method:     "POST",
							Header:     header,
							PostData:   url.Values{"method": {"next"}, "params": {`{"limit":20,"offset":` + strconv.Itoa(offset) + `}`}}.Encode(),
							Reloadable: true,
						})
					}

					return nil
				},
				ParseFunc: func(ctx *spider.Context) {
					type Items struct {
						R   int           `json:"r"`
						Msg []interface{} `json:"msg"`
					}

					content, err := io.ReadAll(ctx.GetResponse().Body)

					ctx.GetResponse().Body.Close()

					if err != nil {
						ctx.Log().Error(err.Error())
					}

					e := new(Items)

					err = json.Unmarshal(content, e)

					html := ""

					for _, v := range e.Msg {
						msg, ok := v.(string)
						if ok {
							html = html + "\n" + msg
						}
					}

					ctx = ctx.ResetText(html)

					query := ctx.GetDom()

					query.Find(".zm-item h2 a").Each(func(i int, selection *goquery.Selection) {
						if url := selection.Attr("href"); url.IsSome() {
							u := changeToAbspath(url.Unwrap())
							if filterZhihuAnswerURL(u) {
								ctx.AddQueue(&request.Request{URL: u, Rule: "解析知乎问答落地页"})
							} else {
								ctx.AddQueue(&request.Request{URL: u, Rule: "解析知乎文章落地页"})
							}
						}
					})

				},
			},

			"解析知乎问答落地页": {
				ItemFields: []string{
					"标题",
					"提问内容",
					"回答内容",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					questionHeader := query.Find(".QuestionPage .QuestionHeader .QuestionHeader-content")
					//headerSide := questionHeader.Find(".QuestionHeader-side")
					headerMain := questionHeader.Find(".QuestionHeader-main")

					// get question title
					title := headerMain.Find(".QuestionHeader-title").Text()

					// get question description
					content := headerMain.Find(".QuestionHeader-detail span").Text()

					answerMain := query.Find(".QuestionPage .Question-main")

					answer, _ := answerMain.Find(".AnswerCard .QuestionAnswer-content .ContentItem .RichContent .RichContent-inner").First().Html()

					// store results in Response
					ctx.Output(map[int]interface{}{
						0: title,
						1: content,
						2: answer,
					})

				},
			},

			"解析知乎文章落地页": {
				ItemFields: []string{
					"标题",
					"内容",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()

					// get question title
					title, _ := query.Find(".PostIndex-title.av-paddingSide.av-titleFont").Html()

					// get question description
					content, _ := query.Find(".RichText.PostIndex-content.av-paddingSide.av-card").Html()

					// store results in Response
					ctx.Output(map[int]interface{}{
						0: title,
						1: content,
					})

				},
			},
		},
	},
}

// convert relative path to absolute path
func changeToAbspath(url string) string {
	if strings.HasPrefix(url, "https://") {
		return url
	}
	return "https://www.zhihu.com" + url
}

// determine if URL is user answer or zhihu column article
func filterZhihuAnswerURL(url string) bool {
	return regexp.MustCompile(`^https:\/\/www\.zhihu\.com\/question\/\d{1,}(\/answer\/\d{1,})?$`).MatchString(url)
}
