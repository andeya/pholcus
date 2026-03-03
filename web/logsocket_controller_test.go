package web

import (
	"testing"
)

func TestLogSocketControllerWrite(t *testing.T) {
	lsc := &LogSocketController{}
	n, err := lsc.Write([]byte("test log"))
	if err != nil {
		t.Errorf("Write() err = %v", err)
	}
	if n != 8 {
		t.Errorf("Write() n = %v, want 8", n)
	}
}

func TestLogSocketControllerAddRemove(t *testing.T) {
	lsc := &LogSocketController{}
	lsc.Add("sess1", nil)
	lsc.Remove("sess1")
}
