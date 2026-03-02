package distribute

import (
	"encoding/json"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/distribute/teleport"
	"github.com/andeya/pholcus/logs"
)

// SlaveApi creates the slave node API.
func SlaveApi(n Distributer) teleport.API {
	return teleport.API{
		"task": &slaveTaskHandle{n},
	}
}

// slaveTaskHandle receives tasks from the master and adds them to the task jar.
type slaveTaskHandle struct {
	Distributer
}

func (self *slaveTaskHandle) Process(receive *teleport.NetData) *teleport.NetData {
	t := &Task{}
	if r := result.RetVoid(json.Unmarshal([]byte(receive.Body.(string)), t)); r.IsErr() {
		logs.Log.Error("JSON decode failed: %v", receive.Body)
		return nil
	}
	self.Receive(t)
	return nil
}
