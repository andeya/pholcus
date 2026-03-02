package util

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestJSONPToJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"object callback", `forbar({"a":"1"})`, `{"a":"1"}`},
		{"object with number", `cb({a:"1",b:2})`, `{"a":"1","b":2}`},
		{"array", `fn([1,2,3])`, `[1,2,3]`},
		{"nested object", `x({a:1,b:{c:2}})`, `{"a":1,"b":{"c":2}}`},
		{"empty object", `wrap({})`, `{}`},
		{"empty array", `wrap([])`, `[]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JSONPToJSON(tt.in)
			if got != tt.want {
				t.Errorf("JSONPToJSON(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMkdir(t *testing.T) {
	tmp := t.TempDir()

	t.Run("success", func(t *testing.T) {
		dir := filepath.Join(tmp, "a", "b", "c")
		filePath := filepath.Join(dir, "file.txt")
		r := Mkdir(filePath)
		if r.IsErr() {
			t.Errorf("Mkdir(%q) failed: %v", filePath, r.UnwrapErr())
		}
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("parent directory not created: %v", err)
		}
	})

	t.Run("already exists", func(t *testing.T) {
		dir := filepath.Join(tmp, "existing")
		if err := os.MkdirAll(dir, 0777); err != nil {
			t.Fatal(err)
		}
		r := Mkdir(dir)
		if r.IsErr() {
			t.Errorf("Mkdir on existing dir failed: %v", r.UnwrapErr())
		}
	})

	t.Run("empty path component", func(t *testing.T) {
		r := Mkdir("justfilename.txt")
		if r.IsErr() {
			t.Errorf("Mkdir with no dir component should be Ok: %v", r.UnwrapErr())
		}
	})
}

func TestIsDirExists(t *testing.T) {
	tmp := t.TempDir()
	subdir := filepath.Join(tmp, "subdir")
	if err := os.MkdirAll(subdir, 0777); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if !IsDirExists(tmp) {
		t.Error("IsDirExists(tmp) = false, want true")
	}
	if !IsDirExists(subdir) {
		t.Error("IsDirExists(subdir) = false, want true")
	}
	if IsDirExists(file) {
		t.Error("IsDirExists(file) = true, want false")
	}
	if IsDirExists(filepath.Join(tmp, "nonexistent")) {
		t.Error("IsDirExists(nonexistent) = true, want false")
	}
}

func TestIsFileExists(t *testing.T) {
	tmp := t.TempDir()
	subdir := filepath.Join(tmp, "subdir")
	if err := os.MkdirAll(subdir, 0777); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if !IsFileExists(file) {
		t.Error("IsFileExists(file) = false, want true")
	}
	if IsFileExists(tmp) {
		t.Error("IsFileExists(tmp) = true, want false")
	}
	if IsFileExists(subdir) {
		t.Error("IsFileExists(subdir) = true, want false")
	}
	if IsFileExists(filepath.Join(tmp, "nonexistent")) {
		t.Error("IsFileExists(nonexistent) = true, want false")
	}
}

func TestWalkFiles(t *testing.T) {
	tmp := t.TempDir()
	dir1 := filepath.Join(tmp, "d1")
	dir2 := filepath.Join(tmp, "d2")
	if err := os.MkdirAll(dir1, 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir2, 0777); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		filepath.Join(tmp, "a.txt"),
		filepath.Join(tmp, "b.go"),
		filepath.Join(dir1, "c.txt"),
		filepath.Join(dir2, "d.json"),
	} {
		if err := os.WriteFile(p, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	r := WalkFiles(tmp)
	if r.IsErr() {
		t.Fatalf("WalkFiles failed: %v", r.UnwrapErr())
	}
	files := r.Unwrap()
	if len(files) != 4 {
		t.Errorf("WalkFiles: got %d files, want 4: %v", len(files), files)
	}

	r2 := WalkFiles(tmp, ".txt")
	if r2.IsErr() {
		t.Fatalf("WalkFiles with suffix failed: %v", r2.UnwrapErr())
	}
	txtFiles := r2.Unwrap()
	if len(txtFiles) != 2 {
		t.Errorf("WalkFiles(.txt): got %d files, want 2: %v", len(txtFiles), txtFiles)
	}
}

func TestWalkDir(t *testing.T) {
	tmp := t.TempDir()
	d1 := filepath.Join(tmp, "d1")
	d2 := filepath.Join(tmp, "d2", "d2a")
	if err := os.MkdirAll(d1, 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(d2, 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "f.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	r := WalkDir(tmp)
	if r.IsErr() {
		t.Fatalf("WalkDir failed: %v", r.UnwrapErr())
	}
	dirs := r.Unwrap()
	if len(dirs) != 4 { // tmp, d1, d2, d2a
		t.Errorf("WalkDir: got %d dirs, want 4: %v", len(dirs), dirs)
	}
}

func TestIsNum(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"123", true},
		{"0", true},
		{"abc", false},
		{"", false},
		{"12a", false},
		{" 123", false},
		{"123 ", false},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := IsNum(tt.s)
			if got != tt.want {
				t.Errorf("IsNum(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestMakeHash(t *testing.T) {
	h := MakeHash("hello")
	if h == "" {
		t.Error("MakeHash returned empty string")
	}
	if MakeHash("hello") != MakeHash("hello") {
		t.Error("MakeHash not deterministic")
	}
}

func TestHashString(t *testing.T) {
	v := HashString("test")
	if v == 0 {
		t.Error("HashString returned 0")
	}
	if HashString("test") != HashString("test") {
		t.Error("HashString not deterministic")
	}
}

func TestMakeUnique(t *testing.T) {
	u := MakeUnique(map[string]int{"a": 1})
	if u == "" {
		t.Error("MakeUnique returned empty string")
	}
	if MakeUnique(map[string]int{"a": 1}) != MakeUnique(map[string]int{"a": 1}) {
		t.Error("MakeUnique not deterministic for same input")
	}
}

func TestMakeMd5(t *testing.T) {
	m := MakeMd5("hello", 16)
	if len(m) != 16 {
		t.Errorf("MakeMd5(_, 16) len = %d, want 16", len(m))
	}
	if MakeMd5("hello", 16) != MakeMd5("hello", 16) {
		t.Error("MakeMd5 not deterministic")
	}
	m32 := MakeMd5("x", 64)
	if len(m32) != 32 {
		t.Errorf("MakeMd5(_, 64) capped at 32: len = %d", len(m32))
	}
}

func TestJSONString(t *testing.T) {
	got := JSONString(map[string]string{"a": "1", "b": "2"})
	if !strings.Contains(got, `"a"`) || !strings.Contains(got, `"1"`) {
		t.Errorf("JSONString output invalid: %q", got)
	}
	got2 := JSONString([]int{1, 2, 3})
	if !strings.Contains(got2, "1") {
		t.Errorf("JSONString array invalid: %q", got2)
	}
}

func TestFileNameReplace(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{`:*?/<>|\"\\`, ""}, // all special chars replaced
		{"normal", "normal"},
		{`file:name`, "file" + "\uff1a" + "name"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := FileNameReplace(tt.in)
			if tt.want != "" && got != tt.want {
				t.Errorf("FileNameReplace(%q) = %q, want %q", tt.in, got, tt.want)
			}
			for _, c := range `:*?/<>|\` {
				if strings.ContainsRune(got, c) {
					t.Errorf("FileNameReplace(%q) still contains invalid char %q: %q", tt.in, c, got)
				}
			}
		})
	}
}

func TestExcelSheetNameReplace(t *testing.T) {
	got := ExcelSheetNameReplace("sheet:name*?/\\[]")
	if strings.ContainsAny(got, ":*?/\\[]") {
		t.Errorf("ExcelSheetNameReplace still has invalid chars: %q", got)
	}
	if got != "sheet_name______" {
		t.Errorf("ExcelSheetNameReplace = %q, want sheet_name______", got)
	}
}

func TestAtoa(t *testing.T) {
	opt := Atoa(nil)
	if opt.IsSome() {
		t.Error("Atoa(nil) should be None")
	}
	opt2 := Atoa(" hello ")
	if !opt2.IsSome() {
		t.Error("Atoa(\" hello \") should be Some")
	}
	if opt2.Unwrap() != "hello" {
		t.Errorf("Atoa Unwrap = %q, want hello", opt2.Unwrap())
	}
}

func TestAtoi(t *testing.T) {
	opt := Atoi(nil)
	if opt.IsSome() {
		t.Error("Atoi(nil) should be None")
	}
	opt2 := Atoi(" 42 ")
	if !opt2.IsSome() {
		t.Error("Atoi(\" 42 \") should be Some")
	}
	if opt2.Unwrap() != 42 {
		t.Errorf("Atoi Unwrap = %d, want 42", opt2.Unwrap())
	}
}

func TestAtoui(t *testing.T) {
	opt := Atoui(nil)
	if opt.IsSome() {
		t.Error("Atoui(nil) should be None")
	}
	opt2 := Atoui(" 99 ")
	if !opt2.IsSome() {
		t.Error("Atoui(\" 99 \") should be Some")
	}
	if opt2.Unwrap() != 99 {
		t.Errorf("Atoui Unwrap = %d, want 99", opt2.Unwrap())
	}
}

func TestBytes2String_String2Bytes(t *testing.T) {
	orig := "hello world"
	b := String2Bytes(orig)
	s := Bytes2String(b)
	if s != orig {
		t.Errorf("roundtrip: got %q, want %q", s, orig)
	}
	orig2 := []byte("foo bar")
	s2 := Bytes2String(orig2)
	b2 := String2Bytes(s2)
	if string(b2) != string(orig2) {
		t.Errorf("roundtrip bytes: got %q, want %q", b2, orig2)
	}
}

func TestKeyinsParse(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"<a><b><c>", []string{"a", "b", "c"}},
		{"", []string{}},
		{"  ", []string{}},
		{"<x>", []string{"x"}},
		{"<a> <b> <c>", []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := KeyinsParse(tt.in)
			sort.Strings(got)
			sort.Strings(tt.want)
			if len(got) != len(tt.want) {
				t.Errorf("KeyinsParse(%q) = %v, want %v", tt.in, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("KeyinsParse(%q) = %v, want %v", tt.in, got, tt.want)
					return
				}
			}
		})
	}
}

func TestRandomCreateBytes(t *testing.T) {
	b := RandomCreateBytes(10)
	if len(b) != 10 {
		t.Errorf("RandomCreateBytes(10) len = %d", len(b))
	}
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for _, c := range b {
		if !strings.ContainsRune(alphanum, rune(c)) {
			t.Errorf("RandomCreateBytes char %q not in alphanum", c)
		}
	}
	custom := RandomCreateBytes(5, 'a', 'b', 'c')
	if len(custom) != 5 {
		t.Errorf("RandomCreateBytes(5, custom) len = %d", len(custom))
	}
	for _, c := range custom {
		if c != 'a' && c != 'b' && c != 'c' {
			t.Errorf("RandomCreateBytes custom char %q not in abc", c)
		}
	}
}

func TestXML2MapStr(t *testing.T) {
	xml := `<root><name>test</name><value>42</value></root>`
	m := XML2MapStr(xml)
	if m["name"] != "test" {
		t.Errorf("XML2MapStr name = %q, want test", m["name"])
	}
	if m["value"] != "42" {
		t.Errorf("XML2MapStr value = %q, want 42", m["value"])
	}
}

func TestRelPath(t *testing.T) {
	absPath, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	r := RelPath(absPath)
	if r.IsErr() {
		t.Errorf("RelPath(%q) failed: %v", absPath, r.UnwrapErr())
	}
	rel := r.Unwrap()
	if rel == "" {
		t.Error("RelPath returned empty string")
	}
	if strings.Contains(rel, "\\") {
		t.Error("RelPath should use forward slashes")
	}
}

func TestWalkRelFiles(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	if err := os.MkdirAll(sub, 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	r := WalkRelFiles(".")
	if r.IsErr() {
		t.Fatalf("WalkRelFiles failed: %v", r.UnwrapErr())
	}
	files := r.Unwrap()
	if len(files) != 2 {
		t.Errorf("WalkRelFiles: got %d files, want 2: %v", len(files), files)
	}
}

func TestWalkRelDir(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	if err := os.MkdirAll(sub, 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "f.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	r := WalkRelDir(".")
	if r.IsErr() {
		t.Fatalf("WalkRelDir failed: %v", r.UnwrapErr())
	}
	dirs := r.Unwrap()
	if len(dirs) < 2 {
		t.Errorf("WalkRelDir: got %d dirs, want at least 2: %v", len(dirs), dirs)
	}
}
