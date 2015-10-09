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

var (
	logicApp   = app.New().SetLog(Lsc).AsyncLog(true)
	spiderMenu = func() (spmenu []map[string]string) {
		// 获取蜘蛛家族
		for _, sp := range logicApp.GetSpiderLib() {
			spmenu = append(spmenu, map[string]string{"name": sp.GetName(), "description": sp.GetDescription()})
		}
		return spmenu
	}()
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

	void["mode"] = logicApp.GetAppConf("mode").(int)

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
		Sc.Write(sessID, tplData(logicApp.GetAppConf("mode").(int)), 1)
	}

	// 初始化运行
	wsApi["init"] = func(sessID string, req map[string]interface{}) {
		var mode = util.Atoi(req["mode"])
		var port = util.Atoi(req["port"])
		var master = util.Atoa(req["ip"]) //服务器(主节点)地址，不含端口
		currMode := logicApp.GetAppConf("mode").(int)
		if currMode == status.UNSET {
			logicApp.Init(mode, port, master, Lsc) // 运行模式初始化，设置log输出目标
		} else {
			logicApp = logicApp.ReInit(mode, port, master) // 切换运行模式
		}

		if mode == status.CLIENT {
			go logicApp.Run()
		}

		// 写入发送通道
		Sc.Write(sessID, tplData(mode))
	}

	wsApi["run"] = func(sessID string, req map[string]interface{}) {
		if logicApp.GetAppConf("mode").(int) != status.CLIENT {
			if !setConf(req) {
				// Sc.Write(sessID, map[string]interface{}{"operate": "stop", "mode": logicApp.GetAppConf("mode").(int), "status": status.UNKNOW}, 1)
				return
			}
		}

		if logicApp.GetAppConf("mode").(int) == status.OFFLINE {
			Sc.Write(sessID, map[string]interface{}{"operate": "run", "status": status.RUN})
		}

		go func() {
			logicApp.Run()
			if logicApp.GetAppConf("mode").(int) == status.OFFLINE {
				Sc.Write(sessID, map[string]interface{}{"operate": "stop", "status": status.STOP})
			}
		}()
	}

	// 终止当前任务，现仅支持单机模式
	wsApi["stop"] = func(sessID string, req map[string]interface{}) {
		if logicApp.GetAppConf("mode").(int) != status.OFFLINE {
			Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
			return
		} else {
			logicApp.Stop()
			Sc.Write(sessID, map[string]interface{}{"operate": "stop"})
		}
	}

	// 终止当前任务，现仅支持单机模式
	wsApi["pauseRecover"] = func(sessID string, req map[string]interface{}) {
		if logicApp.GetAppConf("mode").(int) != status.OFFLINE {
			return
		}
		logicApp.PauseRecover()
		Sc.Write(sessID, map[string]interface{}{"operate": "pauseRecover"})
	}

	// 退出当前模式
	wsApi["exit"] = func(sessID string, req map[string]interface{}) {
		Sc.Write(sessID, map[string]interface{}{"operate": "exit"})
		logicApp = logicApp.ReInit(status.UNSET, 0, "")
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
			l := logicApp.GetSpiderQueue().Len()
			if l == 0 {
				return 0
			}
			var curr = make(map[string]bool, l)
			for _, sp := range logicApp.GetSpiderQueue().GetAll() {
				curr[sp.GetName()] = true
			}

			return curr
		}(),
	}

	// 输出方式清单
	info["outputs"] = map[string]interface{}{
		"menu": logicApp.GetOutputLib(),
		"curr": logicApp.GetAppConf("OutType"),
	}

	// 并发协程上限
	info["threadNum"] = map[string]uint{
		"max":     999999,
		"min":     1,
		"default": logicApp.GetAppConf("ThreadNum").(uint),
	}
	// 暂停时间，单位ms
	info["sleepTime"] = map[string][]uint{
		"base":   {0, 100, 300, 500, 1000, 3000, 5000, 10000, 15000, 20000, 30000, 60000},
		"random": {0, 100, 300, 500, 1000, 3000, 5000, 10000, 15000, 20000, 30000, 60000},
		"default": func() []uint {
			var a = logicApp.GetAppConf("Pausetime").([2]uint)
			return []uint{a[0], a[1]}
		}(),
	}
	// 分批输出的容量
	info["dockerCap"] = map[string]uint{"min": 1, "max": 5000000, "default": logicApp.GetAppConf("DockerCap").(uint)}
	// 最大页数
	info["maxPage"] = logicApp.GetAppConf("maxPage")
	// 关键词
	info["keywords"] = logicApp.GetAppConf("Keywords")
	// 继承之前的去重记录
	info["inheritDeduplication"] = logicApp.GetAppConf("InheritDeduplication")
	// 去重记录保存位置,"file"或"mgo"
	info["deduplicationTarget"] = map[string]interface{}{
		"menu": []string{status.FILE, status.MGO},
		"curr": logicApp.GetAppConf("DeduplicationTarget"),
	}
	// 运行状态
	info["status"] = logicApp.Status()

	return info
}

// 配置运行参数
func setConf(req map[string]interface{}) bool {
	if tn := util.Atoui(req["threadNum"]); tn == 0 {
		logicApp.SetAppConf("threadNum", 1)
	} else {
		logicApp.SetAppConf("threadNum", tn)
	}

	logicApp.
		SetAppConf("Pausetime", [2]uint{(util.Atoui(req["baseSleeptime"])), util.Atoui(req["randomSleepPeriod"])}).
		SetAppConf("OutType", util.Atoa(req["output"])).
		SetAppConf("DockerCap", util.Atoui(req["dockerCap"])).
		SetAppConf("MaxPage", util.Atoi(req["maxPage"])).
		SetAppConf("Keywords", util.Atoa(req["keywords"])).
		SetAppConf("DeduplicationTarget", req["deduplicationTarget"])

	var inheritDeduplication bool
	if req["inheritDeduplication"] == "true" {
		inheritDeduplication = true
	}
	logicApp.SetAppConf("InheritDeduplication", inheritDeduplication)

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
	for _, sp := range logicApp.GetSpiderLib() {
		for _, spName := range spNames {
			if util.Atoa(spName) == sp.GetName() {
				spiders = append(spiders, sp.Gost())
			}
		}
	}
	logicApp.SpiderPrepare(spiders)
	if logicApp.GetSpiderQueue().Len() == 0 {
		logs.Log.Warning(" *     —— 亲，任务列表不能为空哦~")
		return false
	}
	return true
}
