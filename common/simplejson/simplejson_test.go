package simplejson

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

const sampleJSON = `{"name":"John","age":30,"scores":[90,85,92],"address":{"city":"Beijing"},"active":true}`

func sampleJson(t *testing.T) *Json {
	t.Helper()
	r := NewJson([]byte(sampleJSON))
	if r.IsErr() {
		t.Fatalf("NewJson failed: %v", r.UnwrapErr())
	}
	return r.Unwrap()
}

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("Version() returned empty string")
	}
}

func TestNewJson(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		wantOk  bool
	}{
		{"valid", []byte(sampleJSON), true},
		{"empty object", []byte(`{}`), true},
		{"empty array", []byte(`[]`), true},
		{"invalid", []byte(`{invalid`), false},
		{"empty", []byte(``), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewJson(tt.body)
			if tt.wantOk && r.IsErr() {
				t.Errorf("NewJson() unexpected error: %v", r.UnwrapErr())
			}
			if !tt.wantOk && r.IsOk() {
				t.Error("NewJson() expected error, got Ok")
			}
		})
	}
}

func TestNewFromReader(t *testing.T) {
	tests := []struct {
		name   string
		reader io.Reader
		wantOk bool
	}{
		{"valid", strings.NewReader(sampleJSON), true},
		{"invalid", strings.NewReader(`{invalid`), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewFromReader(tt.reader)
			if tt.wantOk && r.IsErr() {
				t.Errorf("NewFromReader() unexpected error: %v", r.UnwrapErr())
			}
			if !tt.wantOk && r.IsOk() {
				t.Error("NewFromReader() expected error, got Ok")
			}
		})
	}
}

func TestNew(t *testing.T) {
	j := New()
	m := j.Map()
	if m.IsErr() {
		t.Fatalf("New() Map failed: %v", m.UnwrapErr())
	}
	if len(m.Unwrap()) != 0 {
		t.Errorf("New() expected empty map, got %d keys", len(m.Unwrap()))
	}
}

func TestInterface(t *testing.T) {
	j := sampleJson(t)
	if j.Interface() == nil {
		t.Error("Interface() returned nil")
	}
}

func TestGet(t *testing.T) {
	j := sampleJson(t)
	tests := []struct {
		key    string
		exists bool
	}{
		{"name", true},
		{"age", true},
		{"scores", true},
		{"address", true},
		{"active", true},
		{"missing", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := j.Get(tt.key)
			if tt.exists && got.Interface() == nil {
				t.Errorf("Get(%q) returned nil", tt.key)
			}
			if !tt.exists && got.Interface() != nil {
				t.Errorf("Get(%q) expected nil, got %v", tt.key, got.Interface())
			}
		})
	}
}

func TestGetPath(t *testing.T) {
	j := sampleJson(t)
	tests := []struct {
		name   string
		branch []string
		exists bool
	}{
		{"single", []string{"name"}, true},
		{"nested", []string{"address", "city"}, true},
		{"missing", []string{"x"}, false},
		{"nested missing", []string{"address", "country"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := j.GetPath(tt.branch...)
			if tt.exists && got.Interface() == nil {
				t.Errorf("GetPath(%v) returned nil", tt.branch)
			}
			if !tt.exists && got.Interface() != nil {
				t.Errorf("GetPath(%v) expected nil, got %v", tt.branch, got.Interface())
			}
		})
	}
}

func TestGetIndex(t *testing.T) {
	j := sampleJson(t)
	scores := j.Get("scores")
	tests := []struct {
		name   string
		json   *Json
		index  int
		exists bool
	}{
		{"valid 0", scores, 0, true},
		{"valid 1", scores, 1, true},
		{"valid 2", scores, 2, true},
		{"out of range", scores, 10, false},
		{"not array", j, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.json.GetIndex(tt.index)
			if tt.exists && got.Interface() == nil {
				t.Errorf("GetIndex(%d) returned nil", tt.index)
			}
			if !tt.exists && got.Interface() != nil {
				t.Errorf("GetIndex(%d) expected nil, got %v", tt.index, got.Interface())
			}
		})
	}
}

