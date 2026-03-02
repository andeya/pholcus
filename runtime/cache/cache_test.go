package cache

import (
	"testing"
)

func TestResetAndGetPageCount(t *testing.T) {
	ResetPageCount()
	if got := GetPageCount(0); got != 0 {
		t.Errorf("after reset, total = %d, want 0", got)
	}

	PageSuccCount()
	PageSuccCount()
	PageFailCount()

	if got := GetPageCount(1); got != 2 {
		t.Errorf("success count = %d, want 2", got)
	}
	if got := GetPageCount(-1); got != 1 {
		t.Errorf("failure count = %d, want 1", got)
	}
	if got := GetPageCount(0); got != 3 {
		t.Errorf("total count = %d, want 3", got)
	}

	ResetPageCount()
	if got := GetPageCount(0); got != 0 {
		t.Errorf("after second reset, total = %d, want 0", got)
	}
}

func TestExecInitAndWaitInit(t *testing.T) {
	ExecInit(42)
	done := make(chan struct{})
	go func() {
		WaitInit(42)
		close(done)
	}()
	<-done
}

func TestAppConfDefaults(t *testing.T) {
	if Task == nil {
		t.Fatal("Task should be initialized")
	}
	if Task.Mode != 0 {
		t.Errorf("default Mode = %d, want 0", Task.Mode)
	}
}

func TestReportChanInitialized(t *testing.T) {
	if ReportChan == nil {
		t.Fatal("ReportChan should be initialized by init()")
	}
}
