package websocket

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestProtocolError(t *testing.T) {
	err := &ProtocolError{"test"}
	if err.Error() != "test" {
		t.Errorf("Error() = %q, want test", err.Error())
	}
}

func TestAddr_Network(t *testing.T) {
	u, _ := url.Parse("ws://example.com/path")
	addr := &Addr{u}
	if addr.Network() != "websocket" {
		t.Errorf("Network() = %q, want websocket", addr.Network())
	}
}

func TestConn_IsClientConn_IsServerConn(t *testing.T) {
	srv := newWSServer(t)
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()

	if !ws.IsClientConn() {
		t.Error("IsClientConn() = false, want true")
	}
	if ws.IsServerConn() {
		t.Error("IsServerConn() = true, want false")
	}
}

func TestConn_LocalAddr_RemoteAddr(t *testing.T) {
	srv := newWSServer(t)
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()

	loc := ws.LocalAddr()
	if loc.Network() != "websocket" {
		t.Errorf("LocalAddr().Network() = %q", loc.Network())
	}
	rem := ws.RemoteAddr()
	if rem.Network() != "websocket" {
		t.Errorf("RemoteAddr().Network() = %q", rem.Network())
	}
}

func TestConn_Config_Request(t *testing.T) {
	srv := newWSServer(t)
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()

	cfg := ws.Config()
	if cfg == nil {
		t.Fatal("Config() returned nil")
	}
	if cfg.Location == nil {
		t.Error("Config().Location is nil")
	}
	if ws.Request() != nil {
		t.Error("Request() should be nil for client conn")
	}
}

func TestConn_SetDeadline(t *testing.T) {
	srv := newWSServer(t)
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()

	err := ws.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetDeadline: %v", err)
	}
	err = ws.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetReadDeadline: %v", err)
	}
	err = ws.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetWriteDeadline: %v", err)
	}
}

func TestConn_Read_Write_Close(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			defer close(done)
			buf := make([]byte, 256)
			n, err := ws.Read(buf)
			if err != nil {
				t.Errorf("server Read: %v", err)
				return
			}
			_, err = ws.Write(buf[:n])
			if err != nil {
				t.Errorf("server Write: %v", err)
				return
			}
			ws.Close()
		},
	})
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	msg := []byte("hello")
	n, err := ws.Write(msg)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(msg) {
		t.Errorf("Write returned %d, want %d", n, len(msg))
	}
	buf := make([]byte, 256)
	n, err = ws.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "hello" {
		t.Errorf("Read = %q, want hello", buf[:n])
	}
	ws.Close()
	<-done
}

func TestMessage_Send_Receive(t *testing.T) {
	tests := []struct {
		name  string
		send  interface{}
		recv  interface{}
		check func(t *testing.T, recv interface{})
	}{
		{
			name: "text",
			send: "hello",
			recv: new(string),
			check: func(t *testing.T, recv interface{}) {
				if *recv.(*string) != "hello" {
					t.Errorf("received %q, want hello", *recv.(*string))
				}
			},
		},
		{
			name: "binary",
			send: []byte{1, 2, 3},
			recv: new([]byte),
			check: func(t *testing.T, recv interface{}) {
				b := *recv.(*[]byte)
				if len(b) != 3 || b[0] != 1 || b[1] != 2 || b[2] != 3 {
					t.Errorf("received %v, want [1,2,3]", b)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})
			srv := httptest.NewServer(Server{
				Handler: func(ws *Conn) {
					defer close(done)
					if err := Message.Receive(ws, tt.recv); err != nil {
						t.Errorf("Receive: %v", err)
						return
					}
					tt.check(t, tt.recv)
					if _, err := Message.Send(ws, tt.send); err != nil {
						t.Errorf("Send: %v", err)
					}
					ws.Close()
				},
			})
			defer srv.Close()

			client := dialWS(t, srv.URL, "http://example.com")
			defer client.Close()

			if _, err := Message.Send(client, tt.send); err != nil {
				t.Fatalf("Send: %v", err)
			}
			if err := Message.Receive(client, tt.recv); err != nil {
				t.Fatalf("Receive: %v", err)
			}
			tt.check(t, tt.recv)
			<-done
		})
	}
}

func TestMessage_Send_Unsupported(t *testing.T) {
	srv := newWSServer(t)
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()

	_, err := Message.Send(ws, 123)
	if err != ErrNotSupported {
		t.Errorf("Send(123) err = %v, want ErrNotSupported", err)
	}
}

func TestMessage_Receive_Unsupported(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			defer close(done)
			Message.Send(ws, "x")
			ws.Close()
		},
	})
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()

	var v int
	err := Message.Receive(ws, &v)
	if err != ErrNotSupported {
		t.Errorf("Receive(int) err = %v, want ErrNotSupported", err)
	}
	<-done
}

func TestMarshal_Unsupported(t *testing.T) {
	_, _, err := marshal(123)
	if err != ErrNotSupported {
		t.Errorf("marshal(123) err = %v", err)
	}
}

func TestUnmarshal_Unsupported(t *testing.T) {
	err := unmarshal([]byte("x"), TextFrame, new(int))
	if err != ErrNotSupported {
		t.Errorf("unmarshal(int) err = %v", err)
	}
}

