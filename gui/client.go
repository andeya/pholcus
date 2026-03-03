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
		Title:    config.FullName + "                                                          [ Run Mode -> Client ]",
		MinSize:  declarative.Size{1100, 600},
		Layout:   declarative.VBox{MarginsZero: true},
		Children: []declarative.Widget{
			// Composite{
			// 	Layout:  HBox{},
			// 	MaxSize: Size{1100, 150},
			// 	Children: []Widget{
			// 		PushButton{
			// 			MaxSize:  Size{1000, 150},
			// 			Text:     "Disconnect from server",
			// 			AssignTo: &runStopBtn,
			// 		},
			// 	},
			// },
		},
	}.Create()); err != nil {
		panic(err)
	}

	setWindow()

	// Initialize app
	Init()

	// Run task
	go app.LogicApp.Run()

	// Run window
	mw.Run()
}
