package mgo

import (
	"gopkg.in/mgo.v2/bson"
)

// 插入新数据
type Insert struct {
	Database   string                   // 数据库
	Collection string                   // 集合
	Docs       []map[string]interface{} // 文档
	// Result     struct {
	// 	Ids []string
	// }
}

func (self *Insert) Exec() (interface{}, error) {
	s, c, err := Open(self.Database, self.Collection)
	defer Close(s)
	if err != nil || c == nil {
		return nil, err
	}
	result := []string{}
	var docs []interface{}
	for i, _ := range self.Docs {
		var _id string
		if self.Docs[i]["_id"] == nil || self.Docs[i]["_id"] == interface{}("") || self.Docs[i]["_id"] == interface{}(0) {
			objId := bson.NewObjectId()
			_id = objId.Hex()
			self.Docs[i]["_id"] = objId
		} else {
			_id = self.Docs[i]["_id"].(string)
		}
		result = append(result, _id)
		docs = append(docs, self.Docs[i])
	}
	err = c.Insert(docs...)
	return result, err
}
