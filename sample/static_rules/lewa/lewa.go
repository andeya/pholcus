package rules

// 基础包
import (
	// "github.com/andeya/pholcus/common/goquery" //DOM解析
	"github.com/andeya/pholcus/app/downloader/request" //必需
	// "github.com/andeya/pholcus/logs"           //信息输出
	spider "github.com/andeya/pholcus/app/spider"              //必需
	spidercommon "github.com/andeya/pholcus/app/spider/common" //选用

	// net包
	"net/http" //设置http.Header
	// "net/url"
	// 编码包
	// "encoding/xml"
	// "encoding/json"
	// 字符串处理包
	// "regexp"
	// "strconv"
	// "strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Lewa.Register()
}

var Lewa = &spider.Spider{
	Name:        "乐蛙登录测试",
	Description: "乐蛙登录测试 [Auto Page] [http://accounts.lewaos.com]",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: true,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{URL: "http://accounts.lewaos.com/", Rule: "登录页"})
		},

		Trunk: map[string]*spider.Rule{

			"登录页": {
				ParseFunc: func(ctx *spider.Context) {
					// ctx.AddQueue(&request.Request{
					// 	URL:    "http://accounts.lewaos.com",
					// 	Rule:   "登录后",
					// 	Method: "POST",
					// 	PostData: "username=123456@qq.com&password=123456&login_btn=login_btn&submit=login_btn",
					// })
					spidercommon.NewForm(
						ctx,
						"登录后",
						"http://accounts.lewaos.com",
						ctx.GetDom().Find(".userlogin.lw-pl40"),
					).Inputs(map[string]string{
						"username": "",
						"password": "",
					}).Submit()
				},
			},
			"登录后": {
				ParseFunc: func(ctx *spider.Context) {
					// 结果存入Response中转
					ctx.Output(map[string]interface{}{
						"Body":   ctx.GetText(),
						"Cookie": ctx.GetCookie(),
					})
					ctx.AddQueue(&request.Request{
						URL:    "http://accounts.lewaos.com/member",
						Rule:   "个人中心",
						Header: http.Header{"Referer": []string{ctx.GetURL()}},
					})
				},
			},
			"个人中心": {
				ParseFunc: func(ctx *spider.Context) {
					// 结果存入Response中转
					ctx.Output(map[string]interface{}{
						"Body":   ctx.GetText(),
						"Cookie": ctx.GetCookie(),
					})
				},
			},
		},
	},
}
