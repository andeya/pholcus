// [spider frame (golang)] Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能；
//（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）。
// GUI界面版。
package gui

import (
	"log"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/app/spider"
	. "github.com/henrylee2cn/pholcus/gui/model"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

// 执行入口
func Run() {
	app.LogicApp.SetAppConf("Mode", status.OFFLINE)

	outputList = func() (o []declarative.RadioButton) {
		// 设置默认选择
		Input.AppConf.OutType = app.LogicApp.GetOutputLib()[0]
		// 获取输出选项
		for _, out := range app.LogicApp.GetOutputLib() {
			o = append(o, declarative.RadioButton{Text: out, Value: out})
		}
		return
	}()

	spiderMenu = NewSpiderMenu(spider.Species)

	runmodeWindow()
}

func Init() {
	app.LogicApp.Init(Input.Mode, Input.Port, Input.Master)
}

func SetTaskConf() {
	// 纠正协程数
	if Input.ThreadNum == 0 {
		Input.ThreadNum = 1
	}
	app.LogicApp.SetAppConf("ThreadNum", Input.ThreadNum).
		SetAppConf("Pausetime", Input.Pausetime).
		SetAppConf("ProxyMinute", Input.ProxyMinute).
		SetAppConf("OutType", Input.OutType).
		SetAppConf("DockerCap", Input.DockerCap).
		SetAppConf("Limit", Input.Limit).
		SetAppConf("Keyins", Input.Keyins)
}

func SpiderPrepare() {
	sps := []*spider.Spider{}
	for _, sp := range Input.Spiders {
		sps = append(sps, sp.Spider)
	}
	app.LogicApp.SpiderPrepare(sps)
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
	app.LogicApp.SetLog(lv)
	log.SetOutput(lv)
	// 设置左上角图标
	if icon, err := walk.NewIconFromResourceId(3); err == nil {
		mw.SetIcon(icon)
	}
}
