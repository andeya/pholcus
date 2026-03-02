package request

import (
	"encoding/json"
	"reflect"

	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/logs"
)

type Temp map[string]interface{}

// get returns temporary cached data by deserializing from JSON.
func (t Temp) get(key string, defaultValue interface{}) interface{} {
	defer func() {
		if p := recover(); p != nil {
			logs.Log().Error(" *     Request.Temp.Get(%v): %v", key, p)
		}
	}()

	var (
		err error
		b   = util.String2Bytes(t[key].(string))
	)

	if reflect.TypeOf(defaultValue).Kind() == reflect.Ptr {
		err = json.Unmarshal(b, defaultValue)
	} else {
		err = json.Unmarshal(b, &defaultValue)
	}
	if err != nil {
		logs.Log().Error(" *     Request.Temp.Get(%v): %v", key, err)
	}
	return defaultValue
}

func (t Temp) set(key string, value interface{}) Temp {
	b, err := json.Marshal(value)
	if err != nil {
		logs.Log().Error(" *     Request.Temp.Set(%v): %v", key, err)
	}
	t[key] = util.Bytes2String(b)
	return t
}
