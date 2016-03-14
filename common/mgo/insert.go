package mgo

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/henrylee2cn/pholcus/common/pool"
)

// 插入新数据
type Insert struct {
	Database   string                   // 数据库
	Collection string                   // 集合
	Docs       []map[string]interface{} // 文档
}

const (
	MaxLen int = 5000 //分配插入
)

func (self *Insert) Exec(resultPtr interface{}) (err error) {
	defer func() {
		if re := recover(); re != nil {
			err = fmt.Errorf("%v", re)
		}
	}()
	var (
		resultPtr2 = new([]string)
		count      = len(self.Docs)
		docs       = make([]interface{}, count)
	)
	if resultPtr != nil {
		resultPtr2 = resultPtr.(*[]string)
	}
	*resultPtr2 = make([]string, count)

	return Call(func(src pool.Src) error {
		c := src.(*MgoSrc).DB(self.Database).C(self.Collection)
		for i, doc := range self.Docs {
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
	})
}
