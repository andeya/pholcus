//go:build windows

package gui

import (
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"

	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/status"
)

func runmodeWindow() {
	if err := (declarative.MainWindow{
		AssignTo: &mw,
		DataBinder: declarative.DataBinder{
			AssignTo:       &db,
			DataSource:     Input,
			ErrorPresenter: declarative.ErrorPresenterRef{&ep},
		},
		Title:   config.FullName,
		MinSize: declarative.Size{450, 350},
		Layout:  declarative.VBox{ /*MarginsZero: true*/ },
		Children: []declarative.Widget{

			declarative.RadioButtonGroupBox{
				AssignTo: &mode,
				Title:    "*运行模式",
				Layout:   declarative.HBox{},
				MinSize:  declarative.Size{0, 70},

				DataMember: "Mode",
				Buttons: []declarative.RadioButton{
					{Text: GuiOpt.Mode[0].Key, Value: GuiOpt.Mode[0].Int},
					{Text: GuiOpt.Mode[1].Key, Value: GuiOpt.Mode[1].Int},
					{Text: GuiOpt.Mode[2].Key, Value: GuiOpt.Mode[2].Int},
				},
			},

			declarative.VSplitter{
				AssignTo: &host,
				MaxSize:  declarative.Size{0, 120},
				Children: []declarative.Widget{
					declarative.Label{
						Text: "分布式端口：（单机模式不填）",
					},
					declarative.NumberEdit{
						Value:    declarative.Bind("Port"),
						Suffix:   "",
						Decimals: 0,
					},

					declarative.Label{
						Text: "主节点 URL：（客户端模式必填）",
					},
					declarative.LineEdit{
						Text: declarative.Bind("Master"),
					},
				},
			},

			declarative.PushButton{
				Text:     "确认开始",
				MinSize:  declarative.Size{0, 30},
				AssignTo: &runStopBtn,
				OnClicked: func() {
					if err := db.Submit(); err != nil {
						logs.Log().Error("%v", err)
						return
					}

					switch Input.Mode {
					case status.OFFLINE:
						offlineWindow()

					case status.SERVER:
						serverWindow()

					case status.CLIENT:
						clientWindow()
					}

				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	if icon, err := walk.NewIconFromResourceId(3); err == nil {
		mw.SetIcon(icon)
	}
	// 运行窗体程序
	mw.Run()
}
