package distribute

import (
	"encoding/json"
	"testing"

	"github.com/andeya/pholcus/app/distribute/teleport"
)

func TestSlaveAPI(t *testing.T) {
	tj := NewTaskJar()
	api := SlaveAPI(tj)
	if api == nil {
		t.Fatal("SlaveAPI returned nil")
	}
	if _, ok := api["task"]; !ok {
		t.Error("API missing task handler")
	}
}

func TestSlaveTaskHandle_Process(t *testing.T) {
	tj := NewTaskJar()
	task := Task{ID: 2, Limit: 200, OutType: "csv"}
	body, _ := json.Marshal(task)
	handle := &slaveTaskHandle{tj}
	req := &teleport.NetData{From: "master", Body: string(body)}

	resp := handle.Process(req)
	if resp != nil {
		t.Errorf("Process returned %v, want nil", resp)
	}
	got := tj.Pull()
	if got.ID != task.ID || got.Limit != task.Limit {
		t.Errorf("got Task %+v, want %+v", got, task)
	}
}

func TestSlaveTaskHandle_Process_InvalidJSON(t *testing.T) {
	tj := NewTaskJar()
	handle := &slaveTaskHandle{tj}
	req := &teleport.NetData{From: "master", Body: "invalid json {"}

	resp := handle.Process(req)
	if resp != nil {
		t.Errorf("Process returned %v, want nil", resp)
	}
	if tj.Len() != 0 {
		t.Errorf("Len() = %d, want 0", tj.Len())
	}
}
