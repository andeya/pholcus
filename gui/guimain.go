package gui

import (
	"github.com/henrylee2cn/pholcus/gui/app"
	. "github.com/henrylee2cn/pholcus/gui/model"
	"github.com/henrylee2cn/pholcus/spider"
	_ "github.com/henrylee2cn/pholcus/spider/spiders"
)

var LogicApp = app.New()

func Run() {
	runmodeWindow()
}

func WTaskConf1() {
	LogicApp.SetRunMode(Input.RunMode) // 节点角色
	LogicApp.SetPort(Input.Port)       // 主节点端口
	LogicApp.SetMaster(Input.Master)   //服务器(主节点)地址，不含端口
}

func WTaskConf2() {
	// 纠正协程数
	if Input.ThreadNum == 0 {
		Input.ThreadNum = 1
	}
	LogicApp.SetThreadNum(Input.ThreadNum)
	LogicApp.SetBaseSleeptime(Input.BaseSleeptime)
	LogicApp.SetRandomSleepPeriod(Input.RandomSleepPeriod) //随机暂停最大增益时长
	LogicApp.SetOutType(Input.OutType)
	LogicApp.SetDockerCap(Input.DockerCap) //分段转储容器容量
	// 选填项
	LogicApp.SetMaxPage(Input.MaxPage)
}

func SetSpiderQueue() {
	sps := []spider.Spider{}
	for _, sp := range Input.Spiders {
		sps = append(sps, *sp.Spider)
	}
	LogicApp.SetSpiderQueue(sps, Input.Keywords)
}

func SpiderNames() (names []string) {
	for _, sp := range Input.Spiders {
		names = append(names, sp.Spider.GetName())
	}
	return
}
