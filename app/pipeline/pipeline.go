// Package pipeline provides the data collection and output pipeline.
package pipeline

import (
	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/pipeline/collector"
	"github.com/andeya/pholcus/app/pipeline/collector/data"
	"github.com/andeya/pholcus/app/spider"
)

// Pipeline collects spider results and writes them to the configured output.
type Pipeline interface {
	Start()
	Stop()
	CollectData(data.DataCell) result.VoidResult
	CollectFile(data.FileCell) result.VoidResult
}

// New creates a new Pipeline for the given spider.
func New(sp *spider.Spider) Pipeline {
	return collector.NewCollector(sp)
}
