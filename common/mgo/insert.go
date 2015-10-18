package mgo

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

// 插入新数据
type Insert struct {
	Database   string                   // 数据库
	Collection string                   // 集合
	Docs       []map[string]interface{} // 文档
}

func (self *Insert) Exec(resultPtr interface{}) (err error) {
	defer func() {
		if re := recover(); re != nil {
			err = fmt.Errorf("%v", re)
		}
	}()
	var resultPtr2 *[]string
	if resultPtr != nil {
		resultPtr2 = resultPtr.(*[]string)
		*resultPtr2 = []string{}
	}

	s, c, err := Open(self.Database, self.Collection)
	defer Close(s)
	if err != nil || c == nil {
		return err
	}

	var docs []interface{}
	for _, doc := range self.Docs {
		var _id string
		if doc["_id"] == nil || doc["_id"] == interface{}("") || doc["_id"] == interface{}(0) {
			objId := bson.NewObjectId()
			_id = objId.Hex()
			doc["_id"] = objId
		} else {
			_id = doc["_id"].(string)
		}

		if resultPtr != nil {
			*resultPtr2 = append(*resultPtr2, _id)
		}

		docs = append(docs, doc)
	}

	err = c.Insert(docs...)

	return err
}
