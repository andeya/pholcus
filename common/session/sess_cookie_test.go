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

func TestCookie(t *testing.T) {
	config := `{"cookieName":"gosessionid","enableSetCookie":false,"gclifetime":3600,"ProviderConfig":"{\"cookieName\":\"gosessionid\",\"securityKey\":\"beegocookiehashkey\"}"}`
	globalSessions, err := NewManager("cookie", config)
	if err != nil {
		t.Fatal("init cookie session err", err)
	}
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess := globalSessions.SessionStart(w, r).Unwrap()
	sess.Set("username", "astaxie")
	if username := sess.Get("username").UnwrapOr(nil); username != "astaxie" {
		t.Fatal("get username error")
	}
	sess.SessionRelease(w)
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

func TestCookieSessionStore(t *testing.T) {
	config := `{"cookieName":"gosessionid","gclifetime":3600,"ProviderConfig":"{\"cookieName\":\"gosessionid\",\"securityKey\":\"key\",\"blockKey\":\"1234567890123456\"}"}`
	m, err := NewManager("cookie", config)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := http.NewRequest("GET", "/", nil)
	sess := m.SessionStart(httptest.NewRecorder(), r).Unwrap()

	sess.Set("k1", "v1")
	sess.Delete("k1")
	if sess.Get("k1").IsSome() {
		t.Error("Delete: k1 should be gone")
	}

	sess.Set("k2", "v2")
	sess.Flush()
	if sess.Get("k2").IsSome() {
		t.Error("Flush: k2 should be gone")
	}
}

func TestCookieProvider(t *testing.T) {
	config := `{"cookieName":"gosessionid","gclifetime":3600,"ProviderConfig":"{\"cookieName\":\"gosessionid\",\"securityKey\":\"key\",\"blockKey\":\"1234567890123456\"}"}`
	m, err := NewManager("cookie", config)
	if err != nil {
		t.Fatal(err)
	}
	_ = m
	cookiepder.SessionGC()
	if n := cookiepder.SessionAll(); n != 0 {
		t.Errorf("SessionAll() = %d, want 0", n)
	}
	_, err = cookiepder.SessionRegenerate("old", "new")
	if err != nil {
		t.Error("SessionRegenerate:", err)
	}
	cookiepder.SessionUpdate("sid")
}

func TestDestorySessionCookie(t *testing.T) {
	config := `{"cookieName":"gosessionid","enableSetCookie":true,"gclifetime":3600,"ProviderConfig":"{\"cookieName\":\"gosessionid\",\"securityKey\":\"beegocookiehashkey\"}"}`
	globalSessions, err := NewManager("cookie", config)
	if err != nil {
		t.Fatal("init cookie session err", err)
	}

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	session := globalSessions.SessionStart(w, r).Unwrap()

	// request again ,will get same sesssion id .
	r1, _ := http.NewRequest("GET", "/", nil)
	r1.Header.Set("Cookie", w.Header().Get("Set-Cookie"))
	w = httptest.NewRecorder()
	newSession := globalSessions.SessionStart(w, r1).Unwrap()
	if newSession.SessionID() != session.SessionID() {
		t.Fatal("get cookie session id is not the same again.")
	}

	// After destroy session , will get a new session id .
	globalSessions.SessionDestroy(w, r1)
	r2, _ := http.NewRequest("GET", "/", nil)
	r2.Header.Set("Cookie", w.Header().Get("Set-Cookie"))

	w = httptest.NewRecorder()
	newSession = globalSessions.SessionStart(w, r2).Unwrap()
	if newSession.SessionID() == session.SessionID() {
		t.Fatal("after destroy session and reqeust again ,get cookie session id is same.")
	}
}
