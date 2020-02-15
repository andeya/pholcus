package gc

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

const (
	GC_SIZE = 50 << 20 //默认50MB
)

var (
	gcOnce sync.Once
)

// 手动释放堆中准备重用的一些内存
func ManualGC() {
	go gcOnce.Do(func() {
		tick := time.Tick(2 * time.Minute)
		for {
			<-tick
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			if mem.HeapReleased >= GC_SIZE {
				debug.FreeOSMemory()
				// runtime.GC()
			}
		}
	})
}
