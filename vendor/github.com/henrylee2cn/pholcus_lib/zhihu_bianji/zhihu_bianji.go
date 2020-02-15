package zhihu_bianji

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	// . "github.com/henrylee2cn/pholcus/app/spider/common"    //选用
	//"github.com/henrylee2cn/pholcus/logs" //信息输出

	// net包
	"net/http" //设置http.Header
	"net/url"

	// 编码包
	// "encoding/xml"
	"encoding/json"

	// 字符串处理包
	//"strconv"

	// 其他包
	// "fmt"
	// "time"
	//"strconv"
	"io/ioutil"
	"strings"
	"strconv"
	"regexp"
	"math"
)

func init() {
	ZhihuBianji.Register()
}

var urlList []string

var ZhihuBianji = &Spider{
	Name:        "知乎编辑推荐",
	Description: "知乎编辑推荐",
	Pausetime:    300,
	//Keyin:        KEYIN,
	Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url: "https://www.zhihu.com/explore/recommendations",
				Rule: "知乎编辑推荐",
			})


		},

		Trunk: map[string]*Rule{
			"知乎编辑推荐": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					regular := "#zh-recommend-list-full .zh-general-list .zm-item h2 a";
					query.Find(regular).
						Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Attr("href"); ok {
							url = changeToAbspath(url)
							ctx.AddQueue(&request.Request{Url: url, Rule: "解析落地页"})
						}})

					limit := ctx.GetLimit()

					if len(query.Find(regular).Nodes) < limit	{
						total := int(math.Ceil(float64(limit) / float64(20)))
						ctx.Aid(map[string]interface{}{
							"loop": [2]int{1, total},
							"Rule": "知乎编辑推荐翻页",
						}, "知乎编辑推荐翻页")
					}
				},
			},

			"知乎编辑推荐翻页": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
					for loop := aid["loop"].([2]int); loop[0] < loop[1]; loop[0]++ {
						offset := loop[0] * 20
						header := make(http.Header)
						header.Set("Content-Type", "application/x-www-form-urlencoded")
						ctx.AddQueue(&request.Request{
							Url:  "https://www.zhihu.com/node/ExploreRecommendListV2",
							Rule: aid["Rule"].(string),
							Method: "POST",
							Header: header,
							PostData: url.Values{"method":{"next"}, "params":{`{"limit":20,"offset":` + strconv.Itoa(offset) + `}`}}.Encode(),
							Reloadable: true,
						})
					}

					return nil
				},
				ParseFunc: func(ctx *Context) {
					type Items struct {
						R int `json:"r"`
						Msg []interface{} `json:"msg"`
					}

					content, err := ioutil.ReadAll(ctx.GetResponse().Body)

					ctx.GetResponse().Body.Close()

					if err != nil {
						ctx.Log().Error(err.Error());
					}

					e := new(Items)

					err = json.Unmarshal(content, e)

					html := ""

					for _, v := range e.Msg{
						msg, ok := v.(string)
						if ok {
							html = html + "\n" + msg
						}
					}


					ctx = ctx.ResetText(html)

					query := ctx.GetDom()

					query.Find(".zm-item h2 a").Each(func(i int, selection *goquery.Selection){
						if url, ok := selection.Attr("href"); ok {
							url = changeToAbspath(url)
							if filterZhihuAnswerURL(url){
								ctx.AddQueue(&request.Request{Url: url, Rule: "解析知乎问答落地页"})
							}else{
								ctx.AddQueue(&request.Request{Url: url, Rule: "解析知乎文章落地页"})
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
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					questionHeader := query.Find(".QuestionPage .QuestionHeader .QuestionHeader-content")
					//headerSide := questionHeader.Find(".QuestionHeader-side")
					headerMain := questionHeader.Find(".QuestionHeader-main")

					// 获取问题标题
					title := headerMain.Find(".QuestionHeader-title").Text()

					// 获取问题描述
					content := headerMain.Find(".QuestionHeader-detail span").Text()

					answerMain := query.Find(".QuestionPage .Question-main")

					answer, _ := answerMain.Find(".AnswerCard .QuestionAnswer-content .ContentItem .RichContent .RichContent-inner").First().Html()

					// 结果存入Response中转
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
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					// 获取问题标题
					title,_ := query.Find(".PostIndex-title.av-paddingSide.av-titleFont").Html()

					// 获取问题描述
					content, _ := query.Find(".RichText.PostIndex-content.av-paddingSide.av-card").Html()

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: content,
					})

				},
			},
		},
	},
}

//将相对路径替换为绝对路径
func changeToAbspath(url string)string{
	if strings.HasPrefix(url, "https://"){
		return url
	}
	return "https://www.zhihu.com" + url
}

//判断是用户回答的问题，还是知乎专栏作家书写的文章
func filterZhihuAnswerURL(url string) bool{
	return regexp.MustCompile(`^https:\/\/www\.zhihu\.com\/question\/\d{1,}(\/answer\/\d{1,})?$`).MatchString(url)
}