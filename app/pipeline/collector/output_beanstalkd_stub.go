//go:build coverage

package collector

import (
	"github.com/andeya/gust/result"
)

func init() {
	DataOutput["beanstalkd"] = func(*Collector) result.VoidResult { return result.OkVoid() }
}
