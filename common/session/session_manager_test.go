package session

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestManagerGetSid(t *testing.T) {
	m, err := NewManager("memory", `{"cookieName":"sid","gclifetime":3600}`)
	if err != nil {
		t.Fatal(err)
	}
	mempder.SessionRead("abc123")
	r, _ := http.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "sid", Value: url.QueryEscape("abc123"), MaxAge: 3600})
	sess := m.SessionStart(httptest.NewRecorder(), r).Unwrap()
	if got := sess.SessionID(); got != "abc123" {
		t.Errorf("SessionID() = %q, want abc123", got)
	}
}

func TestManagerSessionStartExisting(t *testing.T) {
	m, err := NewManager("memory", `{"cookieName":"sid","gclifetime":3600}`)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	sess1 := m.SessionStart(w, r).Unwrap()
	sid1 := sess1.SessionID()
	sess1.Set("k", "v")

	r2, _ := http.NewRequest("GET", "/", nil)
	r2.AddCookie(&http.Cookie{Name: "sid", Value: url.QueryEscape(sid1), MaxAge: 3600})
	sess2 := m.SessionStart(httptest.NewRecorder(), r2).Unwrap()
	if sess2.SessionID() != sid1 {
		t.Errorf("SessionID = %q, want %q", sess2.SessionID(), sid1)
	}
	if v := sess2.Get("k").UnwrapOr(nil); v != "v" {
		t.Errorf("Get(k) = %v, want v", v)
	}
}

func TestManagerGetSessionStore(t *testing.T) {
	m, err := NewManager("memory", `{"cookieName":"sid","gclifetime":3600}`)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	sess := m.SessionStart(w, r).Unwrap()
	sid := sess.SessionID()

	got := m.GetSessionStore(sid).Unwrap()
	if got.SessionID() != sid {
		t.Errorf("GetSessionStore() = %q, want %q", got.SessionID(), sid)
	}

	gotErr := m.GetSessionStore("nonexistent")
	if gotErr.IsErr() {
		t.Error("GetSessionStore(nonexistent) should return store (memory creates on read)")
	}
}

func TestManagerSessionDestroy(t *testing.T) {
	m, err := NewManager("memory", `{"cookieName":"sid","enableSetCookie":true,"gclifetime":3600}`)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	sess := m.SessionStart(w, r).Unwrap()
	sess.Set("x", 1)

	r2, _ := http.NewRequest("GET", "/", nil)
	r2.AddCookie(&http.Cookie{Name: "sid", Value: url.QueryEscape(sess.SessionID()), MaxAge: 3600})
	w2 := httptest.NewRecorder()
	m.SessionDestroy(w2, r2)
}

func TestManagerSessionDestroyNoCookie(t *testing.T) {
	m, err := NewManager("memory", `{"cookieName":"sid","gclifetime":3600}`)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	m.SessionDestroy(w, r)
}

func TestManagerSessionRegenerateID(t *testing.T) {
	m, err := NewManager("memory", `{"cookieName":"sid","enableSetCookie":true,"gclifetime":3600}`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("no cookie", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		sess := m.SessionRegenerateID(w, r).Unwrap()
		if sess.SessionID() == "" {
			t.Error("SessionRegenerateID() returned empty sid")
		}
	})

	t.Run("with cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/", nil)
		orig := m.SessionStart(w, r).Unwrap()
		oldSid := orig.SessionID()
		orig.Set("data", "val")

		r2, _ := http.NewRequest("GET", "/", nil)
		r2.AddCookie(&http.Cookie{Name: "sid", Value: url.QueryEscape(oldSid), MaxAge: 3600})
		w2 := httptest.NewRecorder()
		newSess := m.SessionRegenerateID(w2, r2).Unwrap()
		if newSess.SessionID() == oldSid {
			t.Error("SessionRegenerateID() should produce new sid")
		}
		if v := newSess.Get("data").UnwrapOr(nil); v != "val" {
			t.Errorf("Get(data) = %v, want val", v)
		}
	})
}

func TestManagerGetActiveSession(t *testing.T) {
	m, err := NewManager("memory", `{"cookieName":"sid","gclifetime":3600}`)
	if err != nil {
		t.Fatal(err)
	}
	before := m.GetActiveSession()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	m.SessionStart(w, r)
	if n := m.GetActiveSession(); n <= before {
		t.Errorf("GetActiveSession() = %d, want > %d", n, before)
	}
}

func TestManagerSetSecure(t *testing.T) {
	m, err := NewManager("memory", `{"cookieName":"sid","gclifetime":3600,"secure":false}`)
	if err != nil {
		t.Fatal(err)
	}
	m.SetSecure(true)
}

func TestManagerIsSecure(t *testing.T) {
	m, err := NewManager("memory", `{"cookieName":"sid","gclifetime":3600}`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("secure false", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "http://example.com/", nil)
		if m.isSecure(r) {
			t.Error("isSecure() want false when config.Secure is false")
		}
	})

	t.Run("secure true with https", func(t *testing.T) {
		m.SetSecure(true)
		r, _ := http.NewRequest("GET", "https://example.com/", nil)
		if !m.isSecure(r) {
			t.Error("isSecure() want true for https URL")
		}
	})

	t.Run("secure true with http", func(t *testing.T) {
		m.SetSecure(true)
		r, _ := http.NewRequest("GET", "http://example.com/", nil)
		if m.isSecure(r) {
			t.Error("isSecure() want false for http URL")
		}
	})
}

func TestManagerEnableSidInUrlQuery(t *testing.T) {
	config := `{"cookieName":"sid","gclifetime":3600,"enableSidInUrlQuery":true,"ProviderConfig":"{\"securityKey\":\"key\",\"blockKey\":\"1234567890123456\"}"}`
	m, err := NewManager("cookie", config)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := http.NewRequest("GET", "/?sid=from-query", nil)
	sess := m.SessionStart(httptest.NewRecorder(), r).Unwrap()
	if got := sess.SessionID(); got != "from-query" {
		t.Errorf("SessionID = %q, want from-query", got)
	}
}
