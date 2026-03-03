package teleport

import (
	"testing"
)

func TestNewNetData(t *testing.T) {
	tests := []struct {
		from, to, op, flag string
		body               interface{}
	}{
		{"a", "b", "task", "", "body"},
		{"", "", "heartbeat", "f", nil},
	}
	for i, tt := range tests {
		t.Run("", func(t *testing.T) {
			d := NewNetData(tt.from, tt.to, tt.op, tt.flag, tt.body)
			if d == nil {
				t.Fatal("NewNetData returned nil")
			}
			if d.From != tt.from {
				t.Errorf("From = %q, want %q", d.From, tt.from)
			}
			if d.To != tt.to {
				t.Errorf("To = %q, want %q", d.To, tt.to)
			}
			if d.Operation != tt.op {
				t.Errorf("Operation = %q, want %q", d.Operation, tt.op)
			}
			if d.Flag != tt.flag {
				t.Errorf("Flag = %q, want %q", d.Flag, tt.flag)
			}
			if d.Status != SUCCESS {
				t.Errorf("Status = %d, want SUCCESS", d.Status)
			}
			_ = i
		})
	}
}
