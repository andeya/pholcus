package scheduler

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/common/deduplicate"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"sort"
	"sync"
)

type Scheduler interface {
	Init()
	// 注册资源队列
	RegSpider(spiderId int)
	// 注销资源队列
	CancelSpider(spiderId int)
	// 暂停\恢复所有爬行任务
	PauseRecover()
	// 终止运行
	Stop()
	// 判断是否已停止运行
	IsStop() bool
	// 存入
	Push(*context.Request)
	// 取出
	Use(int) *context.Request
	// 释放一个资源
	Free()
	// 资源队列是否闲置
	IsEmpty(int) bool

	// 对比是否存在重复项
	Deduplicate(key interface{}) bool
	// 保存去重记录
	SaveDeduplication()
	// 取消指定去重样本
	DelDeduplication(key interface{})
}

type scheduler struct {
	// 全局并发量计数
	count chan bool
	// map[spiderId](map[请求优先级priority]([]请求))，优先级默认为0
	queue map[int](map[int][]*context.Request)
	// map[spiderId]int[存在的请求优先级],优先级从小到大排序
	index map[int][]int
	// map[spiderId]读写锁
	rwMutexes map[int]*sync.RWMutex
	// 全局读写锁
	sync.RWMutex
	// 运行状态
	status int
	// 全局去重
	deduplication deduplicate.Deduplicate
}

// 定义全局调度
var Sdl Scheduler

func init() {
	Sdl = &scheduler{
		deduplication: deduplicate.New(),
		status:        status.RUN,
	}
}

func SaveDeduplication() {
	Sdl.SaveDeduplication()
}

func (self *scheduler) Init() {
	self.count = make(chan bool, cache.Task.ThreadNum)
	self.queue = make(map[int]map[int][]*context.Request)
	self.index = make(map[int][]int)
	self.rwMutexes = make(map[int]*sync.RWMutex)

	self.deduplication.Update(cache.Task.OutType, cache.Task.InheritDeduplication)

	self.status = status.RUN
}

// 注册资源队列
func (self *scheduler) RegSpider(spiderId int) {
	self.Lock()
	defer self.Unlock()
	if _, found := self.queue[spiderId]; !found {
		self.queue[spiderId] = make(map[int][]*context.Request)
		self.index[spiderId] = []int{}
		self.rwMutexes[spiderId] = new(sync.RWMutex)
	}
}

// 注销资源队列
func (self *scheduler) CancelSpider(spiderId int) {
	self.Lock()
	defer self.Unlock()
	delete(self.queue, spiderId)
}

// 添加请求到队列
func (self *scheduler) Push(req *context.Request) {
	// 初始化该蜘蛛的队列
	spiderId, ok := req.GetSpiderId()
	if !ok {
		return
	}

	self.RLock()
	defer self.RUnlock()

	if self.status == status.STOP {
		return
	}

	// 当req不可重复时，有重复则返回
	if !req.GetDuplicatable() && self.Deduplicate(req.GetUrl()+req.GetMethod()) {
		return
	}

	// 初始化该蜘蛛下该优先级队列
	priority := req.GetPriority()
	if !self.foundPriority(spiderId, priority) {
		self.addPriority(spiderId, priority)
	}

	defer func() {
		recover()
	}()

	// 添加请求到队列
	self.queue[spiderId][priority] = append(self.queue[spiderId][priority], req)
}

func (self *scheduler) Use(spiderId int) (req *context.Request) {
	self.RLock()
	defer self.RUnlock()

	if self.status != status.RUN {
		return
	}

	// 按优先级从高到低取出请求
	for i := len(self.queue[spiderId]) - 1; i >= 0; i-- {
		idx := self.index[spiderId][i]
		if len(self.queue[spiderId][idx]) > 0 {
			req = self.queue[spiderId][idx][0]
			self.queue[spiderId][idx] = self.queue[spiderId][idx][1:]
			self.count <- true
			return
		}
	}
	return
}

func (self *scheduler) Free() {
	<-self.count
}

// 暂停\恢复所有爬行任务
func (self *scheduler) PauseRecover() {
	self.Lock()
	defer self.Unlock()

	switch self.status {
	case status.PAUSE:
		self.status = status.RUN
	case status.RUN:
		self.status = status.PAUSE
	}
}

// 终止任务
func (self *scheduler) Stop() {
	self.Lock()
	defer self.Unlock()

	self.status = status.STOP

	// 删除队列中未执行请求的去重记录
	for _, v := range self.queue {
		for _, vv := range v {
			for _, req := range vv {
				self.DelDeduplication(req.GetUrl() + req.GetMethod())
			}
		}
	}
	// 清空队列
	self.count = make(chan bool, cap(self.count))
	self.queue = make(map[int]map[int][]*context.Request)
	self.index = make(map[int][]int)
	self.rwMutexes = make(map[int]*sync.RWMutex)
}

func (self *scheduler) IsStop() bool {
	self.RLock()
	defer self.RUnlock()
	return self.status == status.STOP
}

func (self *scheduler) IsEmpty(spiderId int) bool {
	self.RLock()
	defer self.RUnlock()

	if self.status == status.STOP {
		return true
	}

	for i, count := 0, len(self.queue[spiderId]); i < count; i++ {
		idx := self.index[spiderId][i]
		if len(self.queue[spiderId][idx]) > 0 {
			return false
		}
	}
	return true
}

func (self *scheduler) Deduplicate(key interface{}) bool {
	return self.deduplication.Compare(key)
}

func (self *scheduler) DelDeduplication(key interface{}) {
	self.deduplication.Remove(key)
}

func (self *scheduler) SaveDeduplication() {
	self.deduplication.Submit(cache.Task.OutType)
}

// 检查指定蜘蛛的优先级是否存在
func (self *scheduler) foundPriority(spiderId, priority int) (found bool) {
	defer func() {
		if recover() != nil {
			found = false
		}
	}()
	_, found = self.queue[spiderId][priority]
	return
}

// 登记指定蜘蛛的优先级
func (self *scheduler) addPriority(spiderId, priority int) {
	defer func() {
		recover()
	}()

	self.rwMutexes[spiderId].Lock()
	defer self.rwMutexes[spiderId].Unlock()

	self.index[spiderId] = append(self.index[spiderId], priority)
	sort.Ints(self.index[spiderId]) // 从小到大排序

	self.queue[spiderId][priority] = []*context.Request{}
}
