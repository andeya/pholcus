package deduplicate

import (
	"github.com/henrylee2cn/pholcus/common/util"
	"sync"
)

type Deduplicate interface {
	// 采集非重复样本并返回对比结果，重复为true
	Compare(obj interface{}) bool
}

type Deduplication struct {
	sampling map[string]bool
	sync.Mutex
}

func New() Deduplicate {
	return &Deduplication{
		sampling: make(map[string]bool),
	}
}

// 对比是否已存在，不存在则采样
func (self *Deduplication) Compare(obj interface{}) bool {
	self.Mutex.Lock()
	defer self.Unlock()

	s := util.MakeUnique(obj)
	if !self.sampling[s] {
		self.sampling[s] = true
		return false
	}
	return true
}
