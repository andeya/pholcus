package scheduler

import (
	"sort"
	"sync"
	"sync/atomic"

	"github.com/henrylee2cn/pholcus/app/aid/history"
	"github.com/henrylee2cn/pholcus/app/aid/proxy"
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
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
	// 全局代理IP
	proxy *proxy.Proxy
	// 标记是否使用代理IP
	useProxy bool
	// 全局历史记录
	history history.Historier
	// 全局读写锁
	sync.RWMutex
	sync.Once
}

// 定义全局调度
var sdl = &scheduler{
	history: history.New(),
	status:  status.RUN,
	count:   make(chan bool, cache.Task.ThreadNum),
}

func proxyInit() {
	sdl.proxy = proxy.New()
}

func Init() {
	sdl.Once.Do(proxyInit)
	sdl.matrices = make(map[int]*Matrix)
	sdl.count = make(chan bool, cache.Task.ThreadNum)
	sdl.history.ReadSuccess(cache.Task.OutType, cache.Task.SuccessInherit)
	sdl.history.ReadFailure(cache.Task.OutType, cache.Task.FailureInherit)
	if cache.Task.ProxyMinute > 0 {
		if sdl.proxy.Count() > 0 {
			sdl.useProxy = true
			sdl.proxy.UpdateTicker(cache.Task.ProxyMinute)
			logs.Log.Informational(" *     使用代理IP，代理IP更换频率为 %v 分钟\n", cache.Task.ProxyMinute)
		} else {
			sdl.useProxy = false
			logs.Log.Informational(" *     代理IP列表为空，无法使用代理IP\n")
		}
	} else {
		sdl.useProxy = false
		logs.Log.Informational(" *     不使用代理IP\n")
	}
	sdl.status = status.RUN
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
	sdl.RLock()
	defer sdl.RUnlock()

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
				proxy, _ := sdl.proxy.GetOne()
				req.SetProxy(proxy)
			} else {
				req.SetProxy("")
			}
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

func (self *Matrix) Free() {
	<-sdl.count
	atomic.AddInt32(&self.resCount, -1)
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
