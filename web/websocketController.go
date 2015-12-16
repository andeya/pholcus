package web

import (
	"sync"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/status"
	ws "github.com/henrylee2cn/websocket.google"
)

type SocketController struct {
	connPool  map[string]*ws.Conn
	wchanPool map[string]*Wchan
	rwMutex   sync.RWMutex
}

func (self *SocketController) Add(sessID string, conn *ws.Conn) {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	self.connPool[sessID] = conn
	self.wchanPool[sessID] = newWchan()
}

func (self *SocketController) Remove(sessID string, conn *ws.Conn) {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

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
	self.rwMutex.RLock()
	defer self.rwMutex.RUnlock()

	// to为1时，只向当前连接发送；to为-1时，向除当前连接外的其他所有连接发送；to为0时或为空时，向所有连接发送
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

type Wchan struct {
	wchan chan interface{}
}

func newWchan() *Wchan {
	return &Wchan{
		wchan: make(chan interface{}, 1024),
	}
}

var (
	Sc = &SocketController{
		connPool:  make(map[string]*ws.Conn),
		wchanPool: make(map[string]*Wchan),
	}

	wsApi = map[string]func(string, map[string]interface{}){}
)

func wsHandle(conn *ws.Conn) {
	sess, _ := globalSessions.SessionStart(nil, conn.Request())
	sessID := sess.SessionID()
	if Sc.connPool[sessID] == nil {
		Sc.Add(sessID, conn)
	}

	defer Sc.Remove(sessID, conn)

	go func() {
		var err error
		for info := range Sc.wchanPool[sessID].wchan {
			if _, err = ws.JSON.Send(conn, info); err != nil {
				return
			}
		}
	}()

	for {
		var req map[string]interface{}

		if err := ws.JSON.Receive(conn, &req); err != nil {
			// logs.Log.Debug("websocket接收出错断开 (%v) !", err)
			return
		}

		// log.Log.Debug("Received from web: %v", req)
		wsApi[util.Atoa(req["operate"])](sessID, req)
	}
}

func init() {

	// 初始化运行
	wsApi["refresh"] = func(sessID string, req map[string]interface{}) {
		// 写入发送通道
		Sc.Write(sessID, tplData(app.LogicApp.GetAppConf("mode").(int)), 1)
	}

	// 初始化运行
	wsApi["init"] = func(sessID string, req map[string]interface{}) {
		var mode = util.Atoi(req["mode"])
		var port = util.Atoi(req["port"])
		var master = util.Atoa(req["ip"]) //服务器(主节点)地址，不含端口
		currMode := app.LogicApp.GetAppConf("mode").(int)
		if currMode == status.UNSET {
			app.LogicApp.Init(mode, port, master, Lsc) // 运行模式初始化，设置log输出目标
		} else {
			app.LogicApp = app.LogicApp.ReInit(mode, port, master) // 切换运行模式
		}

		if mode == status.CLIENT {
			go app.LogicApp.Run()
		}

		// 写入发送通道
		Sc.Write(sessID, tplData(mode))
	}

	wsApi["run"] = func(sessID string, req map[string]interface{}) {
		if app.LogicApp.GetAppConf("mode").(int) != status.CLIENT {
			if !setConf(req) {
				return
			}
		}

		if app.LogicApp.GetAppConf("mode").(int) == status.OFFLINE {
			Sc.Write(sessID, map[string]interface{}{"operate": "run", "status": status.RUN})
		}

		go func() {
			app.LogicApp.Run()
			if app.LogicApp.GetAppConf("mode").(int) == status.OFFLINE {
				Sc.Write(sessID, map[string]interface{}{"operate": "stop", "status": status.STOP})
			}
		}()
	}

	// 终止当前任务，现仅支持单机模式
	wsApi["stop"] = func(sessID string, req map[string]interface{}) {
		if app.LogicApp.GetAppConf("mode").(int) != status.OFFLINE {
			Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
			return
		} else {
			app.LogicApp.Stop()
			Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
		}
	}

	// 任务暂停与恢复，目前仅支持单机模式
	wsApi["pauseRecover"] = func(sessID string, req map[string]interface{}) {
		if app.LogicApp.GetAppConf("mode").(int) != status.OFFLINE {
			return
		}
		app.LogicApp.PauseRecover()
		Sc.Write(sessID, map[string]interface{}{"operate": "pauseRecover"})
	}

	// 退出当前模式
	wsApi["exit"] = func(sessID string, req map[string]interface{}) {
		app.LogicApp = app.LogicApp.ReInit(status.UNSET, 0, "")
		Sc.Write(sessID, map[string]interface{}{"operate": "exit"})
	}
}

func tplData(mode int) map[string]interface{} {
	var info = map[string]interface{}{"operate": "init", "mode": mode}

	// 运行模式标题
	switch mode {
	case status.OFFLINE:
		info["title"] = config.APP_FULL_NAME + "                                                          【 运行模式 ->  单机 】"
	case status.SERVER:
		info["title"] = config.APP_FULL_NAME + "                                                          【 运行模式 ->  服务端 】"
	case status.CLIENT:
		info["title"] = config.APP_FULL_NAME + "                                                          【 运行模式 ->  客户端 】"
	}

	if mode == status.CLIENT {
		return info
	}

	// 蜘蛛家族清单
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

	// 输出方式清单
	info["OutType"] = map[string]interface{}{
		"menu": app.LogicApp.GetOutputLib(),
		"curr": app.LogicApp.GetAppConf("OutType"),
	}

	// 并发协程上限
	info["ThreadNum"] = map[string]int{
		"max":  999999,
		"min":  1,
		"curr": app.LogicApp.GetAppConf("ThreadNum").(int),
	}

	// 暂停区间/ms(随机: Pausetime/2 ~ Pausetime*2)
	info["Pausetime"] = map[string][]int64{
		"menu": {0, 100, 300, 500, 1000, 3000, 5000, 10000, 15000, 20000, 30000, 60000},
		"curr": []int64{app.LogicApp.GetAppConf("Pausetime").(int64)},
	}

	// 代理IP更换的间隔分钟数
	info["ProxyMinute"] = map[string][]int64{
		"menu": {0, 1, 3, 5, 10, 15, 20, 30, 45, 60, 120, 180},
		"curr": []int64{app.LogicApp.GetAppConf("ProxyMinute").(int64)},
	}

	// 分批输出的容量
	info["DockerCap"] = map[string]int{
		"min":  1,
		"max":  5000000,
		"curr": app.LogicApp.GetAppConf("DockerCap").(int),
	}

	// 最大页数
	if app.LogicApp.GetAppConf("MaxPage").(int64) != spider.MAXPAGE {
		info["MaxPage"] = app.LogicApp.GetAppConf("MaxPage")
	} else {
		info["MaxPage"] = 0
	}

	// 关键词
	info["Keywords"] = app.LogicApp.GetAppConf("Keywords")

	// 继承历史记录
	info["SuccessInherit"] = app.LogicApp.GetAppConf("SuccessInherit")
	info["FailureInherit"] = app.LogicApp.GetAppConf("FailureInherit")

	// 运行状态
	info["status"] = app.LogicApp.Status()

	return info
}

// 配置运行参数
func setConf(req map[string]interface{}) bool {
	if tn := util.Atoi(req["ThreadNum"]); tn == 0 {
		app.LogicApp.SetAppConf("ThreadNum", 1)
	} else {
		app.LogicApp.SetAppConf("ThreadNum", tn)
	}

	app.LogicApp.
		SetAppConf("Pausetime", int64(util.Atoi(req["Pausetime"]))).
		SetAppConf("ProxyMinute", int64(util.Atoi(req["ProxyMinute"]))).
		SetAppConf("OutType", util.Atoa(req["OutType"])).
		SetAppConf("DockerCap", util.Atoi(req["DockerCap"])).
		SetAppConf("MaxPage", int64(util.Atoi(req["MaxPage"]))).
		SetAppConf("Keywords", util.Atoa(req["Keywords"])).
		SetAppConf("SuccessInherit", req["SuccessInherit"] == "true").
		SetAppConf("FailureInherit", req["FailureInherit"] == "true")

	if !setSpiderQueue(req) {
		return false
	}
	return true
}

func setSpiderQueue(req map[string]interface{}) bool {
	spNames, ok := req["spiders"].([]interface{})
	if !ok {
		logs.Log.Warning(" *     —— 亲，任务列表不能为空哦~")
		return false
	}
	spiders := []*spider.Spider{}
	for _, sp := range app.LogicApp.GetSpiderLib() {
		for _, spName := range spNames {
			if util.Atoa(spName) == sp.GetName() {
				spiders = append(spiders, sp.Copy())
			}
		}
	}
	app.LogicApp.SpiderPrepare(spiders)
	if app.LogicApp.GetSpiderQueue().Len() == 0 {
		logs.Log.Warning(" *     —— 亲，任务列表不能为空哦~")
		return false
	}
	return true
}
