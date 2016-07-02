package spider

import (
	"github.com/henrylee2cn/pholcus/common/pinyin"
)

// 蜘蛛种类列表
type (
	Traversal interface {
		Add(*Spider) *Spider
		Get() []*Spider
		GetByName(string) *Spider
	}

	menu struct {
		list   []*Spider
		sorted bool
	}
)

var Menu = newTraversal()

func newTraversal() Traversal {
	return &menu{
		list: []*Spider{},
	}
}

// 向蜘蛛种类清单添加新种类
func (self *menu) Add(sp *Spider) *Spider {
	self.list = append(self.list, sp)
	return sp
}

// 获取全部蜘蛛种类
func (self *menu) Get() []*Spider {
	if !self.sorted {
		initials := make([]string, len(self.list))
		spMap := map[string]*Spider{}
		for i, sp := range self.list {
			initials[i] = sp.GetName()
			spMap[initials[i]] = sp
		}
		pinyin.SortInitials(initials)
		for i, initail := range initials {
			self.list[i] = spMap[initail]
		}
		self.sorted = true
	}
	return self.list
}

func (self *menu) GetByName(n string) *Spider {
	for _, sp := range self.list {
		if sp.GetName() == n {
			return sp
		}
	}
	return nil
}
