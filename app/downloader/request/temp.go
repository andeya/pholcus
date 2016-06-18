package request

import (
	"encoding/json"
	"reflect"

	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
)

type Temp map[string]interface{}

// 返回临时缓存数据
// 强烈建议数据接收者receive为指针类型
func (self Temp) get(key string, receive interface{}) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error(" *     Request.Temp.Get(%v): %v", key, p)
		}
	}()
	b := util.String2Bytes(self[key].(string))
	var err error
	if reflect.ValueOf(receive).Kind() != reflect.Ptr {
		err = json.Unmarshal(b, &receive)
	} else {
		err = json.Unmarshal(b, receive)
	}
	if err != nil {
		logs.Log.Error(" *     Request.Temp.Get(%v): %v", key, err)
	}
}

func (self Temp) set(key string, value interface{}) Temp {
	b, err := json.Marshal(value)
	if err != nil {
		logs.Log.Error(" *     Request.Temp.Set(%v): %v", key, err)
	}
	self[key] = util.Bytes2String(b)
	return self
}
