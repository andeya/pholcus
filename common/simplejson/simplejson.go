// Package simplejson 提供了简化的 JSON 解析和操作功能。
package simplejson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"reflect"
	"strconv"

	"github.com/andeya/gust/option"
	"github.com/andeya/gust/result"
)

// Version returns the current implementation version.
func Version() string {
	return "0.5.0-alpha"
}

// Json represents a mutable JSON object for parsing and querying.
type Json struct {
	data interface{}
}

// NewJson returns a result.Result[*Json] after unmarshaling `body` bytes
func NewJson(body []byte) result.Result[*Json] {
	j := new(Json)
	err := j.UnmarshalJSON(body)
	return result.Ret(j, err)
}

// NewFromReader returns a result.Result[*Json] by decoding from an io.Reader
func NewFromReader(r io.Reader) result.Result[*Json] {
	j := new(Json)
	dec := json.NewDecoder(r)
	dec.UseNumber()
	err := dec.Decode(&j.data)
	return result.Ret(j, err)
}

// New returns a pointer to a new, empty `Json` object
func New() *Json {
	return &Json{
		data: make(map[string]interface{}),
	}
}

// Interface returns the underlying data
func (j *Json) Interface() interface{} {
	return j.data
}

// Encode returns its marshaled data as `[]byte`
func (j *Json) Encode() ([]byte, error) {
	return j.MarshalJSON()
}

// EncodePretty returns its marshaled data as `[]byte` with indentation
func (j *Json) EncodePretty() ([]byte, error) {
	return json.MarshalIndent(&j.data, "", "  ")
}

// Implements the json.Marshaler interface.
func (j *Json) MarshalJSON() ([]byte, error) {
	return json.Marshal(&j.data)
}

// Implements the json.Unmarshaler interface.
func (j *Json) UnmarshalJSON(p []byte) error {
	dec := json.NewDecoder(bytes.NewBuffer(p))
	dec.UseNumber()
	return dec.Decode(&j.data)
}

// Set modifies `Json` map by `key` and `value`
// Useful for changing single key/value in a `Json` object easily.
func (j *Json) Set(key string, val interface{}) {
	m := j.Map()
	if m.IsErr() {
		return
	}
	m.Unwrap()[key] = val
}

// SetPath modifies `Json`, recursively checking/creating map keys for the supplied path,
// and then finally writing in the value
func (j *Json) SetPath(branch []string, val interface{}) {
	if len(branch) == 0 {
		j.data = val
		return
	}

	// in order to insert our branch, we need map[string]interface{}
	if _, ok := (j.data).(map[string]interface{}); !ok {
		// have to replace with something suitable
		j.data = make(map[string]interface{})
	}
	curr := j.data.(map[string]interface{})

	for i := 0; i < len(branch)-1; i++ {
		b := branch[i]
		// key exists?
		if _, ok := curr[b]; !ok {
			n := make(map[string]interface{})
			curr[b] = n
			curr = n
			continue
		}

		// make sure the value is the right sort of thing
		if _, ok := curr[b].(map[string]interface{}); !ok {
			// have to replace with something suitable
			n := make(map[string]interface{})
			curr[b] = n
		}

		curr = curr[b].(map[string]interface{})
	}

	// add remaining k/v
	curr[branch[len(branch)-1]] = val
}

// Del modifies `Json` map by deleting `key` if it is present.
func (j *Json) Del(key string) {
	m := j.Map()
	if m.IsErr() {
		return
	}
	delete(m.Unwrap(), key)
}

// Get returns a pointer to a new `Json` object
// for `key` in its `map` representation
//
// useful for chaining operations (to traverse a nested JSON):
//
//	js.Get("top_level").Get("dict").Get("value").Int()
func (j *Json) Get(key string) *Json {
	m := j.Map()
	if m.IsOk() {
		mp := m.Unwrap()
		if val, ok := mp[key]; ok {
			return &Json{val}
		}
	}
	return &Json{nil}
}

// GetPath searches for the item as specified by the branch
// without the need to deep dive using Get()'s.
//
//	js.GetPath("top_level", "dict")
func (j *Json) GetPath(branch ...string) *Json {
	jin := j
	for _, p := range branch {
		jin = jin.Get(p)
	}
	return jin
}

// GetIndex returns a pointer to a new `Json` object
// for `index` in its `array` representation
//
// this is the analog to Get when accessing elements of
// a json array instead of a json object:
//
//	js.Get("top_level").Get("array").GetIndex(1).Get("key").Int()
func (j *Json) GetIndex(index int) *Json {
	a := j.Array()
	if a.IsOk() {
		arr := a.Unwrap()
		if len(arr) > index {
			return &Json{arr[index]}
		}
	}
	return &Json{nil}
}

