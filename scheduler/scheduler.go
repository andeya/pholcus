package scheduler

import (
	"github.com/henrylee2cn/pholcus/common/deduplicate"
	"github.com/henrylee2cn/pholcus/downloader/context"
)

const (
	STOP = 0
	RUN  = 1
)

type Scheduler interface {
	// 采集非重复url并返回对比结果，重复为true
	Compare(string) bool
	SrcManager
	Stop()
	IsStop() bool
}

type scheduler struct {
	*SrcManage
	*deduplicate.Deduplication
	status int
}

func New(capacity uint) Scheduler {
	return &scheduler{
		SrcManage:     NewSrcManage(capacity).(*SrcManage),
		Deduplication: deduplicate.New().(*deduplicate.Deduplication),
		status:        RUN,
	}
}

func (self *scheduler) Push(req *context.Request) {
	if self.status == STOP {
		return
	}
	is := self.Compare(req.GetUrl() + req.GetMethod())
	// 有重复则返回
	if is {
		return
	}
	self.SrcManage.Push(req)
}

func (self *scheduler) Compare(url string) bool {
	return self.Deduplication.Compare(url)
}

func (self *scheduler) Use(spiderId int) (req *context.Request) {
	if self.status == STOP {
		return nil
	}
	return self.SrcManage.Use(spiderId)
}

func (self *scheduler) Stop() {
	self.status = STOP
	self.SrcManage.ClearAll()
}

func (self *scheduler) IsStop() bool {
	if self.status == STOP {
		return true
	}
	return false
}

// 定义全局调度
var Sdl Scheduler

func Init(capacity uint) {
	Sdl = New(capacity)
}
