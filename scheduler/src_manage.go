package scheduler

import (
	"github.com/henrylee2cn/pholcus/downloader/context"
	// "github.com/henrylee2cn/pholcus/reporter"
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
}

type SrcManage struct {
	count chan bool
	queue map[int][]*context.Request
}

func NewSrcManage(capacity uint) SrcManager {
	return &SrcManage{
		count: make(chan bool, int(capacity)),
		queue: make(map[int][]*context.Request),
	}
}

func (self *SrcManage) Push(req *context.Request) {
	if spiderId, ok := req.GetSpiderId(); ok {
		self.queue[spiderId] = append(self.queue[spiderId], req)
	}
}

func (self *SrcManage) Use(spiderId int) *context.Request {
	if len(self.queue[spiderId]) == 0 {
		return nil
	}
	req := self.queue[spiderId][0]
	self.queue[spiderId] = self.queue[spiderId][1:]
	self.count <- true
	return req
}

func (self *SrcManage) Free() {
	<-self.count
}

func (self *SrcManage) IsEmpty(spiderId int) bool {
	if len(self.queue[spiderId]) == 0 {
		return true
	}
	return false
}

func (self *SrcManage) IsAllEmpty() bool {
	if len(self.count) == 0 {
		for _, v := range self.queue {
			if len(v) != 0 {
				return false
			}
		}
		return true
	}
	return false
}
