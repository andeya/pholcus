package teleport

import (
	"bytes"
	"testing"
)

func TestNewProtocol(t *testing.T) {
	tests := []struct {
		header string
	}{
		{""},
		{"andeya"},
		{"custom-header"},
	}
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			p := NewProtocol(tt.header)
			if p == nil {
				t.Fatal("NewProtocol returned nil")
			}
			if p.header != tt.header {
				t.Errorf("header = %q, want %q", p.header, tt.header)
			}
			wantLen := len([]byte(tt.header))
			if p.headerLen != wantLen {
				t.Errorf("headerLen = %d, want %d", p.headerLen, wantLen)
			}
		})
	}
}

func TestProtocol_ReSet(t *testing.T) {
	p := NewProtocol("old")
	p.ReSet("new")
	if p.header != "new" {
		t.Errorf("header = %q, want new", p.header)
	}
	if p.headerLen != 3 {
		t.Errorf("headerLen = %d, want 3", p.headerLen)
	}
}

func TestProtocol_Packet(t *testing.T) {
	p := NewProtocol("andeya")
	msg := []byte("hello")
	got := p.Packet(msg)
	want := append(append([]byte("andeya"), IntToBytes(len(msg))...), msg...)
	if !bytes.Equal(got, want) {
		t.Errorf("Packet() = %v, want %v", got, want)
	}
}

func TestProtocol_Unpack(t *testing.T) {
	p := NewProtocol("andeya")
	msg := []byte("hello")
	packed := p.Packet(msg)
	slice, rest := p.Unpack(packed)
	if len(slice) != 1 {
		t.Fatalf("len(slice) = %d, want 1", len(slice))
	}
	if !bytes.Equal(slice[0], msg) {
		t.Errorf("Unpack()[0] = %v, want %v", slice[0], msg)
	}
	if len(rest) != 0 {
		t.Errorf("rest = %v, want empty", rest)
	}
}

func TestProtocol_Unpack_Multiple(t *testing.T) {
	p := NewProtocol("ab")
	m1 := []byte("x")
	m2 := []byte("yz")
	packed := append(p.Packet(m1), p.Packet(m2)...)
	slice, rest := p.Unpack(packed)
	if len(slice) != 2 {
		t.Fatalf("len(slice) = %d, want 2", len(slice))
	}
	if !bytes.Equal(slice[0], m1) {
		t.Errorf("slice[0] = %v, want %v", slice[0], m1)
	}
	if !bytes.Equal(slice[1], m2) {
		t.Errorf("slice[1] = %v, want %v", slice[1], m2)
	}
	if len(rest) != 0 {
		t.Errorf("rest len = %d, want 0", len(rest))
	}
}

func TestProtocol_Unpack_Partial(t *testing.T) {
	p := NewProtocol("ab")
	msg := []byte("full")
	packed := p.Packet(msg)
	partial := packed[:len(packed)-2]
	slice, rest := p.Unpack(partial)
	if len(slice) != 0 {
		t.Errorf("len(slice) = %d, want 0", len(slice))
	}
	if !bytes.Equal(rest, partial) {
		t.Errorf("rest = %v, want %v", rest, partial)
	}
}

func TestProtocol_Unpack_GarbageBeforeHeader(t *testing.T) {
	p := NewProtocol("ab")
	msg := []byte("x")
	packed := p.Packet(msg)
	buf := append([]byte("xx"), packed...)
	slice, rest := p.Unpack(buf)
	if len(slice) != 1 {
		t.Fatalf("len(slice) = %d, want 1", len(slice))
	}
	if !bytes.Equal(slice[0], msg) {
		t.Errorf("slice[0] = %v, want %v", slice[0], msg)
	}
	if len(rest) != 0 {
		t.Errorf("rest len = %d, want 0", len(rest))
	}
}

func TestProtocol_Unpack_EmptyBuffer(t *testing.T) {
	p := NewProtocol("ab")
	slice, rest := p.Unpack([]byte{})
	if len(slice) != 0 {
		t.Errorf("len(slice) = %d, want 0", len(slice))
	}
	if len(rest) != 0 {
		t.Errorf("rest = %v, want empty", rest)
	}
}

func TestProtocol_Unpack_TooShort(t *testing.T) {
	p := NewProtocol("andeya")
	slice, rest := p.Unpack([]byte("and"))
	if len(slice) != 0 {
		t.Errorf("len(slice) = %d, want 0", len(slice))
	}
	if !bytes.Equal(rest, []byte("and")) {
		t.Errorf("rest = %v", rest)
	}
}

func TestIntToBytes_BytesToInt(t *testing.T) {
	tests := []int{0, 1, 42, 1024, -1}
	for _, n := range tests {
		b := IntToBytes(n)
		got := BytesToInt(b)
		if got != n {
			t.Errorf("BytesToInt(IntToBytes(%d)) = %d", n, got)
		}
	}
}
