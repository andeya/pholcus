package crawler

import (
	"github.com/andeya/gust/option"
	spider "github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/logs"
)

// SpiderQueue holds the spider rule queue for the crawler engine.
type (
	SpiderQueue interface {
		Reset() // Reset clears the queue
		Add(*spider.Spider)
		AddAll([]*spider.Spider)
		AddKeyins(string) // AddKeyins assigns Keyin to queue members that have not been assigned yet
		GetByIndex(int) *spider.Spider
		GetByIndexOpt(int) option.Option[*spider.Spider]
		GetByName(string) *spider.Spider
		GetByNameOpt(string) option.Option[*spider.Spider]
		GetAll() []*spider.Spider
		Len() int // Len returns the queue length
	}
	sq struct {
		list []*spider.Spider
	}
)

// NewSpiderQueue creates a new spider queue.
func NewSpiderQueue() SpiderQueue {
	return &sq{
		list: []*spider.Spider{},
	}
}

// Reset clears the spider queue.
func (self *sq) Reset() {
	self.list = []*spider.Spider{}
}

// Add appends a spider to the queue.
func (self *sq) Add(sp *spider.Spider) {
	sp.SetId(self.Len())
	self.list = append(self.list, sp)
}

// AddAll appends all spiders in the list to the queue.
func (self *sq) AddAll(list []*spider.Spider) {
	for _, v := range list {
		self.Add(v)
	}
}

// AddKeyins iterates over the spider queue and assigns Keyin values.
// Spiders that already have an explicit Keyin are not reassigned.
func (self *sq) AddKeyins(keyins string) {
	keyinSlice := util.KeyinsParse(keyins)
	if len(keyinSlice) == 0 {
		return
	}

	unit1 := []*spider.Spider{} // spiders that cannot receive custom config
	unit2 := []*spider.Spider{} // spiders that can receive custom config
	for _, v := range self.GetAll() {
		if v.GetKeyin() == spider.KEYIN {
			unit2 = append(unit2, v)
			continue
		}
		unit1 = append(unit1, v)
	}

	if len(unit2) == 0 {
		logs.Log.Warning("This batch of tasks does not require custom configuration.\n")
		return
	}

	self.Reset()

	for _, keyin := range keyinSlice {
		for _, v := range unit2 {
			v.Keyin = keyin
			self.Add(v.Copy())
		}
	}
	if self.Len() == 0 {
		self.AddAll(append(unit1, unit2...))
	}

	self.AddAll(unit1)
}

// GetByIndex returns the spider at the given index.
func (self *sq) GetByIndex(idx int) *spider.Spider {
	return self.GetByIndexOpt(idx).UnwrapOr(nil)
}

// GetByIndexOpt returns the spider at the given index as Option; None if out of range.
func (self *sq) GetByIndexOpt(idx int) option.Option[*spider.Spider] {
	if idx >= 0 && idx < len(self.list) {
		return option.Some(self.list[idx])
	}
	return option.None[*spider.Spider]()
}

// GetByName returns the spider with the given name, or nil if not found.
func (self *sq) GetByName(n string) *spider.Spider {
	return self.GetByNameOpt(n).UnwrapOr(nil)
}

// GetByNameOpt returns the spider with the given name as Option.
func (self *sq) GetByNameOpt(n string) option.Option[*spider.Spider] {
	for _, sp := range self.list {
		if sp.GetName() == n {
			return option.Some(sp)
		}
	}
	return option.None[*spider.Spider]()
}

// GetAll returns all spiders in the queue.
func (self *sq) GetAll() []*spider.Spider {
	return self.list
}

// Len returns the number of spiders in the queue.
func (self *sq) Len() int {
	return len(self.list)
}
