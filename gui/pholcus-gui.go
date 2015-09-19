// [spider frame (golang)] Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能；
//（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）。
// GUI界面版。
package gui

import (
	"log"
	"runtime"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/lxn/walk"
)

var LogicApp = app.New().AsyncLog(true)

func Run() {
	// 开启最大核心数运行
	runtime.GOMAXPROCS(runtime.NumCPU())
	runmodeWindow()
}

func Init() {
	LogicApp.Init(Input.RunMode, Input.Port, Input.Master)
}

func SetTaskConf() {
	// 纠正协程数
	if Input.ThreadNum == 0 {
		Input.ThreadNum = 1
	}
	LogicApp.SetThreadNum(Input.ThreadNum)
	LogicApp.SetPausetime([2]uint{Input.BaseSleeptime, Input.RandomSleepPeriod})
	LogicApp.SetOutType(Input.OutType)
	LogicApp.SetDockerCap(Input.DockerCap) //分段转储容器容量
	// 选填项
	LogicApp.SetMaxPage(Input.MaxPage)
}

func SpiderPrepare() {
	sps := []*spider.Spider{}
	for _, sp := range Input.Spiders {
		sps = append(sps, sp.Spider)
	}
	LogicApp.SpiderPrepare(sps, Input.Keywords)
}

func SpiderNames() (names []string) {
	for _, sp := range Input.Spiders {
		names = append(names, sp.Spider.GetName())
	}
	return
}

func setWindow() {
	// 绑定log输出界面
	lv, err := NewLogView(mw)
	if err != nil {
		panic(err)
	}
	LogicApp.SetLog(lv)
	log.SetOutput(lv)
	// 设置左上角图标
	if icon, err := walk.NewIconFromResource("ICON"); err == nil {
		mw.SetIcon(icon)
	}
}
