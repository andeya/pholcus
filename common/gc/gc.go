// Package gc provides manual garbage collection to release heap memory.
package gc

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

const (
	GC_SIZE = 50 << 20 // default 50MB
)

var (
	gcOnce sync.Once
)

// ManualGC periodically frees memory from the heap for reuse.
// Skipped for gust adoption: runtime.ReadMemStats, debug.FreeOSMemory, and
// time.Tick do not return errors; no error-returning functions to convert.
func ManualGC() {
	go gcOnce.Do(func() {
		tick := time.Tick(2 * time.Minute)
		for {
			<-tick
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			if mem.HeapReleased >= GC_SIZE {
				debug.FreeOSMemory()
			}
		}
	})
}
