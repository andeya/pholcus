package mgo

import (
	"reflect"
	"strings"

	"github.com/andeya/gust/result"
)

// Mgo is the unified entry for CRUD operations.
// resultPtr types: count=*int, list=*map[string][]string, find=*map[string]interface{},
// insert=*[]string (may be nil to skip IDs), update/remove=nil.
func Mgo(resultPtr interface{}, operate string, option map[string]interface{}) result.Result[any] {
	o := getOperator(operate)
	if o == nil {
		return result.FmtErr[any]("the db-operate %s does not exist!", operate)
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

// Operator defines the interface for CRUD operations.
type Operator interface {
	Exec(resultPtr interface{}) result.Result[any]
}

// getOperator returns the Operator for the given operation name.
func getOperator(operate string) Operator {
	switch strings.ToLower(operate) {
	case "list":
		return new(List)
	case "count":
		return new(Count)
	case "find":
		return new(Find)
	case "insert":
		return new(Insert)
	case "update":
		return new(Update)
	case "update_all":
		return new(UpdateAll)
	case "upsert":
		return new(Upsert)
	case "remove":
		return new(Remove)

	default:
		return nil
	}
}
