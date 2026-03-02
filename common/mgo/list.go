package mgo

import (
	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
)

// List returns a map of database names to their collection names.
type List struct {
	Dbs []string // list of database names to query (empty = all)
}

func (self *List) Exec(resultPtr interface{}) (r result.Result[any]) {
	defer r.Catch()
	resultPtr2 := resultPtr.(*map[string][]string)
	*resultPtr2 = map[string][]string{}

	Call(func(src pool.Src) error {
		s := src.(*MgoSrc)
		dbs, err := s.DatabaseNames()
		if err != nil {
			return err
		}

		if len(self.Dbs) == 0 {
			for _, dbname := range dbs {
				(*resultPtr2)[dbname], err = s.DB(dbname).CollectionNames()
				if err != nil {
					return err
				}
			}
			return nil
		}

		for _, dbname := range self.Dbs {
			(*resultPtr2)[dbname], err = s.DB(dbname).CollectionNames()
			if err != nil {
				return err
			}
		}
		return nil
	}).Unwrap()
	return result.Ok[any](*resultPtr2)
}
