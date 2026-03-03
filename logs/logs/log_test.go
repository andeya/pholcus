package logs

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		channellen  int64
		stealLevel  []int
		wantStealLv int
	}{
		{"default", 100, nil, LevelNothing},
		{"with steal", 100, []int{LevelError}, LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl := NewLogger(tt.channellen, tt.stealLevel...)
			if bl == nil {
				t.Fatal("NewLogger returned nil")
			}
			if tt.stealLevel != nil && bl.stealLevelPreset != tt.wantStealLv {
				t.Errorf("stealLevelPreset = %v, want %v", bl.stealLevelPreset, tt.wantStealLv)
			}
		})
	}
}

func TestSetLogger(t *testing.T) {
	bl := NewLogger(100)
	err := bl.SetLogger("console", map[string]interface{}{"level": LevelDebug})
	if err != nil {
		t.Fatalf("SetLogger: %v", err)
	}
	err = bl.SetLogger("unknown", nil)
	if err == nil {
		t.Error("SetLogger unknown adapter: want error")
	}
	err = bl.SetLogger("console", map[string]interface{}{"level": "invalid"})
	if err == nil {
		t.Error("SetLogger invalid level: want error")
	}
}

func TestDelLogger(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	err := bl.DelLogger("console")
	if err != nil {
		t.Errorf("DelLogger: %v", err)
	}
	err = bl.DelLogger("unknown")
	if err == nil {
		t.Error("DelLogger unknown: want error")
	}
}

func TestBeeLoggerLevels(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", map[string]interface{}{"level": LevelDebug})

	tests := []struct {
		name string
		fn   func()
	}{
		{"App", func() { bl.App("msg") }},
		{"Emergency", func() { bl.Emergency("msg") }},
		{"Alert", func() { bl.Alert("msg") }},
		{"Critical", func() { bl.Critical("msg") }},
		{"Error", func() { bl.Error("msg") }},
		{"Warning", func() { bl.Warning("msg") }},
		{"Notice", func() { bl.Notice("msg") }},
		{"Informational", func() { bl.Informational("msg") }},
		{"Debug", func() { bl.Debug("msg") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn()
		})
	}
}

func TestSetLevel(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	bl.SetLevel(LevelError)
	bl.Debug("should not appear")
}

func TestSetStealLevel(t *testing.T) {
	bl := NewLogger(100, LevelError)
	bl.SetLogger("console", nil)
	bl.SetStealLevel(LevelError)
}

func TestSetLogFuncCallDepth(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogFuncCallDepth(3)
	if bl.GetLogFuncCallDepth() != 3 {
		t.Errorf("GetLogFuncCallDepth = %v, want 3", bl.GetLogFuncCallDepth())
	}
}

func TestEnableFuncCallDepth(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	bl.EnableFuncCallDepth(true)
	bl.Debug("with depth")
}

func TestSyncWrite(t *testing.T) {
	bl := NewLogger(100)
	bl.Async(false)
	bl.SetLogger("console", map[string]interface{}{"level": LevelDebug})
	bl.Debug("sync")
}

func TestPauseOutputGoOn(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	bl.PauseOutput()
	code, s := bl.Status()
	if code != REST {
		t.Errorf("Status after Pause = %v, want REST", code)
	}
	_ = s
	bl.GoOn()
}

func TestEnableStealOne(t *testing.T) {
	bl := NewLogger(100, LevelError)
	bl.SetLogger("console", nil)
	bl.EnableStealOne(true)
	bl.EnableStealOne(false)
}

func TestStatus(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	code, s := bl.Status()
	if code != WORK {
		t.Errorf("Status = %v, want WORK", code)
	}
	if s != "WORK" {
		t.Errorf("Status string = %v, want WORK", s)
	}
}


func TestFlush(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	bl.Flush()
}

func TestClose(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	bl.Close()
	code, _ := bl.Status()
	if code != CLOSE {
		t.Errorf("Status after Close = %v, want CLOSE", code)
	}
}


func TestWriterMsgNonWork(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	bl.PauseOutput()
	bl.Error("during pause")
}

func TestWriterMsgWithFuncCallDepth(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	bl.EnableFuncCallDepth(true)
	bl.Debug("with depth")
}

func TestStealOne(t *testing.T) {
	bl := NewLogger(100, LevelError)
	bl.SetLogger("console", nil)
	bl.EnableStealOne(true)
	bl.Async(false)
	bl.Error("steal")
	level, msg, ok := bl.StealOne()
	if !ok {
		t.Error("StealOne() ok = false")
	}
	if level != LevelError {
		t.Errorf("StealOne level = %v, want LevelError", level)
	}
	if msg == "" {
		t.Error("StealOne msg empty")
	}
}

func TestStealOneAfterClose(t *testing.T) {
	bl := NewLogger(100, LevelError)
	bl.SetLogger("console", nil)
	bl.EnableStealOne(true)
	bl.Close()
	_, _, ok := bl.StealOne()
	if ok {
		t.Error("StealOne after close: want ok=false")
	}
}

