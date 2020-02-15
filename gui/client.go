package gui

import (
	. "github.com/lxn/walk/declarative"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/config"
)

func clientWindow() {
	mw.Close()
	if err := (MainWindow{
		AssignTo: &mw,
		DataBinder: DataBinder{
			AssignTo:       &db,
			DataSource:     Input,
			ErrorPresenter: ErrorPresenterRef{&ep},
		},
		Title:    config.FULL_NAME + "                                                          【 运行模式 -> 客户端 】",
		MinSize:  Size{1100, 600},
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
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

// 点击开始事件
// func clientStart() {

// 	if runStopBtn.Text() == "重新连接服务器" {
// 		runStopBtn.SetEnabled(false)
// 		runStopBtn.SetText("正在连接服务器…")
// 		clientStop()
// 		return
// 	}

// 	runStopBtn.SetText("断开服务器连接")

// }