func TestJSON_Send_Receive(t *testing.T) {
	type T struct {
		Msg   string `json:"msg"`
		Count int    `json:"count"`
	}
	send := T{Msg: "hi", Count: 42}

	done := make(chan struct{})
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			defer close(done)
			var recv T
			if err := JSON.Receive(ws, &recv); err != nil {
				t.Errorf("Receive: %v", err)
				return
			}
			if recv.Msg != send.Msg || recv.Count != send.Count {
				t.Errorf("received %+v, want %+v", recv, send)
			}
			JSON.Send(ws, recv)
			ws.Close()
		},
	})
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()

	if _, err := JSON.Send(ws, send); err != nil {
		t.Fatalf("Send: %v", err)
	}
	var recv T
	if err := JSON.Receive(ws, &recv); err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if recv.Msg != send.Msg || recv.Count != send.Count {
		t.Errorf("received %+v, want %+v", recv, send)
	}
	<-done
}

func TestConn_SetDeadline_NonNetConn(t *testing.T) {
	ws := &Conn{rwc: &mockRWC{}}
	err := ws.SetDeadline(time.Now())
	if err != errSetDeadline {
		t.Errorf("SetDeadline = %v, want errSetDeadline", err)
	}
	err = ws.SetReadDeadline(time.Now())
	if err != errSetDeadline {
		t.Errorf("SetReadDeadline = %v, want errSetDeadline", err)
	}
	err = ws.SetWriteDeadline(time.Now())
	if err != errSetDeadline {
		t.Errorf("SetWriteDeadline = %v, want errSetDeadline", err)
	}
}

func TestConn_ServerSide_Addrs_Request(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			defer close(done)
			if !ws.IsServerConn() {
				t.Error("IsServerConn() = false for server conn")
			}
			if ws.IsClientConn() {
				t.Error("IsClientConn() = true for server conn")
			}
			loc := ws.LocalAddr()
			if loc == nil || loc.Network() != "websocket" {
				t.Errorf("LocalAddr() = %v", loc)
			}
			rem := ws.RemoteAddr()
			if rem == nil || rem.Network() != "websocket" {
				t.Errorf("RemoteAddr() = %v", rem)
			}
			if ws.Request() == nil {
				t.Error("Request() is nil for server conn")
			}
			ws.Close()
		},
	})
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	ws.Close()
	<-done
}

func TestCodec_Receive_WithPartialRead(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			defer close(done)
			Message.Send(ws, "first")
			Message.Send(ws, "second")
			ws.Close()
		},
	})
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()

	buf := make([]byte, 2)
	n, err := ws.Read(buf)
	if err != nil || n != 2 {
		t.Fatalf("Read: n=%d err=%v", n, err)
	}
	var s string
	if err := Message.Receive(ws, &s); err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if s != "second" {
		t.Errorf("Receive = %q, want second (partial first discarded)", s)
	}
	<-done
}

func TestConn_Read_UntilClose(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			ws.Write([]byte("hi"))
			ws.Close()
			close(done)
		},
	})
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	buf := make([]byte, 256)
	n, err := ws.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "hi" {
		t.Errorf("Read = %q, want hi", buf[:n])
	}
	_, err = ws.Read(buf)
	if err == nil {
		t.Error("Read after close: expected error")
	}
	ws.Close()
	<-done
}

func TestConn_Write_BinaryFrame(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			defer close(done)
			ws.PayloadType = BinaryFrame
			buf := make([]byte, 256)
			n, _ := ws.Read(buf)
			ws.Write(buf[:n])
			ws.Close()
		},
	})
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()
	ws.PayloadType = BinaryFrame
	msg := []byte{0x00, 0x01, 0x02}
	n, err := ws.Write(msg)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(msg) {
		t.Errorf("Write returned %d, want %d", n, len(msg))
	}
	buf := make([]byte, 256)
	n, err = ws.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n != 3 || buf[0] != 0 || buf[1] != 1 || buf[2] != 2 {
		t.Errorf("Read = %v", buf[:n])
	}
	<-done
}

func TestConn_LargePayload(t *testing.T) {
	payload := make([]byte, 200)
	for i := range payload {
		payload[i] = byte(i)
	}
	done := make(chan struct{})
	srv := httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			defer close(done)
			buf := make([]byte, 512)
			n, err := ws.Read(buf)
			if err != nil {
				t.Errorf("Read: %v", err)
				return
			}
			ws.Write(buf[:n])
			ws.Close()
		},
	})
	defer srv.Close()

	ws := dialWS(t, srv.URL, "http://example.com")
	defer ws.Close()
	n, err := ws.Write(payload)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(payload) {
		t.Errorf("Write returned %d, want %d", n, len(payload))
	}
	buf := make([]byte, 512)
	n, err = ws.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n != len(payload) {
		t.Errorf("Read returned %d, want %d", n, len(payload))
	}
	for i := 0; i < n; i++ {
		if buf[i] != byte(i) {
			t.Errorf("buf[%d] = %d, want %d", i, buf[i], i)
			break
		}
	}
	<-done
}

type mockRWC struct{}

func (m *mockRWC) Read([]byte) (int, error)  { return 0, nil }
func (m *mockRWC) Write([]byte) (int, error) { return 0, nil }
func (m *mockRWC) Close() error              { return nil }

func newWSServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(Server{
		Handler: func(ws *Conn) {
			ws.Close()
		},
	})
}

func dialWS(t *testing.T, httpURL, origin string) *Conn {
	t.Helper()
	wsURL := "ws" + httpURL[4:]
	ws, err := Dial(wsURL, "", origin)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	return ws
}