func TestWriterMsgLevelFilter(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", map[string]interface{}{"level": LevelError})
	bl.SetLevel(LevelError)
	bl.Debug("filtered")
	bl.Informational("filtered")
	bl.Error("shown")
}

func TestWriterMsgWithWriter(t *testing.T) {
	bl := NewLogger(100)
	err := bl.SetLogger("console", map[string]interface{}{"writer": &bytes.Buffer{}, "level": LevelDebug})
	if err != nil {
		t.Fatalf("SetLogger: %v", err)
	}
	bl.Async(false)
	bl.Debug("to buffer")
}

func TestFileAdapter(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.log")
	bl := NewLogger(100)
	err := bl.SetLogger("file", map[string]interface{}{"filename": fpath})
	if err != nil {
		t.Fatalf("SetLogger file: %v", err)
	}
	bl.Async(false)
	bl.Debug("file debug")
	bl.Error("file error")
	bl.Flush()
	bl.DelLogger("file")
	data, _ := os.ReadFile(fpath)
	if len(data) == 0 {
		t.Error("file adapter wrote nothing")
	}
}

func TestFileAdapterAppendToExisting(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "exist.log")
	os.WriteFile(fpath, []byte("line1\nline2\n"), 0644)
	bl := NewLogger(100)
	err := bl.SetLogger("file", map[string]interface{}{"filename": fpath})
	if err != nil {
		t.Fatalf("SetLogger file: %v", err)
	}
	bl.Async(false)
	bl.Debug("append")
	bl.Flush()
	bl.DelLogger("file")
}

func TestFileAdapterInitErrors(t *testing.T) {
	bl := NewLogger(100)
	err := bl.SetLogger("file", nil)
	if err == nil {
		t.Error("SetLogger file nil config: want error")
	}
	err = bl.SetLogger("file", map[string]interface{}{})
	if err == nil {
		t.Error("SetLogger file empty filename: want error")
	}
	err = bl.SetLogger("file", map[string]interface{}{"filename": ""})
	if err == nil {
		t.Error("SetLogger file empty string filename: want error")
	}
}

func TestFileAdapterMaxsize(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "maxsize.log")
	bl := NewLogger(100)
	err := bl.SetLogger("file", map[string]interface{}{"filename": fpath, "maxsize": 100})
	if err != nil {
		t.Fatalf("SetLogger: %v", err)
	}
	bl.Async(false)
	for i := 0; i < 20; i++ {
		bl.Debug("padding line to trigger rotate %d", i)
	}
	bl.Flush()
	bl.DelLogger("file")
}

func TestConnAdapter(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Skip("cannot create listener:", err)
	}
	defer ln.Close()
	addr := ln.Addr().String()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	bl := NewLogger(100)
	err = bl.SetLogger("conn", map[string]interface{}{"net": "tcp", "addr": addr})
	if err != nil {
		t.Fatalf("SetLogger conn: %v", err)
	}
	bl.Async(false)
	bl.Informational("conn test")
	bl.Flush()
	bl.DelLogger("conn")
}

func TestConnAdapterReconnectOnMsg(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Skip("cannot create listener:", err)
	}
	defer ln.Close()
	addr := ln.Addr().String()
	go func() {
		for i := 0; i < 2; i++ {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	bl := NewLogger(100)
	err = bl.SetLogger("conn", map[string]interface{}{"net": "tcp", "addr": addr, "reconnectOnMsg": true})
	if err != nil {
		t.Fatalf("SetLogger conn: %v", err)
	}
	bl.Async(false)
	bl.Informational("reconnect test")
	bl.Flush()
	bl.DelLogger("conn")
}

func TestSmtpAdapterInit(t *testing.T) {
	bl := NewLogger(100)
	err := bl.SetLogger("smtp", map[string]interface{}{
		"Username": "test@test.com",
		"password": "pass",
		"Host":     "invalid:25",
		"sendTos":  []string{"a@b.com"},
	})
	if err != nil {
		t.Fatalf("SetLogger smtp: %v", err)
	}
	bl.Async(false)
	bl.Critical("smtp test")
	bl.Flush()
	bl.DelLogger("smtp")
}


func TestAsyncStartLogger(t *testing.T) {
	bl := NewLogger(100)
	bl.SetLogger("console", nil)
	bl.Async(true)
	bl.Debug("async")
	bl.Close()
}

func TestCloseWithPendingMessages(t *testing.T) {
	bl := NewLogger(1000)
	bl.SetLogger("console", map[string]interface{}{"writer": &bytes.Buffer{}})
	bl.Async(true)
	for i := 0; i < 5; i++ {
		bl.Debug("pending %d", i)
	}
	bl.Close()
}

func TestConsoleDestroyFlush(t *testing.T) {
	cw := NewConsole()
	cw.Init(map[string]interface{}{"writer": &bytes.Buffer{}})
	cw.WriteMsg("test", LevelDebug)
	cw.Flush()
	cw.Destroy()
}
