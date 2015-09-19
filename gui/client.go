package gui

import (
	"github.com/henrylee2cn/pholcus/config"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
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
		Title:   config.APP_FULL_NAME + "                                                          【 运行模式 -> 客户端 】",
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

					// 必填项错误检查
					LineErrorPresenter{
						AssignTo: &ep,
					},

					PushButton{
						MinSize:  Size{110, 0},
						Text:     "断开服务器连接",
						AssignTo: &runStopBtn,
					},
				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	setWindow()

	// 禁用配置
	setting.SetEnabled(false)
	runStopBtn.SetEnabled(false)

	// 初始化应用
	Init()

	// 执行任务
	go LogicApp.Run()

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
