package mgo

import (
	"encoding/json"
	"errors"
)

// 基本的增删改查
type mgoOperator interface {
	Exec() (result interface{}, err error)
}
type mgoType func() mgoOperator

var mgoRouter = make(map[string]mgoType)

// 增删改查操作路由
func init() {
	// 传入数据库列表 | 返回数据库及其集合树
	mgoRouter["list"] = func() mgoOperator { return new(List) }
	// 传入数据库与集合名 | 返回文档总数
	mgoRouter["count"] = func() mgoOperator { return new(Count) }
	// 在指定集合进行条件查询
	mgoRouter["find"] = func() mgoOperator { return new(Find) }
	// 插入新数据
	mgoRouter["insert"] = func() mgoOperator { return new(Insert) }
	// 更新数据
	mgoRouter["update"] = func() mgoOperator { return new(Update) }
	// 删除数据
	mgoRouter["remove"] = func() mgoOperator { return new(Remove) }
}

func Mgo(operate string, option map[string]interface{}) (result interface{}, err error) {
	creat, ok := mgoRouter[operate]
	if !ok {
		return nil, errors.New("the mgo-operate " + operate + " does not exist!")
	}

	b, err := json.Marshal(option)
	if err != nil {
		return nil, err
	}

	o := creat()

	err = json.Unmarshal(b, o)
	if err != nil {
		return nil, err
	}
	return o.Exec()
}