func TestCheckGet(t *testing.T) {
	j := sampleJson(t)
	tests := []struct {
		key    string
		wantOk bool
	}{
		{"name", true},
		{"missing", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			opt := j.CheckGet(tt.key)
			if tt.wantOk && opt.IsNone() {
				t.Errorf("CheckGet(%q) expected Some", tt.key)
			}
			if !tt.wantOk && opt.IsSome() {
				t.Errorf("CheckGet(%q) expected None", tt.key)
			}
			if opt.IsSome() && opt.Unwrap().Interface() == nil {
				t.Errorf("CheckGet(%q) Unwrap returned nil", tt.key)
			}
		})
	}
}

func TestSet(t *testing.T) {
	j := sampleJson(t)
	j.Set("newkey", "newval")
	got := j.Get("newkey").String()
	if got.IsErr() || got.Unwrap() != "newval" {
		t.Errorf("Set/Get roundtrip failed: got %v", got)
	}
}

func TestSetPath(t *testing.T) {
	j := sampleJson(t)
	j.SetPath([]string{"address", "country"}, "China")
	got := j.GetPath("address", "country").String()
	if got.IsErr() || got.Unwrap() != "China" {
		t.Errorf("SetPath failed: got %v", got)
	}
	j.SetPath([]string{"a", "b", "c"}, 42)
	got2 := j.GetPath("a", "b", "c").Int()
	if got2.IsErr() || got2.Unwrap() != 42 {
		t.Errorf("SetPath nested failed: got %v", got2)
	}
}

func TestDel(t *testing.T) {
	j := sampleJson(t)
	j.Del("age")
	if j.Get("age").Interface() != nil {
		t.Error("Del did not remove key")
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		wantOk bool
	}{
		{"object", sampleJSON, true},
		{"array", `[1,2,3]`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewJson([]byte(tt.json))
			if r.IsErr() {
				t.Fatal(r.UnwrapErr())
			}
			m := r.Unwrap().Map()
			if tt.wantOk && m.IsErr() {
				t.Errorf("Map() unexpected error: %v", m.UnwrapErr())
			}
			if !tt.wantOk && m.IsOk() {
				t.Error("Map() expected error")
			}
		})
	}
}

func TestArray(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		path   []string
		wantOk bool
	}{
		{"scores", sampleJSON, []string{"scores"}, true},
		{"not array", sampleJSON, []string{"name"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := sampleJson(t)
			sub := j.GetPath(tt.path...)
			a := sub.Array()
			if tt.wantOk && a.IsErr() {
				t.Errorf("Array() unexpected error: %v", a.UnwrapErr())
			}
			if !tt.wantOk && a.IsOk() {
				t.Error("Array() expected error")
			}
		})
	}
}

func TestFloat64(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		path   []string
		want   float64
		wantOk bool
	}{
		{"age as float", sampleJSON, []string{"age"}, 30, true},
		{"not number", sampleJSON, []string{"name"}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := sampleJson(t)
			sub := j.GetPath(tt.path...)
			f := sub.Float64()
			if tt.wantOk {
				if f.IsErr() {
					t.Errorf("Float64() error: %v", f.UnwrapErr())
				} else if f.Unwrap() != tt.want {
					t.Errorf("Float64() = %v, want %v", f.Unwrap(), tt.want)
				}
			} else if f.IsOk() {
				t.Error("Float64() expected error")
			}
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		path   []string
		want   int
		wantOk bool
	}{
		{"age", sampleJSON, []string{"age"}, 30, true},
		{"not number", sampleJSON, []string{"name"}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := sampleJson(t)
			sub := j.GetPath(tt.path...)
			i := sub.Int()
			if tt.wantOk {
				if i.IsErr() {
					t.Errorf("Int() error: %v", i.UnwrapErr())
				} else if i.Unwrap() != tt.want {
					t.Errorf("Int() = %v, want %v", i.Unwrap(), tt.want)
				}
			} else if i.IsOk() {
				t.Error("Int() expected error")
			}
		})
	}
}

func TestInt64(t *testing.T) {
	j := sampleJson(t)
	r := j.Get("age").Int64()
	if r.IsErr() {
		t.Fatalf("Int64() error: %v", r.UnwrapErr())
	}
	if r.Unwrap() != 30 {
		t.Errorf("Int64() = %v, want 30", r.Unwrap())
	}
}

