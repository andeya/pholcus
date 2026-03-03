//go:build windows

package gui

import (
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"

	"github.com/andeya/pholcus/gui/model"
	"github.com/andeya/pholcus/runtime/cache"
	"github.com/andeya/pholcus/runtime/status"
)

// InputConfig holds GUI input.
type InputConfig struct {
	Spiders []*model.GUISpider
	*cache.AppConf
	Pausetime   int64
	ProxyMinute int64
}

var (
	runStopBtn      *walk.PushButton
	pauseRecoverBtn *walk.PushButton
	setting         *walk.Composite
	mw              *walk.MainWindow
	runMode         *walk.GroupBox
	db              *walk.DataBinder
	ep              walk.ErrorPresenter
	mode            *walk.GroupBox
	host            *walk.Splitter
	spiderMenu      *model.SpiderMenu
)

var Input = &InputConfig{
	AppConf:     cache.Task,
	Pausetime:   cache.Task.Pausetime,
	ProxyMinute: cache.Task.ProxyMinute,
}

//****************************************GUI display config*******************************************\\

// Output options
var outputList []declarative.RadioButton

// KV is a key-value helper for dropdown menus.
type KV struct {
	Key   string
	Int   int
	Int64 int64
}

// GuiOpt holds pause time and run mode options.
var GuiOpt = struct {
	Mode        []*KV
	Pausetime   []*KV
	ProxyMinute []*KV
}{
	Mode: []*KV{
		{Key: "Standalone", Int: status.OFFLINE},
		{Key: "Server", Int: status.SERVER},
		{Key: "Client", Int: status.CLIENT},
	},
	Pausetime: []*KV{
		{Key: "No pause", Int64: 0},
		{Key: "0.1 sec", Int64: 100},
		{Key: "0.3 sec", Int64: 300},
		{Key: "0.5 sec", Int64: 500},
		{Key: "1 sec", Int64: 1000},
		{Key: "3 sec", Int64: 3000},
		{Key: "5 sec", Int64: 5000},
		{Key: "10 sec", Int64: 10000},
		{Key: "15 sec", Int64: 15000},
		{Key: "20 sec", Int64: 20000},
		{Key: "30 sec", Int64: 30000},
		{Key: "60 sec", Int64: 60000},
	},
	ProxyMinute: []*KV{
		{Key: "No proxy", Int64: 0},
		{Key: "1 min", Int64: 1},
		{Key: "3 min", Int64: 3},
		{Key: "5 min", Int64: 5},
		{Key: "10 min", Int64: 10},
		{Key: "15 min", Int64: 15},
		{Key: "20 min", Int64: 20},
		{Key: "30 min", Int64: 30},
		{Key: "45 min", Int64: 45},
		{Key: "60 min", Int64: 60},
		{Key: "120 min", Int64: 120},
		{Key: "180 min", Int64: 180},
	},
}
