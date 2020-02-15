package goutil

import (
	"reflect"
	"runtime"
	"unicode"
	"unicode/utf8"
)

// IsExportedOrBuiltinType is this type exported or a builtin?
func IsExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return IsExportedName(t.Name()) || t.PkgPath() == ""
}

// IsExportedName is this an exported - upper case - name?
func IsExportedName(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// ObjectName gets the type name of the object
func ObjectName(obj interface{}) string {
	v := reflect.ValueOf(obj)
	t := v.Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(v.Pointer()).Name()
	}
	return t.String()
}
