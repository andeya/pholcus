package rules

// base packages
import (
	// "github.com/andeya/pholcus/common/goquery"                          // DOM parsing
	"github.com/andeya/pholcus/app/downloader/request" // required
	spider "github.com/andeya/pholcus/app/spider"      // required
	// . "github.com/andeya/pholcus/app/spider/common" // optional
	// "github.com/andeya/pholcus/logs"
	// net packages
	// "net/http" // set http.Header
	// "net/url"
	// encoding packages
	// "encoding/xml"
	//"encoding/json"
	// string processing packages
	//"regexp"
	// "strconv"
	//	"strings"
	// other packages
	// "fmt"
	// "math"
	// "time"
)

func init() {
	FileTest.Register()
}

var FileTest = &spider.Spider{
	Name:        "文件下载测试",
	Description: "文件下载测试",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			ctx.AddQueue(&request.Request{
				URL:          "https://www.baidu.com/img/bd_logo1.png",
				Rule:         "百度图片",
				ConnTimeout:  -1,
				DownloaderID: 0, // media files must use 0 (surfer: native Go downloader)
			})
			ctx.AddQueue(&request.Request{
				URL:          "https://github.com/andeya/pholcus",
				Rule:         "Pholcus页面",
				ConnTimeout:  -1,
				DownloaderID: 0, // text files can use 0 or 1 (0: surfer surf go native; 1: surfer phantomjs kernel)
			})
		},

		Trunk: map[string]*spider.Rule{

			"百度图片": {
				ParseFunc: func(ctx *spider.Context) {
					ctx.FileOutput("baidu") // equivalent to ctx.AddFile("baidu")
				},
			},
			"Pholcus页面": {
				ParseFunc: func(ctx *spider.Context) {
					ctx.FileOutput() // equivalent to ctx.AddFile()
				},
			},
		},
	},
}
