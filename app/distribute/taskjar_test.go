package distribute

import (
	"sync"
	"testing"
)

func TestNewTaskJar(t *testing.T) {
	tj := NewTaskJar()
	if tj == nil {
		t.Fatal("NewTaskJar() returned nil")
	}
	if tj.Tasks == nil {
		t.Fatal("Tasks channel is nil")
	}
	if cap(tj.Tasks) != 1024 {
		t.Errorf("cap(Tasks) = %d, want 1024", cap(tj.Tasks))
	}
}

func TestTaskJar_PushPull(t *testing.T) {
	tj := NewTaskJar()
	task := &Task{ID: 0, Limit: 10}
	tj.Push(task)
	if tj.Len() != 1 {
		t.Errorf("Len() = %d, want 1", tj.Len())
	}
	got := tj.Pull()
	if got != task {
		t.Errorf("Pull() = %p, want %p", got, task)
	}
	if tj.Len() != 0 {
		t.Errorf("Len() after Pull = %d, want 0", tj.Len())
	}
}

func TestTaskJar_SendReceive(t *testing.T) {
	tests := []struct {
		name      string
		clientNum int
	}{
		{"single", 1},
		{"multi", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tj := NewTaskJar()
			task := &Task{ID: 0, Limit: 5}
			tj.Receive(task)
			if tj.Len() != 1 {
				t.Errorf("Len() = %d, want 1", tj.Len())
			}
			got := tj.Send(tt.clientNum)
			if got.Limit != 5 {
				t.Errorf("Send() Limit = %d, want 5", got.Limit)
			}
		})
	}
}

func TestTaskJar_PushAssignsID(t *testing.T) {
	tj := NewTaskJar()
	t1 := &Task{Limit: 1}
	t2 := &Task{Limit: 2}
	tj.Push(t1)
	tj.Push(t2)
	got1 := tj.Pull()
	got2 := tj.Pull()
	if got1.ID != 0 {
		t.Errorf("first task ID = %d, want 0", got1.ID)
	}
	if got2.ID != 1 {
		t.Errorf("second task ID = %d, want 1", got2.ID)
	}
}

func TestTaskJar_CountNodes(t *testing.T) {
	tj := NewTaskJar()
	if got := tj.CountNodes(); got != 0 {
		t.Errorf("CountNodes() = %d, want 0", got)
	}
}

func TestTaskJar_Concurrent(t *testing.T) {
	tj := NewTaskJar()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tj.Push(&Task{ID: id, Limit: int64(id)})
		}(i)
	}
	wg.Wait()
	if tj.Len() != 10 {
		t.Errorf("Len() = %d, want 10", tj.Len())
	}
	for tj.Len() > 0 {
		_ = tj.Pull()
	}
}
