package scheduler

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/henrylee2cn/pholcus/app/aid/history"
	"github.com/henrylee2cn/pholcus/app/downloader/request"
	"github.com/henrylee2cn/pholcus/common/mns"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"github.com/souriki/ali_mns"
	"log"
)

type PresistentMatrix interface {
	Push(req *request.Request)
	Pull() (req *request.Request)
}

type PresistentMatrixStoreHandler interface {
	Push([]byte)
	Pull() []byte
}

type PresistentMatrixStore struct {
	Handler PresistentMatrixStoreHandler
}

func (self *PresistentMatrixStore) Push(req *request.Request) {
	self.Handler.Push([]byte(req.Serialize()))
}

func (self *PresistentMatrixStore) Pull() (req *request.Request) {
	bytes := self.Handler.Pull()
	req, err := request.UnSerialize(string(bytes))
	if err != nil {
		return
	}
	req.Unique()
	return

}

// 一个Spider实例的请求矩阵
type Matrix struct {
	maxPage         int64  // 最大采集页数，以负数形式表示
	resCount        int32  // 资源使用情况计数
	spiderName      string // 所属Spider
	presistent      PresistentMatrix
	reqs            map[int][]*request.Request  // [优先级]队列，优先级默认为0
	priorities      []int                       // 优先级顺序，从低到高
	history         history.Historier           // 历史记录
	tempHistory     map[string]bool             // 临时记录 [reqUnique(url+method)]true
	failures        map[string]*request.Request // 历史及本次失败请求
	tempHistoryLock sync.RWMutex
	failureLock     sync.Mutex
	sync.Mutex
}

func newMatrix(spiderName, spiderSubName string, maxPage int64) *Matrix {
	var presistent PresistentMatrix

	if config.PRESISTENT == "mns" {
		factory := mns.NewMNSPresistentFactory(config.MNS_PREFIX, ali_mns.NewAliMNSClient(config.MNS_ROOT, config.MNS_KEY, config.MNS_SECRET))
		mnsPresistent, err := factory.New(spiderSubName)
		if err == nil {
			presistent = &PresistentMatrixStore{mnsPresistent}
		}
		log.Println(err,presistent)
	}
	matrix := &Matrix{
		spiderName:  spiderName,
		maxPage:     maxPage,
		reqs:        make(map[int][]*request.Request),
		presistent:  presistent,
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

// 添加请求到队列，并发安全
func (self *Matrix) Push(req *request.Request) {
	if sdl.checkStatus(status.STOP) {
		return
	}
	// 禁止并发，降低请求积存量
	self.Lock()
	defer self.Unlock()
	// 达到请求上限，停止该规则运行
	if self.maxPage >= 0 {
		return
	}

	// 暂停状态时等待，降低请求积存量
	waited := false
	for sdl.checkStatus(status.PAUSE) {
		waited = true
		runtime.Gosched()
	}
	if waited && sdl.checkStatus(status.STOP) {
		return
	}

	// 资源使用过多时等待，降低请求积存量
	waited = false
	for self.resCount > sdl.avgRes() {
		waited = true
		runtime.Gosched()
	}
	if waited && sdl.checkStatus(status.STOP) {
		return
	}

	// 不可重复下载的req
	if !req.IsReloadable() {
		// 已存在成功记录时退出
		if self.hasHistory(req.Unique()) {
			return
		}
		// 添加到临时记录
		self.insertTempHistory(req.Unique())
	}

	// 持久化下载队列
	if self.presistent != nil {
		self.presistent.Push(req)
		// 大致限制加入队列的请求量，并发情况下应该会比maxPage多
		atomic.AddInt64(&self.maxPage, 1)
		return
	}

	var priority = req.GetPriority()

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

	defer func() {
		if req == nil {
			return
		}
		if sdl.useProxy {
			req.SetProxy(sdl.proxy.GetOne(req.GetUrl()))
		} else {
			req.SetProxy("")
		}
	}()

	log.Println(self.presistent,"pull")
	// 持久化下载队列
	if self.presistent != nil {

		req = self.presistent.Pull()

		log.Println("pull")
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

func (self *Matrix) Free() {
	<-sdl.count
	atomic.AddInt32(&self.resCount, -1)
}

// 返回是否作为新的失败请求被添加至队列尾部
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
		// 首次失败时，在任务队列末尾重新执行一次
		self.failures[req.Unique()] = req
		logs.Log.Informational(" *     + 失败请求: [%v]\n", req.GetUrl())
		return true
	}
	// 失败两次后，加入历史失败记录
	self.history.UpsertFailure(req)
	return false
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
	log.Println("canStop")

	self.failureLock.Lock()
	defer self.failureLock.Unlock()
	if len(self.failures) > 0 {
		// 重新下载历史记录中失败的请求
		var goon bool
		for reqUnique, req := range self.failures {
			if req == nil {
				continue
			}
			self.failures[reqUnique] = nil
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

// 非服务器模式下保存历史成功记录
func (self *Matrix) TryFlushSuccess() {
	if cache.Task.Mode != status.SERVER && cache.Task.SuccessInherit {
		self.history.FlushSuccess(cache.Task.OutType)
	}
}

// 非服务器模式下保存历史失败记录
func (self *Matrix) TryFlushFailure() {
	if cache.Task.Mode != status.SERVER && cache.Task.FailureInherit {
		self.history.FlushFailure(cache.Task.OutType)
	}
}

// 等待处理中的请求完成
func (self *Matrix) Wait() {
	for self.resCount != 0 {
		runtime.Gosched()
	}
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
		logs.Log.Informational(" *     + 失败请求: [%v]\n", req.GetUrl())
	}
}
