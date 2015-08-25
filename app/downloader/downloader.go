package downloader

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"time"
)

// The Downloader interface.
// You can implement the interface by implement function Download.
// Function Download need to return Page instance pointer that has request result downloaded from Request.
type Downloader interface {
	Download(req *context.Request) *context.Response
	SetUseCookie(use bool) Downloader
	SetPaseTime(paseTime time.Duration) Downloader
	SetProxy(proxy string) Downloader
}