// CheckGet returns an option.Option[*Json] identifying success or failure
//
// useful for chained operations when success is important:
//
//	if data := js.Get("top_level").CheckGet("inner"); data.IsSome() {
//	    log.Println(data.Unwrap())
//	}
func (j *Json) CheckGet(key string) option.Option[*Json] {
	m := j.Map()
	if m.IsOk() {
		mp := m.Unwrap()
		if val, ok := mp[key]; ok {
			return option.Some(&Json{val})
		}
	}
	return option.None[*Json]()
}

// Map type asserts to `map`
func (j *Json) Map() result.Result[map[string]interface{}] {
	if m, ok := (j.data).(map[string]interface{}); ok {
		return result.Ok(m)
	}
	return result.TryErr[map[string]interface{}](errors.New("type assertion to map[string]interface{} failed"))
}

// Array type asserts to an `array`
func (j *Json) Array() result.Result[[]interface{}] {
	if a, ok := (j.data).([]interface{}); ok {
		return result.Ok(a)
	}
	return result.TryErr[[]interface{}](errors.New("type assertion to []interface{} failed"))
}

// Float64 coerces into a float64
func (j *Json) Float64() result.Result[float64] {
	switch j.data.(type) {
	case json.Number:
		f, err := j.data.(json.Number).Float64()
		return result.Ret(f, err)
	case float32, float64:
		return result.Ok(reflect.ValueOf(j.data).Float())
	case int, int8, int16, int32, int64:
		return result.Ok(float64(reflect.ValueOf(j.data).Int()))
	case uint, uint8, uint16, uint32, uint64:
		return result.Ok(float64(reflect.ValueOf(j.data).Uint()))
	}
	return result.TryErr[float64](errors.New("invalid value type"))
}

// Int coerces into an int
func (j *Json) Int() result.Result[int] {
	switch j.data.(type) {
	case json.Number:
		i, err := j.data.(json.Number).Int64()
		return result.Ret(int(i), err)
	case float32, float64:
		return result.Ok(int(reflect.ValueOf(j.data).Float()))
	case int, int8, int16, int32, int64:
		return result.Ok(int(reflect.ValueOf(j.data).Int()))
	case uint, uint8, uint16, uint32, uint64:
		return result.Ok(int(reflect.ValueOf(j.data).Uint()))
	}
	return result.TryErr[int](errors.New("invalid value type"))
}

// Int64 coerces into an int64
func (j *Json) Int64() result.Result[int64] {
	switch j.data.(type) {
	case json.Number:
		return result.Ret(j.data.(json.Number).Int64())
	case float32, float64:
		return result.Ok(int64(reflect.ValueOf(j.data).Float()))
	case int, int8, int16, int32, int64:
		return result.Ok(reflect.ValueOf(j.data).Int())
	case uint, uint8, uint16, uint32, uint64:
		return result.Ok(int64(reflect.ValueOf(j.data).Uint()))
	}
	return result.TryErr[int64](errors.New("invalid value type"))
}

// Uint64 coerces into an uint64
func (j *Json) Uint64() result.Result[uint64] {
	switch j.data.(type) {
	case json.Number:
		u, err := strconv.ParseUint(j.data.(json.Number).String(), 10, 64)
		return result.Ret(u, err)
	case float32, float64:
		return result.Ok(uint64(reflect.ValueOf(j.data).Float()))
	case int, int8, int16, int32, int64:
		return result.Ok(uint64(reflect.ValueOf(j.data).Int()))
	case uint, uint8, uint16, uint32, uint64:
		return result.Ok(reflect.ValueOf(j.data).Uint())
	}
	return result.TryErr[uint64](errors.New("invalid value type"))
}

// Bool type asserts to `bool`
func (j *Json) Bool() result.Result[bool] {
	if s, ok := (j.data).(bool); ok {
		return result.Ok(s)
	}
	return result.TryErr[bool](errors.New("type assertion to bool failed"))
}

// String type asserts to `string`
func (j *Json) String() result.Result[string] {
	if s, ok := (j.data).(string); ok {
		return result.Ok(s)
	}
	return result.TryErr[string](errors.New("type assertion to string failed"))
}

// Bytes returns []byte from string data
func (j *Json) Bytes() result.Result[[]byte] {
	if s, ok := (j.data).(string); ok {
		return result.Ok([]byte(s))
	}
	return result.TryErr[[]byte](errors.New("type assertion to []byte failed"))
}

// StringArray type asserts to an `array` of `string`
func (j *Json) StringArray() result.Result[[]string] {
	arr := j.Array()
	if arr.IsErr() {
		return result.TryErr[[]string](arr.UnwrapErr())
	}
	retArr := make([]string, 0, len(arr.Unwrap()))
	for _, a := range arr.Unwrap() {
		if a == nil {
			retArr = append(retArr, "")
			continue
		}
		s, ok := a.(string)
		if !ok {
			return result.TryErr[[]string](errors.New("array element is not string"))
		}
		retArr = append(retArr, s)
	}
	return result.Ok(retArr)
}

