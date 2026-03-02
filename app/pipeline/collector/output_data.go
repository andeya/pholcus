package collector

import (
	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/logs"
)

// Refresher is an optional interface that output backends can implement
// to refresh their state (e.g., reconnect) before a new task run.
type Refresher interface {
	Refresh()
}

var (
	// DataOutput maps output type names to their implementation functions.
	DataOutput = make(map[string]func(col *Collector) result.VoidResult)

	// DataOutputLib lists the names of supported text data output backends.
	DataOutputLib []string

	// dataRefreshers maps output type names to optional Refresher implementations.
	dataRefreshers = make(map[string]Refresher)
)

// outputData writes collected text data to the configured output backend.
func (c *Collector) outputData() {
	defer func() {
		c.resetDataBuf()
	}()

	dataLen := uint64(len(c.dataBuf))
	if dataLen == 0 {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			logs.Log().Informational(" * ")
			logs.Log().App(" *     Panic  [Data output: %v | KEYIN: %v | Batch: %v]  %v records! [ERROR]  %v\n",
				c.Spider.GetName(), c.Spider.GetKeyin(), c.dataBatch, dataLen, p)
		}
	}()

	c.addDataSum(dataLen)

	r := DataOutput[c.outType](c)

	logs.Log().Informational(" * ")
	if r.IsErr() {
		logs.Log().App(" *     Fail  [Data output: %v | KEYIN: %v | Batch: %v]  %v records! [ERROR]  %v\n",
			c.Spider.GetName(), c.Spider.GetKeyin(), c.dataBatch, dataLen, r.UnwrapErr())
	} else {
		logs.Log().App(" *     [Data output: %v | KEYIN: %v | Batch: %v]  %v records!\n",
			c.Spider.GetName(), c.Spider.GetKeyin(), c.dataBatch, dataLen)
		c.Spider.TryFlushSuccess()
	}
}

// Register adds an output backend for the given type name.
func Register(outType string, outFunc func(col *Collector) result.VoidResult) {
	DataOutput[outType] = outFunc
}

// RegisterRefresher associates a Refresher with an output type.
func RegisterRefresher(outType string, r Refresher) {
	dataRefreshers[outType] = r
}

// RefreshBackend calls the Refresher for the given output type, if registered.
func RefreshBackend(outType string) {
	if r, ok := dataRefreshers[outType]; ok {
		r.Refresh()
	}
}
