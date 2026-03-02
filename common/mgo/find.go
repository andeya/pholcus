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

func (f *Find) Exec(resultPtr interface{}) (r result.Result[any]) {
	defer r.Catch()
	resultPtr2 := resultPtr.(*map[string]interface{})
	*resultPtr2 = map[string]interface{}{}

	Call(func(src pool.Src) error {
		c := src.(*MgoSrc).DB(f.Database).C(f.Collection)

		if id, ok := f.Query["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "parameter _id must be of string type")
			} else {
				f.Query["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		q := c.Find(f.Query)

		total, err := q.Count()
		if err != nil {
			return err
		}
		(*resultPtr2)["Total"] = total

		if len(f.Sort) > 0 {
			q.Sort(f.Sort...)
		}

		if f.Skip > 0 {
			q.Skip(f.Skip)
		}

		if f.Limit > 0 {
			q.Limit(f.Limit)
		}

		if f.Select != nil {
			q.Select(f.Select)
		}
		docs := []interface{}{}
		err = q.All(&docs)

		(*resultPtr2)["Docs"] = docs

		return err
	}).Unwrap()
	return result.Ok[any](*resultPtr2)
}
