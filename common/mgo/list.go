package mgo

// 传入数据库列表 | 返回数据库及其集合树
type List struct {
	Dbs []string //数据库名称列表
	// Result struct {
	// 	Tree map[string][]string
	// }
}

func (self *List) Exec() (interface{}, error) {
	s := MgoPool.GetOne().(*MgoFish)
	defer MgoPool.Free(s)
	var err error
	var dbs []string
	if dbs, err = s.DatabaseNames(); err != nil {
		return nil, err
	}

	var result = make(map[string][]string)

	if len(self.Dbs) == 0 {
		for _, dbname := range dbs {
			result[dbname], _ = s.DB(dbname).CollectionNames()
		}
		return result, err
	}

	for _, dbname := range self.Dbs {
		result[dbname], _ = s.DB(dbname).CollectionNames()
	}
	return result, err
}
