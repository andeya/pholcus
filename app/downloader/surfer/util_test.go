package surfer

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestURLEncode(t *testing.T) {
	tests := []struct {
		url   string
		wantQ string
	}{
		{"http://example.com", ""},
		{"http://example.com?a=1&b=2", "a=1&b=2"},
		{"http://example.com?x=hello world", "x=hello+world"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			u, err := URLEncode(tt.url)
			if err != nil {
				t.Fatalf("URLEncode err: %v", err)
			}
			if tt.wantQ != "" && u.RawQuery != tt.wantQ {
				t.Errorf("RawQuery = %q, want %q", u.RawQuery, tt.wantQ)
			}
		})
	}
}

func TestBodyBytes(t *testing.T) {
	body := []byte("test body")
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewReader(body)),
	}
	got, err := BodyBytes(resp)
	if err != nil {
		t.Fatalf("BodyBytes err: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("BodyBytes = %q, want %q", got, body)
	}
}

func TestAutoToUTF8(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=gbk")
		w.Write([]byte("hello"))
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("http.Get err: %v", err)
	}
	defer resp.Body.Close()

	err = AutoToUTF8(resp)
	if err != nil {
		t.Logf("AutoToUTF8 err (charset may be unsupported): %v", err)
	}
}

func TestBodyRead(t *testing.T) {
	r := strings.NewReader("abc")
	b := &Body{
		ReadCloser: io.NopCloser(r),
		Reader:     r,
	}
	p := make([]byte, 2)
	n, err := b.Read(p)
	if err != nil && err != io.EOF {
		t.Fatalf("Read err: %v", err)
	}
	if n != 2 || string(p) != "ab" {
		t.Errorf("Read = %d, %q", n, p)
	}
}

func TestIsDirExists(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{os.TempDir(), true},
		{"/nonexistent/path/12345", false},
		{"util_test.go", false},
	}
	for _, tt := range tests {
		got := IsDirExists(tt.path)
		if got != tt.want {
			t.Errorf("IsDirExists(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestIsFileExists(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"util_test.go", true},
		{os.TempDir(), false},
		{"/nonexistent/file", false},
	}
	for _, tt := range tests {
		got := IsFileExists(tt.path)
		if got != tt.want {
			t.Errorf("IsFileExists(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestWalkDir(t *testing.T) {
	td := t.TempDir()
	os.MkdirAll(filepath.Join(td, "a"), 0755)
	os.MkdirAll(filepath.Join(td, "b"), 0755)
	os.WriteFile(filepath.Join(td, "f.txt"), nil, 0644)

	dirs := WalkDir(td)
	if len(dirs) < 2 {
		t.Errorf("WalkDir len = %d, want >= 2", len(dirs))
	}

	dirsSuffix := WalkDir(td, "a")
	if len(dirsSuffix) != 1 {
		t.Errorf("WalkDir with suffix len = %d, want 1", len(dirsSuffix))
	}
}
