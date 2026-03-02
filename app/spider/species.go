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
func (ss *SpiderSpecies) Add(sp *Spider) *Spider {
	name := sp.Name
	for i := 2; true; i++ {
		if _, ok := ss.hash[name]; !ok {
			sp.Name = name
			ss.hash[sp.Name] = sp
			break
		}
		name = fmt.Sprintf("%s(%d)", sp.Name, i)
	}
	sp.Name = name
	ss.list = append(ss.list, sp)
	return sp
}

// Get returns all registered spiders, sorted by pinyin initials on first call.
// Dynamic spiders are lazily registered on first access.
func (ss *SpiderSpecies) Get() []*Spider {
	RegisterDynamicSpiders()
	if !ss.sorted {
		l := len(ss.list)
		initials := make([]string, l)
		newlist := map[string]*Spider{}
		for i := 0; i < l; i++ {
			initials[i] = ss.list[i].GetName()
			newlist[initials[i]] = ss.list[i]
		}
		pinyin.SortInitials(initials)
		for i := 0; i < l; i++ {
			ss.list[i] = newlist[initials[i]]
		}
		ss.sorted = true
	}
	return ss.list
}

// GetByNameOpt returns the spider with the given name as Option.
func (ss *SpiderSpecies) GetByNameOpt(name string) option.Option[*Spider] {
	if sp, ok := ss.hash[name]; ok {
		return option.Some(sp)
	}
	return option.None[*Spider]()
}
