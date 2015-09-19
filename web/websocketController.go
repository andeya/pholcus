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
	spiderMenu = make([]map[string]string, 0)
	wsApi      = map[string]func(*ws.Conn, map[string]interface{}){}
)

func wsHandle(conn *ws.Conn) {
	wchanClosed = false
	defer func() {
		// 连接断开前关闭正在运行的任务
		if isRunning {
			isRunning = false
			logicApp.LogRest().Stop()
		}
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
	wsApi["init"] = func(conn *ws.Conn, req map[string]interface{}) {
		var mode = util.Atoi(req["mode"])
		var port = util.Atoi(req["port"])
		var master = util.Atoa(req["ip"]) //服务器(主节点)地址，不含端口
		currMode := logicApp.GetRunMode()
		if currMode == status.UNSET {
			logicApp.Init(mode, port, master, Log) // 运行模式初始化，设置log输出目标
			// 获取蜘蛛家族
			for _, sp := range logicApp.GetAllSpiders() {
				spiderMenu = append(spiderMenu, map[string]string{"name": sp.GetName(), "description": sp.GetDescription()})
			}
		} else {
			logicApp = logicApp.ReInit(mode, port, master) // 切换运行模式
		}

		// 输出到前端的信息
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
			go logicApp.Run()
			goto send
		}

		// 蜘蛛家族清单
		info["spiderMenu"] = spiderMenu
		// 输出方式清单
		info["outputMenu"] = logicApp.GetOutputLib()
		// 并发协程上限
		info["threadNum"] = map[string]uint{
			"max":     999999,
			"min":     1,
			"default": defaultConfig.ThreadNum,
		}
		// 暂停时间，单位ms
		info["sleepTime"] = map[string][]uint{
			"base":    {0, 100, 300, 500, 1000, 3000, 5000, 10000, 15000, 20000, 30000, 60000},
			"random":  {0, 100, 300, 500, 1000, 3000, 5000, 10000, 15000, 20000, 30000, 60000},
			"default": {defaultConfig.Pausetime[0], defaultConfig.Pausetime[1]},
		}
		// 分批输出的容量
		info["dockerCap"] = map[string]uint{"min": 1, "max": 5000000, "default": defaultConfig.DockerCap}

	send:
		// 写入发送通道
		if !wchanClosed {
			wchan <- info
		}
	}

	wsApi["run"] = func(conn *ws.Conn, req map[string]interface{}) {
		if logicApp.GetRunMode() != status.CLIENT && !wchanClosed {
			if !setConf(req) {
				wchan <- map[string]interface{}{"mode": logicApp.GetRunMode(), "status": 0}
				return
			}
		}

		if logicApp.GetRunMode() == status.OFFLINE && !wchanClosed {
			wchan <- map[string]interface{}{"operate": "run", "mode": status.OFFLINE, "status": 1}
		}

		go func() {
			isRunning = true
			logicApp.Run()
			if logicApp.GetRunMode() == status.OFFLINE && !wchanClosed {
				isRunning = false
				wchan <- map[string]interface{}{"operate": "stop", "mode": status.OFFLINE, "status": 1}
			}
		}()
	}

	// 终止当前任务，现仅支持单机模式
	wsApi["stop"] = func(conn *ws.Conn, req map[string]interface{}) {
		if logicApp.GetRunMode() != status.OFFLINE && !wchanClosed {
			wchan <- map[string]interface{}{"operate": "stop", "mode": logicApp.GetRunMode(), "status": 0}
			return
		} else if !wchanClosed {
			wchan <- map[string]interface{}{"operate": "stop", "mode": status.OFFLINE, "status": 1}
		}
		logicApp.Stop()
	}

	// 终止当前任务，现仅支持单机模式
	wsApi["pauseRecover"] = func(conn *ws.Conn, req map[string]interface{}) {
		if logicApp.GetRunMode() != status.OFFLINE {
			return
		}
		logicApp.PauseRecover()
	}
}

// 配置运行参数
func setConf(req map[string]interface{}) bool {
	if tn := util.Atoui(req["threadNum"]); tn == 0 {
		logicApp.SetThreadNum(1)
	} else {
		logicApp.SetThreadNum(tn)
	}
	logicApp.SetPausetime([2]uint{(util.Atoui(req["baseSleeptime"])), util.Atoui(req["randomSleepPeriod"])})
	logicApp.SetOutType(util.Atoa(req["output"]))
	logicApp.SetDockerCap(util.Atoui(req["dockerCap"])) //分段转储容器容量
	// 选填项
	logicApp.SetMaxPage(util.Atoi(req["maxPage"]))
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
	for _, sp := range logicApp.GetAllSpiders() {
		for _, spName := range spNames {
			if util.Atoa(spName) == sp.GetName() {
				spiders = append(spiders, sp.Gost())
			}
		}
	}
	logicApp.SpiderPrepare(spiders, util.Atoa(req["keywords"]))
	if logicApp.SpiderQueueLen() == 0 {
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
