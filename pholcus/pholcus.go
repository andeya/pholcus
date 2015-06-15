package pholcus

import (
	"github.com/henrylee2cn/pholcus/downloader/context"
	// "github.com/henrylee2cn/pholcus/pholcus/node"
	// "github.com/henrylee2cn/pholcus/pholcus/status"
	"github.com/henrylee2cn/pholcus/scheduler"
	"sync"
)

type Pholcus struct {
	// *node.Node
	// *Status
	isOutsource bool
}

var pushMutex sync.Mutex

func (self *Pholcus) Push(req *context.Request) {
	pushMutex.Lock()
	defer func() {
		pushMutex.Unlock()
	}()
	if !self.TryOutsource(req) {
		scheduler.Sdl.Push(req)
	}
}

func (self *Pholcus) TryOutsource(req *context.Request) bool {
	if self.IsOutsource() && req.TryOutsource() {
		self.Send(*req)
		return true
	}
	return false
}

func (self *Pholcus) SetOutsource(serve bool) {
	self.isOutsource = serve
}

func (self *Pholcus) IsOutsource() bool {
	return self.isOutsource
}

func (self *Pholcus) Send(req context.Request) {

}

func (self *Pholcus) Receive(req context.Request) {
	scheduler.Sdl.Push(&req)
}

// 初始化
var Self *Pholcus

func init() {
	Self = &Pholcus{}
}
