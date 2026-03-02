//go:build windows

package gui

import (
	"strconv"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"

	"github.com/andeya/pholcus/app"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
)

var serverCount int

func serverWindow() {
	mw.Close()

	if err := (declarative.MainWindow{
		AssignTo: &mw,
		DataBinder: declarative.DataBinder{
			AssignTo:       &db,
			DataSource:     Input,
			ErrorPresenter: declarative.ErrorPresenterRef{ErrorPresenter: &ep},
		},
		Title:   config.FullName + "                                                          【 运行模式 -> 服务器 】",
		MinSize: declarative.Size{Width: 1100, Height: 700},
		Layout:  declarative.VBox{MarginsZero: true},
		Children: []declarative.Widget{

			declarative.Composite{
				AssignTo: &setting,
				Layout:   declarative.Grid{Columns: 2},
				Children: []declarative.Widget{
					// 任务列表
					declarative.TableView{
						ColumnSpan:            1,
						MinSize:               declarative.Size{Width: 550, Height: 450},
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
						MinSize:    declarative.Size{Width: 550, Height: 450},
						Children: []declarative.Widget{

							declarative.VSplitter{
								Children: []declarative.Widget{
									declarative.Label{
										Text: "自定义配置（多任务请分别多包一层“<>”）",
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

					// 必填项错误检查
					declarative.LineErrorPresenter{
						AssignTo: &ep,
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

					declarative.PushButton{
						MinSize:   declarative.Size{90, 0},
						Text:      serverBtnTxt(),
						AssignTo:  &runStopBtn,
						OnClicked: serverStart,
					},
				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	setWindow()

	// 初始化应用
	Init()

	// 运行窗体程序
	mw.Run()
}

// 点击开始事件
func serverStart() {
	if err := db.Submit(); err != nil {
		logs.Log().Error("%v", err)
		return
	}

	// 读取任务
	Input.Spiders = spiderMenu.GetChecked()

	if len(Input.Spiders) == 0 {
		logs.Log().Warning(" *     —— 亲，任务列表不能为空哦~")
		return
	}

	// 记录配置信息
	SetTaskConf()

	runStopBtn.SetEnabled(false)
	runStopBtn.SetText("分发任务 (···)")

	// 重置spiders队列
	SpiderPrepare()

	// 生成分发任务
	app.LogicApp.Run()

	serverCount++

	runStopBtn.SetText(serverBtnTxt())
	runStopBtn.SetEnabled(true)
}

// 更新按钮文字
func serverBtnTxt() string {
	return "分发任务 (" + strconv.Itoa(serverCount) + ")"
}
