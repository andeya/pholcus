package session

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileProvider_InitAndRead(t *testing.T) {
	tmp := t.TempDir()
	fp := &FileProvider{}
	if err := fp.SessionInit(3600, tmp); err != nil {
		t.Fatal(err)
	}

	sid := "abcdef1234567890"
	store, err := fp.SessionRead(sid)
	if err != nil {
		t.Fatal(err)
	}
	if store.SessionID() != sid {
		t.Errorf("SessionID = %q, want %q", store.SessionID(), sid)
	}
}

func TestFileProvider_Exist(t *testing.T) {
	tmp := t.TempDir()
	fp := &FileProvider{}
	fp.SessionInit(3600, tmp)

	sid := "abcdef1234567890"
	if fp.SessionExist(sid) {
		t.Error("SessionExist should be false before read")
	}

	fp.SessionRead(sid)
	if !fp.SessionExist(sid) {
		t.Error("SessionExist should be true after read")
	}
}

func TestFileProvider_Destroy(t *testing.T) {
	tmp := t.TempDir()
	fp := &FileProvider{}
	fp.SessionInit(3600, tmp)

	sid := "abcdef1234567890"
	fp.SessionRead(sid)
	if !fp.SessionExist(sid) {
		t.Fatal("session should exist")
	}

	fp.SessionDestroy(sid)
	if fp.SessionExist(sid) {
		t.Error("session should not exist after destroy")
	}
}

func TestFileProvider_GC(t *testing.T) {
	tmp := t.TempDir()
	fp := &FileProvider{}
	fp.SessionInit(0, tmp)

	sid := "abcdef1234567890"
	fp.SessionRead(sid)

	sessionDir := filepath.Join(tmp, string(sid[0]), string(sid[1]))
	sessionFile := filepath.Join(sessionDir, sid)
	past := time.Now().Add(-2 * time.Hour)
	os.Chtimes(sessionFile, past, past)

	fp.SessionGC()

	if fp.SessionExist(sid) {
		t.Error("session should be GC'd")
	}
}

func TestFileProvider_SessionAll(t *testing.T) {
	tmp := t.TempDir()
	fp := &FileProvider{}
	fp.SessionInit(3600, tmp)

	if fp.SessionAll() != 0 {
		t.Errorf("SessionAll = %d, want 0", fp.SessionAll())
	}

	fp.SessionRead("abcdef1234567890")
	fp.SessionRead("bcdefg2345678901")
	if fp.SessionAll() != 2 {
		t.Errorf("SessionAll = %d, want 2", fp.SessionAll())
	}
}

func TestFileSessionStore_SetGetDelete(t *testing.T) {
	tmp := t.TempDir()
	fp := &FileProvider{}
	fp.SessionInit(3600, tmp)

	sid := "abcdef1234567890"
	store, _ := fp.SessionRead(sid)

	store.Set("key1", "value1")
	opt := store.Get("key1")
	if !opt.IsSome() || opt.Unwrap() != "value1" {
		t.Error("Get after Set failed")
	}

	store.Delete("key1")
	opt = store.Get("key1")
	if opt.IsSome() {
		t.Error("Get after Delete should be None")
	}
}

func TestFileSessionStore_Flush(t *testing.T) {
	tmp := t.TempDir()
	fp := &FileProvider{}
	fp.SessionInit(3600, tmp)

	store, _ := fp.SessionRead("abcdef1234567890")
	store.Set("a", 1)
	store.Set("b", 2)
	store.Flush()
	if store.Get("a").IsSome() {
		t.Error("Flush should clear all values")
	}
}

func TestFileProvider_Regenerate(t *testing.T) {
	tmp := t.TempDir()
	fp := &FileProvider{}
	fp.SessionInit(3600, tmp)

	oldsid := "abcdef1234567890"
	store, _ := fp.SessionRead(oldsid)
	store.Set("key", "val")
	w := httptest.NewRecorder()
	oldSavePath := filepder.savePath
	filepder.savePath = tmp
	filepder.maxlifetime = 3600
	defer func() { filepder.savePath = oldSavePath }()
	store.SessionRelease(w)

	newsid := "bcdefg2345678901"
	newStore, err := fp.SessionRegenerate(oldsid, newsid)
	if err != nil {
		t.Fatal(err)
	}
	if newStore.SessionID() != newsid {
		t.Errorf("SessionID = %q, want %q", newStore.SessionID(), newsid)
	}
}

func TestFileSessionStore_SessionRelease(t *testing.T) {
	tmp := t.TempDir()

	oldSavePath := filepder.savePath
	filepder.savePath = tmp
	filepder.maxlifetime = 3600
	defer func() { filepder.savePath = oldSavePath }()

	sid := "abcdef1234567890"
	fp := filepder
	store, _ := fp.SessionRead(sid)
	store.Set("user", "test")

	w := httptest.NewRecorder()
	store.SessionRelease(w)

	store2, _ := fp.SessionRead(sid)
	opt := store2.Get("user")
	if !opt.IsSome() || opt.Unwrap() != "test" {
		t.Error("SessionRelease should persist data")
	}
}
