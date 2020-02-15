package mgo

import (
	"fmt"

	"github.com/henrylee2cn/pholcus/common/pool"
)

// 传入数据库列表 | 返回数据库及其集合树
type List struct {
	Dbs []string //数据库名称列表
}

func (self *List) Exec(resultPtr interface{}) (err error) {
	defer func() {
		if re := recover(); re != nil {
			err = fmt.Errorf("%v", re)
		}
	}()
	resultPtr2 := resultPtr.(*map[string][]string)
	*resultPtr2 = map[string][]string{}

	err = Call(func(src pool.Src) error {
		var (
			s   = src.(*MgoSrc)
			dbs []string
		)

		if dbs, err = s.DatabaseNames(); err != nil {
			return err
		}

		if len(self.Dbs) == 0 {
			for _, dbname := range dbs {
				(*resultPtr2)[dbname], _ = s.DB(dbname).CollectionNames()
			}
			return err
		}

		for _, dbname := range self.Dbs {
			(*resultPtr2)[dbname], _ = s.DB(dbname).CollectionNames()
		}
		return err
	})

	return
}
