package logs

import (
	"bytes"
	"testing"
)

func TestLog(t *testing.T) {
	l := Log()
	if l == nil {
		t.Fatal("Log() returned nil")
	}
}

func TestSetOutput(t *testing.T) {
	l := Log()
	buf := &bytes.Buffer{}
	got := l.SetOutput(buf)
	if got != l {
		t.Errorf("SetOutput() = %v, want %v", got, l)
	}
}

func TestLogLevels(t *testing.T) {
	l := Log()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	tests := []struct {
		name string
		fn   func()
	}{
		{"Debug", func() { l.Debug("msg") }},
		{"Informational", func() { l.Informational("msg") }},
		{"App", func() { l.App("msg") }},
		{"Notice", func() { l.Notice("msg") }},
		{"Warning", func() { l.Warning("msg") }},
		{"Error", func() { l.Error("msg") }},
		{"Critical", func() { l.Critical("msg") }},
		{"Alert", func() { l.Alert("msg") }},
		{"Emergency", func() { l.Emergency("msg") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.fn()
		})
	}
}

func TestPauseOutputGoOn(t *testing.T) {
	l := Log()
	l.PauseOutput()
	l.GoOn()
}

func TestStatus(t *testing.T) {
	l := Log()
	_, s := l.Status()
	if s == "" {
		t.Error("Status() returned empty string")
	}
}
