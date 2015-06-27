package crawl

import (
	"github.com/henrylee2cn/pholcus/spider"
)

type Crawler interface {
	Init(*spider.Spider) Crawler
	Start()
	GetId() int
}
