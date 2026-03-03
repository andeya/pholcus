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
		Title:   config.FullName + "                                                          [ Run Mode -> Server ]",
		MinSize: declarative.Size{Width: 1100, Height: 700},
		Layout:  declarative.VBox{MarginsZero: true},
		Children: []declarative.Widget{

			declarative.Composite{
				AssignTo: &setting,
				Layout:   declarative.Grid{Columns: 2},
				Children: []declarative.Widget{
					// Task list
					declarative.TableView{
						ColumnSpan:            1,
						MinSize:               declarative.Size{Width: 550, Height: 450},
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
						MinSize:    declarative.Size{Width: 550, Height: 450},
						Children: []declarative.Widget{

							declarative.VSplitter{
								Children: []declarative.Widget{
									declarative.Label{
										Text: "Custom config (wrap each task in \"<>\" for multiple tasks)",
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

					// Required field validation
					declarative.LineErrorPresenter{
						AssignTo: &ep,
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

	// Initialize app
	Init()

	// Run window
	mw.Run()
}

// Start button click handler
func serverStart() {
	if err := db.Submit(); err != nil {
		logs.Log().Error("%v", err)
		return
	}

	// Read tasks
	Input.Spiders = spiderMenu.GetChecked()

	if len(Input.Spiders) == 0 {
		logs.Log().Warning(" *     Task list cannot be empty")
		return
	}

	// Save config
	SetTaskConf()

	runStopBtn.SetEnabled(false)
	runStopBtn.SetText("Dispatch tasks (···)")

	// Reset spider queue
	SpiderPrepare()

	// Dispatch tasks
	app.LogicApp.Run()

	serverCount++

	runStopBtn.SetText(serverBtnTxt())
	runStopBtn.SetEnabled(true)
}

// Update button text
func serverBtnTxt() string {
	return "Dispatch tasks (" + strconv.Itoa(serverCount) + ")"
}
