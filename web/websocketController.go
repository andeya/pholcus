package web

import (
	"sync"

	"github.com/andeya/pholcus/app"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/common/util"
	ws "github.com/andeya/pholcus/common/websocket"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/status"
)

// SocketController manages WebSocket connections and message channels.
type SocketController struct {
	connPool     map[string]*ws.Conn
	wchanPool    map[string]*Wchan
	connRWMutex  sync.RWMutex
	wchanRWMutex sync.RWMutex
}

func (self *SocketController) GetConn(sessID string) *ws.Conn {
	self.connRWMutex.RLock()
	defer self.connRWMutex.RUnlock()
	return self.connPool[sessID]
}

func (self *SocketController) GetWchan(sessID string) *Wchan {
	self.wchanRWMutex.RLock()
	defer self.wchanRWMutex.RUnlock()
	return self.wchanPool[sessID]
}

func (self *SocketController) Add(sessID string, conn *ws.Conn) {
	self.connRWMutex.Lock()
	self.wchanRWMutex.Lock()
	defer self.connRWMutex.Unlock()
	defer self.wchanRWMutex.Unlock()

	self.connPool[sessID] = conn
	self.wchanPool[sessID] = newWchan()
}

func (self *SocketController) Remove(sessID string, conn *ws.Conn) {
	self.connRWMutex.Lock()
	self.wchanRWMutex.Lock()
	defer self.connRWMutex.Unlock()
	defer self.wchanRWMutex.Unlock()

	if self.connPool[sessID] == nil {
		return
	}
	wc := self.wchanPool[sessID]
	close(wc.wchan)
	conn.Close()
	delete(self.connPool, sessID)
	delete(self.wchanPool, sessID)
}

func (self *SocketController) Write(sessID string, void map[string]interface{}, to ...int) {
	self.wchanRWMutex.RLock()
	defer self.wchanRWMutex.RUnlock()

	// When to is 1: send only to current connection; -1: send to all except current; 0 or empty: send to all.
	var t int = 0
	if len(to) > 0 {
		t = to[0]
	}

	void["mode"] = app.LogicApp.GetAppConf("mode").(int)

	switch t {
	case 1:
		wc := self.wchanPool[sessID]
		if wc == nil {
			return
		}
		void["initiative"] = true
		wc.wchan <- void

	case 0, -1:
		l := len(self.wchanPool)
		for _sessID, wc := range self.wchanPool {
			if t == -1 && _sessID == sessID {
				continue
			}
			_void := make(map[string]interface{}, l)
			for k, v := range void {
				_void[k] = v
			}
			if _sessID == sessID {
				_void["initiative"] = true
			} else {
				_void["initiative"] = false
			}
			wc.wchan <- _void
		}
	}
}

// Wchan is a channel for WebSocket message delivery.
type Wchan struct {
	wchan chan interface{}
}

func newWchan() *Wchan {
	return &Wchan{
		wchan: make(chan interface{}, 1024),
	}
}

var (
	wsApi = map[string]func(string, map[string]interface{}){}
	// Sc is the global SocketController for WebSocket API connections.
	Sc = &SocketController{
		connPool:  make(map[string]*ws.Conn),
		wchanPool: make(map[string]*Wchan),
	}
)

func wsHandle(conn *ws.Conn) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error("%v", p)
		}
	}()
	r := globalSessions.SessionStart(nil, conn.Request())
	if r.IsErr() {
		logs.Log.Error("session start: %v", r.UnwrapErr())
		return
	}
	sess := r.Unwrap()
	sessID := sess.SessionID()
	if Sc.GetConn(sessID) == nil {
		Sc.Add(sessID, conn)
	}

	defer Sc.Remove(sessID, conn)

	go func() {
		var err error
		for info := range Sc.GetWchan(sessID).wchan {
			if _, err = ws.JSON.Send(conn, info); err != nil {
				return
			}
		}
	}()

	for {
		var req map[string]interface{}

		if err := ws.JSON.Receive(conn, &req); err != nil {
			return
		}

		wsApi[util.Atoa(req["operate"]).UnwrapOr("")](sessID, req)
	}
}

func init() {
	wsApi["refresh"] = func(sessID string, req map[string]interface{}) {
		Sc.Write(sessID, tplData(app.LogicApp.GetAppConf("mode").(int)), 1)
	}

	wsApi["init"] = func(sessID string, req map[string]interface{}) {
		var mode = util.Atoi(req["mode"]).UnwrapOr(0)
		var port = util.Atoi(req["port"]).UnwrapOr(0)
		var master = util.Atoa(req["ip"]).UnwrapOr("") // master address without port
		currMode := app.LogicApp.GetAppConf("mode").(int)
		if currMode == status.UNSET {
			app.LogicApp.Init(mode, port, master, Lsc)
		} else {
			app.LogicApp = app.LogicApp.ReInit(mode, port, master)
		}

		if mode == status.CLIENT {
			go app.LogicApp.Run()
		}

		Sc.Write(sessID, tplData(mode))
	}

	wsApi["run"] = func(sessID string, req map[string]interface{}) {
		if app.LogicApp.GetAppConf("mode").(int) != status.CLIENT {
			setConf(req)
		}

		if app.LogicApp.GetAppConf("mode").(int) == status.OFFLINE {
			Sc.Write(sessID, map[string]interface{}{"operate": "run"})
		}

		go func() {
			app.LogicApp.Run()
			if app.LogicApp.GetAppConf("mode").(int) == status.OFFLINE {
				Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
			}
		}()
	}

	// Stop current task; only supported in standalone mode.
	wsApi["stop"] = func(sessID string, req map[string]interface{}) {
		if app.LogicApp.GetAppConf("mode").(int) != status.OFFLINE {
			Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
			return
		} else {
			app.LogicApp.Stop()
			Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
		}
	}

	// Pause and resume task; only supported in standalone mode.
	wsApi["pauseRecover"] = func(sessID string, req map[string]interface{}) {
		if app.LogicApp.GetAppConf("mode").(int) != status.OFFLINE {
			return
		}
		app.LogicApp.PauseRecover()
		Sc.Write(sessID, map[string]interface{}{"operate": "pauseRecover"})
	}

	// Exit current mode.
	wsApi["exit"] = func(sessID string, req map[string]interface{}) {
		app.LogicApp = app.LogicApp.ReInit(status.UNSET, 0, "")
		Sc.Write(sessID, map[string]interface{}{"operate": "exit"})
	}
}

