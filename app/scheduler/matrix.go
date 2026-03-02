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
	history         history.HistoryStore        // history
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
func (m *Matrix) Push(req *request.Request) {
	m.Lock()
	defer m.Unlock()

	if sched.checkStatus(status.STOP) {
		return
	}

	if m.maxPage >= 0 {
		return
	}

	waited := false
	for sched.checkStatus(status.PAUSE) {
		waited = true
		time.Sleep(time.Second)
	}
	if waited && sched.checkStatus(status.STOP) {
		return
	}

	waited = false
	for atomic.LoadInt32(&m.resCount) > sched.avgRes() {
		waited = true
		time.Sleep(100 * time.Millisecond)
	}
	if waited && sched.checkStatus(status.STOP) {
		return
	}

	if !req.IsReloadable() {
		if m.hasHistory(req.Unique()) {
			return
		}
		m.insertTempHistory(req.Unique())
	}

	var priority = req.GetPriority()

	if _, found := m.reqs[priority]; !found {
		m.priorities = append(m.priorities, priority)
		sort.Ints(m.priorities)
		m.reqs[priority] = []*request.Request{}
	}

	m.reqs[priority] = append(m.reqs[priority], req)
	atomic.AddInt64(&m.maxPage, 1)
}

// Pull removes and returns a request from the queue, or nil if empty. Concurrency-safe.
func (m *Matrix) Pull() (req *request.Request) {
	m.Lock()
	defer m.Unlock()
	if !sched.checkStatus(status.RUN) {
		return
	}
	for i := len(m.reqs) - 1; i >= 0; i-- {
		idx := m.priorities[i]
		if len(m.reqs[idx]) > 0 {
			req = m.reqs[idx][0]
			m.reqs[idx] = m.reqs[idx][1:]
			if req.GetProxy() != "" {
				return
			}
			if sched.useProxy {
				req.SetProxy(sched.proxy.GetOne(req.GetURL()).UnwrapOr(""))
			} else {
				req.SetProxy("")
			}
			return
		}
	}
	return
}

// Use acquires a resource slot for this Matrix.
func (m *Matrix) Use() {
	defer func() {
		if p := recover(); p != nil {
			logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
		}
	}()
	sched.count <- true
	atomic.AddInt32(&m.resCount, 1)
}

// Free releases a resource slot.
func (m *Matrix) Free() {
	<-sched.count
	atomic.AddInt32(&m.resCount, -1)
}

// DoHistory records success/failure and returns true if the request was requeued as a new failure.
func (m *Matrix) DoHistory(req *request.Request, ok bool) bool {
	if !req.IsReloadable() {
		m.tempHistoryLock.Lock()
		delete(m.tempHistory, req.Unique())
		m.tempHistoryLock.Unlock()

		if ok {
			m.history.UpsertSuccess(req.Unique())
			return false
		}
	}

	if ok {
		return false
	}

	m.failureLock.Lock()
	defer m.failureLock.Unlock()
	if _, ok := m.failures[req.Unique()]; !ok {
		m.failures[req.Unique()] = req
		logs.Log().Informational(" *     + Failed request: [%v]\n", req.GetURL())
		return true
	}
	m.history.UpsertFailure(req)
	return false
}

// CanStop reports whether this Matrix can stop (no pending work).
func (m *Matrix) CanStop() bool {
	if sched.checkStatus(status.STOP) {
		return true
	}
	if m.maxPage >= 0 {
		return true
	}
	if atomic.LoadInt32(&m.resCount) != 0 {
		return false
	}
	if m.Len() > 0 {
		return false
	}

	m.failureLock.Lock()
	defer m.failureLock.Unlock()
	if len(m.failures) > 0 {
		var goon bool
		for reqUnique, req := range m.failures {
			if req == nil {
				continue
			}
			m.failures[reqUnique] = nil
			goon = true
			logs.Log().Informational(" *     - Failed request: [%v]\n", req.GetURL())
			m.Push(req)
		}
		if goon {
			return false
		}
	}
	return true
}

// TryFlushSuccess flushes success history in non-server mode.
func (m *Matrix) TryFlushSuccess() {
	if cache.Task.Mode != status.SERVER && cache.Task.SuccessInherit {
		m.history.FlushSuccess(cache.Task.OutType)
	}
}

// TryFlushFailure flushes failure history in non-server mode.
func (m *Matrix) TryFlushFailure() {
	if cache.Task.Mode != status.SERVER && cache.Task.FailureInherit {
		m.history.FlushFailure(cache.Task.OutType)
	}
}

// Wait blocks until all in-flight requests complete.
func (m *Matrix) Wait() {
	if sched.checkStatus(status.STOP) {
		return
	}
	for atomic.LoadInt32(&m.resCount) != 0 {
		time.Sleep(500 * time.Millisecond)
	}
}

// Len returns the number of queued requests.
func (m *Matrix) Len() int {
	m.Lock()
	defer m.Unlock()
	var l int
	for _, reqs := range m.reqs {
		l += len(reqs)
	}
	return l
}

func (m *Matrix) hasHistory(reqUnique string) bool {
	if m.history.HasSuccess(reqUnique) {
		return true
	}
	m.tempHistoryLock.RLock()
	has := m.tempHistory[reqUnique]
	m.tempHistoryLock.RUnlock()
	return has
}

func (m *Matrix) insertTempHistory(reqUnique string) {
	m.tempHistoryLock.Lock()
	m.tempHistory[reqUnique] = true
	m.tempHistoryLock.Unlock()
}

func (m *Matrix) setFailures(reqs map[string]*request.Request) {
	m.failureLock.Lock()
	defer m.failureLock.Unlock()
	for key, req := range reqs {
		m.failures[key] = req
		logs.Log().Informational(" *     + Failed request: [%v]\n", req.GetURL())
	}
}
