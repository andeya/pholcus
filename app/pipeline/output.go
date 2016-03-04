package pipeline

import (
	"sort"

	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

// 初始化输出方式列表collector.OutputLib
func init() {
	for out, _ := range collector.Output {
		collector.OutputLib = append(collector.OutputLib, out)
	}
	sort.Strings(collector.OutputLib)
}

// 刷新输出方式的状态
func RefreshOutput() {
	switch cache.Task.OutType {
	case "mgo":
		mgo.Refresh()
	case "mysql":
		mysql.Refresh()
	}
}
