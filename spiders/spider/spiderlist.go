package spider

type Spiders struct {
	list []*Spider
}

func (self *Spiders) Init() {
	self.list = []*Spider{}
}

func (self *Spiders) Add(sp *Spider) {
	self.list = append(self.list, sp)
}

func (self *Spiders) Len() int {
	return len(self.list)
}

func (self *Spiders) Get(idx int) *Spider {
	return self.list[idx]
}

func (self *Spiders) GetAll() []*Spider {
	return self.list
}

func (self *Spiders) ReSet(list []*Spider) {
	self.list = list
}

var (
	// 任务队列
	List = &Spiders{}

	// GUI菜单列表
	Menu = &Spiders{}
)
