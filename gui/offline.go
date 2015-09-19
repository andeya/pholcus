package gui

import (
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func offlineWindow() {
	mw.Close()

	if err := (MainWindow{
		AssignTo: &mw,
		DataBinder: DataBinder{
			AssignTo:       &db,
			DataSource:     Input,
			ErrorPresenter: ErrorPresenterRef{&ep},
		},
		Title:   config.APP_FULL_NAME + "                                                          【 运行模式 ->  单机 】",
		MinSize: Size{1100, 700},
		Layout:  VBox{MarginsZero: true},
		Children: []Widget{

			Composite{
				AssignTo: &setting,
				Layout:   Grid{Columns: 2},
				Children: []Widget{
					// 任务列表
					TableView{
						ColumnSpan:            1,
						MinSize:               Size{550, 450},
						AlternatingRowBGColor: walk.RGB(255, 255, 224),
						CheckBoxes:            true,
						ColumnsOrderable:      true,
						Columns: []TableViewColumn{
							{Title: "#", Width: 45},
							{Title: "任务", Width: 110 /*, Format: "%.2f", Alignment: AlignFar*/},
							{Title: "描述", Width: 370},
						},
						Model: spiderMenu,
					},

					VSplitter{
						ColumnSpan: 1,
						MinSize:    Size{550, 450},
						Children: []Widget{

							VSplitter{
								Children: []Widget{
									Label{
										Text: "自定义输入：（多任务之间以 | 隔开，选填）",
									},
									LineEdit{
										Text: Bind("Keywords"),
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "*并发协程：（1~99999）",
									},
									NumberEdit{
										Value:    Bind("ThreadNum", Range{1, 99999}),
										Suffix:   "",
										Decimals: 0,
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "采集页数：（选填）",
									},
									NumberEdit{
										Value:    Bind("MaxPage"),
										Suffix:   "",
										Decimals: 0,
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "*分批输出大小：（1~5,000,000 条数据）",
									},
									NumberEdit{
										Value:    Bind("DockerCap", Range{1, 5000000}),
										Suffix:   "",
										Decimals: 0,
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "*间隔基准:",
									},
									ComboBox{
										Value:         Bind("BaseSleeptime", SelRequired{}),
										BindingMember: "Uint",
										DisplayMember: "Key",
										Model:         GuiOpt.SleepTime,
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "*随机延迟:",
									},
									ComboBox{
										Value:         Bind("RandomSleepPeriod", SelRequired{}),
										BindingMember: "Uint",
										DisplayMember: "Key",
										Model:         GuiOpt.SleepTime,
									},
								},
							},

							RadioButtonGroupBox{
								ColumnSpan: 1,
								Title:      "*输出方式",
								Layout:     HBox{},
								DataMember: "OutType",
								Buttons:    outputList,
							},
						},
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					VSplitter{
						Children: []Widget{
							// 必填项错误检查
							LineErrorPresenter{
								AssignTo: &ep,
							},
						},
					},
					VSplitter{
						MaxSize: Size{90, 50},
						Children: []Widget{
							PushButton{
								Text:      "暂停/恢复",
								AssignTo:  &pauseRecoverBtn,
								OnClicked: offlinePauseRecover,
							},
						},
					},
					VSplitter{
						MaxSize: Size{90, 50},
						Children: []Widget{
							PushButton{
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
	switch LogicApp.Status() {
	case status.RUN:
		pauseRecoverBtn.SetText("恢复运行")
	case status.PAUSE:
		pauseRecoverBtn.SetText("暂停")
	}
	LogicApp.PauseRecover()
}

// 开始\停止控制
func offlineRunStop() {
	if LogicApp.Status() != status.STOP {
		runStopBtn.SetEnabled(false)
		runStopBtn.SetText("停止中…")
		pauseRecoverBtn.SetVisible(false)
		pauseRecoverBtn.SetText("暂停")
		LogicApp.Stop()
		offlineResetBtn()
		return
	}

	if err := db.Submit(); err != nil {
		logs.Log.Error("%v", err)
		return
	}

	// 读取任务
	Input.Spiders = spiderMenu.GetChecked()

	if len(Input.Spiders) == 0 {
		logs.Log.Warning(" *     —— 亲，任务列表不能为空哦~")
		return
	}

	runStopBtn.SetText("停止")

	// 记录配置信息
	SetTaskConf()

	// 更新蜘蛛队列
	SpiderPrepare()

	go func() {
		pauseRecoverBtn.SetText("暂停")
		pauseRecoverBtn.SetVisible(true)
		LogicApp.Run()
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
