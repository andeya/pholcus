package crawl

import (
	"errors"
	. "github.com/henrylee2cn/pholcus/app/spider"
	"strings"
)

// 运行时蜘蛛队列
type SpiderQueue interface {
	Reset() //重置清空队列
	Add(*Spider)
	AddAll([]*Spider)
	AddKeywords(string) error //为队列成员遍历添加Keyword属性，但前提必须是队列成员未被添加过keyword
	GetByIndex(int) *Spider
	GetByName(string) *Spider
	GetAll() []*Spider
	Len() int // 返回队列长度
}

type sq struct {
	list       []*Spider
	hasKeyWord bool
}

func NewSpiderQueue() SpiderQueue {
	return &sq{
		list: []*Spider{},
	}
}

func (self *sq) Reset() {
	self.list = []*Spider{}
}

func (self *sq) Add(sp *Spider) {
	sp.Id = self.Len()
	self.list = append(self.list, sp)
}

func (self *sq) AddAll(list []*Spider) {
	for _, v := range list {
		self.Add(v)
	}
}

// 添加keyword，遍历蜘蛛队列得到新的队列（已被显式赋值过的spider将不再重新分配Keyword）
func (self *sq) AddKeywords(keywords string) error {
	if keywords == "" {
		return errors.New("遍历关键词失败：keywords 不能为空！")
	}
	// 不可被添加kw的蜘蛛
	unit1 := []*Spider{}
	// 可被添加kw的蜘蛛
	unit2 := []*Spider{}
	for _, v := range self.GetAll() {
		if v.GetKeyword() == USE {
			unit2 = append(unit2, v)
			continue
		}
		unit1 = append(unit1, v)
	}

	if len(unit2) == 0 {
		return errors.New("遍历关键词失败：没有可被添加的蜘蛛！")
	}

	self.Reset()

	keywordSlice := strings.Split(keywords, "|")
	for _, keyword := range keywordSlice {
		keyword = strings.Trim(keyword, " ")
		if keyword == "" {
			continue
		}
		for _, v := range unit2 {
			v.Keyword = keyword
			nv := *v
			self.Add((&nv).Gost())
		}
	}
	if self.Len() == 0 {
		self.AddAll(append(unit1, unit2...))
	}

	self.AddAll(unit1)
	return nil
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
