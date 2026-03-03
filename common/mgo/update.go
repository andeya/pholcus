package mgo

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
)

// Update updates the first document matching the selector.
type Update struct {
	Database   string                 // database name
	Collection string                 // collection name
	Selector   map[string]interface{} // document selector
	Change     map[string]interface{} // update document
}

func (u *Update) Exec(_ interface{}) (r result.Result[any]) {
	defer r.Catch()
	Call(func(src pool.Src) error {
		c := getSessionFunc(src).DB(u.Database).C(u.Collection)

		if id, ok := u.Selector["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "parameter _id must be of string type")
			} else {
				u.Selector["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		return c.Update(u.Selector, u.Change)
	}).Unwrap()
	return result.Ok[any](nil)
}
