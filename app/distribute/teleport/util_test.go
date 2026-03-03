package teleport

import (
	"testing"
)

func TestMakeHash(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"", "0"},
		{"a", "e8b7be43"},
		{"hello", "3610a686"},
	}
	for _, tt := range tests {
		got := MakeHash(tt.s)
		if got != tt.want {
			t.Errorf("MakeHash(%q) = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		s string
	}{
		{""},
		{"x"},
		{"hello world"},
	}
	for _, tt := range tests {
		got := HashString(tt.s)
		if tt.s != "" && got == 0 {
			t.Errorf("HashString(%q) = 0", tt.s)
		}
	}
}

func TestMakeUnique(t *testing.T) {
	tests := []struct {
		obj interface{}
	}{
		{nil},
		{"s"},
		{map[string]int{"a": 1}},
	}
	for _, tt := range tests {
		got := MakeUnique(tt.obj)
		if got == "" {
			t.Errorf("MakeUnique(%v) = empty", tt.obj)
		}
	}
}

func TestMakeMd5(t *testing.T) {
	tests := []struct {
		obj    interface{}
		length int
	}{
		{"x", 8},
		{123, 16},
		{[]int{1, 2}, 32},
		{"y", 64},
	}
	for _, tt := range tests {
		got := MakeMd5(tt.obj, tt.length)
		wantLen := tt.length
		if wantLen > 32 {
			wantLen = 32
		}
		if len(got) != wantLen {
			t.Errorf("MakeMd5(%v, %d) len = %d, want %d", tt.obj, tt.length, len(got), wantLen)
		}
	}
}
