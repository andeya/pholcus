package pipeline

import (
	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"github.com/henrylee2cn/pholcus/common/util"
)

// 初始化输出方式列表collector.OutputLib
func init() {
	for out, _ := range collector.Output {
		collector.OutputLib = append(collector.OutputLib, out)
	}
	util.QSortT(collector.OutputLib)
}
