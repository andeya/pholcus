// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// taken from http://golang.org/src/pkg/net/ipraw_test.go

package ping

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

type emptyBody struct{}

func (*emptyBody) Len() int                    { return 0 }
func (*emptyBody) Marshal() ([]byte, error)    { return nil, nil }

type errorBody struct{}

func (*errorBody) Len() int                    { return 4 }
func (*errorBody) Marshal() ([]byte, error)    { return nil, errors.New("marshal error") }

func TestPing(t *testing.T) {
	t.Log(Ping("www.baidu.com", 5))
}

func TestIcmpEchoLen(t *testing.T) {
	tests := []struct {
		name string
		echo *icmpEcho
		want int
	}{
		{"nil", nil, 0},
		{"empty data", &icmpEcho{ID: 1, Seq: 1, Data: nil}, 4},
		{"with data", &icmpEcho{ID: 1, Seq: 1, Data: []byte("abc")}, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			if tt.echo != nil {
				got = tt.echo.Len()
			} else {
				got = (*icmpEcho)(nil).Len()
			}
			if got != tt.want {
				t.Errorf("Len() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIcmpEchoMarshal(t *testing.T) {
	tests := []struct {
		name string
		echo *icmpEcho
		want []byte
	}{
		{"empty", &icmpEcho{ID: 0, Seq: 0, Data: nil}, []byte{0, 0, 0, 0}},
		{"basic", &icmpEcho{ID: 0x1234, Seq: 0x5678, Data: nil}, []byte{0x12, 0x34, 0x56, 0x78}},
		{"with data", &icmpEcho{ID: 1, Seq: 2, Data: []byte("xyz")}, []byte{0, 1, 0, 2, 'x', 'y', 'z'}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.echo.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseICMPEcho(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		want *icmpEcho
	}{
		{"minimal", []byte{0, 0, 0, 0}, &icmpEcho{ID: 0, Seq: 0, Data: nil}},
		{"with id seq", []byte{0x12, 0x34, 0x56, 0x78}, &icmpEcho{ID: 0x1234, Seq: 0x5678, Data: nil}},
		{"with data", []byte{0, 1, 0, 2, 'a', 'b', 'c'}, &icmpEcho{ID: 1, Seq: 2, Data: []byte("abc")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseICMPEcho(tt.b)
			if err != nil {
				t.Fatalf("parseICMPEcho() error = %v", err)
			}
			if got.ID != tt.want.ID || got.Seq != tt.want.Seq || !bytes.Equal(got.Data, tt.want.Data) {
				t.Errorf("parseICMPEcho() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseICMPMessage(t *testing.T) {
	tests := []struct {
		name    string
		b       []byte
		wantErr bool
		check   func(*testing.T, *icmpMessage)
	}{
		{
			name:    "too short",
			b:       []byte{0, 0, 0},
			wantErr: true,
		},
		{
			name: "header only",
			b:    []byte{8, 0, 0, 0},
			check: func(t *testing.T, m *icmpMessage) {
				if m.Type != 8 || m.Code != 0 || m.Body != nil {
					t.Errorf("unexpected header: %+v", m)
				}
			},
		},
		{
			name: "icmpv4 echo request",
			b:    []byte{8, 0, 0, 0, 0, 1, 0, 2, 'x'},
			check: func(t *testing.T, m *icmpMessage) {
				if m.Type != 8 || m.Code != 0 {
					t.Errorf("unexpected type/code: %d/%d", m.Type, m.Code)
				}
				if m.Body == nil {
					t.Fatal("Body is nil")
				}
				echo := m.Body.(*icmpEcho)
				if echo.ID != 1 || echo.Seq != 2 || !bytes.Equal(echo.Data, []byte{'x'}) {
					t.Errorf("unexpected body: %+v", echo)
				}
			},
		},
		{
			name: "icmpv4 echo reply",
			b:    []byte{0, 0, 0, 0, 0, 1, 0, 2},
			check: func(t *testing.T, m *icmpMessage) {
				if m.Type != 0 || m.Body == nil {
					t.Errorf("unexpected: %+v", m)
				}
			},
		},
		{
			name: "icmpv6 echo request",
			b:    []byte{128, 0, 0, 0, 0, 1, 0, 2},
			check: func(t *testing.T, m *icmpMessage) {
				if m.Type != 128 || m.Body == nil {
					t.Errorf("unexpected: %+v", m)
				}
			},
		},
		{
			name: "icmpv6 echo reply",
			b:    []byte{129, 0, 0, 0, 0, 1, 0, 2},
			check: func(t *testing.T, m *icmpMessage) {
				if m.Type != 129 || m.Body == nil {
					t.Errorf("unexpected: %+v", m)
				}
			},
		},
		{
			name: "other type no body",
			b:    []byte{3, 1, 0, 0, 0xff, 0xff},
			check: func(t *testing.T, m *icmpMessage) {
				if m.Type != 3 || m.Code != 1 || m.Body != nil {
					t.Errorf("unexpected: %+v", m)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseICMPMessage(tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseICMPMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestIcmpMessageMarshal(t *testing.T) {
	tests := []struct {
		name string
		msg  *icmpMessage
	}{
		{
			name: "icmpv4 with body",
			msg: &icmpMessage{
				Type: icmpv4EchoRequest,
				Code: 0,
				Body: &icmpEcho{ID: 1, Seq: 1, Data: []byte("test")},
			},
		},
		{
			name: "icmpv4 header only",
			msg:  &icmpMessage{Type: icmpv4EchoRequest, Code: 0, Body: nil},
		},
		{
			name: "icmpv6 with body",
			msg: &icmpMessage{
				Type: icmpv6EchoRequest,
				Code: 0,
				Body: &icmpEcho{ID: 2, Seq: 3, Data: []byte("v6")},
			},
		},
		{
			name: "icmpv6 reply",
			msg: &icmpMessage{
				Type: icmpv6EchoReply,
				Code: 0,
				Body: &icmpEcho{ID: 1, Seq: 0, Data: nil},
			},
		},
		{
			name: "icmpv4 body len zero",
			msg:  &icmpMessage{Type: icmpv4EchoRequest, Code: 0, Body: &emptyBody{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.msg.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if len(got) < 4 {
				t.Fatalf("Marshal() too short: %d", len(got))
			}
			if got[0] != byte(tt.msg.Type) || got[1] != byte(tt.msg.Code) {
				t.Errorf("Marshal() header = %v, want type=%d code=%d", got[:4], tt.msg.Type, tt.msg.Code)
			}
			parsed, err := parseICMPMessage(got)
			if err != nil {
				t.Fatalf("parseICMPMessage(Marshal()) error = %v", err)
			}
			if parsed.Type != tt.msg.Type || parsed.Code != tt.msg.Code {
				t.Errorf("roundtrip type/code mismatch: got %d/%d", parsed.Type, parsed.Code)
			}
			if tt.msg.Body != nil && parsed.Body != nil {
				wantEcho := tt.msg.Body.(*icmpEcho)
				gotEcho := parsed.Body.(*icmpEcho)
				if wantEcho.ID != gotEcho.ID || wantEcho.Seq != gotEcho.Seq || !bytes.Equal(wantEcho.Data, gotEcho.Data) {
					t.Errorf("roundtrip body mismatch: got %+v, want %+v", gotEcho, wantEcho)
				}
			}
		})
	}
}

func TestIcmpMessageMarshalBodyError(t *testing.T) {
	msg := &icmpMessage{Type: icmpv4EchoRequest, Code: 0, Body: &errorBody{}}
	_, err := msg.Marshal()
	if err == nil {
		t.Error("Marshal() expected error from Body.Marshal()")
	}
}

func TestIcmpMessageMarshalRoundtrip(t *testing.T) {
	echo := &icmpEcho{ID: 1234, Seq: 5678, Data: []byte("Go Go Gadget Ping!!!")}
	msg := &icmpMessage{
		Type: icmpv4EchoRequest,
		Code: 0,
		Body: echo,
	}
	b, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	m, err := parseICMPMessage(b)
	if err != nil {
		t.Fatalf("parseICMPMessage() error = %v", err)
	}
	if m.Type != msg.Type || m.Code != msg.Code {
		t.Errorf("type/code mismatch")
	}
	gotEcho := m.Body.(*icmpEcho)
	if !reflect.DeepEqual(gotEcho, echo) {
		t.Errorf("body mismatch: got %+v, want %+v", gotEcho, echo)
	}
}

func TestIpv4Payload(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		want []byte
	}{
		{"short", []byte{1, 2, 3}, []byte{1, 2, 3}},
		{"len 19", make([]byte, 19), make([]byte, 19)},
		{"hdrlen 20", append(append([]byte{0x45}, make([]byte, 19)...), 1, 2, 3, 4), []byte{1, 2, 3, 4}},
		{"hdrlen 24", append(append([]byte{0x46}, make([]byte, 23)...), 5, 6, 7, 8, 9, 10), []byte{5, 6, 7, 8, 9, 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ipv4Payload(tt.b)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("ipv4Payload() = %v (len=%d), want %v (len=%d)", got, len(got), tt.want, len(tt.want))
			}
		})
	}
}

func TestIpv4PayloadHdrlen(t *testing.T) {
	hdrlen := 20
	payload := []byte("payload123")
	b := make([]byte, hdrlen+len(payload))
	b[0] = 0x45
	copy(b[hdrlen:], payload)
	got := ipv4Payload(b)
	if !bytes.Equal(got, payload) {
		t.Errorf("ipv4Payload() = %v, want %v", got, payload)
	}
}
