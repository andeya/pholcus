package scheduler

import (
	"github.com/henrylee2cn/pholcus/common/deduplicate"
	"github.com/henrylee2cn/pholcus/downloader/context"
)

type Scheduler interface {
	// 采集非重复url并返回对比结果，重复为true
	Compare(string) bool

	SrcManager
	// 以下为具体方法列表
	// 存入
	// Push(*context.Request)
	// 取出
	// Use(string) *context.Request
	// 释放一个资源
	// Free()
	// 资源队列是否闲置
	// IsEmpty(string) bool
	// IsAllEmpty() bool

}

type scheduler struct {
	*SrcManage
	*deduplicate.Deduplication
}

func New(capacity uint) Scheduler {
	return &scheduler{
		SrcManage:     NewSrcManage(capacity).(*SrcManage),
		Deduplication: deduplicate.New().(*deduplicate.Deduplication),
	}
}

func (self *scheduler) Push(req *context.Request) {
	is := self.Compare(req.GetUrl())
	// 有重复则返回
	if is {
		return
	}
	self.SrcManage.Push(req)
}

func (self *scheduler) Compare(url string) bool {
	return self.Deduplication.Compare(url)
}

// 定义全局调度
var Self Scheduler

func Init(capacity uint) Scheduler {
	Self = New(capacity)
	return Self
}
