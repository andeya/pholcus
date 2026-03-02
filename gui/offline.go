//go:build windows

package gui

import (
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"

	"github.com/andeya/pholcus/app"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/status"
)

func offlineWindow() {
	mw.Close()

	if err := (declarative.MainWindow{
		AssignTo: &mw,
		DataBinder: declarative.DataBinder{
			AssignTo:       &db,
			DataSource:     Input,
			ErrorPresenter: declarative.ErrorPresenterRef{&ep},
		},
		Title:   config.FullName + "                                                          【 运行模式 ->  单机 】",
		MinSize: declarative.Size{1100, 700},
		Layout:  declarative.VBox{MarginsZero: true},
		Children: []declarative.Widget{

			declarative.Composite{
				AssignTo: &setting,
				Layout:   declarative.Grid{Columns: 2},
				Children: []declarative.Widget{
					// 任务列表
					declarative.TableView{
						ColumnSpan:            1,
						MinSize:               declarative.Size{550, 450},
						AlternatingRowBGColor: walk.RGB(255, 255, 224),
						CheckBoxes:            true,
						ColumnsOrderable:      true,
						Columns: []declarative.TableViewColumn{
							{Title: "#", Width: 45},
							{Title: "任务", Width: 110 /*, Format: "%.2f", Alignment: AlignFar*/},
							{Title: "描述", Width: 370},
						},
						Model: spiderMenu,
					},

					declarative.VSplitter{
						ColumnSpan: 1,
						MinSize:    declarative.Size{550, 450},
						Children: []declarative.Widget{

							declarative.VSplitter{
								Children: []declarative.Widget{
									declarative.Label{
										Text: "自定义配置（多任务请分别多包一层“<>”）：",
									},
									declarative.LineEdit{
										Text: declarative.Bind("Keyins"),
									},
								},
							},

							declarative.VSplitter{
								Children: []declarative.Widget{
									declarative.Label{
										Text: "*采集上限（默认限制URL数）：",
									},
									declarative.NumberEdit{
										Value:    declarative.Bind("Limit"),
										Suffix:   "",
										Decimals: 0,
									},
								},
							},

							declarative.VSplitter{
								Children: []declarative.Widget{
									declarative.Label{
										Text: "*并发协程：（1~99999）",
									},
									declarative.NumberEdit{
										Value:    declarative.Bind("ThreadNum", declarative.Range{1, 99999}),
										Suffix:   "",
										Decimals: 0,
									},
								},
							},

							declarative.VSplitter{
								Children: []declarative.Widget{
									declarative.Label{
										Text: "*分批输出大小：（1~5,000,000 条数据）",
									},
									declarative.NumberEdit{
										Value:    declarative.Bind("BatchCap", declarative.Range{1, 5000000}),
										Suffix:   "",
										Decimals: 0,
									},
								},
							},

							declarative.VSplitter{
								Children: []declarative.Widget{
									declarative.Label{
										Text: "*暂停时长参考:",
									},
									declarative.ComboBox{
										Value:         declarative.Bind("Pausetime", declarative.SelRequired{}),
										DisplayMember: "Key",
										BindingMember: "Int64",
										Model:         GuiOpt.Pausetime,
									},
								},
							},

							declarative.VSplitter{
								Children: []declarative.Widget{
									declarative.Label{
										Text: "*代理IP更换频率:",
									},
									declarative.ComboBox{
										Value:         declarative.Bind("ProxyMinute", declarative.SelRequired{}),
										DisplayMember: "Key",
										BindingMember: "Int64",
										Model:         GuiOpt.ProxyMinute,
									},
								},
							},

							declarative.RadioButtonGroupBox{
								ColumnSpan: 1,
								Title:      "*输出方式",
								Layout:     declarative.HBox{},
								DataMember: "OutType",
								Buttons:    outputList,
							},
						},
					},
				},
			},

			declarative.Composite{
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.VSplitter{
						Children: []declarative.Widget{
							// 必填项错误检查
							declarative.LineErrorPresenter{
								AssignTo: &ep,
							},
						},
					},

					declarative.HSplitter{
						MaxSize: declarative.Size{220, 50},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "继承并保存成功记录",
							},
							declarative.CheckBox{
								Checked: declarative.Bind("SuccessInherit"),
							},
						},
					},

					declarative.HSplitter{
						MaxSize: declarative.Size{220, 50},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "继承并保存失败记录",
							},
							declarative.CheckBox{
								Checked: declarative.Bind("FailureInherit"),
							},
						},
					},

					declarative.VSplitter{
						MaxSize: declarative.Size{90, 50},
						Children: []declarative.Widget{
							declarative.PushButton{
								Text:      "暂停/恢复",
								AssignTo:  &pauseRecoverBtn,
								OnClicked: offlinePauseRecover,
							},
						},
					},
					declarative.VSplitter{
						MaxSize: declarative.Size{90, 50},
						Children: []declarative.Widget{
							declarative.PushButton{
								Text:      "开始运行",
								AssignTo:  &runStopBtn,
								OnClicked: offlineRunStop,
							},
						},
					},
				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	setWindow()

	pauseRecoverBtn.SetVisible(false)

	// 初始化应用
	Init()

	// 运行窗体程序
	mw.Run()
}

// 暂停\恢复
func offlinePauseRecover() {
	switch app.LogicApp.Status() {
	case status.RUN:
		pauseRecoverBtn.SetText("恢复运行")
	case status.PAUSE:
		pauseRecoverBtn.SetText("暂停")
	}
	app.LogicApp.PauseRecover()
}

// 开始\停止控制
func offlineRunStop() {
	if !app.LogicApp.IsStopped() {
		go func() {
			runStopBtn.SetEnabled(false)
			runStopBtn.SetText("停止中…")
			pauseRecoverBtn.SetVisible(false)
			pauseRecoverBtn.SetText("暂停")
			app.LogicApp.Stop()
			offlineResetBtn()
		}()
		return
	}

	if err := db.Submit(); err != nil {
		logs.Log().Error("%v", err)
		return
	}

	// 读取任务
	Input.Spiders = spiderMenu.GetChecked()

	// if len(Input.Spiders) == 0 {
	// 	logs.Log().Warning(" *     —— 亲，任务列表不能为空哦~")
	// 	return
	// }

	runStopBtn.SetText("停止")

	// 记录配置信息
	SetTaskConf()

	// 更新蜘蛛队列
	SpiderPrepare()

	go func() {
		pauseRecoverBtn.SetText("暂停")
		pauseRecoverBtn.SetVisible(true)
		app.LogicApp.Run()
		offlineResetBtn()
		pauseRecoverBtn.SetVisible(false)
		pauseRecoverBtn.SetText("暂停")
	}()
}

// Offline 模式下按钮状态控制
func offlineResetBtn() {
	runStopBtn.SetEnabled(true)
	runStopBtn.SetText("开始运行")
}
