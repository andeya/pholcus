// 数据收集
package pipeline

import (
	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"github.com/henrylee2cn/pholcus/app/pipeline/collector/data"
	"github.com/henrylee2cn/pholcus/app/spider"
)

// 数据收集/输出管道
type Pipeline interface {
	Start()                          //启动
	Stop()                           //停止
	CollectData(data.DataCell) error //收集数据单元
	CollectFile(data.FileCell) error //收集文件
}

func New(sp *spider.Spider) Pipeline {
	return collector.NewCollector(sp)
}
