// 支持优先级的矩阵型队列
package scheduler

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"sort"
	"sync"
)

// SrcManage is an interface that who want implement an management object can realize these functions.
type SrcManager interface {
	// 注册资源队列
	RegSpider(spiderId int)
	// 存入
	Push(*context.Request)
	// 取出
	Use(int) *context.Request
	// 释放一个资源
	Free()
	// 资源队列是否闲置
	IsEmpty(int) bool
	// IsAllEmpty() bool
	// 情况全部队列
	ClearAll()
}

type SrcManage struct {
	// 全局并发量计数
	count chan bool
	// map[spiderId](map[请求优先级priority]([]请求))，优先级默认为0
	queue map[int](map[int][]*context.Request)
	// map[spiderId]int[存在的请求优先级],优先级从小到大排序
	index map[int][]int
	// map[spiderId]读写锁
	rwMutex map[int]*sync.RWMutex
	// 全局读写锁
	sync.RWMutex
}

func NewSrcManage(capacity uint) SrcManager {
	return &SrcManage{
		count:   make(chan bool, capacity),
		queue:   make(map[int]map[int][]*context.Request),
		index:   make(map[int][]int),
		rwMutex: make(map[int]*sync.RWMutex),
	}
}

// 注册资源队列
func (self *SrcManage) RegSpider(spiderId int) {
	if !self.foundSpider(spiderId) {
		self.addSpider(spiderId)
	}
}

func (self *SrcManage) Push(req *context.Request) {
	// 初始化该蜘蛛的队列
	spiderId, ok := req.GetSpiderId()
	if !ok {
		return
	}

	// 初始化该蜘蛛下该优先级队列
	priority := req.GetPriority()
	if !self.foundPriority(spiderId, priority) {
		self.addPriority(spiderId, priority)
	}

	// 添加请求到队列
	self.queue[spiderId][priority] = append(self.queue[spiderId][priority], req)
}

func (self *SrcManage) Use(spiderId int) (req *context.Request) {
	self.rwMutex[spiderId].Lock()
	defer self.rwMutex[spiderId].Unlock()
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

func (self *SrcManage) Free() {
	<-self.count
}

func (self *SrcManage) IsEmpty(spiderId int) bool {
	self.rwMutex[spiderId].RLock()
	defer self.rwMutex[spiderId].RUnlock()
	for i, count := 0, len(self.queue[spiderId]); i < count; i++ {
		idx := self.index[spiderId][i]
		if len(self.queue[spiderId][idx]) > 0 {
			return false
		}
	}
	return true
}

// func (self *SrcManage) IsAllEmpty() bool {
// 	self.RLock()
// 	defer self.RUnlock()
// 	if len(self.count) > 0 {
// 		return false
// 	}
// 	for k, v := range self.queue {
// 		self.rwMutex[k].RLock()
// 		for _, vv := range v {
// 			if len(vv) > 0 {
// 				return false
// 			}
// 		}
// 		self.rwMutex[k].RUnlock()
// 	}
// 	return true
// }

func (self *SrcManage) GetQueue() map[int]map[int][]*context.Request {
	self.RLock()
	defer self.RUnlock()
	return self.queue
}

func (self *SrcManage) ClearAll() {
	self.Lock()
	defer self.Unlock()
	self.count = make(chan bool, cap(self.count))
	self.queue = make(map[int]map[int][]*context.Request)
	self.index = make(map[int][]int)
	self.rwMutex = make(map[int]*sync.RWMutex)
}

// 检查指定蜘蛛是否存在
func (self *SrcManage) foundSpider(spiderId int) (found bool) {
	self.RLock()
	defer self.RUnlock()
	_, found = self.queue[spiderId]
	return
}

// 登记指定蜘蛛
func (self *SrcManage) addSpider(spiderId int) {
	self.Lock()
	defer self.Unlock()
	self.queue[spiderId] = map[int][]*context.Request{}
	self.index[spiderId] = []int{}
	self.rwMutex[spiderId] = new(sync.RWMutex)
}

// 检查指定蜘蛛的优先级是否存在
func (self *SrcManage) foundPriority(spiderId, priority int) (found bool) {
	self.rwMutex[spiderId].RLock()
	defer self.rwMutex[spiderId].RUnlock()
	_, found = self.queue[spiderId][priority]
	return
}

// 登记指定蜘蛛的优先级
func (self *SrcManage) addPriority(spiderId, priority int) {
	self.rwMutex[spiderId].Lock()
	defer self.rwMutex[spiderId].Unlock()

	self.index[spiderId] = append(self.index[spiderId], priority)
	sort.Ints(self.index[spiderId]) // 从小到大排序

	self.queue[spiderId][priority] = []*context.Request{}
}
