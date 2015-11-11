package spider

import (
	"github.com/henrylee2cn/pholcus/common/pinyin"
)

// 蜘蛛种类接口
type Traversal interface {
	Add(*Spider)
	Get() []*Spider
	GetByName(string) *Spider
}

// 蜘蛛种类清单
type menu struct {
	list   []*Spider
	sorted bool
}

func newTraversal() Traversal {
	return &menu{
		list: []*Spider{},
	}
}

var Menu = newTraversal()

// 向蜘蛛种类清单添加新种类
func (self *menu) Add(sp *Spider) {
	self.list = append(self.list, sp)
}

// 获取全部蜘蛛种类
func (self *menu) Get() []*Spider {
	if !self.sorted {
		l := len(self.list)
		initials := make([]string, l)
		newlist := map[string]*Spider{}
		for i := 0; i < l; i++ {
			initials[i] = self.list[i].GetName()
			newlist[initials[i]] = self.list[i]
		}
		pinyin.SortInitials(initials)
		for i := 0; i < l; i++ {
			self.list[i] = newlist[initials[i]]
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
