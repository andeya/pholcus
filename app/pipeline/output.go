package pipeline

import (
	"sort"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/pipeline/collector"
	"github.com/andeya/pholcus/common/kafka"
	"github.com/andeya/pholcus/common/mgo"
	"github.com/andeya/pholcus/common/mysql"
	"github.com/andeya/pholcus/runtime/cache"
)

// init populates the output library and registers refreshers for stateful backends.
func init() {
	for out := range collector.DataOutput {
		collector.DataOutputLib = append(collector.DataOutputLib, out)
	}
	sort.Strings(collector.DataOutputLib)

	collector.RegisterRefresher("mgo", refresherFunc(func() { mgo.Refresh() }))
	collector.RegisterRefresher("mysql", refresherFunc(func() { mysql.Refresh() }))
	collector.RegisterRefresher("kafka", refresherFunc(func() { kafka.Refresh() }))
}

// refresherFunc adapts a plain function to the Refresher interface.
type refresherFunc func()

func (f refresherFunc) Refresh() { f() }

// RegisterOutput registers an output backend at the pipeline level.
func RegisterOutput(name string, fn func(*collector.Collector) result.VoidResult) {
	collector.Register(name, fn)
	collector.DataOutputLib = append(collector.DataOutputLib, name)
	sort.Strings(collector.DataOutputLib)
}

// GetOutputLib returns a sorted list of all registered output backend names.
func GetOutputLib() []string {
	return collector.DataOutputLib
}

// RefreshOutput refreshes the state of the configured output backend via the registry.
func RefreshOutput() {
	collector.RefreshBackend(cache.Task.OutType)
}
