package spider

import (
	"fmt"

	"github.com/andeya/gust/option"
	"github.com/andeya/pholcus/common/pinyin"
)

// SpiderSpecies is the global registry of available spider types.
type SpiderSpecies struct {
	list   []*Spider
	hash   map[string]*Spider
	sorted bool
}

// Species is the singleton spider registry.
var Species = &SpiderSpecies{
	list: []*Spider{},
	hash: map[string]*Spider{},
}

// Add registers a spider. If the name already exists, a numeric suffix is appended.
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

// Get returns all registered spiders, sorted by pinyin initials on first call.
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

// GetByNameOpt returns the spider with the given name as Option.
func (self *SpiderSpecies) GetByNameOpt(name string) option.Option[*Spider] {
	if sp, ok := self.hash[name]; ok {
		return option.Some(sp)
	}
	return option.None[*Spider]()
}
