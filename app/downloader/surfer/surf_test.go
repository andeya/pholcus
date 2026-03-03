package surfer

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if _, ok := s.(*Surf); !ok {
		t.Errorf("New() = %T, want *Surf", s)
	}

	jar, _ := cookiejar.New(nil)
	s2 := New(jar)
	if s2 == nil {
		t.Fatal("New(jar) returned nil")
	}
}

func TestSurfDownload(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct {
		name   string
		method string
		url    string
		want   string
	}{
		{"GET", "GET", srv.URL, "hello"},
		{"HEAD", "HEAD", srv.URL, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			req := &DefaultRequest{
				URL:         tt.url,
				Method:      tt.method,
				TryTimes:    3,
				RetryPause:  time.Millisecond,
				DialTimeout: time.Second,
				ConnTimeout: time.Second,
			}
			r := s.Download(req)
			if r.IsErr() {
				t.Fatalf("Download() err: %v", r.UnwrapErr())
			}
			resp := r.Unwrap()
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
			}
			body, _ := io.ReadAll(resp.Body)
			if !strings.Contains(string(body), tt.want) && tt.want != "" {
				t.Errorf("body = %q, want to contain %q", body, tt.want)
			}
		})
	}
}

func TestSurfDownloadGzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		gz.Write([]byte("gzip body"))
		gz.Close()
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:         srv.URL,
		Method:      "GET",
		TryTimes:    3,
		RetryPause:  time.Millisecond,
		DialTimeout: time.Second,
		ConnTimeout: time.Second,
	}
	r := s.Download(req)
	if r.IsErr() {
		t.Fatalf("Download() err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "gzip body" {
		t.Errorf("body = %q, want %q", body, "gzip body")
	}
}

func TestSurfDownloadPOST(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:         srv.URL,
		Method:      "POST",
		PostData:    "a=1&b=2",
		TryTimes:    3,
		RetryPause:  time.Millisecond,
		DialTimeout: time.Second,
		ConnTimeout: time.Second,
	}
	r := s.Download(req)
	if r.IsErr() {
		t.Fatalf("Download() err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "a=1&b=2" {
		t.Errorf("body = %q, want a=1&b=2", body)
	}
}

func TestSurfDownloadPOSTM(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("Content-Type = %s, want multipart", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:         srv.URL,
		Method:      "POST-M",
		PostData:    "k=v",
		TryTimes:    3,
		RetryPause:  time.Millisecond,
		DialTimeout: time.Second,
		ConnTimeout: time.Second,
	}
	r := s.Download(req)
	if r.IsErr() {
		t.Fatalf("Download() err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d", resp.StatusCode)
	}
}

func TestSurfDownloadRetry(t *testing.T) {
	var attempt int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "no hijack", 500)
				return
			}
			conn, _, _ := hj.Hijack()
			conn.Close()
			return
		}
		w.Write([]byte("ok"))
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:          srv.URL,
		Method:       "GET",
		TryTimes:     3,
		RetryPause:   time.Millisecond,
		EnableCookie: false,
		DialTimeout:  time.Second,
		ConnTimeout:  time.Second,
	}
	r := s.Download(req)
	if r.IsErr() {
		t.Fatalf("Download err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("body = %q", body)
	}
	if attempt < 2 {
		t.Errorf("expected retry, got %d attempts", attempt)
	}
}

func TestSurfDownloadWithCookie(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:          srv.URL,
		Method:       "GET",
		TryTimes:     3,
		RetryPause:   time.Millisecond,
		EnableCookie: true,
		DialTimeout:  time.Second,
		ConnTimeout:  time.Second,
	}
	r := s.Download(req)
	if r.IsErr() {
		t.Fatalf("Download err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
}

func TestSurfDownloadHTTPS(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("https ok"))
	})
	srv := httptest.NewTLSServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:         srv.URL,
		Method:      "GET",
		TryTimes:    3,
		RetryPause:  time.Millisecond,
		DialTimeout: time.Second,
		ConnTimeout: time.Second,
	}
	r := s.Download(req)
	if r.IsErr() {
		t.Fatalf("Download() err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "https ok" {
		t.Errorf("body = %q, want https ok", body)
	}
}

func TestSurfDownloadDeflate(t *testing.T) {
	var buf bytes.Buffer
	fw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	fw.Write([]byte("deflate body"))
	fw.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "deflate")
		w.Write(buf.Bytes())
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:         srv.URL,
		Method:      "GET",
		TryTimes:    3,
		RetryPause:  time.Millisecond,
		DialTimeout: time.Second,
		ConnTimeout: time.Second,
	}
	r := s.Download(req)
	if r.IsErr() {
		t.Fatalf("Download err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "deflate body" {
		t.Errorf("deflate body = %q, want deflate body", body)
	}
}

func TestSurfDownloadZlib(t *testing.T) {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	zw.Write([]byte("zlib body"))
	zw.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "zlib")
		w.Write(buf.Bytes())
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:         srv.URL,
		Method:      "GET",
		TryTimes:    3,
		RetryPause:  time.Millisecond,
		DialTimeout: time.Second,
		ConnTimeout: time.Second,
	}
	r := s.Download(req)
	if r.IsErr() {
		t.Fatalf("Download err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "zlib body" {
		t.Errorf("zlib body = %q, want zlib body", body)
	}
}

func TestDownloadSurfID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req := &DefaultRequest{
		URL:          srv.URL,
		Method:       "GET",
		DownloaderID: SurfID,
		TryTimes:     3,
		RetryPause:   time.Millisecond,
		DialTimeout:  time.Second,
		ConnTimeout:  time.Second,
	}
	r := Download(req)
	if r.IsErr() {
		t.Fatalf("Download err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("body = %q", body)
	}
}

func TestDestroyJsFiles(t *testing.T) {
	DestroyJsFiles()
}

func TestDownloadUnknownID(t *testing.T) {
	req := &mockRequest{downloaderID: 99}
	r := Download(req)
	if r.IsOk() {
		t.Error("Download expected error for unknown ID")
	}
}

type mockRequest struct {
	downloaderID int
}

func (m *mockRequest) GetURL() string                { return "http://example.com" }
func (m *mockRequest) GetMethod() string             { return "GET" }
func (m *mockRequest) GetPostData() string           { return "" }
func (m *mockRequest) GetHeader() http.Header        { return nil }
func (m *mockRequest) GetEnableCookie() bool         { return false }
func (m *mockRequest) GetDialTimeout() time.Duration { return time.Second }
func (m *mockRequest) GetConnTimeout() time.Duration { return time.Second }
func (m *mockRequest) GetTryTimes() int              { return 1 }
func (m *mockRequest) GetRetryPause() time.Duration  { return time.Millisecond }
func (m *mockRequest) GetProxy() string              { return "" }
func (m *mockRequest) GetRedirectTimes() int         { return 0 }
func (m *mockRequest) GetDownloaderID() int          { return m.downloaderID }

func TestDnsCache(t *testing.T) {
	dc := &DnsCache{}
	dc.Reg("host:80", "127.0.0.1:80")
	opt := dc.Query("host:80")
	if !opt.IsSome() || opt.Unwrap() != "127.0.0.1:80" {
		t.Errorf("Query = %v, want Some(127.0.0.1:80)", opt)
	}
	dc.Del("host:80")
	opt2 := dc.Query("host:80")
	if opt2.IsSome() {
		t.Errorf("Query after Del = %v, want None", opt2)
	}
}
