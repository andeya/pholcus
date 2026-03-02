package queue

import (
	"testing"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue(5)
	if q.PoolSize != 5 {
		t.Errorf("PoolSize = %d, want 5", q.PoolSize)
	}
	if cap(q.PoolChan) != 5 {
		t.Errorf("cap(PoolChan) = %d, want 5", cap(q.PoolChan))
	}
}

func TestPushAndPull(t *testing.T) {
	q := NewQueue(3)
	if ok := q.Push("a"); !ok {
		t.Error("Push should succeed on empty queue")
	}
	if ok := q.Push("b"); !ok {
		t.Error("Push should succeed when queue not full")
	}
	if ok := q.Push("c"); !ok {
		t.Error("Push should succeed on last slot")
	}
	if ok := q.Push("d"); ok {
		t.Error("Push should fail on full queue")
	}

	got := q.Pull()
	if got != "a" {
		t.Errorf("Pull() = %v, want %q", got, "a")
	}
	got = q.Pull()
	if got != "b" {
		t.Errorf("Pull() = %v, want %q", got, "b")
	}
}

func TestPushSlice(t *testing.T) {
	q := NewQueue(5)
	q.PushSlice([]interface{}{"x", "y", "z"})
	if len(q.PoolChan) != 3 {
		t.Errorf("len after PushSlice = %d, want 3", len(q.PoolChan))
	}
}

func TestInit(t *testing.T) {
	q := NewQueue(2)
	q.Push("a")
	q2 := q.Init(10)
	if q2 != q {
		t.Error("Init should return the same queue")
	}
	if q.PoolSize != 10 {
		t.Errorf("PoolSize after Init = %d, want 10", q.PoolSize)
	}
	if cap(q.PoolChan) != 10 {
		t.Errorf("cap(PoolChan) after Init = %d, want 10", cap(q.PoolChan))
	}
}

func TestExchange(t *testing.T) {
	q := NewQueue(3)
	q.Push("a")
	q.Push("b")

	add := q.Exchange(5)
	if add != 3 {
		t.Errorf("Exchange(5) with 2 items: add = %d, want 3", add)
	}
	if q.PoolSize != 5 {
		t.Errorf("PoolSize after Exchange = %d, want 5", q.PoolSize)
	}

	q2 := NewQueue(10)
	q2.Push("x")
	q2.Push("y")
	q2.Push("z")
	add2 := q2.Exchange(2)
	if add2 != 0 {
		t.Errorf("Exchange(2) with 3 items: add = %d, want 0", add2)
	}
}
