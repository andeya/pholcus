package web

import (
	"testing"

	"github.com/andeya/pholcus/runtime/cache"
	"github.com/andeya/pholcus/runtime/status"
	ws "github.com/andeya/pholcus/common/websocket"
)

func init() {
	cache.Task.Mode = status.OFFLINE
}

func TestSocketController(t *testing.T) {
	sc := &SocketController{
		connPool:  make(map[string]*ws.Conn),
		wchanPool: make(map[string]*Wchan),
	}

	tests := []struct {
		name   string
		sessID string
	}{
		{"GetConn nil", "sess1"},
		{"GetWchan nil", "sess1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "GetConn nil" {
				if got := sc.GetConn(tt.sessID); got != nil {
					t.Errorf("GetConn() = %v, want nil", got)
				}
			}
			if tt.name == "GetWchan nil" {
				if got := sc.GetWchan(tt.sessID); got != nil {
					t.Errorf("GetWchan() = %v, want nil", got)
				}
			}
		})
	}
}

func TestNewWchan(t *testing.T) {
	wc := newWchan()
	if wc == nil || wc.wchan == nil {
		t.Error("newWchan() returned nil")
	}
}

func TestTplData(t *testing.T) {
	cache.Task.ThreadNum = 20
	cache.Task.Pausetime = 300
	cache.Task.ProxyMinute = 10
	cache.Task.BatchCap = 1000
	cache.Task.OutType = "csv"
	cache.Task.Limit = 0
	cache.Task.Keyins = ""
	cache.Task.SuccessInherit = true
	cache.Task.FailureInherit = true

	tests := []struct {
		name string
		mode int
	}{
		{"offline", status.OFFLINE},
		{"server", status.SERVER},
		{"client", status.CLIENT},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := tplData(tt.mode)
			if info["operate"] != "init" || info["mode"] != tt.mode {
				t.Errorf("tplData() = %v", info)
			}
		})
	}
}

func TestSetConf(t *testing.T) {
	cache.Task.Mode = status.OFFLINE

	tests := []struct {
		name string
		req  map[string]interface{}
	}{
		{"zero ThreadNum", map[string]interface{}{"ThreadNum": "0"}},
		{"with ThreadNum", map[string]interface{}{"ThreadNum": "10", "Pausetime": "100", "ProxyMinute": "5", "OutType": "csv", "BatchCap": "100", "Limit": "0", "Keyins": "", "SuccessInherit": "true", "FailureInherit": "false"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setConf(tt.req)
		})
	}
}

func TestSocketControllerWrite(t *testing.T) {
	WSController.Write("sess1", map[string]interface{}{"k": "v"})
	WSController.Write("sess1", map[string]interface{}{"k": "v"}, 1)
	WSController.Write("sess1", map[string]interface{}{"k": "v"}, -1)
}

func TestSetSpiderQueue(t *testing.T) {
	tests := []struct {
		name string
		req  map[string]interface{}
	}{
		{"no spiders key", map[string]interface{}{}},
		{"spiders not slice", map[string]interface{}{"spiders": "invalid"}},
		{"spiders empty", map[string]interface{}{"spiders": []interface{}{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setSpiderQueue(tt.req)
		})
	}
}
