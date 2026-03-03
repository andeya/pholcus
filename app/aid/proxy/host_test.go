package proxy

import (
	"testing"
	"time"
)

func TestProxyForHost_Len(t *testing.T) {
	tests := []struct {
		proxys []string
		want   int
	}{
		{nil, 0},
		{[]string{}, 0},
		{[]string{"a"}, 1},
		{[]string{"a", "b", "c"}, 3},
	}
	for _, tt := range tests {
		ph := &ProxyForHost{proxys: tt.proxys}
		if got := ph.Len(); got != tt.want {
			t.Errorf("Len() = %v, want %v", got, tt.want)
		}
	}
}

func TestProxyForHost_Less(t *testing.T) {
	ph := &ProxyForHost{
		proxys:    []string{"a", "b", "c"},
		timedelay: []time.Duration{10 * time.Millisecond, 5 * time.Millisecond, 20 * time.Millisecond},
	}
	tests := []struct {
		i, j int
		want bool
	}{
		{0, 1, false},
		{1, 0, true},
		{1, 2, true},
		{2, 1, false},
	}
	for _, tt := range tests {
		if got := ph.Less(tt.i, tt.j); got != tt.want {
			t.Errorf("Less(%d,%d) = %v, want %v", tt.i, tt.j, got, tt.want)
		}
	}
}

func TestProxyForHost_Swap(t *testing.T) {
	ph := &ProxyForHost{
		proxys:    []string{"a", "b"},
		timedelay: []time.Duration{10 * time.Millisecond, 5 * time.Millisecond},
	}
	ph.Swap(0, 1)
	if ph.proxys[0] != "b" || ph.proxys[1] != "a" {
		t.Errorf("Swap proxys = %v", ph.proxys)
	}
	if ph.timedelay[0] != 5*time.Millisecond || ph.timedelay[1] != 10*time.Millisecond {
		t.Errorf("Swap timedelay = %v", ph.timedelay)
	}
}
