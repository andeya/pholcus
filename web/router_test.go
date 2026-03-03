package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andeya/pholcus/runtime/cache"
	"github.com/andeya/pholcus/runtime/status"
)

func TestRouter(t *testing.T) {
	cache.Task.Mode = status.OFFLINE
	cache.Task.Port = 9090
	cache.Task.Master = "127.0.0.1"

	Router()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"GET root", "GET", "/", http.StatusOK},
		{"GET public css", "GET", "/public/css/pholcus.css", http.StatusOK},
		{"GET public index", "GET", "/public/index.html", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(nil)
			defer server.Close()

			req, _ := http.NewRequest(tt.method, server.URL+tt.path, nil)
			resp, err := server.Client().Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Router() %s %s status = %v, want %v", tt.method, tt.path, resp.StatusCode, tt.wantStatus)
			}
		})
	}
}
