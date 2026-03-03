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
		Title:   config.FullName + "                                                          [ Run Mode -> Standalone ]",
		MinSize: declarative.Size{1100, 700},
		Layout:  declarative.VBox{MarginsZero: true},
		Children: []declarative.Widget{

			declarative.Composite{
				AssignTo: &setting,
				Layout:   declarative.Grid{Columns: 2},
				Children: []declarative.Widget{
					// Task list
					declarative.TableView{
						ColumnSpan:            1,
						MinSize:               declarative.Size{550, 450},
						AlternatingRowBGColor: walk.RGB(255, 255, 224),
						CheckBoxes:            true,
						ColumnsOrderable:      true,
						Columns: []declarative.TableViewColumn{
							{Title: "#", Width: 45},
							{Title: "Task", Width: 110 /*, Format: "%.2f", Alignment: AlignFar*/},
							{Title: "Description", Width: 370},
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
										Text: "Custom config (wrap each task in \"<>\" for multiple tasks):",
									},
									declarative.LineEdit{
										Text: declarative.Bind("Keyins"),
									},
								},
							},

							declarative.VSplitter{
								Children: []declarative.Widget{
									declarative.Label{
										Text: "*Crawl limit (default URL count):",
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
										Text: "*Concurrency: (1~99999)",
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
										Text: "*Batch output size: (1~5,000,000 records)",
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
										Text: "*Pause duration reference:",
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
										Text: "*Proxy rotation interval:",
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
								Title:      "*Output type",
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
							// Required field validation
							declarative.LineErrorPresenter{
								AssignTo: &ep,
							},
						},
					},

					declarative.HSplitter{
						MaxSize: declarative.Size{220, 50},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Inherit success records",
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
								Text: "Inherit failure records",
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
								Text:      "Pause/Resume",
								AssignTo:  &pauseRecoverBtn,
								OnClicked: offlinePauseRecover,
							},
						},
					},
					declarative.VSplitter{
						MaxSize: declarative.Size{90, 50},
						Children: []declarative.Widget{
							declarative.PushButton{
								Text:      "Start",
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

	// Initialize app
	Init()

	// Run window
	mw.Run()
}

// Pause/Resume task
func offlinePauseRecover() {
	switch app.LogicApp.Status() {
	case status.RUN:
		pauseRecoverBtn.SetText("Resume")
	case status.PAUSE:
		pauseRecoverBtn.SetText("Pause")
	}
	app.LogicApp.PauseRecover()
}

// Start/Stop control
func offlineRunStop() {
	if !app.LogicApp.IsStopped() {
		go func() {
			runStopBtn.SetEnabled(false)
			runStopBtn.SetText("Stopping…")
			pauseRecoverBtn.SetVisible(false)
			pauseRecoverBtn.SetText("Pause")
			app.LogicApp.Stop()
			offlineResetBtn()
		}()
		return
	}

	if err := db.Submit(); err != nil {
		logs.Log().Error("%v", err)
		return
	}

	// Read tasks
	Input.Spiders = spiderMenu.GetChecked()

	runStopBtn.SetText("Stop")

	// Save config
	SetTaskConf()

	// Update spider queue
	SpiderPrepare()

	go func() {
		pauseRecoverBtn.SetText("Pause")
		pauseRecoverBtn.SetVisible(true)
		app.LogicApp.Run()
		offlineResetBtn()
		pauseRecoverBtn.SetVisible(false)
		pauseRecoverBtn.SetText("Pause")
	}()
}

// Reset button state in offline mode
func offlineResetBtn() {
	runStopBtn.SetEnabled(true)
	runStopBtn.SetText("Start")
}
