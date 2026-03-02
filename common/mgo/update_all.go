package mgo

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
)

// UpdateAll updates all documents matching the selector.
type UpdateAll struct {
	Database   string                 // database name
	Collection string                 // collection name
	Selector   map[string]interface{} // document selector
	Change     map[string]interface{} // update document
}

func (self *UpdateAll) Exec(resultPtr interface{}) (r result.Result[any]) {
	defer r.Catch()
	resultPtr2 := resultPtr.(*map[string]interface{})
	*resultPtr2 = map[string]interface{}{}

	Call(func(src pool.Src) error {
		c := src.(*MgoSrc).DB(self.Database).C(self.Collection)

		if id, ok := self.Selector["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "parameter _id must be of string type")
			} else {
				self.Selector["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		info, err := c.UpdateAll(self.Selector, self.Change)
		if err != nil {
			return err
		}

		(*resultPtr2)["Updated"] = info.Updated
		(*resultPtr2)["Removed"] = info.Removed
		(*resultPtr2)["UpsertedId"] = info.UpsertedId

		return err
	}).Unwrap()
	return result.Ok[any](nil)
}
