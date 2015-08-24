package spider

// 蜘蛛种类接口
type Traversal interface {
	Add(*Spider)
	Get() []*Spider
	GetByName(string) *Spider
}

// 蜘蛛种类清单
type menu struct {
	list []*Spider
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
