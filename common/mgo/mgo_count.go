package mgo

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
)

// 条件查询
type Count struct {
	Database   string                 // 数据库
	Collection string                 // 集合
	Query      map[string]interface{} // 查询语句
}

func (self *Count) Exec() (result int, err error) {
	s, c, err := Open(self.Database, self.Collection)
	defer Close(s)
	if err != nil {
		return 0, err
	}

	if id, ok := self.Query["_id"]; ok {
		if idStr, ok2 := id.(string); !ok2 {
			return 0, errors.New("参数 _id 必须为string类型！")
		} else {
			self.Query["_id"] = bson.ObjectIdHex(idStr)
		}
	}

	return c.Find(self.Query).Count()
}
