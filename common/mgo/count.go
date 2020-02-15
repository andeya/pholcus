package mgo

//基础查询
import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/henrylee2cn/pholcus/common/pool"
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

	err = Call(func(src pool.Src) error {
		c := src.(*MgoSrc).DB(self.Database).C(self.Collection)

		if id, ok := self.Query["_id"]; ok {
			if idStr, ok2 := id.(string); !ok2 {
				return fmt.Errorf("%v", "参数 _id 必须为 string 类型！")
			} else {
				self.Query["_id"] = bson.ObjectIdHex(idStr)
			}
		}

		*resultPtr2, err = c.Find(self.Query).Count()
		return err
	})
	return
}
