package websocket

import (
	"io"
	"net"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDialError_Error(t *testing.T) {
	cfg, _ := NewConfig("ws://example.com", "http://example.com")
	err := &DialError{cfg, ErrBadStatus}
	s := err.Error()
	if s == "" || len(s) < 20 {
		t.Errorf("Error() = %q", s)
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		server  string
		origin  string
		wantErr bool
	}{
		{"ok", "ws://example.com/path", "http://example.com", false},
		{"bad server", "://invalid", "http://example.com", true},
		{"bad origin", "ws://example.com", "://invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewConfig(tt.server, tt.origin)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewConfig: %v", err)
			}
			if cfg.Version != ProtocolVersionHybi13 {
				t.Errorf("Version = %d, want 13", cfg.Version)
			}
			if cfg.Header == nil {
				t.Error("Header is nil")
			}
		})
	}
}

func TestDialConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  func() *Config
		wantErr bool
	}{
		{
			name: "nil location",
			config: func() *Config {
				cfg, _ := NewConfig("ws://x.com", "http://x.com")
				cfg.Location = nil
				return cfg
			},
			wantErr: true,
		},
		{
			name: "nil origin",
			config: func() *Config {
				cfg, _ := NewConfig("ws://x.com", "http://x.com")
				cfg.Origin = nil
				return cfg
			},
			wantErr: true,
		},
		{
			name: "bad scheme",
			config: func() *Config {
				cfg, _ := NewConfig("http://x.com", "http://x.com")
				return cfg
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws, err := DialConfig(tt.config())
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				if ws != nil {
					t.Error("expected nil conn")
				}
				return
			}
			if err != nil {
				t.Errorf("DialConfig: %v", err)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			ws.Close()
		},
	})
	defer srv.Close()

	cfg, err := NewConfig("ws"+srv.URL[4:], "http://example.com")
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}

	conn := dialConn(t, srv.URL)
	defer conn.Close()

	ws, err := NewClient(cfg, conn)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ws.Close()
}

func TestDial_WithProtocol(t *testing.T) {
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			ws.Close()
		},
	})
	defer srv.Close()

	ws, err := Dial("ws"+srv.URL[4:], "proto1", "http://example.com")
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	ws.Close()
}

func dialConn(t *testing.T, httpURL string) io.ReadWriteCloser {
	t.Helper()
	u, err := url.Parse(httpURL)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}
