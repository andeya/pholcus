package scheduler

import (
	"runtime/debug"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andeya/pholcus/app/aid/history"
	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/cache"
	"github.com/andeya/pholcus/runtime/status"
)

// Matrix is the request queue for a single Spider instance.
type Matrix struct {
	maxPage         int64                       // max pages to collect (negative value)
	resCount        int32                       // resource usage count
	spiderName      string                      // associated Spider name
	reqs            map[int][]*request.Request  // [priority] queues, default priority 0
	priorities      []int                       // priority order, low to high
	history         history.Historier           // history
	tempHistory     map[string]bool             // temp record [reqUnique(url+method)]true
	failures        map[string]*request.Request // historical and current failed requests
	tempHistoryLock sync.RWMutex
	failureLock     sync.Mutex
	sync.Mutex
}

func newMatrix(spiderName, spiderSubName string, maxPage int64) *Matrix {
	matrix := &Matrix{
		spiderName:  spiderName,
		maxPage:     maxPage,
		reqs:        make(map[int][]*request.Request),
		priorities:  []int{},
		history:     history.New(spiderName, spiderSubName),
		tempHistory: make(map[string]bool),
		failures:    make(map[string]*request.Request),
	}
	if cache.Task.Mode != status.SERVER {
		matrix.history.ReadSuccess(cache.Task.OutType, cache.Task.SuccessInherit)
		matrix.history.ReadFailure(cache.Task.OutType, cache.Task.FailureInherit)
		matrix.setFailures(matrix.history.PullFailure())
	}
	return matrix
}

// Push adds a request to the queue. Concurrency-safe.
func (self *Matrix) Push(req *request.Request) {
	self.Lock()
	defer self.Unlock()

	if sdl.checkStatus(status.STOP) {
		return
	}

	if self.maxPage >= 0 {
		return
	}

	waited := false
	for sdl.checkStatus(status.PAUSE) {
		waited = true
		time.Sleep(time.Second)
	}
	if waited && sdl.checkStatus(status.STOP) {
		return
	}

	waited = false
	for atomic.LoadInt32(&self.resCount) > sdl.avgRes() {
		waited = true
		time.Sleep(100 * time.Millisecond)
	}
	if waited && sdl.checkStatus(status.STOP) {
		return
	}

	if !req.IsReloadable() {
		if self.hasHistory(req.Unique()) {
			return
		}
		self.insertTempHistory(req.Unique())
	}

	var priority = req.GetPriority()

	if _, found := self.reqs[priority]; !found {
		self.priorities = append(self.priorities, priority)
		sort.Ints(self.priorities)
		self.reqs[priority] = []*request.Request{}
	}

	self.reqs[priority] = append(self.reqs[priority], req)
	atomic.AddInt64(&self.maxPage, 1)
}

// Pull removes and returns a request from the queue, or nil if empty. Concurrency-safe.
func (self *Matrix) Pull() (req *request.Request) {
	self.Lock()
	defer self.Unlock()
	if !sdl.checkStatus(status.RUN) {
		return
	}
	for i := len(self.reqs) - 1; i >= 0; i-- {
		idx := self.priorities[i]
		if len(self.reqs[idx]) > 0 {
			req = self.reqs[idx][0]
			self.reqs[idx] = self.reqs[idx][1:]
			if req.GetProxy() != "" {
				return
			}
			if sdl.useProxy {
				req.SetProxy(sdl.proxy.GetOne(req.GetUrl()).UnwrapOr(""))
			} else {
				req.SetProxy("")
			}
			return
		}
	}
	return
}

// Use acquires a resource slot for this Matrix.
func (self *Matrix) Use() {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error("panic recovered: %v\n%s", p, debug.Stack())
		}
	}()
	sdl.count <- true
	atomic.AddInt32(&self.resCount, 1)
}

// Free releases a resource slot.
func (self *Matrix) Free() {
	<-sdl.count
	atomic.AddInt32(&self.resCount, -1)
}

// DoHistory records success/failure and returns true if the request was requeued as a new failure.
func (self *Matrix) DoHistory(req *request.Request, ok bool) bool {
	if !req.IsReloadable() {
		self.tempHistoryLock.Lock()
		delete(self.tempHistory, req.Unique())
		self.tempHistoryLock.Unlock()

		if ok {
			self.history.UpsertSuccess(req.Unique())
			return false
		}
	}

	if ok {
		return false
	}

	self.failureLock.Lock()
	defer self.failureLock.Unlock()
	if _, ok := self.failures[req.Unique()]; !ok {
		self.failures[req.Unique()] = req
		logs.Log.Informational(" *     + Failed request: [%v]\n", req.GetUrl())
		return true
	}
	self.history.UpsertFailure(req)
	return false
}

// CanStop reports whether this Matrix can stop (no pending work).
func (self *Matrix) CanStop() bool {
	if sdl.checkStatus(status.STOP) {
		return true
	}
	if self.maxPage >= 0 {
		return true
	}
	if atomic.LoadInt32(&self.resCount) != 0 {
		return false
	}
	if self.Len() > 0 {
		return false
	}

	self.failureLock.Lock()
	defer self.failureLock.Unlock()
	if len(self.failures) > 0 {
		var goon bool
		for reqUnique, req := range self.failures {
			if req == nil {
				continue
			}
			self.failures[reqUnique] = nil
			goon = true
			logs.Log.Informational(" *     - Failed request: [%v]\n", req.GetUrl())
			self.Push(req)
		}
		if goon {
			return false
		}
	}
	return true
}

// TryFlushSuccess flushes success history in non-server mode.
func (self *Matrix) TryFlushSuccess() {
	if cache.Task.Mode != status.SERVER && cache.Task.SuccessInherit {
		self.history.FlushSuccess(cache.Task.OutType)
	}
}

// TryFlushFailure flushes failure history in non-server mode.
func (self *Matrix) TryFlushFailure() {
	if cache.Task.Mode != status.SERVER && cache.Task.FailureInherit {
		self.history.FlushFailure(cache.Task.OutType)
	}
}

// Wait blocks until all in-flight requests complete.
func (self *Matrix) Wait() {
	if sdl.checkStatus(status.STOP) {
		return
	}
	for atomic.LoadInt32(&self.resCount) != 0 {
		time.Sleep(500 * time.Millisecond)
	}
}

// Len returns the number of queued requests.
func (self *Matrix) Len() int {
	self.Lock()
	defer self.Unlock()
	var l int
	for _, reqs := range self.reqs {
		l += len(reqs)
	}
	return l
}

func (self *Matrix) hasHistory(reqUnique string) bool {
	if self.history.HasSuccess(reqUnique) {
		return true
	}
	self.tempHistoryLock.RLock()
	has := self.tempHistory[reqUnique]
	self.tempHistoryLock.RUnlock()
	return has
}

func (self *Matrix) insertTempHistory(reqUnique string) {
	self.tempHistoryLock.Lock()
	self.tempHistory[reqUnique] = true
	self.tempHistoryLock.Unlock()
}

func (self *Matrix) setFailures(reqs map[string]*request.Request) {
	self.failureLock.Lock()
	defer self.failureLock.Unlock()
	for key, req := range reqs {
		self.failures[key] = req
		logs.Log.Informational(" *     + Failed request: [%v]\n", req.GetUrl())
	}
}

// // windup performs cleanup when stopping tasks.
// func (self *Matrix) windup() {
// 	self.Lock()

// 	self.reqs = make(map[int][]*request.Request)
// 	self.priorities = []int{}
// 	self.tempHistory = make(map[string]bool)

// 	self.failures = make(map[string]*request.Request)

// 	self.Unlock()
// }
