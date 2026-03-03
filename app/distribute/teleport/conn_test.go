package teleport

import (
	"net"
	"testing"
)

func TestNewConnect(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	k, v := NewConnect(client, 1024, 256)
	if k != client.RemoteAddr().String() {
		t.Errorf("key = %q, want %q", k, client.RemoteAddr().String())
	}
	if v == nil {
		t.Fatal("Connect is nil")
	}
	if v.WriteChan == nil {
		t.Error("WriteChan is nil")
	}
	if len(v.Buffer) != 1024 {
		t.Errorf("Buffer len = %d, want 1024", len(v.Buffer))
	}
	if cap(v.WriteChan) != 256 {
		t.Errorf("WriteChan cap = %d, want 256", cap(v.WriteChan))
	}
}

func TestConnect_Addr(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	_, conn := NewConnect(client, 64, 16)
	addr := conn.Addr()
	want := client.RemoteAddr().String()
	if addr != want {
		t.Errorf("Addr() = %q, want %q", addr, want)
	}
}
