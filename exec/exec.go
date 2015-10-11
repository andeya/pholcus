package exec

import (
	"runtime"
)

func init() {
	// 开启最大核心数运行
	runtime.GOMAXPROCS(runtime.NumCPU())
}
