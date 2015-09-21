package mgo

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
)

type Update struct {
	Database   string                 // 数据库
	Collection string                 // 集合
	Selector   map[string]interface{} // 文档选择器
	Change     map[string]interface{} // 文档更新内容
}

func (self *Update) Exec() error {
	s, c, err := Open(self.Database, self.Collection)
	defer Close(s)
	if err != nil {
		return err
	}

	if id, ok := self.Selector["_id"]; ok {
		if idStr, ok2 := id.(string); !ok2 {
			return errors.New("参数 _id 必须为string类型！")
		} else {
			self.Selector["_id"] = bson.ObjectIdHex(idStr)
		}
	}

	return c.Update(self.Selector, self.Change)
}
