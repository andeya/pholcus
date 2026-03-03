package kafka

import (
	"errors"
	"sync"
	"testing"

	"github.com/Shopify/sarama"
)

type mockSyncProducer struct {
	sendErr error
	closed  bool
	mu      sync.Mutex
}

func (m *mockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	if m.sendErr != nil {
		return 0, 0, m.sendErr
	}
	return 0, 0, nil
}

func (m *mockSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	return nil
}

func (m *mockSyncProducer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
}

func TestSetTopic(t *testing.T) {
	s := New()
	s.SetTopic("test-topic")
	if s.topic != "test-topic" {
		t.Errorf("topic = %q, want test-topic", s.topic)
	}
}

func TestGetProducer_BeforeRefresh(t *testing.T) {
	r := GetProducer()
	if r.IsErr() {
		t.Errorf("GetProducer before Refresh should not return error when producer and err are nil")
	}
}

func TestRefresh(t *testing.T) {
	Refresh()
}

func TestGetProducer_AfterRefresh(t *testing.T) {
	Refresh()
	r := GetProducer()
	if r.IsErr() {
		if r.UnwrapErr() == nil {
			t.Error("IsErr but UnwrapErr is nil")
		}
		return
	}
	_ = r.Unwrap()
}

func TestPush_NilProducer(t *testing.T) {
	old := producer
	producer = nil
	defer func() { producer = old }()

	s := New()
	s.SetTopic("test")
	r := s.Push(map[string]interface{}{"a": 1})
	if r.IsOk() {
		t.Error("Push with nil producer should return error")
	}
	if r.UnwrapErr().Error() != "kafka producer not initialized" {
		t.Errorf("Push err = %v", r.UnwrapErr())
	}
}

func TestPush_Success(t *testing.T) {
	old := producer
	producer = &mockSyncProducer{}
	defer func() { producer = old }()

	s := New()
	s.SetTopic("test-topic")
	r := s.Push(map[string]interface{}{"key": "value"})
	if r.IsErr() {
		t.Errorf("Push err = %v", r.UnwrapErr())
	}
}

func TestPush_SendError(t *testing.T) {
	old := producer
	producer = &mockSyncProducer{sendErr: errors.New("send failed")}
	defer func() { producer = old }()

	s := New()
	s.SetTopic("test")
	r := s.Push(map[string]interface{}{"x": 1})
	if r.IsOk() {
		t.Error("Push with send error should return error")
	}
	if r.UnwrapErr().Error() != "send failed" {
		t.Errorf("Push err = %v", r.UnwrapErr())
	}
}

func TestPush_EmptyData(t *testing.T) {
	old := producer
	producer = &mockSyncProducer{}
	defer func() { producer = old }()

	s := New()
	s.SetTopic("t")
	r := s.Push(map[string]interface{}{})
	if r.IsErr() {
		t.Errorf("Push empty map err = %v", r.UnwrapErr())
	}
}
