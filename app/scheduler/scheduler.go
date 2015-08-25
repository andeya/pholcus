package scheduler

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/common/deduplicate"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"sync"
)

type Scheduler interface {
	// 采集非重复url并返回对比结果，重复为true
	Compare(string) bool
	PauseRecover() // 暂停\恢复所有爬行任务
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
	Sdl = newScheduler(capacity)
}

func newScheduler(capacity uint) Scheduler {
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

	// 有重复则返回
	if self.Compare(req.GetUrl() + req.GetMethod()) {
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
	if self.status != status.RUN {
		return nil
	}
	return self.SrcManage.Use(spiderId)
}

// 暂停\恢复所有爬行任务
func (self *scheduler) PauseRecover() {
	switch self.status {
	case status.PAUSE:
		self.status = status.RUN
	case status.RUN:
		self.status = status.PAUSE
	}
}

func (self *scheduler) Stop() {
	self.status = status.STOP
	self.SrcManage.ClearAll()
}

func (self *scheduler) IsStop() bool {
	return self.status == status.STOP
}
