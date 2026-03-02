package rules

// 基础包
import (
	"github.com/andeya/pholcus/app/downloader/request" //必需
	spider "github.com/andeya/pholcus/app/spider"      //必需

	//. "github.com/andeya/pholcus/app/spider/common"    //选用
	"github.com/andeya/pholcus/common/goquery" //DOM解析

	//信息输出
	// net包
	//设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	// "regexp"

	"strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	AreaCodes2018.Register()
}

/*
-- 数据清洗

SET SQL_SAFE_UPDATES = 0;
-- 去重
delete from 2018年统计用区划代码和城乡划分代码__0__市 where id not in (select temp.id from (select min(id) as id from 2018年统计用区划代码和城乡划分代码__0__市 group by 代码) as temp);

-- 合并表
CREATE TABLE area_codes
select 名称 as name,RPAD(代码,12,'0') as area_code,级别 as level,RPAD(上级,12,'0') as parent from 2018年统计用区划代码和城乡划分代码__0__省
UNION
select 名称 as name,RPAD(代码,12,'0') as area_code,级别 as level,RPAD(上级,12,'0') as parent from 2018年统计用区划代码和城乡划分代码__0__市;
*/

// AreaCodes2018 2018年统计用区划代码和城乡划分代码
//
// creatTime:2019年09月06日 09:23:55
// author:hailaz
var AreaCodes2018 = &spider.Spider{
	Name:        "2018年统计用区划代码和城乡划分代码",
	Description: "2018年统计用区划代码和城乡划分代码。间隔不要小于100ms，不然容易触发验证码导致失败。总数据大概71万（暂停时长100ms，耗时2小时），所以适当做数据分批输出，不然出现内存溢出。",
	// Pausetime:   50,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{
				Url:  "http://www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/2018/index.html",
				Rule: "省",
			})
		},

		Trunk: map[string]*spider.Rule{
			"省": {
				ItemFields: []string{
					"名称",
					"代码",
					"级别",
					"上级",
				},
				ParseFunc: func(ctx *spider.Context) {
					baseUrl := ctx.GetRequest().Url
					baseUrl = baseUrl[:strings.LastIndex(baseUrl, "/")+1]
					query := ctx.GetDom()
					//cc := 0
					query.Find("tr.provincetr").Each(func(i int, tr *goquery.Selection) {
						//cc++
						tr.Find("td a").Each(func(j int, a *goquery.Selection) {
							if url := a.Attr("href"); url.IsSome() {
								u := url.Unwrap()
								code := strings.Split(u, ".")[0]
								u = baseUrl + u
								//fmt.Println("0", a.Text()+":"+url)
								ctx.Output(map[int]interface{}{
									0: a.Text(),
									1: code,
									2: 0,
									3: 0,
								})
								ctx.AddQueue(&request.Request{Url: u, Rule: "市", Temp: request.Temp{"level": 0, "parent": code}})
							}
						})
					})
					//fmt.Println(cc) // 等于零，证明请求太过频繁，需要输入验证码
				},
			},
			"市": {
				ItemFields: []string{
					"名称",
					"代码",
					"级别",
					"上级",
				},
				ParseFunc: func(ctx *spider.Context) {
					baseUrl := ctx.GetRequest().Url
					baseUrl = baseUrl[:strings.LastIndex(baseUrl, "/")+1]
					level := ctx.GetRequest().Temp["level"].(int) + 1
					parent := ctx.GetRequest().Temp["parent"].(string)
					query := ctx.GetDom()
					if level == 4 {
						myCode := ""
						query.Find("tr.villagetr td").Each(func(i int, td *goquery.Selection) {
							if i%3 == 0 {
								myCode = td.Text()
							}
							if i%3 == 2 {
								ctx.Output(map[int]interface{}{
									0: td.Text(),
									1: myCode,
									2: level,
									3: parent,
								})
								//fmt.Println(level, td.Text(), myCode)
							}
						})
					} else {
						myCode := ""
						query.Find("tr td a").Each(func(i int, a *goquery.Selection) {
							if i%2 == 0 {
								myCode = a.Text()
							}
							if i%2 == 1 {
								if url := a.Attr("href"); url.IsSome() {
									u := url.Unwrap()
									code := strings.Split(strings.Split(u, "/")[1], ".")[0]
									u = baseUrl + u
									ctx.Output(map[int]interface{}{
										0: a.Text(),
										1: myCode,
										2: level,
										3: parent,
									})
									//fmt.Println(level, a.Text(), myCode)
									ctx.AddQueue(&request.Request{Url: u, Rule: "市", Temp: request.Temp{"level": level, "parent": code}})
								}
							}
						})
					}
				},
			},
		},
	},
}
