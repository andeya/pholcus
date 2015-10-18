package mgo

//基础查询
import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

// 传入数据库与集合名 | 返回文档总数
type Count struct {
	Database   string                 // 数据库
	Collection string                 // 集合
	Query      map[string]interface{} // 查询语句
}

func (self *Count) Exec(resultPtr interface{}) (err error) {
	defer func() {
		if re := recover(); re != nil {
			err = fmt.Errorf("%v", re)
		}
	}()
	resultPtr2 := resultPtr.(*int)
	*resultPtr2 = 0

	s, c, err := Open(self.Database, self.Collection)
	defer Close(s)
	if err != nil {
		return err
	}

	if id, ok := self.Query["_id"]; ok {
		if idStr, ok2 := id.(string); !ok2 {
			err = fmt.Errorf("%v", "参数 _id 必须为 string 类型！")
			return err
		} else {
			self.Query["_id"] = bson.ObjectIdHex(idStr)
		}
	}

	*resultPtr2, err = c.Find(self.Query).Count()

	return err
}
