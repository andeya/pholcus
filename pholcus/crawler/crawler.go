package crawler

import (
	"github.com/henrylee2cn/pholcus/spiders/spider"
)

type Crawler interface {
	Init(*spider.Spider) Crawler
	Start()
}
