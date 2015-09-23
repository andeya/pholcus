package mgo

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
)

// 在指定集合进行条件查询
type Find struct {
	Database   string                 // 数据库
	Collection string                 // 集合
	Query      map[string]interface{} // 查询语句
	Sort       []string               // 排序，用法如Sort("firstname", "-lastname")，优先按firstname正序排列，其次按lastname倒序排列
	Skip       int                    // 跳过前n个文档
	Limit      int                    // 返回最多n个文档
	Select     interface{}            // 只查询、返回指定字段，如{"name":1}
	// Result     struct {
	// 	Docs  []interface{}
	// 	Total int
	// }
}

func (self *Find) Exec() (interface{}, error) {
	s, c, err := Open(self.Database, self.Collection)
	defer Close(s)
	if err != nil {
		return nil, err
	}

	if id, ok := self.Query["_id"]; ok {
		if idStr, ok2 := id.(string); !ok2 {
			return nil, errors.New("参数 _id 必须为 string 类型！")
		} else {
			self.Query["_id"] = bson.ObjectIdHex(idStr)
		}
	}

	var result = make(map[string]interface{})

	q := c.Find(self.Query)

	result["total"], _ = q.Count()

	if len(self.Sort) > 0 {
		q.Sort(self.Sort...)
	}

	if self.Skip > 0 {
		q.Skip(self.Skip)
	}

	if self.Limit > 0 {
		q.Limit(self.Limit)
	}

	if self.Select != nil {
		q.Select(self.Select)
	}

	err = q.All(result["docs"])
	return result, err
}
