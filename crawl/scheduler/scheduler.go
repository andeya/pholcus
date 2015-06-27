package scheduler

import (
	"github.com/henrylee2cn/pholcus/common/deduplicate"
	"github.com/henrylee2cn/pholcus/crawl/downloader/context"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"sync"
)

type Scheduler interface {
	// 采集非重复url并返回对比结果，重复为true
	Compare(string) bool
	Stop()
	IsStop() bool
	SrcManager
	// // 存入
	// Push(*context.Request)
	// // 取出
	// Use(int) *context.Request
	// // 释放一个资源
	// Free()
	// // 资源队列是否闲置
	// IsEmpty(int) bool
	// IsAllEmpty() bool

	// // 情况全部队列
	// ClearAll()
}

type scheduler struct {
	*SrcManage
	*deduplicate.Deduplication
	status int
}

// 定义全局调度
var Sdl Scheduler

func Init(capacity uint) {
	Sdl = new(capacity)
}

func new(capacity uint) Scheduler {
	return &scheduler{
		SrcManage:     NewSrcManage(capacity).(*SrcManage),
		Deduplication: deduplicate.New().(*deduplicate.Deduplication),
		status:        status.RUN,
	}
}

var pushMutex sync.Mutex

// 添加请求到队列
func (self *scheduler) Push(req *context.Request) {
	pushMutex.Lock()
	defer func() {
		pushMutex.Unlock()
	}()

	if self.status == status.STOP {
		return
	}
	is := self.Compare(req.GetUrl() + req.GetMethod())
	// 有重复则返回
	if is {
		return
	}

	// 留作未来分发请求用
	// if pholcus.Self.GetRunMode() == config.SERVER || req.CanOutsource() {
	// 	return
	// }

	self.SrcManage.Push(req)
}

func (self *scheduler) Compare(url string) bool {
	return self.Deduplication.Compare(url)
}

func (self *scheduler) Use(spiderId int) (req *context.Request) {
	if self.status == status.STOP {
		return nil
	}
	return self.SrcManage.Use(spiderId)
}

func (self *scheduler) Stop() {
	self.status = status.STOP
	self.SrcManage.ClearAll()
}

func (self *scheduler) IsStop() bool {
	if self.status == status.STOP {
		return true
	}
	return false
}
