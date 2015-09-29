package web

import (
	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/status"
	ws "github.com/henrylee2cn/websocket.google"
)

var (
	wchan       chan interface{}
	wchanClosed bool
	isRunning   bool

	logicApp   = app.New().SetLog(Log).AsyncLog(true)
	spiderMenu = func() (spmenu []map[string]string) {
		// 获取蜘蛛家族
		for _, sp := range logicApp.GetSpiderLib() {
			spmenu = append(spmenu, map[string]string{"name": sp.GetName(), "description": sp.GetDescription()})
		}
		return spmenu
	}()

	wsApi = map[string]func(*ws.Conn, map[string]interface{}){}
)

func wsHandle(conn *ws.Conn) {
	wchanClosed = false
	defer func() {
		// 连接断开前关闭正在运行的任务
		// if isRunning {
		// 	isRunning = false
		// 	logicApp.LogRest().Stop()
		// }
		wchanClosed = true
		close(wchan)
		conn.Close()
	}()
	wchan = make(chan interface{}, 1024)

	go func(conn *ws.Conn) {
		var err error
		defer func() {
			// logs.Log.Debug("websocket发送出错断开 (%v) !", err)
		}()
		for info := range wchan {
			if _, err = ws.JSON.Send(conn, info); err != nil {
				return
			}
		}
	}(conn)

	for {
		var req map[string]interface{}

		if err := ws.JSON.Receive(conn, &req); err != nil {
			// logs.Log.Debug("websocket接收出错断开 (%v) !", err)
			return
		}

		// log.Log.Debug("Received from web: %v", req)
		wsApi[util.Atoa(req["operate"])](conn, req)
	}
}

func init() {
	// 初始化运行
	wsApi["goon"] = func(conn *ws.Conn, req map[string]interface{}) {
		// 写入发送通道
		if !wchanClosed {
			wchan <- tplData(logicApp.GetAppConf("mode").(int))
		}
	}

	// 初始化运行
	wsApi["init"] = func(conn *ws.Conn, req map[string]interface{}) {
		var mode = util.Atoi(req["mode"])
		var port = util.Atoi(req["port"])
		var master = util.Atoa(req["ip"]) //服务器(主节点)地址，不含端口
		currMode := logicApp.GetAppConf("mode").(int)
		if currMode == status.UNSET {
			logicApp.Init(mode, port, master, Log) // 运行模式初始化，设置log输出目标
		} else {
			logicApp = logicApp.ReInit(mode, port, master) // 切换运行模式
		}

		if mode == status.CLIENT {
			go logicApp.Run()
		}

		// 写入发送通道
		if !wchanClosed {
			wchan <- tplData(mode)
		}
	}

	wsApi["run"] = func(conn *ws.Conn, req map[string]interface{}) {
		if logicApp.GetAppConf("mode").(int) != status.CLIENT && !wchanClosed {
			if !setConf(req) {
				wchan <- map[string]interface{}{"mode": logicApp.GetAppConf("mode").(int), "status": status.UNKNOW}
				return
			}
		}

		if logicApp.GetAppConf("mode").(int) == status.OFFLINE && !wchanClosed {
			wchan <- map[string]interface{}{"operate": "run", "mode": status.OFFLINE, "status": status.RUN}
		}

		go func() {
			isRunning = true
			logicApp.Run()
			if logicApp.GetAppConf("mode").(int) == status.OFFLINE && !wchanClosed {
				isRunning = false
				wchan <- map[string]interface{}{"operate": "stop", "mode": status.OFFLINE, "status": status.STOP}
			}
		}()
	}

	// 终止当前任务，现仅支持单机模式
	wsApi["stop"] = func(conn *ws.Conn, req map[string]interface{}) {
		if logicApp.GetAppConf("mode").(int) != status.OFFLINE && !wchanClosed {
			wchan <- map[string]interface{}{"operate": "stop", "mode": logicApp.GetAppConf("mode").(int), "status": status.UNKNOW}
			return
		} else if !wchanClosed {
			logicApp.Stop()
			wchan <- map[string]interface{}{"operate": "stop", "mode": status.OFFLINE, "status": status.STOP}
		}
	}

	// 终止当前任务，现仅支持单机模式
	wsApi["pauseRecover"] = func(conn *ws.Conn, req map[string]interface{}) {
		if logicApp.GetAppConf("mode").(int) != status.OFFLINE {
			return
		}
		logicApp.PauseRecover()
	}

	// 退出当前模式
	wsApi["exit"] = func(conn *ws.Conn, req map[string]interface{}) {
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
		"memu": spiderMenu,
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
		"memu": logicApp.GetOutputLib(),
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
		SetAppConf("Keywords", util.Atoa(req["keywords"]))

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

// log发送api
func wsLogHandle(conn *ws.Conn) {
	var err error
	defer func() {
		// logs.Log.Debug("websocket log发送出错断开 (%v) !", err)
	}()

	// 新建web前端log输出
	Log.Open()

	go func(conn *ws.Conn) {
		defer func() {
			// 关闭web前端log输出
			Log.Close()
			// 关闭websocket连接
			conn.Close()
		}()
		for {
			if err := ws.JSON.Receive(conn, nil); err != nil {
				// logs.Log.Debug("websocket log接收出错断开 (%v) !", err)
				return
			}
		}
	}(conn)

	for msg := range Log.logChan {
		if _, err = ws.Message.Send(conn, msg); err != nil {
			return
		}
	}
}
