package mgo

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
)

// Count returns the number of documents matching the query.
type Count struct {
	Database   string                 // database name
	Collection string                 // collection name
	Query      map[string]interface{} // query filter
}

func (cnt *Count) Exec(resultPtr interface{}) (r result.Result[any]) {
	defer r.Catch()
	resultPtr2 := resultPtr.(*int)
	*resultPtr2 = 0

	Call(func(src pool.Src) error {
		c := src.(*MgoSrc).DB(cnt.Database).C(cnt.Collection)

		if id, ok := cnt.Query["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "parameter _id must be of string type")
			} else {
				cnt.Query["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		var err error
		*resultPtr2, err = c.Find(cnt.Query).Count()
		return err
	}).Unwrap()
	return result.Ok[any](*resultPtr2)
}
