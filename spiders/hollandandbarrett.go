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
// "net/http" //http.Header
)

// 编码包
import (
	// "encoding/xml"
	"encoding/json"
)

// 字符串处理包
import (
	"regexp"
	// "strconv"
	"strings"
)

// 其他包
import (
	"fmt"
	// "math"
)

func init() {
	Hollandandbarrett.AddMenu()
}

var Hollandandbarrett = &Spider{
	Name:        "Hollandandbarrett",
	Description: "Hollandand&Barrett商品数据 [Auto Page] [www.Hollandandbarrett.com]",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AddQueue(
				map[string]interface{}{
					"url":  "http://www.hollandandbarrett.com/",
					"rule": "获取版块URL",
				},
			)
		},

		Nodes: map[string]*Rule{

			"获取版块URL": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					lis := query.Find(".footer-links nav.l-one-half a")

					lis.Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Attr("href"); ok {
							tit, _ := s.Attr("title")
							self.AddQueue(
								map[string]interface{}{
									"url":  "http://www.hollandandbarrett.com" + url + "?showAll=1&pageHa=1&es=true&vm=grid&imd=true&format=json&single=true",
									"rule": "获取总数",
									"temp": map[string]interface{}{
										"type":    tit,
										"baseUrl": url,
									},
								},
							)
						}
					})
				},
			},

			"获取总数": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {

					query := resp.GetDom()

					re, _ := regexp.Compile(`(?U)"totalNumRecs":[\d]+,`)
					total := re.FindString(query.Text())
					re, _ = regexp.Compile(`[\d]+`)
					total = re.FindString(total)
					total = strings.Trim(total, " \t\n")

					if total == "0" {
						reporter.Log.Printf("[消息提示：| 任务：%v | 关键词：%v | 规则：%v] 没有抓取到任何数据！!!\n", self.GetName(), self.GetKeyword(), resp.GetRuleName())
					} else {

						self.AddQueue(
							map[string]interface{}{
								"url":  "http://www.hollandandbarrett.com" + resp.GetTemp("baseUrl").(string) + "?showAll=" + total + "&pageHa=1&es=true&vm=grid&imd=true&format=json&single=true",
								"rule": "商品详情",
								"temp": map[string]interface{}{
									"type": resp.GetTemp("type").(string),
								},
							},
						)

					}
				},
			},

			"商品详情": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"标题",
					"原价",
					"折后价",
					"打折",
					"星级",
					"分类",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()

					src := query.Text()

					infos := map[string]interface{}{}

					err := json.Unmarshal([]byte(src), &infos)

					if err != nil {
						reporter.Log.Printf("error is %v\n", err)
						return
					} else {
						for _, info1 := range infos["contents"].([]interface{})[0].(map[string]interface{})["mainContent"].([]interface{})[0].(map[string]interface{})["records"].([]interface{}) {

							info2 := info1.(map[string]interface{})["records"].([]interface{})[0].(map[string]interface{})["attributes"].(map[string]interface{})

							var n, price1, price2, prm, level string

							if info2["Name"] == nil {
								n = ""
							} else {
								n = fmt.Sprint(info2["Name"])
								n = strings.TrimRight(n, "]")
								n = strings.TrimLeft(n, "[")
							}

							if info2["lp"] == nil {
								price1 = ""
							} else {
								price1 = fmt.Sprint(info2["lp"])
								price1 = strings.TrimRight(price1, "]")
								price1 = strings.TrimLeft(price1, "[")
							}

							if info2["sp"] == nil {
								price2 = ""
							} else {
								price2 = fmt.Sprint(info2["sp"])
								price2 = strings.TrimRight(price2, "]")
								price2 = strings.TrimLeft(price2, "[")
							}

							if info2["prm"] == nil {
								prm = ""
							} else {
								prm = fmt.Sprint(info2["prm"])
								prm = strings.TrimRight(prm, "]")
								prm = strings.TrimLeft(prm, "[")
							}

							if info2["ratingCount"] == nil {
								level = "0"
							} else {
								level = fmt.Sprint(info2["ratingCount"])
								level = strings.TrimRight(level, "]")
								level = strings.TrimLeft(level, "[")
							}

							// 结果存入Response中转
							resp.AddItem(map[string]interface{}{
								self.GetOutFeild(resp, 0): n,
								self.GetOutFeild(resp, 1): price1,
								self.GetOutFeild(resp, 2): price2,
								self.GetOutFeild(resp, 3): prm,
								self.GetOutFeild(resp, 4): level,
								self.GetOutFeild(resp, 5): resp.GetTemp("type"),
							})
						}
					}
				},
			},
		},
	},
}
