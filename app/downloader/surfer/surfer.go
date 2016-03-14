// surfer是一款Go语言编写的高并发爬虫下载器，支持 GET/POST/HEAD 方法及 http/https 协议，同时支持固定UserAgent自动保存cookie与随机大量UserAgent禁用cookie两种模式，高度模拟浏览器行为，可实现模拟登录等功能。
package surfer

import (
	"net/http"
	"os"
	"sync"
)

var (
	surf          Surfer
	phantom       Surfer
	once_surf     sync.Once
	once_phantom  sync.Once
	tempJsDir     = "./tmp"
	phantomjsFile = os.Getenv("GOPATH") + `\src\github.com\henrylee2cn\surfer\phantomjs\phantomjs`
)

func Download(req Request) (resp *http.Response, err error) {
	switch req.GetDownloaderID() {
	case SurfID:
		once_surf.Do(func() { surf = New() })
		resp, err = surf.Download(req)
	case PhomtomJsID:
		once_phantom.Do(func() { phantom = NewPhantom(phantomjsFile, tempJsDir) })
		resp, err = phantom.Download(req)
	}
	return
}

//销毁Phantomjs的js临时文件
func DestroyJsFiles() {
	if pt, ok := phantom.(*Phantom); ok {
		pt.DestroyJsFiles()
	}
}

// Downloader represents an core of HTTP web browser for crawler.
type Surfer interface {
	// GET @param url string, header http.Header, cookies []*http.Cookie
	// HEAD @param url string, header http.Header, cookies []*http.Cookie
	// POST PostForm @param url, referer string, values url.Values, header http.Header, cookies []*http.Cookie
	// POST-M PostMultipart @param url, referer string, values url.Values, header http.Header, cookies []*http.Cookie
	Download(Request) (resp *http.Response, err error)
}
