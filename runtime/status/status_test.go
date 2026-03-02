package status

import (
	"testing"
)

func TestModeConstants(t *testing.T) {
	if UNSET != -1 {
		t.Errorf("UNSET = %d, want -1", UNSET)
	}
	if OFFLINE != 0 {
		t.Errorf("OFFLINE = %d, want 0", OFFLINE)
	}
	if SERVER != 1 {
		t.Errorf("SERVER = %d, want 1", SERVER)
	}
	if CLIENT != 2 {
		t.Errorf("CLIENT = %d, want 2", CLIENT)
	}
}

func TestHeaderConstants(t *testing.T) {
	if REQTASK != 1 {
		t.Errorf("REQTASK = %d, want 1", REQTASK)
	}
	if TASK != 2 {
		t.Errorf("TASK = %d, want 2", TASK)
	}
	if LOG != 3 {
		t.Errorf("LOG = %d, want 3", LOG)
	}
}

func TestStatusConstants(t *testing.T) {
	if STOPPED != -1 {
		t.Errorf("STOPPED = %d, want -1", STOPPED)
	}
	if STOP != 0 {
		t.Errorf("STOP = %d, want 0", STOP)
	}
	if RUN != 1 {
		t.Errorf("RUN = %d, want 1", RUN)
	}
	if PAUSE != 2 {
		t.Errorf("PAUSE = %d, want 2", PAUSE)
	}
}
