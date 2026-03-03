package distribute

import (
	"testing"
)

func TestTask_Fields(t *testing.T) {
	tests := []struct {
		name       string
		task       Task
		wantID     int
		wantLimit  int64
		wantOutType string
	}{
		{"zero", Task{}, 0, 0, ""},
		{"with_values", Task{
			ID:        1,
			Limit:     100,
			OutType:   "mgo",
			ThreadNum: 10,
		}, 1, 100, "mgo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.task.ID != tt.wantID {
				t.Errorf("ID = %d, want %d", tt.task.ID, tt.wantID)
			}
			if tt.task.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", tt.task.Limit, tt.wantLimit)
			}
			if tt.task.OutType != tt.wantOutType {
				t.Errorf("OutType = %q, want %q", tt.task.OutType, tt.wantOutType)
			}
		})
	}
}
