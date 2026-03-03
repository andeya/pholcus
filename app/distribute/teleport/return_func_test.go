package teleport

import (
	"testing"
)

func TestReturnData(t *testing.T) {
	tests := []struct {
		body                     string
		args                     []string
		wantOp, wantTo, wantFrom string
	}{
		{"ok", nil, "", "", ""},
		{"x", []string{"op1"}, "op1", "", ""},
		{"y", []string{"op2", "to2"}, "op2", "to2", ""},
		{"z", []string{"op3", "to3", "from3"}, "op3", "to3", "from3"},
	}
	for _, tt := range tests {
		t.Run(tt.body, func(t *testing.T) {
			var d *NetData
			if len(tt.args) == 0 {
				d = ReturnData(tt.body)
			} else {
				d = ReturnData(tt.body, tt.args...)
			}
			if d == nil {
				t.Fatal("ReturnData returned nil")
			}
			if d.Status != SUCCESS {
				t.Errorf("Status = %d, want SUCCESS", d.Status)
			}
			if d.Body != tt.body {
				t.Errorf("Body = %v, want %v", d.Body, tt.body)
			}
			if d.Operation != tt.wantOp {
				t.Errorf("Operation = %q, want %q", d.Operation, tt.wantOp)
			}
			if d.To != tt.wantTo {
				t.Errorf("To = %q, want %q", d.To, tt.wantTo)
			}
			if d.From != tt.wantFrom {
				t.Errorf("From = %q, want %q", d.From, tt.wantFrom)
			}
		})
	}
}

func TestReturnError(t *testing.T) {
	req := &NetData{From: "a", To: "b", Operation: "task", Body: "orig"}
	resp := ReturnError(req, FAILURE, "err msg", "target")
	if resp != req {
		t.Error("ReturnError should return same pointer")
	}
	if req.Status != FAILURE {
		t.Errorf("Status = %d, want FAILURE", req.Status)
	}
	if req.Body != "err msg" {
		t.Errorf("Body = %q, want err msg", req.Body)
	}
	if req.From != "" {
		t.Errorf("From = %q, want empty", req.From)
	}
	if req.To != "target" {
		t.Errorf("To = %q, want target", req.To)
	}
}

func TestReturnError_NoNodeUID(t *testing.T) {
	req := &NetData{To: "x"}
	ReturnError(req, LLLEGAL, "bad")
	if req.To != "" {
		t.Errorf("To = %q, want empty", req.To)
	}
}
