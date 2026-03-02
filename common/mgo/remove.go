package mgo

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
)

// Remove deletes documents matching the selector.
type Remove struct {
	Database   string                 // database name
	Collection string                 // collection name
	Selector   map[string]interface{} // document selector
}

func (self *Remove) Exec(_ interface{}) (r result.Result[any]) {
	defer r.Catch()
	Call(func(src pool.Src) error {
		c := src.(*MgoSrc).DB(self.Database).C(self.Collection)

		if id, ok := self.Selector["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "parameter _id must be of string type")
			} else {
				self.Selector["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		return c.Remove(self.Selector)
	}).Unwrap()
	return result.Ok[any](nil)
}
