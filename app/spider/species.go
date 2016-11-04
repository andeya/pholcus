package spider

import (
	"fmt"

	"github.com/henrylee2cn/pholcus/common/pinyin"
)

// 蜘蛛种类列表
type SpiderSpecies struct {
	list   []*Spider
	hash   map[string]*Spider
	sorted bool
}

// 全局蜘蛛种类实例
var Species = &SpiderSpecies{
	list: []*Spider{},
	hash: map[string]*Spider{},
}

// 向蜘蛛种类清单添加新种类
func (self *SpiderSpecies) Add(sp *Spider) *Spider {
	name := sp.Name
	for i := 2; true; i++ {
		if _, ok := self.hash[name]; !ok {
			sp.Name = name
			self.hash[sp.Name] = sp
			break
		}
		name = fmt.Sprintf("%s(%d)", sp.Name, i)
	}
	sp.Name = name
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

func (self *SpiderSpecies) GetByName(name string) *Spider {
	return self.hash[name]
}
