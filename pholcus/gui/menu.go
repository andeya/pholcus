package gui

// GUI输入
type Inputor struct {
	ThreadNum         uint
	OutType           string
	BaseSleeptime     uint
	RandomSleepPeriod uint //随机暂停最大增益时长
	MaxPage           int
	Keywords          string //后期split()为slice
	Spiders           []*GUISpider
	DockerCap         uint
}

var Input = &Inputor{
	// 默认值
	ThreadNum:         20,
	OutType:           "csv",
	BaseSleeptime:     1000,
	RandomSleepPeriod: 3000,
	MaxPage:           100,
	DockerCap:         10000,
}

// GUI内容
// 下拉菜单辅助结构体
type KV struct {
	Key    string
	Uint   uint
	String string
}

// 暂停时间选项及输出类型选项
var GUIOpt = struct {
	OutType   []*KV
	SleepTime []*KV
}{
	OutType: []*KV{
		{Key: "csv", String: "csv"},
		{Key: "excel", String: "excel"},
		{Key: "mongoDB", String: "mongoDB"},
	},
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
}
