package context

import (
	"encoding/json"
	"reflect"
)

type Temp map[string]interface{}

// 返回临时缓存数据
// 强烈建议数据接收者receive为指针类型
func (self Temp) Get(key string, receive interface{}) interface{} {
	b, _ := json.Marshal(self[key])
	if reflect.ValueOf(receive).Kind() != reflect.Ptr {
		json.Unmarshal(b, &receive)
	} else {
		json.Unmarshal(b, receive)
	}
	return receive
}

func (self Temp) Set(key string, value interface{}) Temp {
	b, _ := json.Marshal(value)
	self[key] = string(b)
	return self
}

func (self *Temp) MarshalJSON() ([]byte, error) {
	for k, v := range *self {
		b, _ := json.Marshal(v)
		(*self)[k] = string(b)
	}
	return json.Marshal(*self)
}

func (self *Temp) UnmarshalJSON(jsonByte []byte) (err error) {
	var t = make(map[string]string)
	err = json.Unmarshal(jsonByte, &t)
	(*self) = make(map[string]interface{})
	for k, v := range t {
		(*self)[k] = v
	}
	return
}
