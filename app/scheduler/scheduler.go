package scheduler

import (
	"sort"
	"sync"
	"sync/atomic"

	"github.com/henrylee2cn/pholcus/app/aid/history"
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

type scheduler struct {
	// Spider实例的请求矩阵列表
	matrices map[int]*Matrix
	// 总并发量计数
	count chan bool
	// 运行状态
	status int
	// 全局历史记录
	history history.Historier
	// 全局读写锁
	sync.RWMutex
}

// 定义全局调度
var sdl = &scheduler{
	history: history.New(),
	status:  status.RUN,
	count:   make(chan bool, cache.Task.ThreadNum),
}

func Init() {
	sdl.matrices = make(map[int]*Matrix)
	sdl.count = make(chan bool, cache.Task.ThreadNum)
	sdl.history.ReadSuccess(cache.Task.OutType, cache.Task.SuccessInherit)
	sdl.history.ReadFailure(cache.Task.OutType, cache.Task.FailureInherit)
	sdl.status = status.RUN
}

// 注册资源队列
func NewMatrix(spiderId int) *Matrix {
	sdl.Lock()
	defer sdl.Unlock()
	matrix := &Matrix{
		reqs:       make(map[int][]*context.Request),
		priorities: []int{},
	}
	sdl.matrices[spiderId] = matrix
	return matrix
}

// 暂停\恢复所有爬行任务
func PauseRecover() {
	sdl.Lock()
	defer sdl.Unlock()
	switch sdl.status {
	case status.PAUSE:
		sdl.status = status.RUN
	case status.RUN:
		sdl.status = status.PAUSE
	}
}

// 终止任务
func Stop() {
	sdl.Lock()
	defer func() {
		recover()
		sdl.Unlock()
	}()

	sdl.status = status.STOP
	for _, v := range sdl.matrices {
		for _, vv := range v.reqs {
			for _, req := range vv {
				// 删除队列中未执行请求的去重记录
				DeleteSuccess(req)
			}
		}
		v.reqs = make(map[int][]*context.Request)
		v.priorities = []int{}
	}

	// 清空
	close(sdl.count)
	sdl.matrices = make(map[int]*Matrix)
}

func UpsertSuccess(record history.Record) bool {
	return sdl.history.UpsertSuccess(record)
}

func DeleteSuccess(record history.Record) {
	sdl.history.DeleteSuccess(record)
}

func UpsertFailure(req *context.Request) bool {
	return sdl.history.UpsertFailure(req)
}

func DeleteFailure(req *context.Request) {
	sdl.history.DeleteFailure(req)
}

// 获取指定蜘蛛在上一次运行时失败的请求
func PullFailure(spiderName string) []*context.Request {
	return sdl.history.PullFailure(spiderName)
}

func TryFlushHistory() {
	if cache.Task.SuccessInherit {
		sdl.history.FlushSuccess(cache.Task.OutType)
	}
	if cache.Task.FailureInherit {
		sdl.history.FlushFailure(cache.Task.OutType)
	}
}

// 一个Spider实例的请求矩阵
type Matrix struct {
	// [优先级]队列，优先级默认为0
	reqs map[int][]*context.Request
	// 优先级顺序，从低到高
	priorities []int
	// 资源使用情况计数
	resCount int32
	// 历史失败请求
	failures []*context.Request
}

// 添加请求到队列
func (self *Matrix) Push(req *context.Request) {
	sdl.RLock()
	defer sdl.RUnlock()

	if sdl.status == status.STOP {
		return
	}
	// 当req不可重复下载时，已存在成功记录则返回
	if !req.IsReloadable() && !UpsertSuccess(req) {
		return
	}

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
			return
		}
	}
	return
}

func (self *Matrix) Use() {
	defer func() {
		recover()
	}()
	sdl.count <- true
	atomic.AddInt32(&self.resCount, 1)
}

func (self *Matrix) CanStop() bool {
	sdl.RLock()
	if sdl.status == status.STOP {
		sdl.RUnlock()
		return true
	}
	sdl.RUnlock()

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
		for _, req := range self.failures {
			self.Push(req)
		}
		self.failures = []*context.Request{}
		return false
	}

	return true
}

func (self *Matrix) Free() {
	<-sdl.count
	atomic.AddInt32(&self.resCount, -1)
}

func (self *Matrix) SetFailures(reqs []*context.Request) {
	self.failures = reqs
}