// IntArray type asserts to an `array` of `int`
func (j *Json) IntArray() result.Result[[]int] {
	arr := j.Array()
	if arr.IsErr() {
		return result.TryErr[[]int](arr.UnwrapErr())
	}
	retArr := make([]int, 0, len(arr.Unwrap()))
	for _, a := range arr.Unwrap() {
		if a == nil {
			retArr = append(retArr, 0)
			continue
		}
		ji := &Json{a}
		ri := ji.Int()
		if ri.IsErr() {
			return result.TryErr[[]int](ri.UnwrapErr())
		}
		retArr = append(retArr, ri.Unwrap())
	}
	return result.Ok(retArr)
}

// Int64Array type asserts to an `array` of `int64`
func (j *Json) Int64Array() result.Result[[]int64] {
	arr := j.Array()
	if arr.IsErr() {
		return result.TryErr[[]int64](arr.UnwrapErr())
	}
	retArr := make([]int64, 0, len(arr.Unwrap()))
	for _, a := range arr.Unwrap() {
		if a == nil {
			retArr = append(retArr, 0)
			continue
		}
		ji := &Json{a}
		ri := ji.Int64()
		if ri.IsErr() {
			return result.TryErr[[]int64](ri.UnwrapErr())
		}
		retArr = append(retArr, ri.Unwrap())
	}
	return result.Ok(retArr)
}

// MustArray guarantees the return of a `[]interface{}` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//
//	for i, v := range js.Get("results").MustArray() {
//		fmt.Println(i, v)
//	}
func (j *Json) MustArray(args ...[]interface{}) []interface{} {
	var def []interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustArray() received too many arguments %d", len(args))
	}

	return j.Array().UnwrapOr(def)
}

// MustMap guarantees the return of a `map[string]interface{}` (with optional default)
//
// useful when you want to interate over map values in a succinct manner:
//
//	for k, v := range js.Get("dictionary").MustMap() {
//		fmt.Println(k, v)
//	}
func (j *Json) MustMap(args ...map[string]interface{}) map[string]interface{} {
	var def map[string]interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustMap() received too many arguments %d", len(args))
	}

	return j.Map().UnwrapOr(def)
}

// MustString guarantees the return of a `string` (with optional default)
//
// useful when you explicitly want a `string` in a single value return context:
//
//	myFunc(js.Get("param1").MustString(), js.Get("optional_param").MustString("my_default"))
func (j *Json) MustString(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustString() received too many arguments %d", len(args))
	}

	return j.String().UnwrapOr(def)
}

// MustStringArray guarantees the return of a `[]string` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//
//	for i, s := range js.Get("results").MustStringArray() {
//		fmt.Println(i, s)
//	}
func (j *Json) MustStringArray(args ...[]string) []string {
	var def []string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustStringArray() received too many arguments %d", len(args))
	}

	return j.StringArray().UnwrapOr(def)
}

// MustInt guarantees the return of an `int` (with optional default)
//
// useful when you explicitly want an `int` in a single value return context:
//
//	myFunc(js.Get("param1").MustInt(), js.Get("optional_param").MustInt(5150))
func (j *Json) MustInt(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt() received too many arguments %d", len(args))
	}

	return j.Int().UnwrapOr(def)
}

// MustFloat64 guarantees the return of a `float64` (with optional default)
//
// useful when you explicitly want a `float64` in a single value return context:
//
//	myFunc(js.Get("param1").MustFloat64(), js.Get("optional_param").MustFloat64(5.150))
func (j *Json) MustFloat64(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustFloat64() received too many arguments %d", len(args))
	}

	return j.Float64().UnwrapOr(def)
}

// MustBool guarantees the return of a `bool` (with optional default)
//
// useful when you explicitly want a `bool` in a single value return context:
//
//	myFunc(js.Get("param1").MustBool(), js.Get("optional_param").MustBool(true))
func (j *Json) MustBool(args ...bool) bool {
	var def bool

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustBool() received too many arguments %d", len(args))
	}

	return j.Bool().UnwrapOr(def)
}

// MustInt64 guarantees the return of an `int64` (with optional default)
//
// useful when you explicitly want an `int64` in a single value return context:
//
//	myFunc(js.Get("param1").MustInt64(), js.Get("optional_param").MustInt64(5150))
func (j *Json) MustInt64(args ...int64) int64 {
	var def int64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt64() received too many arguments %d", len(args))
	}

	return j.Int64().UnwrapOr(def)
}

// MustUInt64 guarantees the return of an `uint64` (with optional default)
//
// useful when you explicitly want an `uint64` in a single value return context:
//
//	myFunc(js.Get("param1").MustUint64(), js.Get("optional_param").MustUint64(5150))
func (j *Json) MustUint64(args ...uint64) uint64 {
	var def uint64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustUint64() received too many arguments %d", len(args))
	}

	return j.Uint64().UnwrapOr(def)
}
