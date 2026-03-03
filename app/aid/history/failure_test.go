package history

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/config"
)

func newTestRequest(url string) *request.Request {
	r := &request.Request{Spider: "s", URL: url, Rule: "r", Method: "GET", Header: make(http.Header)}
	r.Prepare()
	return r
}

func TestFailure_PullFailure(t *testing.T) {
	req := newTestRequest("http://a.com")
	f := &Failure{
		tabName:  "t",
		fileName: "f",
		list:     map[string]*request.Request{req.Unique(): req},
	}
	got := f.PullFailure()
	if len(got) != 1 {
		t.Errorf("PullFailure len = %v, want 1", len(got))
	}
	if len(f.list) != 0 {
		t.Error("PullFailure should clear list")
	}
}

func TestFailure_UpsertFailure(t *testing.T) {
	req := newTestRequest("http://a.com")
	f := &Failure{
		tabName:  "t",
		fileName: "f",
		list:     make(map[string]*request.Request),
	}
	tests := []struct {
		req  *request.Request
		want bool
	}{
		{req, true},
		{req, false},
	}
	for i, tt := range tests {
		if got := f.UpsertFailure(tt.req); got != tt.want {
			t.Errorf("UpsertFailure #%d = %v, want %v", i, got, tt.want)
		}
	}
}

func TestFailure_DeleteFailure(t *testing.T) {
	req := newTestRequest("http://a.com")
	f := &Failure{
		tabName:  "t",
		fileName: "f",
		list:     map[string]*request.Request{req.Unique(): req},
	}
	f.DeleteFailure(req)
	if len(f.list) != 0 {
		t.Error("DeleteFailure should remove from list")
	}
}

func TestFailure_Flush_File(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, config.WorkRoot, config.HistoryTag)
	if err := os.MkdirAll(dir, 0777); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	fileName := filepath.Join(dir, "history__n__test")
	req := newTestRequest("http://b.com")
	f := &Failure{
		tabName:  util.FileNameReplace("history__n__test"),
		fileName: fileName,
		list:     map[string]*request.Request{req.Unique(): req},
	}
	r := f.flush("file")
	if r.IsErr() {
		t.Fatalf("flush: %v", r.UnwrapErr())
	}
	if r.Unwrap() != 1 {
		t.Errorf("flush count = %v, want 1", r.Unwrap())
	}
	if _, err := os.Stat(fileName); err != nil {
		t.Errorf("flush file: %v", err)
	}
}

func TestFailure_Flush_FileEmpty(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, config.WorkRoot, config.HistoryTag)
	if err := os.MkdirAll(dir, 0777); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	fileName := filepath.Join(dir, "history__n__empty")
	f := &Failure{
		tabName:  util.FileNameReplace("history__n__empty"),
		fileName: fileName,
		list:     make(map[string]*request.Request),
	}
	r := f.flush("file")
	if r.IsErr() {
		t.Fatalf("flush empty: %v", r.UnwrapErr())
	}
	if r.Unwrap() != 0 {
		t.Errorf("flush count = %v, want 0", r.Unwrap())
	}
}

func TestFailure_Flush_FileOverwrite(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, config.WorkRoot, config.HistoryTag)
	if err := os.MkdirAll(dir, 0777); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	fileName := filepath.Join(dir, "history__n__overwrite")
	if err := os.WriteFile(fileName, []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	req := newTestRequest("http://c.com")
	f := &Failure{
		tabName:  util.FileNameReplace("history__n__overwrite"),
		fileName: fileName,
		list:     map[string]*request.Request{req.Unique(): req},
	}
	r := f.flush("file")
	if r.IsErr() {
		t.Fatalf("flush: %v", r.UnwrapErr())
	}
	data, _ := os.ReadFile(fileName)
	if len(data) < 10 {
		t.Errorf("flush should overwrite file, got %d bytes", len(data))
	}
}
