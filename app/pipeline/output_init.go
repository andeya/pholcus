package pipeline

import (
	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"sort"
)

// 初始化输出方式列表collector.OutputLib
func init() {
	for out, _ := range collector.Output {
		collector.OutputLib = append(collector.OutputLib, out)
	}
	sort.Strings(collector.OutputLib)
}
