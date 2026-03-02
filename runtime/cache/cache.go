package cache

import (
	"runtime"
	"sync/atomic"
	"time"
)

// --- Task Runtime Common Configuration ---

// AppConf holds the common configuration for task runtime.
type AppConf struct {
	Mode           int    // node role
	Port           int    // master node port
	Master         string // master node address (without port)
	ThreadNum      int    // global max concurrency
	Pausetime      int64  // pause duration reference in ms (random: Pausetime/2 ~ Pausetime*2)
	OutType        string // output method
	DockerCap      int    // segment dump container capacity
	Limit          int64  // crawl limit; 0 means unlimited; if set to LIMIT in rules, uses custom limit; otherwise defaults to request count limit
	ProxyMinute    int64  // proxy IP rotation interval in minutes
	SuccessInherit bool   // inherit historical success records
	FailureInherit bool   // inherit historical failure records
	Keyins         string // custom input; later split into Keyin config for multiple tasks
}

// Task holds the default runtime configuration.
var Task = new(AppConf)

// --- Task Report ---

// Report summarizes task execution results.
type Report struct {
	SpiderName string
	Keyin      string
	DataNum    uint64
	FileNum    uint64
	// DataSize   uint64
	// FileSize uint64
	Time time.Duration
}

var (
	StartTime  time.Time    // timestamp when start button was clicked
	ReportChan chan *Report // text data summary report channel
	pageSum    [2]uint64    // [total count, failure count]
)

// ResetPageCount resets the page counters.
func ResetPageCount() {
	pageSum = [2]uint64{}
}

// GetPageCount returns page counts: i>0 returns success count, i<0 returns failure count, i==0 returns total.
func GetPageCount(i int) uint64 {
	switch {
	case i > 0:
		return pageSum[0]
	case i < 0:
		return pageSum[1]
	case i == 0:
	}
	return pageSum[0] + pageSum[1]
}

// PageSuccCount increments the success page count.
func PageSuccCount() {
	atomic.AddUint64(&pageSum[0], 1)
}

// PageFailCount increments the failure page count.
func PageFailCount() {
	atomic.AddUint64(&pageSum[1], 1)
}

// --- Init Function Execution Order Control ---

var initOrder = make(map[int]bool)

// ExecInit marks the init at the given order as completed.
func ExecInit(order int) {
	initOrder[order] = true
}

// WaitInit blocks until the init at the given order has completed. Must be called from a goroutine.
func WaitInit(order int) {
	for !initOrder[order] {
		runtime.Gosched()
	}
}

// --- Initialization ---

func init() {
	ReportChan = make(chan *Report)
}
