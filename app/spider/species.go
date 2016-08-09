package spider

import (
	"github.com/henrylee2cn/pholcus/common/pinyin"
)

// 蜘蛛种类列表
type SpiderSpecies struct {
	list   []*Spider
	sorted bool
}

// 全局蜘蛛种类实例
var Species = &SpiderSpecies{
	list: []*Spider{},
}

// 向蜘蛛种类清单添加新种类
func (self *SpiderSpecies) Add(sp *Spider) *Spider {
	self.list = append(self.list, sp)
	return sp
}

// 获取全部蜘蛛种类
func (self *SpiderSpecies) Get() []*Spider {
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

func (self *SpiderSpecies) GetByName(n string) *Spider {
	for _, sp := range self.list {
		if sp.GetName() == n {
			return sp
		}
	}
	return nil
}
