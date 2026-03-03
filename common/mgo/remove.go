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

func (r *Remove) Exec(_ interface{}) (res result.Result[any]) {
	defer res.Catch()
	Call(func(src pool.Src) error {
		c := getSessionFunc(src).DB(r.Database).C(r.Collection)

		if id, ok := r.Selector["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "parameter _id must be of string type")
			} else {
				r.Selector["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		return c.Remove(r.Selector)
	}).Unwrap()
	return result.Ok[any](nil)
}
