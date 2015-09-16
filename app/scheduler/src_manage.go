// 支持优先级的矩阵型队列
package scheduler

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"sort"
	"sync"
)

// SrcManage is an interface that who want implement an management object can realize these functions.
type SrcManager interface {
	// 存入
	Push(*context.Request)
	// 取出
	Use(int) *context.Request
	// 释放一个资源
	Free()
	// 资源队列是否闲置
	IsEmpty(int) bool
	IsAllEmpty() bool

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
	// map[spiderId]并发锁
	mutex map[int]*sync.Mutex
}

func NewSrcManage(capacity uint) SrcManager {
	return &SrcManage{
		count: make(chan bool, capacity),
		queue: make(map[int]map[int][]*context.Request),
		index: make(map[int][]int),
		mutex: make(map[int]*sync.Mutex),
	}
}

func (self *SrcManage) Push(req *context.Request) {
	spiderId, ok := req.GetSpiderId()
	if !ok {
		return
	}

	// 初始化该蜘蛛的队列
	if _, ok := self.queue[spiderId]; !ok {
		self.mutex[spiderId] = new(sync.Mutex)
		self.queue[spiderId] = make(map[int][]*context.Request)
	}

	priority := req.GetPriority()

	// 登记该蜘蛛下该优先级队列
	if _, ok := self.queue[spiderId][priority]; !ok {
		self.uIndex(spiderId, priority)
	}

	// 添加请求到队列
	self.queue[spiderId][priority] = append(self.queue[spiderId][priority], req)
}

func (self *SrcManage) Use(spiderId int) (req *context.Request) {
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

func (self *SrcManage) IsEmpty(spiderId int) (empty bool) {
	empty = true
	for i, count := 0, len(self.queue[spiderId]); i < count; i++ {
		idx := self.index[spiderId][i]
		if len(self.queue[spiderId][idx]) > 0 {
			empty = false
			return
		}
	}
	return
}

func (self *SrcManage) IsAllEmpty() bool {
	if len(self.count) > 0 {
		return false
	}
	for _, v := range self.queue {
		for _, vv := range v {
			if len(vv) > 0 {
				return false
			}
		}
	}
	return true
}

func (self *SrcManage) ClearAll() {
	self.count = make(chan bool, cap(self.count))
	self.queue = make(map[int]map[int][]*context.Request)
	self.index = make(map[int][]int)
	self.mutex = make(map[int]*sync.Mutex)
}

// 登记蜘蛛的优先级
func (self *SrcManage) uIndex(spiderId int, priority int) {
	self.mutex[spiderId].Lock()
	defer self.mutex[spiderId].Unlock()

	self.index[spiderId] = append(self.index[spiderId], priority)

	// 从小到大排序
	sort.Ints(self.index[spiderId])
}
