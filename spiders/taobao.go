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
	"strconv"
	"strings"
)

// 其他包
import (
// "fmt"
// "math"
)

func init() {
	Taobao.AddMenu()
}

var cookies_Taobao = SplitCookies("mt=ci%3D-1_0; swfstore=35673; thw=cn; cna=fcr5DRDmwnQCAT2QxZSu3Db6; sloc=%E8%BE%BD%E5%AE%81; _tb_token_=XLlMHhT9BI8IzeA; ck1=; v=0; uc3=nk2=symxAo6NBazVq7cY2z0%3D&id2=UU23CgHxOwgwgA%3D%3D&vt3=F8dAT%2BCFEEyTLicOBEc%3D&lg2=U%2BGCWk%2F75gdr5Q%3D%3D; existShop=MTQzNDM1NDcyNg%3D%3D; lgc=%5Cu5C0F%5Cu7C73%5Cu7C92%5Cu559C%5Cu6B22%5Cu5927%5Cu6D77; tracknick=%5Cu5C0F%5Cu7C73%5Cu7C92%5Cu559C%5Cu6B22%5Cu5927%5Cu6D77; sg=%E6%B5%B721; cookie2=1433b814776e3b3c61f4ba3b8631a81a; cookie1=Bqbn0lh%2FkPm9D0NtnTdFiqggRYia%2FBrNeQpwLWlbyJk%3D; unb=2559173312; t=1a9b12bb535040723808836b32e53507; _cc_=WqG3DMC9EA%3D%3D; tg=5; _l_g_=Ug%3D%3D; _nk_=%5Cu5C0F%5Cu7C73%5Cu7C92%5Cu559C%5Cu6B22%5Cu5927%5Cu6D77; cookie17=UU23CgHxOwgwgA%3D%3D; mt=ci=0_1; x=e%3D1%26p%3D*%26s%3D0%26c%3D0%26f%3D0%26g%3D0%26t%3D0%26__ll%3D-1%26_ato%3D0; whl=-1%260%260%260; uc1=lltime=1434353890&cookie14=UoW0FrfFYp27FQ%3D%3D&existShop=false&cookie16=V32FPkk%2FxXMk5UvIbNtImtMfJQ%3D%3D&cookie21=U%2BGCWk%2F7p4mBoUyTltGF&tag=7&cookie15=Vq8l%2BKCLz3%2F65A%3D%3D&pas=0; isg=C08C1D752BC08A3DCDF1FE6611FA3EE1; l=Ajk53TTUeK0ZKkG8yx7w7svcyasSxC34")