func tplData(mode int) map[string]interface{} {
	var info = map[string]interface{}{"operate": "init", "mode": mode}

	switch mode {
	case status.OFFLINE:
		info["title"] = config.FULL_NAME + "                                                          【 运行模式 ->  单机 】"
	case status.SERVER:
		info["title"] = config.FULL_NAME + "                                                          【 运行模式 ->  服务端 】"
	case status.CLIENT:
		info["title"] = config.FULL_NAME + "                                                          【 运行模式 ->  客户端 】"
	}

	if mode == status.CLIENT {
		return info
	}

	info["spiders"] = map[string]interface{}{
		"menu": spiderMenu,
		"curr": func() interface{} {
			l := app.LogicApp.GetSpiderQueue().Len()
			if l == 0 {
				return 0
			}
			var curr = make(map[string]bool, l)
			for _, sp := range app.LogicApp.GetSpiderQueue().GetAll() {
				curr[sp.GetName()] = true
			}

			return curr
		}(),
	}

	info["OutType"] = map[string]interface{}{
		"menu": app.LogicApp.GetOutputLib(),
		"curr": app.LogicApp.GetAppConf("OutType"),
	}

	info["ThreadNum"] = map[string]int{
		"max":  999999,
		"min":  1,
		"curr": app.LogicApp.GetAppConf("ThreadNum").(int),
	}

	info["Pausetime"] = map[string][]int64{
		"menu": {0, 100, 300, 500, 1000, 3000, 5000, 10000, 15000, 20000, 30000, 60000},
		"curr": []int64{app.LogicApp.GetAppConf("Pausetime").(int64)},
	}

	info["ProxyMinute"] = map[string][]int64{
		"menu": {0, 1, 3, 5, 10, 15, 20, 30, 45, 60, 120, 180},
		"curr": []int64{app.LogicApp.GetAppConf("ProxyMinute").(int64)},
	}

	info["DockerCap"] = map[string]int{
		"min":  1,
		"max":  5000000,
		"curr": app.LogicApp.GetAppConf("DockerCap").(int),
	}

	if app.LogicApp.GetAppConf("Limit").(int64) == spider.LIMIT {
		info["Limit"] = 0
	} else {
		info["Limit"] = app.LogicApp.GetAppConf("Limit")
	}

	info["Keyins"] = app.LogicApp.GetAppConf("Keyins")

	info["SuccessInherit"] = app.LogicApp.GetAppConf("SuccessInherit")
	info["FailureInherit"] = app.LogicApp.GetAppConf("FailureInherit")

	info["status"] = app.LogicApp.Status()

	return info
}

func setConf(req map[string]interface{}) {
	if tn := util.Atoi(req["ThreadNum"]).UnwrapOr(0); tn == 0 {
		app.LogicApp.SetAppConf("ThreadNum", 1)
	} else {
		app.LogicApp.SetAppConf("ThreadNum", tn)
	}

	app.LogicApp.
		SetAppConf("Pausetime", int64(util.Atoi(req["Pausetime"]).UnwrapOr(0))).
		SetAppConf("ProxyMinute", int64(util.Atoi(req["ProxyMinute"]).UnwrapOr(0))).
		SetAppConf("OutType", util.Atoa(req["OutType"]).UnwrapOr("")).
		SetAppConf("DockerCap", util.Atoi(req["DockerCap"]).UnwrapOr(0)).
		SetAppConf("Limit", int64(util.Atoi(req["Limit"]).UnwrapOr(0))).
		SetAppConf("Keyins", util.Atoa(req["Keyins"]).UnwrapOr("")).
		SetAppConf("SuccessInherit", req["SuccessInherit"] == "true").
		SetAppConf("FailureInherit", req["FailureInherit"] == "true")

	setSpiderQueue(req)
}

func setSpiderQueue(req map[string]interface{}) {
	spNames, ok := req["spiders"].([]interface{})
	if !ok {
		return
	}
	spiders := []*spider.Spider{}
	for _, sp := range app.LogicApp.GetSpiderLib() {
		for _, spName := range spNames {
			if util.Atoa(spName).UnwrapOr("") == sp.GetName() {
				spiders = append(spiders, sp.Copy())
			}
		}
	}
	app.LogicApp.SpiderPrepare(spiders)
}
