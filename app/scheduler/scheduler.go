// Package scheduler provides crawl task scheduling and resource allocation.
package scheduler

import (
	"runtime/debug"
	"sync"

	"github.com/andeya/pholcus/app/aid/proxy"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/status"
)

// scheduler coordinates crawl tasks and resource allocation.
type scheduler struct {
	status       int          // running status
	count        chan bool    // total concurrency count
	useProxy     bool         // whether proxy IP is used
	proxy        *proxy.Proxy // global proxy IP
	matrices     []*Matrix    // request matrices per Spider instance
	sync.RWMutex              // global read-write lock
}

// sched is the global scheduler instance.
var sched = &scheduler{
	status: status.RUN,
	count:  make(chan bool, 1),
	proxy:  proxy.New(),
}

// Init initializes the scheduler with the given concurrency and proxy settings.
func Init(threadNum int, proxyMinute int64) {
	sched.matrices = []*Matrix{}
	sched.count = make(chan bool, threadNum)

	if proxyMinute > 0 {
		if sched.proxy.Count() > 0 {
			sched.useProxy = true
			sched.proxy.UpdateTicker(proxyMinute)
			logs.Log().Informational(" *     Using proxy IP, rotation interval: %v minutes\n", proxyMinute)
		} else {
			sched.useProxy = false
			logs.Log().Informational(" *     Proxy IP list is empty, cannot use proxy\n")
		}
	} else {
		sched.useProxy = false
		logs.Log().Informational(" *     Not using proxy IP\n")
	}

	sched.status = status.RUN
}

// ReloadProxyLib reloads the proxy IP list from the config file.
func ReloadProxyLib() {
	sched.proxy.Update()
}

// AddMatrix registers a resource queue for the given spider and returns its Matrix.
func AddMatrix(spiderName, spiderSubName string, maxPage int64) *Matrix {
	matrix := newMatrix(spiderName, spiderSubName, maxPage)
	sched.RLock()
	defer sched.RUnlock()
	sched.matrices = append(sched.matrices, matrix)
	return matrix
}

// PauseRecover toggles pause/resume for all crawl tasks.
func PauseRecover() {
	sched.Lock()
	defer sched.Unlock()
	switch sched.status {
	case status.PAUSE:
		sched.status = status.RUN
	case status.RUN:
		sched.status = status.PAUSE
	}
}

// Stop terminates all crawl tasks.
func Stop() {
	sched.Lock()
	defer sched.Unlock()
	sched.status = status.STOP
	defer func() {
		if p := recover(); p != nil {
			logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
		}
	}()
	close(sched.count)
	sched.matrices = []*Matrix{}
}

// avgRes returns the average resources allocated per spider instance.
func (sched *scheduler) avgRes() int32 {
	avg := int32(cap(sched.count) / len(sched.matrices))
	if avg == 0 {
		avg = 1
	}
	return avg
}

func (sched *scheduler) checkStatus(s int) bool {
	sched.RLock()
	b := sched.status == s
	sched.RUnlock()
	return b
}
