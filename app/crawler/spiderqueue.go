package crawler

import (
	. "github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
)

// 采集引擎中规则队列
type (
	SpiderQueue interface {
		Reset() //重置清空队列
		Add(*Spider)
		AddAll([]*Spider)
		AddKeyins(string) //为队列成员遍历添加Keyin属性，但前提必须是队列成员未被添加过keyin
		GetByIndex(int) *Spider
		GetByName(string) *Spider
		GetAll() []*Spider
		Len() int // 返回队列长度
	}
	sq struct {
		list []*Spider
	}
)

func NewSpiderQueue() SpiderQueue {
	return &sq{
		list: []*Spider{},
	}
}

func (self *sq) Reset() {
	self.list = []*Spider{}
}

func (self *sq) Add(sp *Spider) {
	sp.SetId(self.Len())
	self.list = append(self.list, sp)
}

func (self *sq) AddAll(list []*Spider) {
	for _, v := range list {
		self.Add(v)
	}
}

// 添加keyin，遍历蜘蛛队列得到新的队列（已被显式赋值过的spider将不再重新分配Keyin）
func (self *sq) AddKeyins(keyins string) {
	keyinSlice := util.KeyinsParse(keyins)
	if len(keyinSlice) == 0 {
		return
	}

	unit1 := []*Spider{} // 不可被添加自定义配置的蜘蛛
	unit2 := []*Spider{} // 可被添加自定义配置的蜘蛛
	for _, v := range self.GetAll() {
		if v.GetKeyin() == KEYIN {
			unit2 = append(unit2, v)
			continue
		}
		unit1 = append(unit1, v)
	}

	if len(unit2) == 0 {
		logs.Log.Warning("本批任务无需填写自定义配置！\n")
		return
	}

	self.Reset()

	for _, keyin := range keyinSlice {
		for _, v := range unit2 {
			v.Keyin = keyin
			nv := *v
			self.Add((&nv).Copy())
		}
	}
	if self.Len() == 0 {
		self.AddAll(append(unit1, unit2...))
	}

	self.AddAll(unit1)
}

func (self *sq) GetByIndex(idx int) *Spider {
	return self.list[idx]
}

func (self *sq) GetByName(n string) *Spider {
	for _, sp := range self.list {
		if sp.GetName() == n {
			return sp
		}
	}
	return nil
}

func (self *sq) GetAll() []*Spider {
	return self.list
}

func (self *sq) Len() int {
	return len(self.list)
}
