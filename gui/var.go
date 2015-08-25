package gui

import (
	"github.com/henrylee2cn/pholcus/app/spider"
	. "github.com/henrylee2cn/pholcus/gui/model"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

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
	spiderMenu      = NewSpiderMenu(spider.Menu)
)

// GUI输入
type Inputor struct {
	Keywords string //后期split()为slice
	Spiders  []*GUISpider
	*cache.TaskConf
	BaseSleeptime     uint
	RandomSleepPeriod uint
}

var Input = &Inputor{
	// 默认值
	TaskConf:          cache.Task,
	BaseSleeptime:     cache.Task.Pausetime[0],
	RandomSleepPeriod: cache.Task.Pausetime[1],
}

//****************************************GUI内容显示配置*******************************************\\

// 下拉菜单辅助结构体
type KV struct {
	Key    string
	Int    int
	Uint   uint
	String string
}

// 暂停时间选项及运行模式选项
var GuiOpt = struct {
	SleepTime []*KV
	RunMode   []*KV
}{
	SleepTime: []*KV{
		{Key: "无暂停", Uint: 0},
		{Key: "0.1 秒", Uint: 100},
		{Key: "0.3 秒", Uint: 300},
		{Key: "0.5 秒", Uint: 500},
		{Key: "1 秒", Uint: 1000},
		{Key: "3 秒", Uint: 3000},
		{Key: "5 秒", Uint: 5000},
		{Key: "10 秒", Uint: 10000},
		{Key: "15 秒", Uint: 15000},
		{Key: "20 秒", Uint: 20000},
		{Key: "30 秒", Uint: 30000},
		{Key: "60 秒", Uint: 60000},
	},
	RunMode: []*KV{
		{Key: "单机", Int: status.OFFLINE},
		{Key: "服务器", Int: status.SERVER},
		{Key: "客户端", Int: status.CLIENT},
	},
}

// 输出选项
var outputList = func() (o []declarative.RadioButton) {
	// 设置默认选择
	Input.TaskConf.OutType = LogicApp.GetOutputLib()[0]
	// 获取输出选项
	for _, out := range LogicApp.GetOutputLib() {
		o = append(o, declarative.RadioButton{Text: out, Value: out})
	}
	return
}()
