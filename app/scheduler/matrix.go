package scheduler

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/henrylee2cn/pholcus/app/downloader/request"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

// 一个Spider实例的请求矩阵
type Matrix struct {
	// 资源使用情况计数
	resCount int32
	// 最大采集页数，以负数形式表示
	maxPage int64
	// [优先级]队列，优先级默认为0
	reqs map[int][]*request.Request
	// 优先级顺序，从低到高
	priorities []int
	// 历史及本次失败请求
	failures    map[string]*request.Request
	failureLock sync.Mutex
	sync.Mutex
}

// 注册资源队列
func NewMatrix(spiderId int, maxPage int64) *Matrix {
	matrix := &Matrix{
		reqs:       make(map[int][]*request.Request),
		priorities: []int{},
		maxPage:    maxPage,
		failures:   make(map[string]*request.Request),
	}
	sdl.addMatrix(spiderId, matrix)
	return matrix
}

// 添加请求到队列，并发安全
func (self *Matrix) Push(req *request.Request) {
	// 禁止并发，降低请求积存量
	self.Lock()
	defer self.Unlock()

	// 根据运行状态及资源使用情况，降低请求积存量
	for sdl.checkStatus(status.PAUSE) || self.resCount > sdl.avgRes() {
		runtime.Gosched()
	}

	// 终止添加操作
	if sdl.checkStatus(status.STOP) || self.maxPage >= 0 ||
		// 当req不可重复下载时，已存在成功记录则返回
		!req.IsReloadable() && !UpsertSuccess(req) {
		return
	}

	priority := req.GetPriority()

	// 初始化该蜘蛛下该优先级队列
	if _, found := self.reqs[priority]; !found {
		self.priorities = append(self.priorities, priority)
		sort.Ints(self.priorities) // 从小到大排序
		self.reqs[priority] = []*request.Request{}
	}

	// 添加请求到队列
	self.reqs[priority] = append(self.reqs[priority], req)

	// 大致限制加入队列的请求量，并发情况下应该会比maxPage多
	atomic.AddInt64(&self.maxPage, 1)
}

// 从队列取出请求，不存在时返回nil，并发安全
func (self *Matrix) Pull() (req *request.Request) {
	if !sdl.checkStatus(status.RUN) {
		return
	}
	self.Lock()
	defer self.Unlock()
	// 按优先级从高到低取出请求
	for i := len(self.reqs) - 1; i >= 0; i-- {
		idx := self.priorities[i]
		if len(self.reqs[idx]) > 0 {
			req = self.reqs[idx][0]
			self.reqs[idx] = self.reqs[idx][1:]
			if sdl.useProxy {
				req.SetProxy(sdl.proxy.GetOne(req.GetUrl()))
			} else {
				req.SetProxy("")
			}
			return
		}
	}
	return
}

func (self *Matrix) Len() int {
	self.Lock()
	defer self.Unlock()
	var l int
	for _, reqs := range self.reqs {
		l += len(reqs)
	}
	return l
}

func (self *Matrix) Use() {
	defer func() {
		recover()
	}()
	sdl.count <- true
	atomic.AddInt32(&self.resCount, 1)
}

func (self *Matrix) Free() {
	<-sdl.count
	atomic.AddInt32(&self.resCount, -1)
}

func (self *Matrix) CanStop() bool {
	if sdl.checkStatus(status.STOP) {
		return true
	}
	if self.maxPage >= 0 {
		return true
	}
	if self.resCount != 0 {
		return false
	}
	if self.Len() > 0 {
		return false
	}

	self.failureLock.Lock()
	defer self.failureLock.Unlock()
	if len(self.failures) > 0 {
		// 重新下载历史记录中失败的请求
		var goon bool
		for unique, req := range self.failures {
			if req == nil {
				continue
			}
			self.failures[unique] = nil
			goon = true
			logs.Log.Informational(" *     - 失败请求: [%v]\n", req.GetUrl())
			self.Push(req)
		}
		if goon {
			return false
		}
	}
	return true
}

// 等待处理中的请求完成
func (self *Matrix) Flush() {
	for self.resCount != 0 {
		runtime.Gosched()
	}
}

func (self *Matrix) SetFailures(reqs []*request.Request) {
	self.failureLock.Lock()
	defer self.failureLock.Unlock()
	for _, req := range reqs {
		self.failures[makeUnique(req)] = req
		logs.Log.Informational(" *     + 失败请求: [%v]\n", req.GetUrl())
	}
}

func (self *Matrix) SetFailure(req *request.Request) bool {
	self.failureLock.Lock()
	defer self.failureLock.Unlock()
	unique := makeUnique(req)
	if _, ok := self.failures[unique]; !ok {
		// 首次失败时，在任务队列末尾重新执行一次
		self.failures[unique] = req
		logs.Log.Informational(" *     + 失败请求: [%v]\n", req.GetUrl())
		return true
	}
	// 失败两次后，加入历史失败记录
	UpsertFailure(req)
	return false
}

func makeUnique(req *request.Request) string {
	return util.MakeUnique(req.GetUrl() + req.GetMethod())
}
