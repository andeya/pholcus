package pool

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type mockSrc struct {
	usable bool
	closed bool
	reset  bool
	mu     sync.Mutex
}

func (m *mockSrc) Usable() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.usable
}

func (m *mockSrc) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reset = true
}

func (m *mockSrc) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
}

func newMockSrc(usable bool) *mockSrc {
	return &mockSrc{usable: usable}
}

func TestClassicPool_Creation(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		maxIdle  int
		gctime   time.Duration
	}{
		{"default_gctime", 10, 5, 0},
		{"custom_gctime", 10, 5, 100 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var factory Factory
			if tt.gctime == 0 {
				factory = func() (Src, error) {
					return newMockSrc(true), nil
				}
				p := ClassicPool(tt.capacity, tt.maxIdle, factory)
				if p == nil {
					t.Fatal("pool should not be nil")
				}
				if p.Len() != 0 {
					t.Errorf("Len() = %d, want 0", p.Len())
				}
			} else {
				factory = func() (Src, error) {
					return newMockSrc(true), nil
				}
				p := ClassicPool(tt.capacity, tt.maxIdle, factory, tt.gctime)
				if p == nil {
					t.Fatal("pool should not be nil")
				}
				_ = p.Call(func(src Src) error { return nil })
				p.Close()
			}
		})
	}
}

func TestClassicPool_Call(t *testing.T) {
	cbErr := errors.New("callback error")
	factoryErr := errors.New("factory error")
	tests := []struct {
		name      string
		setup     func() (Pool, func())
		callback  func(Src) error
		wantErr   bool
		wantIsErr error
	}{
		{
			name: "normal",
			setup: func() (Pool, func()) {
				f := func() (Src, error) { return newMockSrc(true), nil }
				p := ClassicPool(2, 1, f, 10*time.Second)
				return p, func() { p.Close() }
			},
			callback: func(src Src) error {
				if src == nil {
					t.Error("src should not be nil")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "callback_error",
			setup: func() (Pool, func()) {
				f := func() (Src, error) { return newMockSrc(true), nil }
				p := ClassicPool(2, 1, f, 10*time.Second)
				return p, func() { p.Close() }
			},
			callback:  func(src Src) error { return cbErr },
			wantErr:   true,
			wantIsErr: cbErr,
		},
		{
			name: "after_close",
			setup: func() (Pool, func()) {
				f := func() (Src, error) { return newMockSrc(true), nil }
				p := ClassicPool(2, 1, f, 10*time.Second)
				_ = p.Call(func(src Src) error { return nil })
				p.Close()
				return p, func() {}
			},
			callback:  func(src Src) error { return nil },
			wantErr:   true,
			wantIsErr: closedError,
		},
		{
			name: "factory_error",
			setup: func() (Pool, func()) {
				f := func() (Src, error) { return nil, factoryErr }
				p := ClassicPool(2, 1, f, 10*time.Second)
				return p, func() { p.Close() }
			},
			callback:  func(src Src) error { return nil },
			wantErr:   true,
			wantIsErr: factoryErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, cleanup := tt.setup()
			defer cleanup()
			r := p.Call(tt.callback)
			if tt.wantErr {
				if !r.IsErr() {
					t.Fatal("expected error")
				}
				if tt.wantIsErr != nil && !errors.Is(r.UnwrapErr(), tt.wantIsErr) {
					t.Errorf("got err %v, want %v", r.UnwrapErr(), tt.wantIsErr)
				}
			} else {
				if r.IsErr() {
					t.Errorf("unexpected err: %v", r.UnwrapErr())
				}
				if tt.name == "normal" && p.Len() != 1 {
					t.Errorf("Len() = %d, want 1", p.Len())
				}
			}
		})
	}
}

func TestClassicPool_Len(t *testing.T) {
	factory := func() (Src, error) { return newMockSrc(true), nil }
	p := ClassicPool(3, 2, factory, 10*time.Second)
	defer p.Close()

	tests := []struct {
		name   string
		action func()
		want   int
	}{
		{"initial", func() {}, 0},
		{"after_1_call", func() { _ = p.Call(func(src Src) error { return nil }) }, 1},
		{"after_2_call_reuse", func() { _ = p.Call(func(src Src) error { return nil }) }, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.action()
			if got := p.Len(); got != tt.want {
				t.Errorf("Len() = %d, want %d", got, tt.want)
			}
		})
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		_ = p.Call(func(src Src) error { time.Sleep(10 * time.Millisecond); return nil })
		wg.Done()
	}()
	go func() {
		_ = p.Call(func(src Src) error { time.Sleep(10 * time.Millisecond); return nil })
		wg.Done()
	}()
	wg.Wait()
	if got := p.Len(); got < 2 {
		t.Errorf("after 2 concurrent Call Len() = %d, want >= 2", got)
	}
}

func TestClassicPool_UsableFalse_Retry(t *testing.T) {
	callCount := 0
	factory := func() (Src, error) {
		callCount++
		if callCount == 1 {
			return newMockSrc(false), nil
		}
		return newMockSrc(true), nil
	}
	p := ClassicPool(2, 1, factory, 10*time.Second)
	defer p.Close()

	p.Call(func(src Src) error {
		return nil
	})

	if callCount < 2 {
		t.Errorf("factory should be called at least 2 times when first src is unusable, got %d", callCount)
	}
}

func TestClassicPool_Call_PanicRecovery(t *testing.T) {
	factory := func() (Src, error) {
		return newMockSrc(true), nil
	}
	p := ClassicPool(2, 1, factory, 10*time.Second)
	defer p.Close()

	r := p.Call(func(src Src) error {
		panic("test panic")
	})
	if !r.IsErr() {
		t.Fatal("Call should return error after panic")
	}
}

func TestClassicPool_ConcurrentCalls(t *testing.T) {
	factory := func() (Src, error) {
		return newMockSrc(true), nil
	}
	p := ClassicPool(10, 5, factory, 10*time.Second)
	defer p.Close()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := p.Call(func(src Src) error {
				time.Sleep(time.Millisecond)
				return nil
			})
			if r.IsErr() {
				t.Errorf("concurrent Call failed: %v", r.UnwrapErr())
			}
		}()
	}
	wg.Wait()
}
