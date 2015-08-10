package gui

import (
	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/spider"
	"github.com/lxn/walk"
	"log"
)

var LogicApp = app.New()

func Run() {
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
		log.Fatal(err)
	}
	LogicApp.SetLog(lv)

	// 设置左上角图标
	if icon, err := walk.NewIconFromResource("ICON"); err == nil {
		mw.SetIcon(icon)
	}
}
