//go:build coverage

package collector

import (
	"github.com/andeya/gust/result"
)

func init() {
	DataOutput["kafka"] = func(*Collector) result.VoidResult { return result.OkVoid() }
}
