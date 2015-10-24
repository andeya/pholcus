package scheduler

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/common/deduplicate"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"sync"
)

type Scheduler interface {
	Init(capacity uint, inheritDeduplication bool, deduplicationTarget string)
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

	// 对比是否存在重复项
	Deduplicate(key interface{}) bool
	// 保存去重记录
	SaveDeduplication()
	// 取消指定去重样本
	DelDeduplication(key interface{})
}

type scheduler struct {
	*SrcManage
	deduplication       deduplicate.Deduplicate
	deduplicationTarget string
	sync.RWMutex
	status int
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

func (self *scheduler) Init(capacity uint, inheritDeduplication bool, deduplicationTarget string) {
	self.SrcManage = NewSrcManage(capacity).(*SrcManage)
	self.status = status.RUN
	if inheritDeduplication {
		if self.deduplicationTarget == deduplicationTarget {
			return
		}
		self.deduplicationTarget = deduplicationTarget
		self.deduplication.ReRead(deduplicationTarget)
	} else {
		self.deduplication.CleanRead()
		self.deduplicationTarget = ""
	}
}

// 添加请求到队列
func (self *scheduler) Push(req *context.Request) {
	self.RWMutex.RLock()
	defer self.RWMutex.RUnlock()

	if self.status == status.STOP {
		return
	}

	// 当req不可重复时，有重复则返回
	if !req.GetDuplicatable() && self.Deduplicate(req.GetUrl()+req.GetMethod()) {
		return
	}

	self.SrcManage.Push(req)
}

func (self *scheduler) Deduplicate(key interface{}) bool {
	return self.deduplication.Compare(key)
}

func (self *scheduler) DelDeduplication(key interface{}) {
	self.deduplication.Remove(key)
}

func (self *scheduler) SaveDeduplication() {
	self.deduplication.Write(cache.Task.DeduplicationTarget)
}

func (self *scheduler) Use(spiderId int) (req *context.Request) {
	self.RWMutex.RLock()
	defer self.RWMutex.RUnlock()

	if self.status != status.RUN {
		return nil
	}
	return self.SrcManage.Use(spiderId)
}

// 暂停\恢复所有爬行任务
func (self *scheduler) PauseRecover() {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	switch self.status {
	case status.PAUSE:
		self.status = status.RUN
	case status.RUN:
		self.status = status.PAUSE
	}
}

func (self *scheduler) Stop() {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	self.status = status.STOP
	if self.SrcManage == nil {
		return
	}
	// 删除队列中未执行的请求的去重记录
	for _, v := range self.SrcManage.GetQueue() {
		for _, vv := range v {
			for _, req := range vv {
				self.DelDeduplication(req.GetUrl() + req.GetMethod())
			}
		}
	}
	// 清空队列
	self.SrcManage.ClearAll()
}

func (self *scheduler) IsStop() bool {
	self.RWMutex.RLock()
	defer self.RWMutex.RUnlock()

	return self.status == status.STOP
}
