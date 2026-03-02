package mgo

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
)

// Find performs a conditional query on the specified collection.
type Find struct {
	Database   string                 // database name
	Collection string                 // collection name
	Query      map[string]interface{} // query filter
	Sort       []string               // sort fields, e.g. Sort("firstname", "-lastname") for asc firstname, desc lastname
	Skip       int                    // skip first n documents
	Limit      int                    // return at most n documents
	Select     interface{}            // projection, e.g. {"name":1} to return only name
}

func (self *Find) Exec(resultPtr interface{}) (r result.Result[any]) {
	defer r.Catch()
	resultPtr2 := resultPtr.(*map[string]interface{})
	*resultPtr2 = map[string]interface{}{}

	Call(func(src pool.Src) error {
		c := src.(*MgoSrc).DB(self.Database).C(self.Collection)

		if id, ok := self.Query["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "parameter _id must be of string type")
			} else {
				self.Query["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		q := c.Find(self.Query)

		total, err := q.Count()
		if err != nil {
			return err
		}
		(*resultPtr2)["Total"] = total

		if len(self.Sort) > 0 {
			q.Sort(self.Sort...)
		}

		if self.Skip > 0 {
			q.Skip(self.Skip)
		}

		if self.Limit > 0 {
			q.Limit(self.Limit)
		}

		if self.Select != nil {
			q.Select(self.Select)
		}
		docs := []interface{}{}
		err = q.All(&docs)

		(*resultPtr2)["Docs"] = docs

		return err
	}).Unwrap()
	return result.Ok[any](*resultPtr2)
}
