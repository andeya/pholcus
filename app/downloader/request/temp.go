package request

import (
	"encoding/json"
	"reflect"

	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
)

type Temp map[string]interface{}

// 返回临时缓存数据
func (self Temp) get(key string, defaultValue interface{}) interface{} {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error(" *     Request.Temp.Get(%v): %v", key, p)
		}
	}()

	var (
		err error
		b   = util.String2Bytes(self[key].(string))
	)

	if reflect.TypeOf(defaultValue).Kind() == reflect.Ptr {
		err = json.Unmarshal(b, defaultValue)
	} else {
		err = json.Unmarshal(b, &defaultValue)
	}
	if err != nil {
		logs.Log.Error(" *     Request.Temp.Get(%v): %v", key, err)
	}
	return defaultValue
}

func (self Temp) set(key string, value interface{}) Temp {
	b, err := json.Marshal(value)
	if err != nil {
		logs.Log.Error(" *     Request.Temp.Set(%v): %v", key, err)
	}
	self[key] = util.Bytes2String(b)
	return self
}
