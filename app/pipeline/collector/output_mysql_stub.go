//go:build coverage

package collector

import (
	"github.com/andeya/gust/result"
)

func init() {
	DataOutput["mysql"] = func(*Collector) result.VoidResult { return result.OkVoid() }
}
