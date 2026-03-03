package websocket

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandler_ServeHTTP_CheckOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		wantErr bool
	}{
		{"valid origin", "http://example.com", false},
		{"null origin", "null", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})
			srv := httptest.NewServer(Handler(func(ws *Conn) {
				close(done)
				ws.Close()
			}))
			defer srv.Close()

			conn := dialConn(t, srv.URL)
			defer conn.Close()

			u, _ := url.Parse(srv.URL)
			path := "/"
			if u.Path != "" {
				path = u.Path
			}
			req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nOrigin: %s\r\nSec-WebSocket-Version: 13\r\n\r\n",
				path, u.Host, base64.StdEncoding.EncodeToString([]byte("1234567890123456")), tt.origin)

			_, err := conn.Write([]byte(req))
			if err != nil {
				t.Fatalf("Write: %v", err)
			}

			br := bufio.NewReader(conn)
			resp, err := http.ReadResponse(br, &http.Request{Method: "GET"})
			if err != nil {
				t.Fatalf("ReadResponse: %v", err)
			}
			resp.Body.Close()

			if tt.wantErr {
				if resp.StatusCode != 403 {
					t.Errorf("status = %d, want 403", resp.StatusCode)
				}
				return
			}
			if resp.StatusCode != 101 {
				t.Errorf("status = %d, want 101", resp.StatusCode)
			}
			if !tt.wantErr {
				<-done
			}
		})
	}
}

func TestNewServerConn_BadHandshake(t *testing.T) {
	tests := []struct {
		name   string
		req    string
		status int
		body   string
	}{
		{
			name:   "bad method",
			req:    "POST / HTTP/1.1\r\nHost: x\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n",
			status: 405,
			body:   ErrBadRequestMethod.Error(),
		},
		{
			name:   "bad upgrade",
			req:    "GET / HTTP/1.1\r\nHost: x\r\nUpgrade: http\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n",
			status: 400,
			body:   ErrNotWebSocket.Error(),
		},
		{
			name:   "bad version",
			req:    "GET / HTTP/1.1\r\nHost: x\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 99\r\n\r\n",
			status: 400,
			body:   ErrBadWebSocketVersion.Error(),
		},
		{
			name:   "missing key",
			req:    "GET / HTTP/1.1\r\nHost: x\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Version: 13\r\n\r\n",
			status: 400,
			body:   ErrChallengeResponse.Error(),
		},
		{
			name:   "bad connection",
			req:    "GET / HTTP/1.1\r\nHost: x\r\nUpgrade: websocket\r\nConnection: keep-alive\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n",
			status: 400,
			body:   ErrNotWebSocket.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})
			srv := httptest.NewServer(Server{
				Handler: func(ws *Conn) {
					close(done)
					ws.Close()
				},
			})
			defer srv.Close()

			u, _ := url.Parse(srv.URL)
			conn, err := net.Dial("tcp", u.Host)
			if err != nil {
				t.Fatalf("Dial: %v", err)
			}
			defer conn.Close()

			req := strings.Replace(tt.req, "Host: x", "Host: "+u.Host, 1)
			path := "/"
			if u.Path != "" {
				path = u.Path
			}
			req = strings.Replace(req, "GET /", "GET "+path, 1)

			_, err = conn.Write([]byte(req))
			if err != nil {
				t.Fatalf("Write: %v", err)
			}

			br := bufio.NewReader(conn)
			resp, err := http.ReadResponse(br, &http.Request{Method: "GET"})
			if err != nil {
				t.Fatalf("ReadResponse: %v", err)
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != tt.status {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.status)
			}
			if tt.body != "" && !bytes.Contains(body, []byte(tt.body)) {
				t.Errorf("body %q does not contain %q", string(body), tt.body)
			}
			select {
			case <-done:
				t.Error("handler should not have been called")
			default:
			}
		})
	}
}

func TestServer_HandshakeReject(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(Server{
		Handshake: func(config *Config, req *http.Request) error {
			return fmt.Errorf("rejected")
		},
		Handler: func(ws *Conn) {
			close(done)
			ws.Close()
		},
	})
	defer srv.Close()

	wsURL := "ws" + srv.URL[4:]
	_, err := Dial(wsURL, "", "http://example.com")
	if err == nil {
		t.Error("expected Dial to fail")
	}
	select {
	case <-done:
		t.Error("handler should not have been called")
	default:
	}
}
