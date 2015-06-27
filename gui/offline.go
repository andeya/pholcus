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

func offlineWindow() {
	mw.Close()

	if err := (MainWindow{
		AssignTo: &mw,
		DataBinder: DataBinder{
			AssignTo:       &db,
			DataSource:     Input,
			ErrorPresenter: ErrorPresenterRef{&ep},
		},
		Title:   config.APP_NAME + "                                                          【 运行模式 ->  单机 】",
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
						Text:      "开始运行",
						AssignTo:  &toggleRunBtn,
						OnClicked: offlineStart,
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

	// 开启报告
	reporter.Log.Run()

	// 运行pholcus核心
	PholcusRun()

	// 运行窗体程序
	mw.Run()
}

// 点击开始事件
func offlineStart() {
	// 每次点击重新开启报告
	reporter.Log.Run()

	if toggleRunBtn.Text() == "取消" {
		toggleRunBtn.SetEnabled(false)
		toggleRunBtn.SetText("取消中…")
		Stop()
		return
	}

	if err := db.Submit(); err != nil {
		log.Println(err)
		return
	}

	// 读取任务
	Input.Spiders = spiderMenu.GetChecked()

	if len(Input.Spiders) == 0 {
		log.Println(" *     —— 亲，任务列表不能为空哦~")
		return
	}

	toggleRunBtn.SetText("取消")

	// 记录配置信息
	WTaskConf2()

	// 初始化蜘蛛列表，返回长度，并执行任务
	Exec(InitSpiders())
}
