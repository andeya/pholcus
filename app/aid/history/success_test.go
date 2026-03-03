package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/config"
)

func TestSuccess_UpsertSuccess(t *testing.T) {
	s := &Success{
		tabName:  "t",
		fileName: "f",
		new:      make(map[string]bool),
		old:      make(map[string]bool),
	}
	tests := []struct {
		unique string
		want   bool
	}{
		{"id1", true},
		{"id1", false},
		{"id2", true},
		{"id2", false},
	}
	for _, tt := range tests {
		if got := s.UpsertSuccess(tt.unique); got != tt.want {
			t.Errorf("UpsertSuccess(%q) = %v, want %v", tt.unique, got, tt.want)
		}
	}
}

func TestSuccess_UpsertSuccess_OldExists(t *testing.T) {
	s := &Success{
		tabName:  "t",
		fileName: "f",
		new:      make(map[string]bool),
		old:      map[string]bool{"id1": true},
	}
	if got := s.UpsertSuccess("id1"); got {
		t.Error("UpsertSuccess when old exists want false")
	}
}

func TestSuccess_HasSuccess(t *testing.T) {
	s := &Success{
		tabName:  "t",
		fileName: "f",
		new:      map[string]bool{"n1": true},
		old:      map[string]bool{"o1": true},
	}
	tests := []struct {
		unique string
		want   bool
	}{
		{"n1", true},
		{"o1", true},
		{"x", false},
	}
	for _, tt := range tests {
		if got := s.HasSuccess(tt.unique); got != tt.want {
			t.Errorf("HasSuccess(%q) = %v, want %v", tt.unique, got, tt.want)
		}
	}
}

func TestSuccess_DeleteSuccess(t *testing.T) {
	s := &Success{
		tabName:  "t",
		fileName: "f",
		new:      map[string]bool{"id1": true},
		old:      make(map[string]bool),
	}
	s.DeleteSuccess("id1")
	if s.HasSuccess("id1") {
		t.Error("DeleteSuccess should remove from new")
	}
}

func TestSuccess_Flush_File(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, config.WorkRoot, config.HistoryTag)
	if err := os.MkdirAll(dir, 0777); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	fileName := filepath.Join(dir, "history__y__test")
	s := &Success{
		tabName:  util.FileNameReplace("history__y__test"),
		fileName: fileName,
		new:      map[string]bool{"a": true, "b": true},
		old:      make(map[string]bool),
	}
	r := s.flush("file")
	if r.IsErr() {
		t.Fatalf("flush: %v", r.UnwrapErr())
	}
	if r.Unwrap() != 2 {
		t.Errorf("flush count = %v, want 2", r.Unwrap())
	}
	if _, err := os.Stat(fileName); err != nil {
		t.Errorf("flush file: %v", err)
	}
}

func TestSuccess_Flush_Empty(t *testing.T) {
	s := &Success{
		tabName:  "t",
		fileName: "/nonexistent",
		new:      make(map[string]bool),
		old:      make(map[string]bool),
	}
	r := s.flush("file")
	if r.IsErr() {
		t.Fatalf("flush empty: %v", r.UnwrapErr())
	}
	if r.Unwrap() != 0 {
		t.Errorf("flush count = %v, want 0", r.Unwrap())
	}
}

func TestSuccess_Flush_FileAppend(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, config.WorkRoot, config.HistoryTag)
	if err := os.MkdirAll(dir, 0777); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	fileName := filepath.Join(dir, "history__y__test")
	s := &Success{
		tabName:  util.FileNameReplace("history__y__test"),
		fileName: fileName,
		new:      map[string]bool{"c": true},
		old:      make(map[string]bool),
	}
	r := s.flush("file")
	if r.IsErr() {
		t.Fatalf("flush: %v", r.UnwrapErr())
	}
	data, _ := os.ReadFile(fileName)
	var m map[string]bool
	if err := json.Unmarshal(append(append([]byte{'{'}, data[1:]...), '}'), &m); err != nil {
		t.Fatalf("unmarshal file: %v, content: %s", err, data)
	}
	if !m["c"] {
		t.Errorf("expected c in file, got %v", m)
	}
}
