package gui

import (
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func runmodeWindow() {
	if err := (MainWindow{
		AssignTo: &mw,
		DataBinder: DataBinder{
			AssignTo:       &db,
			DataSource:     Input,
			ErrorPresenter: ErrorPresenterRef{&ep},
		},
		Title:   config.APP_FULL_NAME,
		MinSize: Size{450, 350},
		Layout:  VBox{ /*MarginsZero: true*/ },
		Children: []Widget{

			RadioButtonGroupBox{
				AssignTo: &mode,
				Title:    "*运行模式",
				Layout:   HBox{},
				MinSize:  Size{0, 70},

				DataMember: "RunMode",
				Buttons: []RadioButton{
					{Text: GuiOpt.RunMode[0].Key, Value: GuiOpt.RunMode[0].Int},
					{Text: GuiOpt.RunMode[1].Key, Value: GuiOpt.RunMode[1].Int},
					{Text: GuiOpt.RunMode[2].Key, Value: GuiOpt.RunMode[2].Int},
				},
			},

			VSplitter{
				AssignTo: &host,
				MaxSize:  Size{0, 120},
				Children: []Widget{
					VSplitter{
						Children: []Widget{
							Label{
								Text: "分布式端口：（单机模式不填）",
							},
							NumberEdit{
								Value:    Bind("Port"),
								Suffix:   "",
								Decimals: 0,
							},
						},
					},

					VSplitter{
						Children: []Widget{
							Label{
								Text: "主节点 URL：（客户端模式必填）",
							},
							LineEdit{
								Text: Bind("Master"),
							},
						},
					},
				},
			},

			PushButton{
				Text:     "确认开始",
				MinSize:  Size{0, 30},
				AssignTo: &runStopBtn,
				OnClicked: func() {
					if err := db.Submit(); err != nil {
						logs.Log.Error("%v", err)
						return
					}

					switch Input.RunMode {
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

	if icon, err := walk.NewIconFromResource("ICON"); err == nil {
		mw.SetIcon(icon)
	}
	// 运行窗体程序
	mw.Run()
}
