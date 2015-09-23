package mgo

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
)

// 删除数据
type Remove struct {
	Database   string                 // 数据库
	Collection string                 // 集合
	Selector   map[string]interface{} // 文档选择器
}

func (self *Remove) Exec() (interface{}, error) {
	s, c, err := Open(self.Database, self.Collection)
	defer Close(s)
	if err != nil {
		return nil, err
	}

	if id, ok := self.Selector["_id"]; ok {
		if idStr, ok2 := id.(string); !ok2 {
			return nil, errors.New("参数 _id 必须为string类型！")
		} else {
			self.Selector["_id"] = bson.ObjectIdHex(idStr)
		}
	}

	return nil, c.Remove(self.Selector)
}