var Taobao = &Spider{
	Name:        "淘宝数据",
	Description: "淘宝天猫商品数据 [Auto Page] [http://list.taobao.com/]",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			if k := strings.Trim(self.GetKeyword(), " "); k != "" {
				cookies_Taobao = SplitCookies(k)
			}
			self.AddQueue(map[string]interface{}{
				"url":     "http://list.taobao.com/browse/cat-0.htm",
				"rule":    "生成请求",
				"cookies": cookies_Taobao,
			})
		},

		Nodes: map[string]*Rule{

			"生成请求": &Rule{
				AidFunc: func(self *Spider, aid map[string]interface{}) interface{} {
					self.LoopAddQueue(
						aid["loop"].([2]int),
						func(i int) []string {
							urls := []string{}
							for _, loc := range loc_Taobao {
								urls = append(urls, "http:"+aid["urlBase"].(string)+"&_input_charset=utf-8&json=on&viewIndex=1&as=0&atype=b&style=grid&same_info=1&tid=0&isnew=2&data-action&module=page&s=0&loc="+loc+"&pSize=96&data-key=s&data-value="+strconv.Itoa(i*96))
							}
							return urls
						},
						map[string]interface{}{
							"rule":     aid["rule"],
							"respType": "text",
							"cookies":  cookies_Taobao,
							"temp":     aid["temp"],
						},
					)
					return nil
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					query.Find(".J_TBMarketCat").Each(func(i int, a *goquery.Selection) {
						type1 := a.Find("h4").Text()
						a.Find(".section").Each(func(i int, b *goquery.Selection) {
							type2 := b.Find(".subtitle a").Text()
							b.Find(".sublist a").Each(func(i int, c *goquery.Selection) {
								type3 := c.Text()
								href3, _ := c.Attr("href")

								self.AidRule("生成请求", map[string]interface{}{
									"loop":    [2]int{0, 1},
									"urlBase": href3,
									"rule":    "列表页数",
									"temp": map[string]interface{}{
										"type1": type1,
										"type2": type2,
										"type3": type3,
									},
								})
							})
						})
					})
				},
			},

			"列表页数": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					json := resp.GetText()
					re, _ := regexp.Compile(`(?U)"totalPage":"[\d]+",`)
					total := re.FindString(json)
					re, _ = regexp.Compile(`[\d]+`)
					total = re.FindString(total)
					total = strings.Trim(total, " \t\n")
					totalPage, _ := strconv.Atoi(total)
					if total == "0" {
						reporter.Log.Printf("[消息提示：| 任务：%v | 关键词：%v | 规则：%v] 没有抓取到任何数据！!!\n", self.GetName(), self.GetKeyword(), resp.GetRuleName())
					} else {
						self.AidRule("生成请求", map[string]interface{}{
							"loop":    [2]int{1, totalPage},
							"urlBase": resp.GetUrl(),
							"rule":    "商品列表",
							"temp":    resp.GetTemps(),
						})
						self.CallRule("商品列表", resp)
					}
				},
			},

			"商品列表": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					j := resp.GetText()
					// re, _ := regexp.Compile(`null`)
					// j = re.ReplaceAllString(j, " ")

					infos := map[string]interface{}{}
					err := json.Unmarshal([]byte(j), &infos)
					if err != nil {
						reporter.Log.Printf("商品列表解析错误： %v\n", err)
						return
					}
					if infos["mallItemList"] == nil {
						reporter.Log.Println("商品列表解析错误： 内容不存在！")
						return
					}
					for _, item := range infos["mallItemList"].([]interface{}) {
						item2 := item.(map[string]interface{})
						temp := map[string]interface{}{
							self.ShowOutFeild("结果", 0):  item2["title"],
							self.ShowOutFeild("结果", 1):  item2["price"],
							self.ShowOutFeild("结果", 2):  item2["currentPrice"],
							self.ShowOutFeild("结果", 3):  item2["vipPrice"],
							self.ShowOutFeild("结果", 4):  item2["unitPrice"],
							self.ShowOutFeild("结果", 5):  item2["unit"],
							self.ShowOutFeild("结果", 6):  item2["isVirtual"],
							self.ShowOutFeild("结果", 7):  item2["ship"],
							self.ShowOutFeild("结果", 8):  item2["tradeNum"],
							self.ShowOutFeild("结果", 9):  item2["formatedNum"],
							self.ShowOutFeild("结果", 10): item2["nick"],
							self.ShowOutFeild("结果", 11): item2["sellerId"],
							self.ShowOutFeild("结果", 12): item2["guarantee"],
							self.ShowOutFeild("结果", 13): item2["itemId"],
							self.ShowOutFeild("结果", 14): item2["isLimitPromotion"],
							self.ShowOutFeild("结果", 15): item2["loc"],
							self.ShowOutFeild("结果", 16): "http:" + item2["storeLink"].(string),
							self.ShowOutFeild("结果", 17): "http:" + item2["href"].(string),
							self.ShowOutFeild("结果", 18): item2["commend"],
							self.ShowOutFeild("结果", 19): item2["source"],
							self.ShowOutFeild("结果", 20): item2["ratesum"],
							self.ShowOutFeild("结果", 21): item2["goodRate"],
							self.ShowOutFeild("结果", 22): item2["dsrScore"],
							self.ShowOutFeild("结果", 23): item2["spSource"],
						}
						self.AddQueue(map[string]interface{}{
							"url":      "http:" + item2["storeLink"].(string),
							"rule":     "商品详情",
							"temp":     temp,
							"priority": uint(1),
						})

						// 去"结果"规则输出结果
						// resp.SetAllTemps(temp)
						// self.CallRule("结果", resp)
					}
				},
			},

			"商品详情": &Rule{

				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()

					// 商品规格参数
					detail := make(map[string]string)

					if li := query.Find(".attributes-list ul li"); len(li.Nodes) != 0 {
						// 天猫店宝贝详情
						li.Each(func(i int, s *goquery.Selection) {
							native := s.Text()
							slice := strings.Split(native, ":&nbsp;")
							//空格替换为分隔号“|”
							slice[1] = strings.Replace(slice[1], "&nbsp;", "&#124;", -1)
							detail[slice[0]] = UnicodeToUTF8(slice[1])
						})

					} else {
						// 淘宝店宝贝详情
						query.Find(".attributes-list li").Each(func(i int, s *goquery.Selection) {
							native := s.Text()
							slice := strings.Split(native, ": ")
							detail[slice[0]] = slice[1]
						})
					}
					temp := resp.GetTemps()
					temp[self.ShowOutFeild("结果", 24)] = detail
					temp[self.ShowOutFeild("结果", 25)] = []interface{}{}
					self.AddQueue(map[string]interface{}{
						"rule":     "商品评论",
						"url":      "http://rate.taobao.com/feedRateList.htm?siteID=4&rateType=&orderType=sort_weight&showContent=1&userNumId=" + resp.GetTemp("sellerId").(string) + "&auctionNumId=" + resp.GetTemp("itemId").(string) + "&currentPageNum=1",
						"temp":     temp,
						"priority": uint(2),
					})
				},
			},

			"商品评论": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					j := resp.GetText()
					j = strings.TrimLeft(j, "(")
					j = strings.TrimRight(j, ")")

					infos := map[string]interface{}{}
					if err := json.Unmarshal([]byte(j), &infos); err != nil {
						reporter.Log.Printf("商品评论解析错误： %v\n", err)
						return
					}
					if infos["comments"] == nil || infos["maxPage"] == nil || infos["currentPageNum"] == nil {
						reporter.Log.Println("商品评论解析错误： 内容不存在！")
						return
					}
					discussSlice := infos["comments"].([]interface{})
					discussAll := resp.GetTemp(self.ShowOutFeild("结果", 25)).([]interface{})
					discussAll = append(discussAll, discussSlice...)
					resp.SetTemp(self.ShowOutFeild("结果", 25), discussAll)

					currentPageNum := infos["currentPageNum"].(int)
					maxPage := infos["maxPage"].(int)
					if currentPageNum < maxPage {
						// 请求下一页
						self.AddQueue(map[string]interface{}{
							"rule": "商品评论",
							"url":  "http://rate.taobao.com/feedRateList.htm?siteID=4&rateType=&orderType=sort_weight&showContent=1&userNumId=" + resp.GetTemp("sellerId").(string) + "&auctionNumId=" + resp.GetTemp("itemId").(string) + "&currentPageNum=" + strconv.Itoa(currentPageNum+1),
							"temp": resp.GetTemps(),
						})
					} else {
						// 输出结果
						self.CallRule("结果", resp)
					}
				},
			},

			"结果": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"标题",               //title
					"原价",               //price
					"现价",               //currentPrice
					"会员价",              //vipPrice
					"单价",               //unitPrice
					"单位",               //unit
					"是否虚拟物品",           //isVirtual
					"ship",             //ship
					"tradeNum",         //tradeNum
					"formatedNum",      //formatedNum
					"店铺",               //nick
					"店铺ID",             //sellerId
					"guarantee",        //guarantee
					"货号",               //itemId
					"isLimitPromotion", //isLimitPromotion
					"发货地",              //loc
					"店铺链接",             //storeLink
					"商品链接",             //href
					"评价",               //commend
					"source",           //source
					"店铺信誉",             //ratesum
					"店铺好评率",            //goodRate
					"dsrScore",         //dsrScore
					"spSource",         //spSource
					"规格参数",
					"评论内容",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					// 结果存入Response中转
					resp.AddItem(resp.GetTemps())
				},
			},
		},
	},
}

