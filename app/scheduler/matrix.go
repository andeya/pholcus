package scheduler

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

// 一个Spider实例的请求矩阵
type Matrix struct {
	// [优先级]队列，优先级默认为0
	reqs map[int][]*context.Request
	// 优先级顺序，从低到高
	priorities []int
	// 资源使用情况计数
	resCount int32
	// 最大采集页数，以负数形式表示
	maxPage int64
	// 历史及本次失败请求
	failures map[string]*context.Request
	sync.Mutex
}

// 注册资源队列
func NewMatrix(spiderId int, maxPage int64) *Matrix {
	sdl.Lock()
	defer sdl.Unlock()
	matrix := &Matrix{
		reqs:       make(map[int][]*context.Request),
		priorities: []int{},
		maxPage:    maxPage,
		failures:   make(map[string]*context.Request),
	}
	sdl.matrices[spiderId] = matrix
	return matrix
}

// 添加请求到队列
func (self *Matrix) Push(req *context.Request) {
	// 禁止并发，降低请求积存量
	self.Lock()
	defer self.Unlock()

	// 根据运行状态及资源使用情况，降低请求积存量
	for sdl.status == status.PAUSE || self.resCount > sdl.avgRes() {
		runtime.Gosched()
	}

	sdl.RLock()
	defer sdl.RUnlock()

	// 终止添加操作
	if sdl.status == status.STOP ||
		self.maxPage >= 0 ||
		// 当req不可重复下载时，已存在成功记录则返回
		!req.IsReloadable() && !UpsertSuccess(req) {
		return
	}

	// 大致限制加入队列的请求量，并发情况下应该会比maxPage多
	atomic.AddInt64(&self.maxPage, 1)

	priority := req.GetPriority()

	// 初始化该蜘蛛下该优先级队列
	if _, found := self.reqs[priority]; !found {
		self.priorities = append(self.priorities, priority)
		sort.Ints(self.priorities) // 从小到大排序
		self.reqs[priority] = []*context.Request{}
	}

	// 添加请求到队列
	self.reqs[priority] = append(self.reqs[priority], req)
}

// 从队列取出请求，不存在时返回nil
func (self *Matrix) Pull() (req *context.Request) {
	sdl.RLock()
	defer sdl.RUnlock()
	if sdl.status != status.RUN {
		return
	}

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
	sdl.RLock()
	if sdl.status == status.STOP {
		sdl.RUnlock()
		return true
	}
	sdl.RUnlock()

	if self.maxPage >= 0 {
		return true
	}

	if self.resCount != 0 {
		return false
	}

	for i, count := 0, len(self.reqs); i < count; i++ {
		if len(self.reqs[i]) > 0 {
			return false
		}
	}

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
		// time.Sleep(5e8)
	}
}

func (self *Matrix) SetFailures(reqs []*context.Request) {
	for _, req := range reqs {
		self.failures[makeUnique(req)] = req
		logs.Log.Informational(" *     + 失败请求: [%v]\n", req.GetUrl())
	}
}

func (self *Matrix) SetFailure(req *context.Request) bool {
	self.Lock()
	defer self.Unlock()
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

func makeUnique(req *context.Request) string {
	return util.MakeUnique(req.GetUrl() + req.GetMethod())
}
