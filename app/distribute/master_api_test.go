package distribute

import (
	"encoding/json"
	"testing"

	"github.com/andeya/pholcus/app/distribute/teleport"
)

type mockDistributor struct {
	sendCount   int
	countNodes  int
	sendTask    Task
	receiveTask *Task
}

func (m *mockDistributor) Send(clientNum int) Task {
	m.sendCount++
	return m.sendTask
}

func (m *mockDistributor) Receive(task *Task) {
	m.receiveTask = task
}

func (m *mockDistributor) CountNodes() int {
	return m.countNodes
}

func TestMasterAPI(t *testing.T) {
	d := &mockDistributor{countNodes: 2, sendTask: Task{ID: 1, Limit: 100}}
	api := MasterAPI(d)
	if api == nil {
		t.Fatal("MasterAPI returned nil")
	}
	if _, ok := api["task"]; !ok {
		t.Error("API missing task handler")
	}
	if _, ok := api["log"]; !ok {
		t.Error("API missing log handler")
	}
}

func TestMasterTaskHandle_Process(t *testing.T) {
	task := Task{ID: 1, Limit: 50, OutType: "mgo"}
	d := &mockDistributor{countNodes: 1, sendTask: task}
	handle := &masterTaskHandle{d}
	req := &teleport.NetData{From: "client1", To: "server", Operation: "task", Body: ""}

	resp := handle.Process(req)
	if resp == nil {
		t.Fatal("Process returned nil")
	}
	if resp.Status != teleport.SUCCESS {
		t.Errorf("Status = %d, want SUCCESS", resp.Status)
	}
	bodyStr, ok := resp.Body.(string)
	if !ok {
		t.Fatalf("Body type = %T, want string", resp.Body)
	}
	var got Task
	if err := json.Unmarshal([]byte(bodyStr), &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if got.ID != task.ID || got.Limit != task.Limit {
		t.Errorf("got Task %+v, want %+v", got, task)
	}
}

func TestMasterLogHandle_Process(t *testing.T) {
	handle := &masterLogHandle{}
	req := &teleport.NetData{From: "slave1", Body: "test log message"}
	resp := handle.Process(req)
	if resp != nil {
		t.Errorf("Process returned %v, want nil", resp)
	}
}
