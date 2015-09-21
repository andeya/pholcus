package mgo

// 数据库及其集合列表
type List []string

func NewList(dbnames ...string) List {
	return dbnames
}

func (self List) Exec() (result map[string][]string, err error) {

	s := MgoPool.GetOne().(*MgoFish)
	defer MgoPool.Free(s)

	var dbs []string
	if dbs, err = s.DatabaseNames(); err != nil {
		return nil, err
	}

	result = make(map[string][]string)

	if len(self) == 0 {
		for _, dbname := range dbs {
			result[dbname], _ = s.DB(dbname).CollectionNames()
		}
		return
	}

	for _, dbname := range self {
		result[dbname], _ = s.DB(dbname).CollectionNames()
	}

	return
}
