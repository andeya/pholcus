package mgo

import (
	"errors"
	"reflect"
	"strings"
)

// 增删改查操作的统一方法
// count操作resultPtr类型为*int
// list操作resultPtr类型为*map[string][]string
// find操作resultPtr类型为*map[string]interface{}
// insert操作resultPtr类型为*[]string，允许为nil(不接收id列表)
// update操作resultPtr为nil
// remove操作resultPtr为nil
func Mgo(resultPtr interface{}, operate string, option map[string]interface{}) error {
	o := getOperator(operate)
	if o == nil {
		return errors.New("the db-operate " + operate + " does not exist!")
	}

	v := reflect.ValueOf(o).Elem()
	for key, val := range option {
		value := v.FieldByName(key)
		if value == (reflect.Value{}) || !value.CanSet() {
			continue
		}
		value.Set(reflect.ValueOf(val))
	}

	return o.Exec(resultPtr)
}

// 增删改查操作
type Operator interface {
	Exec(resultPtr interface{}) (err error)
}

// 增删改查操作列表
func getOperator(operate string) Operator {
	switch strings.ToLower(operate) {
	// 传入数据库列表 | 返回数据库及其集合树
	case "list":
		return new(List)

	// 传入数据库与集合名 | 返回文档总数
	case "count":
		return new(Count)

	// 在指定集合进行条件查询
	case "find":
		return new(Find)

	// 插入新数据
	case "insert":
		return new(Insert)

	// 更新第一个匹配的数据
	case "update":
		return new(Update)

	// 更新全部匹配的数据
	case "update_all":
		return new(UpdateAll)

	// 更新第一个匹配的数据，若无匹配项则插入
	case "upsert":
		return new(Upsert)

	// 删除数据
	case "remove":
		return new(Remove)

	default:
		return nil
	}
}
