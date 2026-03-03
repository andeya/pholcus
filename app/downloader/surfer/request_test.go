package surfer

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultRequestPrepare(t *testing.T) {
	tests := []struct {
		name string
		req  *DefaultRequest
		chk  func(*testing.T, *DefaultRequest)
	}{
		{
			name: "default method",
			req:  &DefaultRequest{URL: "http://a.com"},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetMethod() != DefaultMethod {
					t.Errorf("Method = %q", r.GetMethod())
				}
			},
		},
		{
			name: "default dial timeout",
			req:  &DefaultRequest{URL: "http://a.com"},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetDialTimeout() != DefaultDialTimeout {
					t.Errorf("DialTimeout = %v", r.GetDialTimeout())
				}
			},
		},
		{
			name: "default conn timeout",
			req:  &DefaultRequest{URL: "http://a.com"},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetConnTimeout() != DefaultConnTimeout {
					t.Errorf("ConnTimeout = %v", r.GetConnTimeout())
				}
			},
		},
		{
			name: "default try times",
			req:  &DefaultRequest{URL: "http://a.com"},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetTryTimes() != DefaultTryTimes {
					t.Errorf("TryTimes = %v", r.GetTryTimes())
				}
			},
		},
		{
			name: "default retry pause",
			req:  &DefaultRequest{URL: "http://a.com"},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetRetryPause() != DefaultRetryPause {
					t.Errorf("RetryPause = %v", r.GetRetryPause())
				}
			},
		},
		{
			name: "negative dial timeout",
			req:  &DefaultRequest{URL: "http://a.com", DialTimeout: -1},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetDialTimeout() != 0 {
					t.Errorf("DialTimeout = %v", r.GetDialTimeout())
				}
			},
		},
		{
			name: "negative conn timeout",
			req:  &DefaultRequest{URL: "http://a.com", ConnTimeout: -1},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetConnTimeout() != 0 {
					t.Errorf("ConnTimeout = %v", r.GetConnTimeout())
				}
			},
		},
		{
			name: "method uppercase",
			req:  &DefaultRequest{URL: "http://a.com", Method: "get"},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetMethod() != "GET" {
					t.Errorf("Method = %q", r.GetMethod())
				}
			},
		},
		{
			name: "PhantomJsID preserved",
			req:  &DefaultRequest{URL: "http://a.com", DownloaderID: PhantomJsID},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetDownloaderID() != PhantomJsID {
					t.Errorf("DownloaderID = %v", r.GetDownloaderID())
				}
			},
		},
		{
			name: "ChromeID preserved",
			req:  &DefaultRequest{URL: "http://a.com", DownloaderID: ChromeID},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetDownloaderID() != ChromeID {
					t.Errorf("DownloaderID = %v", r.GetDownloaderID())
				}
			},
		},
		{
			name: "invalid DownloaderID defaults to SurfID",
			req:  &DefaultRequest{URL: "http://a.com", DownloaderID: 99},
			chk: func(t *testing.T, r *DefaultRequest) {
				r.GetURL()
				if r.GetDownloaderID() != SurfID {
					t.Errorf("DownloaderID = %v", r.GetDownloaderID())
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.chk(t, tt.req)
		})
	}
}

func TestDefaultRequestGetters(t *testing.T) {
	h := http.Header{"X-Custom": {"val"}}
	req := &DefaultRequest{
		URL:           "http://example.com/path",
		Method:        "POST",
		PostData:      "a=1",
		Header:        h,
		EnableCookie:  true,
		DialTimeout:   time.Minute,
		ConnTimeout:   time.Minute,
		TryTimes:      5,
		RetryPause:    time.Second,
		RedirectTimes: 3,
		Proxy:         "http://proxy:8080",
		DownloaderID:  SurfID,
	}
	req.GetURL()
	if req.GetURL() != "http://example.com/path" {
		t.Errorf("GetURL = %q", req.GetURL())
	}
	if req.GetMethod() != "POST" {
		t.Errorf("GetMethod = %q", req.GetMethod())
	}
	if req.GetPostData() != "a=1" {
		t.Errorf("GetPostData = %q", req.GetPostData())
	}
	if req.GetHeader().Get("X-Custom") != "val" {
		t.Errorf("GetHeader X-Custom = %q", req.GetHeader().Get("X-Custom"))
	}
	if !req.GetEnableCookie() {
		t.Error("GetEnableCookie = false")
	}
	if req.GetDialTimeout() != time.Minute {
		t.Errorf("GetDialTimeout = %v", req.GetDialTimeout())
	}
	if req.GetConnTimeout() != time.Minute {
		t.Errorf("GetConnTimeout = %v", req.GetConnTimeout())
	}
	if req.GetTryTimes() != 5 {
		t.Errorf("GetTryTimes = %v", req.GetTryTimes())
	}
	if req.GetRetryPause() != time.Second {
		t.Errorf("GetRetryPause = %v", req.GetRetryPause())
	}
	if req.GetProxy() != "http://proxy:8080" {
		t.Errorf("GetProxy = %q", req.GetProxy())
	}
	if req.GetRedirectTimes() != 3 {
		t.Errorf("GetRedirectTimes = %v", req.GetRedirectTimes())
	}
	if req.GetDownloaderID() != SurfID {
		t.Errorf("GetDownloaderID = %v", req.GetDownloaderID())
	}
}
