package collector

import (
	"testing"

	"github.com/andeya/gust/result"
)

func TestRegister(t *testing.T) {
	Register("_test_register", func(*Collector) result.VoidResult { return result.OkVoid() })
	if _, ok := DataOutput["_test_register"]; !ok {
		t.Error("_test_register not registered")
	}
}

func TestRegisterRefresher(t *testing.T) {
	called := false
	RegisterRefresher("_test_refresher", &testRefresher{fn: func() { called = true }})
	RefreshBackend("_test_refresher")
	if !called {
		t.Error("Refresh should have been called")
	}
}

func TestRefreshBackend_Unregistered(t *testing.T) {
	RefreshBackend("_nonexistent_type_")
}

type testRefresher struct {
	fn func()
}

func (t *testRefresher) Refresh() {
	if t.fn != nil {
		t.fn()
	}
}
