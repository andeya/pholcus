package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andeya/pholcus/runtime/cache"
	"github.com/andeya/pholcus/runtime/status"
)

func TestWebHandler(t *testing.T) {
	cache.Task.Mode = status.OFFLINE
	cache.Task.Port = 9090
	cache.Task.Master = "127.0.0.1"

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"GET root", "GET", "/", http.StatusOK},
		{"GET favicon", "GET", "/favicon.ico", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			web(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("web() status = %v, want %v", rec.Code, tt.wantStatus)
			}
		})
	}
}
