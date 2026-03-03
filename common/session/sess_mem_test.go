// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package session

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMem(t *testing.T) {
	globalSessions, _ := NewManager("memory", `{"cookieName":"gosessionid","gclifetime":10}`)
	go globalSessions.GC()
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess := globalSessions.SessionStart(w, r).Unwrap()
	defer sess.SessionRelease(w)
	sess.Set("username", "astaxie")
	if username := sess.Get("username").UnwrapOr(nil); username != "astaxie" {
		t.Fatal("get username error")
	}
	if cookiestr := w.Header().Get("Set-Cookie"); cookiestr == "" {
		t.Fatal("setcookie error")
	} else {
		parts := strings.Split(strings.TrimSpace(cookiestr), ";")
		for k, v := range parts {
			nameval := strings.Split(v, "=")
			if k == 0 && nameval[0] != "gosessionid" {
				t.Fatal("error")
			}
		}
	}
}

func TestMemSessionStore(t *testing.T) {
	m, _ := NewManager("memory", `{"cookieName":"sid","gclifetime":3600}`)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	sess := m.SessionStart(w, r).Unwrap().(*MemSessionStore)

	sess.Set("a", 1)
	sess.Set("b", 2)
	if v := sess.Get("a").UnwrapOr(nil); v != 1 {
		t.Errorf("Get(a) = %v, want 1", v)
	}

	sess.Delete("a")
	if sess.Get("a").IsSome() {
		t.Error("Delete: key a should be gone")
	}

	sess.Flush()
	if sess.Get("b").IsSome() {
		t.Error("Flush: key b should be gone")
	}

	if sess.SessionID() == "" {
		t.Error("SessionID() should not be empty")
	}

	sess.SessionRelease(httptest.NewRecorder())
}

func TestMemProvider(t *testing.T) {
	m, _ := NewManager("memory", `{"cookieName":"sid","gclifetime":3600}`)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	sess1 := m.SessionStart(w, r).Unwrap()
	sid1 := sess1.SessionID()
	sess1.Set("x", "y")

	before := m.GetActiveSession()
	if before < 1 {
		t.Errorf("GetActiveSession() = %d, want >= 1", before)
	}

	sess2, err := mempder.SessionRead(sid1)
	if err != nil {
		t.Fatal(err)
	}
	if sess2.SessionID() != sid1 {
		t.Errorf("SessionRead SessionID = %q, want %q", sess2.SessionID(), sid1)
	}

	if !mempder.SessionExist(sid1) {
		t.Error("SessionExist(sid1) want true")
	}
	nonexistentSid := "sid-that-never-existed-" + sid1
	if mempder.SessionExist(nonexistentSid) {
		t.Error("SessionExist(nonexistent) want false")
	}

	reg, err := mempder.SessionRegenerate(sid1, "new-sid")
	if err != nil {
		t.Fatal(err)
	}
	if reg.SessionID() != "new-sid" {
		t.Errorf("SessionRegenerate SessionID = %q, want new-sid", reg.SessionID())
	}
	if v := reg.Get("x").UnwrapOr(nil); v != "y" {
		t.Errorf("SessionRegenerate Get(x) = %v, want y", v)
	}

	regNew, err := mempder.SessionRegenerate("never-existed", "another-sid")
	if err != nil {
		t.Fatal(err)
	}
	if regNew.SessionID() != "another-sid" {
		t.Errorf("SessionRegenerate(new) SessionID = %q", regNew.SessionID())
	}

	if err := mempder.SessionDestroy("new-sid"); err != nil {
		t.Error("SessionDestroy:", err)
	}
	if mempder.SessionExist("new-sid") {
		t.Error("SessionExist after destroy want false")
	}

	mempder.SessionDestroy("nonexistent")
}

func TestMemProviderGC(t *testing.T) {
	m, _ := NewManager("memory", `{"cookieName":"sid","gclifetime":0}`)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	m.SessionStart(w, r)
	mempder.SessionGC()
}
