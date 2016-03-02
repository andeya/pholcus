package mgo

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/henrylee2cn/pholcus/common/pool"
)

// 删除数据
type Remove struct {
	Database   string                 // 数据库
	Collection string                 // 集合
	Selector   map[string]interface{} // 文档选择器
}

func (self *Remove) Exec(_ interface{}) error {
	return Call(func(src pool.Src) error {
		c := src.(*MgoSrc).DB(self.Database).C(self.Collection)

		if id, ok := self.Selector["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "参数 _id 必须为 string 类型！")
			} else {
				self.Selector["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		return c.Remove(self.Selector)
	})
}
