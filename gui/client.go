//go:build windows

package gui

import (
	"github.com/lxn/walk/declarative"

	"github.com/andeya/pholcus/app"
	"github.com/andeya/pholcus/config"
)

func clientWindow() {
	mw.Close()
	if err := (declarative.MainWindow{
		AssignTo: &mw,
		DataBinder: declarative.DataBinder{
			AssignTo:       &db,
			DataSource:     Input,
			ErrorPresenter: declarative.ErrorPresenterRef{&ep},
		},
		Title:    config.FullName + "                                                          【 运行模式 -> 客户端 】",
		MinSize:  declarative.Size{1100, 600},
		Layout:   declarative.VBox{MarginsZero: true},
		Children: []declarative.Widget{
			// Composite{
			// 	Layout:  HBox{},
			// 	MaxSize: Size{1100, 150},
			// 	Children: []Widget{
			// 		PushButton{
			// 			MaxSize:  Size{1000, 150},
			// 			Text:     "断开服务器连接",
			// 			AssignTo: &runStopBtn,
			// 		},
			// 	},
			// },
		},
	}.Create()); err != nil {
		panic(err)
	}

	setWindow()

	// 初始化应用
	Init()

	// 执行任务
	go app.LogicApp.Run()

	// 运行窗体程序
	mw.Run()
}