func TestUint64(t *testing.T) {
	j := sampleJson(t)
	r := j.Get("age").Uint64()
	if r.IsErr() {
		t.Fatalf("Uint64() error: %v", r.UnwrapErr())
	}
	if r.Unwrap() != 30 {
		t.Errorf("Uint64() = %v, want 30", r.Unwrap())
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		name   string
		path   []string
		want   bool
		wantOk bool
	}{
		{"active", []string{"active"}, true, true},
		{"not bool", []string{"name"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := sampleJson(t)
			sub := j.GetPath(tt.path...)
			b := sub.Bool()
			if tt.wantOk {
				if b.IsErr() {
					t.Errorf("Bool() error: %v", b.UnwrapErr())
				} else if b.Unwrap() != tt.want {
					t.Errorf("Bool() = %v, want %v", b.Unwrap(), tt.want)
				}
			} else if b.IsOk() {
				t.Error("Bool() expected error")
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name   string
		path   []string
		want   string
		wantOk bool
	}{
		{"name", []string{"name"}, "John", true},
		{"city", []string{"address", "city"}, "Beijing", true},
		{"not string", []string{"age"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := sampleJson(t)
			sub := j.GetPath(tt.path...)
			s := sub.String()
			if tt.wantOk {
				if s.IsErr() {
					t.Errorf("String() error: %v", s.UnwrapErr())
				} else if s.Unwrap() != tt.want {
					t.Errorf("String() = %q, want %q", s.Unwrap(), tt.want)
				}
			} else if s.IsOk() {
				t.Error("String() expected error")
			}
		})
	}
}

func TestBytes(t *testing.T) {
	j := sampleJson(t)
	r := j.Get("name").Bytes()
	if r.IsErr() {
		t.Fatalf("Bytes() error: %v", r.UnwrapErr())
	}
	if string(r.Unwrap()) != "John" {
		t.Errorf("Bytes() = %q, want John", r.Unwrap())
	}
}

func TestStringArray(t *testing.T) {
	r := NewJson([]byte(`["a","b","c"]`))
	if r.IsErr() {
		t.Fatal(r.UnwrapErr())
	}
	arr := r.Unwrap().StringArray()
	if arr.IsErr() {
		t.Fatalf("StringArray() error: %v", arr.UnwrapErr())
	}
	want := []string{"a", "b", "c"}
	got := arr.Unwrap()
	if len(got) != len(want) {
		t.Fatalf("StringArray() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("StringArray()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	j := sampleJson(t)
	notStrArr := j.Get("scores").StringArray()
	if notStrArr.IsOk() {
		t.Error("StringArray() on int array expected error")
	}
}

func TestIntArray(t *testing.T) {
	j := sampleJson(t)
	arr := j.Get("scores").IntArray()
	if arr.IsErr() {
		t.Fatalf("IntArray() error: %v", arr.UnwrapErr())
	}
	want := []int{90, 85, 92}
	got := arr.Unwrap()
	if len(got) != len(want) {
		t.Fatalf("IntArray() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("IntArray()[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestInt64Array(t *testing.T) {
	j := sampleJson(t)
	arr := j.Get("scores").Int64Array()
	if arr.IsErr() {
		t.Fatalf("Int64Array() error: %v", arr.UnwrapErr())
	}
	want := []int64{90, 85, 92}
	got := arr.Unwrap()
	if len(got) != len(want) {
		t.Fatalf("Int64Array() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Int64Array()[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestMustArray(t *testing.T) {
	j := sampleJson(t)
	arr := j.Get("scores").MustArray()
	if len(arr) != 3 {
		t.Errorf("MustArray() len = %d, want 3", len(arr))
	}
	def := j.Get("missing").MustArray([]interface{}{"default"})
	if len(def) != 1 || def[0] != "default" {
		t.Errorf("MustArray(default) = %v, want [default]", def)
	}
}

func TestMustMap(t *testing.T) {
	j := sampleJson(t)
	m := j.MustMap()
	if len(m) == 0 {
		t.Error("MustMap() returned empty map")
	}
	def := map[string]interface{}{"x": 1}
	got := j.Get("name").MustMap(def)
	if got["x"] != 1 {
		t.Errorf("MustMap(default) = %v", got)
	}
}

func TestMustString(t *testing.T) {
	j := sampleJson(t)
	s := j.Get("name").MustString()
	if s != "John" {
		t.Errorf("MustString() = %q, want John", s)
	}
	def := j.Get("missing").MustString("default")
	if def != "default" {
		t.Errorf("MustString(default) = %q, want default", def)
	}
}

func TestMustStringArray(t *testing.T) {
	r := NewJson([]byte(`["x","y"]`))
	if r.IsErr() {
		t.Fatal(r.UnwrapErr())
	}
	arr := r.Unwrap().MustStringArray()
	if len(arr) != 2 || arr[0] != "x" || arr[1] != "y" {
		t.Errorf("MustStringArray() = %v, want [x y]", arr)
	}
	def := []string{"a"}
	got := r.Unwrap().Get("missing").MustStringArray(def)
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("MustStringArray(default) = %v", got)
	}
}

func TestMustInt(t *testing.T) {
	j := sampleJson(t)
	i := j.Get("age").MustInt()
	if i != 30 {
		t.Errorf("MustInt() = %d, want 30", i)
	}
	def := j.Get("missing").MustInt(99)
	if def != 99 {
		t.Errorf("MustInt(default) = %d, want 99", def)
	}
}

func TestMustFloat64(t *testing.T) {
	j := sampleJson(t)
	f := j.Get("age").MustFloat64()
	if f != 30 {
		t.Errorf("MustFloat64() = %v, want 30", f)
	}
	def := j.Get("missing").MustFloat64(3.14)
	if def != 3.14 {
		t.Errorf("MustFloat64(default) = %v, want 3.14", def)
	}
}

func TestMustBool(t *testing.T) {
	j := sampleJson(t)
	b := j.Get("active").MustBool()
	if !b {
		t.Errorf("MustBool() = %v, want true", b)
	}
	def := j.Get("missing").MustBool(true)
	if !def {
		t.Errorf("MustBool(default) = %v, want true", def)
	}
}

func TestMustInt64(t *testing.T) {
	j := sampleJson(t)
	i := j.Get("age").MustInt64()
	if i != 30 {
		t.Errorf("MustInt64() = %d, want 30", i)
	}
	def := j.Get("missing").MustInt64(123)
	if def != 123 {
		t.Errorf("MustInt64(default) = %d, want 123", def)
	}
}

func TestMustUint64(t *testing.T) {
	j := sampleJson(t)
	u := j.Get("age").MustUint64()
	if u != 30 {
		t.Errorf("MustUint64() = %d, want 30", u)
	}
	def := j.Get("missing").MustUint64(456)
	if def != 456 {
		t.Errorf("MustUint64(default) = %d, want 456", def)
	}
}

func TestEncode(t *testing.T) {
	j := sampleJson(t)
	b, err := j.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Encode() output invalid JSON: %v", err)
	}
	if decoded["name"] != "John" {
		t.Errorf("Encode() decoded name = %v", decoded["name"])
	}
}

func TestEncodePretty(t *testing.T) {
	j := sampleJson(t)
	b, err := j.EncodePretty()
	if err != nil {
		t.Fatalf("EncodePretty() error: %v", err)
	}
	if !bytes.Contains(b, []byte("\n")) {
		t.Error("EncodePretty() should produce indented output")
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("EncodePretty() output invalid JSON: %v", err)
	}
}

func TestMarshalJSON(t *testing.T) {
	j := sampleJson(t)
	b, err := j.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}
	if len(b) == 0 {
		t.Error("MarshalJSON() returned empty")
	}
}

func TestUnmarshalJSON(t *testing.T) {
	j := New()
	err := j.UnmarshalJSON([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}
	if j.Get("name").String().Unwrap() != "John" {
		t.Error("UnmarshalJSON() did not parse correctly")
	}
}

func TestSetPathEmptyBranch(t *testing.T) {
	j := sampleJson(t)
	j.SetPath([]string{}, "replaced")
	if j.Interface() != "replaced" {
		t.Errorf("SetPath([]) should replace root: got %v", j.Interface())
	}
}

func TestSetPathOverwriteNonMap(t *testing.T) {
	j := sampleJson(t)
	j.SetPath([]string{"name", "nested"}, "x")
	got := j.GetPath("name").Interface()
	if m, ok := got.(map[string]interface{}); !ok || m["nested"] != "x" {
		t.Errorf("SetPath should overwrite non-map: got %v", got)
	}
}
