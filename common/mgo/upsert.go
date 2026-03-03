package mgo

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
)

// Upsert updates the first matching document, or inserts if none match.
type Upsert struct {
	Database   string                 // database name
	Collection string                 // collection name
	Selector   map[string]interface{} // document selector
	Change     map[string]interface{} // update document
}

func (us *Upsert) Exec(resultPtr interface{}) (r result.Result[any]) {
	defer r.Catch()
	resultPtr2 := resultPtr.(*map[string]interface{})
	*resultPtr2 = map[string]interface{}{}

	Call(func(src pool.Src) error {
		c := getSessionFunc(src).DB(us.Database).C(us.Collection)

		if id, ok := us.Selector["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "parameter _id must be of string type")
			} else {
				us.Selector["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		info, err := c.Upsert(us.Selector, us.Change)
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
