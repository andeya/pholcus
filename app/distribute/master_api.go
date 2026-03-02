package distribute

import (
	"encoding/json"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/distribute/teleport"
	"github.com/andeya/pholcus/logs"
)

// MasterAPI creates the master node API.
func MasterAPI(n Distributor) teleport.API {
	return teleport.API{
		"task": &masterTaskHandle{n},
		"log":  &masterLogHandle{},
	}
}

// masterTaskHandle assigns tasks to clients.
type masterTaskHandle struct {
	Distributor
}

func (mth *masterTaskHandle) Process(receive *teleport.NetData) *teleport.NetData {
	b := result.Ret(json.Marshal(mth.Send(mth.CountNodes())))
	if b.IsErr() {
		return teleport.ReturnError(receive, teleport.FAILURE, "marshal error: "+b.UnwrapErr().Error(), receive.From)
	}
	return teleport.ReturnData(string(b.Unwrap()))
}

// masterLogHandle receives and prints log messages from slave nodes.
type masterLogHandle struct{}

func (*masterLogHandle) Process(receive *teleport.NetData) *teleport.NetData {
	logs.Log().Informational(" * ")
	logs.Log().Informational(" *     [ %s ]    %s", receive.From, receive.Body)
	logs.Log().Informational(" * ")
	return nil
}
