package goutil

import (
	"reflect"
	"runtime"
)

// ObjectName gets the type name of the object
func ObjectName(obj interface{}) string {
	v := reflect.ValueOf(obj)
	t := v.Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(v.Pointer()).Name()
	}
	return t.String()
}
