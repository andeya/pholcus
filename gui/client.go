package gui

import (
	"github.com/henrylee2cn/pholcus/config"
	. "github.com/henrylee2cn/pholcus/gui/model"
	. "github.com/henrylee2cn/pholcus/node"
	"github.com/henrylee2cn/pholcus/reporter"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
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
		Title:   config.APP_NAME + "                                                          【 运行模式 -> 客户端 】",
		MinSize: Size{1100, 700},
		Layout:  VBox{ /*MarginsZero: true*/ },
		Children: []Widget{

			Composite{
				AssignTo: &setting,
				Layout:   Grid{Columns: 2},
				Children: []Widget{
					// 任务列表
					TableView{
						ColumnSpan:            1,
						MinSize:               Size{550, 350},
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
						MinSize:    Size{550, 0},
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
								Buttons: []RadioButton{
									{Text: GuiOpt.OutType[0].Key, Value: GuiOpt.OutType[0].String},
									{Text: GuiOpt.OutType[1].Key, Value: GuiOpt.OutType[1].String},
									{Text: GuiOpt.OutType[2].Key, Value: GuiOpt.OutType[2].String},
								},
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
						MinSize:   Size{110, 0},
						Text:      "断开服务器连接",
						AssignTo:  &toggleRunBtn,
						OnClicked: clientStart,
					},
				},
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	// 绑定log输出界面
	lv, err := NewLogView(mw)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(lv)

	if icon, err := walk.NewIconFromResource("ICON"); err == nil {
		mw.SetIcon(icon)
	}

	// 禁用配置
	setting.SetEnabled(false)
	toggleRunBtn.SetEnabled(false)

	// 开启报告
	reporter.Log.Run()

	// 运行pholcus核心
	PholcusRun()

	// 记录配置信息
	WTaskConf2()

	// 执行任务
	go clientExec()

	// 运行窗体程序
	mw.Run()
}

// 点击开始事件
func clientStart() {

	if toggleRunBtn.Text() == "重新连接服务器" {
		toggleRunBtn.SetEnabled(false)
		toggleRunBtn.SetText("正在连接服务器…")
		clientStop()
		return
	}

	toggleRunBtn.SetText("断开服务器连接")

}

func clientExec() {
	Pholcus.Spiders.Reset()
	for {
		// reporter.Log.Println("开始获取任务")

		// 从任务库获取一个任务
		t := Pholcus.DownTask()
		// reporter.Log.Printf("成功获取任务 %#v", t)

		// 准备运行
		TaskToReady(t)

		// 执行任务
		Exec(Pholcus.Spiders.Len())
	}
}

func clientStop() {

}
