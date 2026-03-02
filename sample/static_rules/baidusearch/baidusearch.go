package rules

import (
	"net/url"
	"strings"

	"github.com/andeya/pholcus/app/downloader/request"
	spider "github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/common/goquery"
)

func init() {
	BaiduSearch.Register()
}

var BaiduSearch = &spider.Spider{
	Name:            "百度搜索",
	Description:     "百度搜索结果 [www.baidu.com]",
	Keyin:           spider.KEYIN,
	Limit:           spider.LIMIT,
	EnableCookie:    true,
	NotDefaultField: true,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{
				URL:          "https://www.baidu.com/s?wd=" + url.QueryEscape(ctx.GetKeyin()) + "&pn=0",
				Rule:         "搜索结果",
				DownloaderID: request.ChromeID,
			})
		},

		Trunk: map[string]*spider.Rule{
			"搜索结果": {
				ItemFields: []string{
					"标题",
					"链接",
					"摘要",
				},
				ParseFunc: func(ctx *spider.Context) {
					query := ctx.GetDom()
					query.Find("div.result,div.result-op").Each(func(i int, s *goquery.Selection) {
						title := strings.TrimSpace(s.Find("h3.t a").Text())
						href := s.Find("h3.t a").AttrOr("href", "")
						summary := strings.TrimSpace(s.Find("[data-module=abstract]").Text())

						if title == "" || href == "" {
							return
						}

						ctx.Output(map[int]interface{}{
							0: title,
							1: href,
							2: summary,
						})
					})

					nextHref := query.Find("a.n").Last().AttrOr("href", "")
					if nextHref != "" {
						ctx.AddQueue(&request.Request{
							URL:          "https://www.baidu.com" + nextHref,
							Rule:         "搜索结果",
							DownloaderID: request.ChromeID,
						})
					}
				},
			},
		},
	},
}
