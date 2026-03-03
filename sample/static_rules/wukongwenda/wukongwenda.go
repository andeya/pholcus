package wukongwenda

import (
	// base packages
	"github.com/andeya/pholcus/app/downloader/request" // required
	//"github.com/andeya/pholcus/common/goquery"         // DOM parsing
	// "github.com/andeya/pholcus/logs"           // logging
	spider "github.com/andeya/pholcus/app/spider" // required
	// . "github.com/andeya/pholcus/app/spider/common" // optional

	// net packages
	"net/http" // set http.Header
	// "net/url"

	// encoding packages
	// "encoding/xml"
	// "encoding/json"

	// string processing packages
	// "regexp"
	"strconv"
	"strings"

	// other packages
	// "math"
	"time"

	"github.com/tidwall/gjson" // JSON processing package
)

func init() {
	WukongWenda.Register()
}

var domains = []string{
	"6300775428692904450", // hot
	"6215497896830175745", // entertainment
	"6215497726554016258", // sports
	"6215497898671475202", // auto
	"6215497899594222081", // tech
	"6215497900164647426", // parenting
	"6215497899774577154", // food
	"6215497897518041601", // digital
	"6215497898084272641", // fashion
	"6215847700051528193", // pets
	"6215847700907166210", // collection
	"6215497901804620289", // home
	"6281512530493835777", // psychology
	"6215497897710979586", // more culture
	"6215847700454181377", // more agriculture
	"6215497895248923137", // more health
	"6215848044378720770", // more science
	"6215497899027991042", // more games
	"6215497895852902913", // more anime
	"6215497897312520705", // more education
	"6215497899963320834", // more career
	"6215497897899723265", // more travel
	"6215497900554717698", // more movies
}

const (
	WUKONG_NORMAL_URL = "https://www.wukong.com/wenda/web/nativefeed/brow/?concern_id=" // different column access URL
	UA                = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36"
)

var WukongWenda = &spider.Spider{
	Name:        "悟空问答",
	Description: "悟空问答 各个频道专栏问题",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			// process domains with same parse structure
			for _, domain := range domains {
				url := WUKONG_NORMAL_URL + domain + "&t=" +
					strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
				header := http.Header{}
				header.Add("User-Agent", UA)

				ctx.AddQueue(&request.Request{
					URL:    url,
					Header: header,
					Rule:   "获取结果",
				})

			}
		},

		Trunk: map[string]*spider.Rule{
			"获取结果": {
				// NOTE: field semantics and data output presence must be consistent
				ItemFields: []string{
					"问题标题",
					"问题描述",
					"问题回答",
					"问题url地址",
				},
				ParseFunc: func(ctx *spider.Context) {

					type question struct {
						title   string
						content string
						answer  string
						url     string
						offset  string
					}

					var questionlist []question
					data := gjson.Get(ctx.GetText(), "data")
					more := gjson.Get(ctx.GetText(), "has_more").String()

					data.ForEach(func(key, value gjson.Result) bool {
						questionlist = append(questionlist,
							question{
								title:   gjson.Get(value.String(), "question.title").String(),
								content: gjson.Get(value.String(), "question.content.text").String(),
								answer:  gjson.Get(value.String(), "answer.content").String(),
								url:     "https://www.wukong.com/question/" + gjson.Get(value.String(), "question.qid").String() + "/",
								offset:  gjson.Get(value.String(), "behot_time").String(),
							})
						return true
					})

					if more == "true" {
						newOffset := questionlist[len(questionlist)-1].offset
						header := http.Header{}
						header.Add("User-Agent", UA)

						visitURL := ctx.GetURL()
						if strings.Contains(visitURL, "&max_behot_time=") {
							visitURL = strings.Split(visitURL, "&max_behot_time=")[0]
						}

						ctx.AddQueue(&request.Request{
							URL:    visitURL + "&max_behot_time=" + newOffset,
							Header: header,
							Rule:   "获取结果",
						})

					}

					for _, v := range questionlist {
						ctx.Output(map[int]interface{}{
							0: v.title,
							1: v.content,
							2: v.answer,
							3: v.url,
						})
					}

				},
			},
		},
	},
}
