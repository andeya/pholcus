package mgo

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
)

// Insert inserts new documents into the collection.
type Insert struct {
	Database   string                   // database name
	Collection string                   // collection name
	Docs       []map[string]interface{} // documents to insert
}

const (
	MaxLen int = 5000 // batch size for bulk insert
)

func (i *Insert) Exec(resultPtr interface{}) (r result.Result[any]) {
	defer r.Catch()
	var (
		resultPtr2 = new([]string)
		count      = len(i.Docs)
		docs       = make([]interface{}, count)
	)
	if resultPtr != nil {
		resultPtr2 = resultPtr.(*[]string)
	}
	*resultPtr2 = make([]string, count)

	Call(func(src pool.Src) error {
		c := getSessionFunc(src).DB(i.Database).C(i.Collection)
		for i, doc := range i.Docs {
			var _id string
			if doc["_id"] == nil || doc["_id"] == interface{}("") || doc["_id"] == interface{}(0) {
				objId := bson.NewObjectId()
				_id = objId.Hex()
				doc["_id"] = objId
			} else {
				_id = doc["_id"].(string)
			}

			if resultPtr != nil {
				(*resultPtr2)[i] = _id
			}
			docs[i] = doc
		}
		loop := count / MaxLen
		for i := 0; i < loop; i++ {
			err := c.Insert(docs[i*MaxLen : (i+1)*MaxLen]...)
			if err != nil {
				return err
			}
		}
		if count%MaxLen == 0 {
			return nil
		}
		return c.Insert(docs[loop*MaxLen:]...)
	}).Unwrap()
	return result.Ok[any](nil)
}
