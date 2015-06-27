// 支持优先级的矩阵型队列
package scheduler

import (
	"github.com/henrylee2cn/pholcus/crawl/downloader/context"
)

const (
	// 允许的最高优先级
	MAX_PRIORITY = 5
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
	count chan bool
	// 蜘蛛编号spiderId----->请求优先级priority队列
	queue map[int]([][]*context.Request)
}

func NewSrcManage(capacity uint) SrcManager {
	return &SrcManage{
		count: make(chan bool, int(capacity)),
		queue: make(map[int][][]*context.Request),
	}
}

func (self *SrcManage) Push(req *context.Request) {
	if spiderId, ok := req.GetSpiderId(); ok {
		priority := int(req.GetPriority())
		if priority > MAX_PRIORITY {
			priority = MAX_PRIORITY
		}

		for i, x := 0, priority+1-len(self.queue[spiderId]); i < x; i++ {
			self.queue[spiderId] = append(self.queue[spiderId], []*context.Request{})
		}

		self.queue[spiderId][priority] = append(self.queue[spiderId][priority], req)
	}
}

func (self *SrcManage) Use(spiderId int) (req *context.Request) {
	for i := len(self.queue[spiderId]) - 1; i >= 0; i-- {
		if len(self.queue[spiderId][i]) > 0 {
			req = self.queue[spiderId][i][0]
			self.queue[spiderId][i] = self.queue[spiderId][i][1:]
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
		if len(self.queue[spiderId][i]) > 0 {
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
	self.queue = make(map[int][][]*context.Request)
}
