package surfer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewParam(t *testing.T) {
	tests := []struct {
		name    string
		req     *DefaultRequest
		wantErr bool
	}{
		{
			name: "GET",
			req: &DefaultRequest{
				URL:         "http://example.com",
				Method:      "GET",
				TryTimes:    1,
				RetryPause:  time.Millisecond,
				DialTimeout: time.Second,
			},
			wantErr: false,
		},
		{
			name: "POST",
			req: &DefaultRequest{
				URL:         "http://example.com",
				Method:      "POST",
				PostData:    "a=1",
				TryTimes:    1,
				RetryPause:  time.Millisecond,
				DialTimeout: time.Second,
			},
			wantErr: false,
		},
		{
			name: "POST-M",
			req: &DefaultRequest{
				URL:         "http://example.com",
				Method:      "POST-M",
				PostData:    "k=v",
				TryTimes:    1,
				RetryPause:  time.Millisecond,
				DialTimeout: time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid URL",
			req: &DefaultRequest{
				URL:         "://invalid",
				Method:      "GET",
				TryTimes:    1,
				RetryPause:  time.Millisecond,
				DialTimeout: time.Second,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewParam(tt.req)
			if tt.wantErr && r.IsOk() {
				t.Error("NewParam expected error")
			}
			if !tt.wantErr && r.IsErr() {
				t.Errorf("NewParam err: %v", r.UnwrapErr())
			}
		})
	}
}

func TestNewParamWithProxy(t *testing.T) {
	req := &DefaultRequest{
		URL:         "http://example.com",
		Method:      "GET",
		Proxy:       "http://proxy.example.com:8080",
		TryTimes:    1,
		RetryPause:  time.Millisecond,
		DialTimeout: time.Second,
	}
	r := NewParam(req)
	if r.IsErr() {
		t.Errorf("NewParam with proxy err: %v", r.UnwrapErr())
	}
}

func TestNewParamWithUserAgent(t *testing.T) {
	req := &DefaultRequest{
		URL:         "http://example.com",
		Method:      "GET",
		Header:      http.Header{"User-Agent": {"CustomAgent/1.0"}},
		TryTimes:    1,
		RetryPause:  time.Millisecond,
		DialTimeout: time.Second,
	}
	r := NewParam(req)
	if r.IsErr() {
		t.Errorf("NewParam err: %v", r.UnwrapErr())
	}
}

func TestRedirectUnlimited(t *testing.T) {
	var redirectCount int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		if redirectCount <= 3 {
			http.Redirect(w, r, r.URL.Path, http.StatusFound)
			return
		}
		w.Write([]byte("ok"))
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:           srv.URL,
		Method:        "GET",
		RedirectTimes: 0,
		TryTimes:      3,
		RetryPause:    time.Millisecond,
		DialTimeout:   time.Second,
		ConnTimeout:   time.Second,
	}
	r := s.Download(req)
	if r.IsErr() {
		t.Fatalf("Download err: %v", r.UnwrapErr())
	}
	resp := r.Unwrap()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestRedirectLimited(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loop", http.StatusFound)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:           srv.URL,
		Method:        "GET",
		RedirectTimes: 2,
		TryTimes:      3,
		RetryPause:    time.Millisecond,
		DialTimeout:   time.Second,
		ConnTimeout:   time.Second,
	}
	r := s.Download(req)
	if r.IsOk() {
		t.Error("Download expected redirect error")
	}
}

func TestRedirectNotAllowed(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/other", http.StatusFound)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s := New()
	req := &DefaultRequest{
		URL:           srv.URL,
		Method:        "GET",
		RedirectTimes: -1,
		TryTimes:      3,
		RetryPause:    time.Millisecond,
		DialTimeout:   time.Second,
		ConnTimeout:   time.Second,
	}
	r := s.Download(req)
	if r.IsOk() {
		t.Error("Download expected no-redirect error")
	}
}