var (
	loc_Taobao = map[string]string{
		// "北京": "%E5%8C%97%E4%BA%AC",
		// "上海": "%E4%B8%8A%E6%B5%B7",
		// "广州":   "%E5%B9%BF%E5%B7%9E",
		// "深圳":   "%E6%B7%B1%E5%9C%B3",
		// "杭州":   "%E6%9D%AD%E5%B7%9E",
		// "海外": "%E7%BE%8E%E5%9B%BD%2C%E8%8B%B1%E5%9B%BD%2C%E6%B3%95%E5%9B%BD%2C%E7%91%9E%E5%A3%AB%2C%E6%BE%B3%E6%B4%B2%2C%E6%96%B0%E8%A5%BF%E5%85%B0%2C%E5%8A%A0%E6%8B%BF%E5%A4%A7%2C%E5%A5%A5%E5%9C%B0%E5%88%A9%2C%E9%9F%A9%E5%9B%BD%2C%E6%97%A5%E6%9C%AC%2C%E5%BE%B7%E5%9B%BD%2C%E6%84%8F%E5%A4%A7%E5%88%A9%2C%E8%A5%BF%E7%8F%AD%E7%89%99%2C%E4%BF%84%E7%BD%97%E6%96%AF%2C%E6%B3%B0%E5%9B%BD%2C%E5%8D%B0%E5%BA%A6%2C%E8%8D%B7%E5%85%B0%2C%E6%96%B0%E5%8A%A0%E5%9D%A1%2C%E5%85%B6%E5%AE%83%E5%9B%BD%E5%AE%B6",
		// "江浙沪":  "%E6%B1%9F%E8%8B%8F%2C%E6%B5%99%E6%B1%9F%2C%E4%B8%8A%E6%B5%B7",
		// "珠三角":  "%E5%B9%BF%E5%B7%9E%2C%E6%B7%B1%E5%9C%B3%2C%E4%B8%AD%E5%B1%B1%2C%E7%8F%A0%E6%B5%B7%2C%E4%BD%9B%E5%B1%B1%2C%E4%B8%9C%E8%8E%9E%2C%E6%83%A0%E5%B7%9E",
		// "京津冀":  "%E5%8C%97%E4%BA%AC%2C%E5%A4%A9%E6%B4%A5%2C%E6%B2%B3%E5%8C%97",
		// "东三省":  "%E9%BB%91%E9%BE%99%E6%B1%9F%2C%E5%90%89%E6%9E%97%2C%E8%BE%BD%E5%AE%81",
		// "港澳台":  "%E9%A6%99%E6%B8%AF%2C%E6%BE%B3%E9%97%A8%2C%E5%8F%B0%E6%B9%BE",
		// "江浙沪皖": "%E6%B1%9F%E8%8B%8F%2C%E6%B5%99%E6%B1%9F%2C%E4%B8%8A%E6%B5%B7%2C%E5%AE%89%E5%BE%BD",
		// "长沙":   "%E9%95%BF%E6%B2%99",
		// "长春":   "%E9%95%BF%E6%98%A5",
		// "成都":   "%E6%88%90%E9%83%BD",
		// "重庆": "%E9%87%8D%E5%BA%86",
		// "大连":   "%E5%A4%A7%E8%BF%9E",
		// "东莞":   "%E4%B8%9C%E8%8E%9E",
		// "福州":   "%E7%A6%8F%E5%B7%9E",
		// "合肥":   "%E5%90%88%E8%82%A5",
		// "济南":   "%E6%B5%8E%E5%8D%97",
		// "嘉兴":   "%E5%98%89%E5%85%B4",
		// "昆明":   "51108009&loc=%E6%98%86%E6%98%8E",
		// "宁波":   "%E5%AE%81%E6%B3%A2",
		// "南京":   "%E5%8D%97%E4%BA%AC",
		// "南昌":   "%E5%8D%97%E6%98%8C",
		// "青岛":   "%E9%9D%92%E5%B2%9B",
		// "苏州":   "%E8%8B%8F%E5%B7%9E",
		// "沈阳":   "%E6%B2%88%E9%98%B3",
		// "天津": "%E5%A4%A9%E6%B4%A5",
		// "温州":   "%E6%B8%A9%E5%B7%9E",
		// "无锡":   "%E6%97%A0%E9%94%A1",
		// "武汉":   "%E6%AD%A6%E6%B1%89",
		// "西安":   "%E8%A5%BF%E5%AE%89",
		// "厦门":   "%E5%8E%A6%E9%97%A8",
		// "郑州":   "%E9%83%91%E5%B7%9E",
		// "中山":   "%E4%B8%AD%E5%B1%B1",
		// "石家庄":  "%E7%9F%B3%E5%AE%B6%E5%BA%84",
		// "哈尔滨":  "%E5%93%88%E5%B0%94%E6%BB%A8",
		// 省级
		// "安徽":  "%E5%AE%89%E5%BE%BD",
		// "福建":  "%E7%A6%8F%E5%BB%BA",
		// "甘肃":  "%E7%94%98%E8%82%83",
		// "广东":  "%E5%B9%BF%E4%B8%9C",
		// "广西":  "%E5%B9%BF%E8%A5%BF",
		// "贵州":  "%E8%B4%B5%E5%B7%9E",
		// "河北":  "%E6%B2%B3%E5%8C%97",
		// "河南":  "%E6%B2%B3%E5%8D%97",
		// "湖北":  "%E6%B9%96%E5%8C%97",
		// "湖南":  "%E6%B9%96%E5%8D%97",
		// "海南":  "%E6%B5%B7%E5%8D%97",
		// "江苏":  "%E6%B1%9F%E8%8B%8F",
		// "江西":  "%E6%B1%9F%E8%A5%BF",
		// "吉林":  "%E5%90%89%E6%9E%97",
		// "辽宁":  "%E8%BE%BD%E5%AE%81",
		// "宁夏":  "%E5%AE%81%E5%A4%8F",
		// "青海":  "%E9%9D%92%E6%B5%B7",
		// "山东":  "%E5%B1%B1%E4%B8%9C",
		// "山西":  "%E5%B1%B1%E8%A5%BF",
		// "陕西":  "%E9%99%95%E8%A5%BF",
		// "四川":  "%E5%9B%9B%E5%B7%9D",
		// "西藏":  "%E8%A5%BF%E8%97%8F",
		// "新疆":  "%E6%96%B0%E7%96%86",
		// "云南":  "%E4%BA%91%E5%8D%97",
		// "浙江":  "%E6%B5%99%E6%B1%9F",
		// "澳门":  "%E6%BE%B3%E9%97%A8",
		// "香港":  "%E9%A6%99%E6%B8%AF",
		// "台湾":  "%E5%8F%B0%E6%B9%BE",
		// "内蒙古": "%E5%86%85%E8%92%99%E5%8F%A4",
		// "黑龙江": "%E9%BB%91%E9%BE%99%E6%B1%9F",
		"": "",
	}
)
