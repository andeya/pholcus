//go:build windows

// [spider frame (golang)] Pholcus (Ghost Spider) is a high-concurrency, distributed, heavyweight crawler written in pure Go.
// It supports standalone, server, and client modes with Web, GUI, and CLI interfaces; simple flexible rules;
// batch task concurrency; rich output formats (mysql/mongodb/csv/excel etc.); and shared demos.
// It also supports horizontal and vertical crawling, simulated login, and advanced features like pause/cancel.
// (Official QQ group: Go Big Data 42731170)
// GUI package.
package gui

import (
	"log"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"

	"github.com/andeya/pholcus/app"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/gui/model"
	"github.com/andeya/pholcus/runtime/status"
)

// Run is the entry point for the GUI.
func Run() {
	app.LogicApp.SetAppConf("Mode", status.OFFLINE)

	outputList = func() (o []declarative.RadioButton) {
		// Set default selection
		Input.AppConf.OutType = app.LogicApp.GetOutputLib()[0]
		// Get output options
		for _, out := range app.LogicApp.GetOutputLib() {
			o = append(o, declarative.RadioButton{Text: out, Value: out})
		}
		return
	}()

	spiderMenu = model.NewSpiderMenu(spider.Species)

	runmodeWindow()
}

func Init() {
	app.LogicApp.Init(Input.Mode, Input.Port, Input.Master)
}

func SetTaskConf() {
	// Correct goroutine count
	if Input.ThreadNum == 0 {
		Input.ThreadNum = 1
	}
	app.LogicApp.SetAppConf("ThreadNum", Input.ThreadNum).
		SetAppConf("Pausetime", Input.Pausetime).
		SetAppConf("ProxyMinute", Input.ProxyMinute).
		SetAppConf("OutType", Input.OutType).
		SetAppConf("BatchCap", Input.BatchCap).
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
	// Bind log output
	lv, err := NewLogView(mw)
	if err != nil {
		panic(err)
	}
	app.LogicApp.SetLog(lv)
	log.SetOutput(lv)
	// Set window icon
	if icon, err := walk.NewIconFromResourceId(3); err == nil {
		mw.SetIcon(icon)
	}
}
