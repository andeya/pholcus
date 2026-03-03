//go:build coverage

package collector

import (
	"github.com/andeya/gust/result"
)

func init() {
	DataOutput["mgo"] = func(*Collector) result.VoidResult { return result.OkVoid() }
}
