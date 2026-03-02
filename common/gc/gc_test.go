package gc

import (
	"testing"
)

func TestGCSizeConstant(t *testing.T) {
	expected := 50 << 20
	if GC_SIZE != expected {
		t.Errorf("GC_SIZE = %d, want %d", GC_SIZE, expected)
	}
}

func TestManualGCDoesNotPanic(t *testing.T) {
	// ManualGC launches a background goroutine guarded by sync.Once.
	// Calling it multiple times must not panic.
	ManualGC()
	ManualGC()
}
