package spiders

// 基础包
import (
	// "github.com/PuerkitoBio/goquery" //DOM解析
	"github.com/henrylee2cn/pholcus/crawl/downloader/context" //必需
	// "github.com/henrylee2cn/pholcus/reporter"           //信息输出
	. "github.com/henrylee2cn/pholcus/spider"        //必需
	. "github.com/henrylee2cn/pholcus/spider/common" //选用
)

import (
// "net/http" //http.Header
// "net/url"
)

// 编码包
import (
// "encoding/xml"
// "encoding/json"
)

// 字符串处理包
import (
// "regexp"
// "strconv"
// "strings"
)

// 其他包
import (
// "fmt"
// "math"
)

func init() {
	Lewa.AddMenu()
}

var Lewa = &Spider{
	Name:        "乐蛙登录测试",
	Description: "乐蛙登录测试 [Auto Page] [http://accounts.lewaos.com]",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AddQueue(map[string]interface{}{"url": "http://accounts.lewaos.com/", "rule": "登录页"})
		},

		Nodes: map[string]*Rule{

			"登录页": &Rule{
				ParseFunc: func(self *Spider, resp *context.Response) {
					// self.AddQueue(map[string]interface{}{
					// 	"url":    "http://accounts.lewaos.com",
					// 	"rule":   "登录后",
					// 	"method": "POST",
					// 	"postData": url.Values{
					// 		"username":  []string{""},
					// 		"password":  []string{""},
					// 		"login_btn": []string{"login_btn"},
					// 		"submit":    []string{"login_btn"},
					// 	},
					// })
					NewForm(
						self,
						"登录后",
						"http://accounts.lewaos.com",
						resp.GetDom().Find(".userlogin.lw-pl40"),
					).Inputs(map[string]string{
						"username": "",
						"password": "",
					}).Submit()
				},
			},
			"登录后": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"全部",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					// 结果存入Response中转
					resp.AddItem(map[string]interface{}{
						self.GetOutFeild(resp, 0): resp.GetText(),
					})
				},
			},
		},
	},
}
