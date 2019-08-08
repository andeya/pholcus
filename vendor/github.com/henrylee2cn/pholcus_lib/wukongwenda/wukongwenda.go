package wukongwenda

import (
	// 基础包
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	//"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	// "github.com/henrylee2cn/pholcus/logs"           //信息输出
	. "github.com/henrylee2cn/pholcus/app/spider" //必需
	// . "github.com/henrylee2cn/pholcus/app/spider/common" //选用

	// net包
	"net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	// "regexp"
	"strconv"
	"strings"

	// 其他包
	// "math"
	"time"
	"github.com/tidwall/gjson" //引用的json处理的包

)

func init() {
	WukongWenda.Register()
}

var domains = []string{
	"6300775428692904450",//热门
	"6215497896830175745",//娱乐
	"6215497726554016258",//体育
	"6215497898671475202",//汽车
	"6215497899594222081",//科技
	"6215497900164647426",//育儿
	"6215497899774577154",//美食
	"6215497897518041601",//数码
	"6215497898084272641",//时尚
	"6215847700051528193",//宠物
	"6215847700907166210",//收藏
	"6215497901804620289",//家居
	"6281512530493835777",//心理
	"6215497897710979586",//更多 文化
	"6215847700454181377",//更多 三农
	"6215497895248923137",//更多 健康
	"6215848044378720770",//更多 科学
	"6215497899027991042",//更多 游戏
	"6215497895852902913",//更多 动漫
	"6215497897312520705",//更多 教育
	"6215497899963320834",//更多 职场
	"6215497897899723265",//更多 旅游
	"6215497900554717698",//更多 电影
}

const (
	WUKONG_NORMAL_URL = "https://www.wukong.com/wenda/web/nativefeed/brow/?concern_id=" //不同栏目访问地址
	UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36"
)


var WukongWenda = &Spider{
	Name:        "悟空问答",
	Description: "悟空问答 各个频道专栏问题",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			//处理解析结构相同的领域
			for _,  domain := range domains{
				url := WUKONG_NORMAL_URL + domain + "&t=" +
					strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
				header := http.Header{}
				header.Add("User-Agent", UA)

				ctx.AddQueue(&request.Request{
					Url: url,
					Header: header,
					Rule: "获取结果",
				})

			}
		},


		Trunk: map[string]*Rule{
			"获取结果": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"问题标题",
					"问题描述",
					"问题回答",
					"问题url地址",
				},
				ParseFunc: func(ctx *Context) {

					type question struct{
						title string
						content string
						answer string
						url string
						offset string
					}

					var questionlist []question
					data := gjson.Get(ctx.GetText(), "data")
					more := gjson.Get(ctx.GetText(), "has_more").String()

					data.ForEach(func(key, value gjson.Result) bool{
						questionlist = append(questionlist,
							question{
								title:gjson.Get(value.String(), "question.title").String(),
								content:gjson.Get(value.String(), "question.content.text").String(),
								answer:gjson.Get(value.String(), "answer.content").String(),
								url:"https://www.wukong.com/question/" + gjson.Get(value.String(), "question.qid").String() + "/",
								offset:gjson.Get(value.String(), "behot_time").String(),
							})
						return true
					})

					if more == "true"{
						newOffset := questionlist[len(questionlist) - 1].offset
						header := http.Header{}
						header.Add("User-Agent", UA)

						visit_url := ctx.GetUrl()
						if strings.Contains(visit_url, "&max_behot_time="){
							visit_url = strings.Split(visit_url, "&max_behot_time=")[0]
						}

						ctx.AddQueue(&request.Request{
							Url: visit_url + "&max_behot_time=" + newOffset,
							Header: header,
							Rule: "获取结果",
						})

					}

					for _, v := range questionlist{
						ctx.Output(map[int]interface{}{
							0:v.title,
							1:v.content,
							2:v.answer,
							3:v.url,
						})
					}

				},
			},
		},
	},
}

