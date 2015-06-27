package model

import (
	. "github.com/henrylee2cn/pholcus/node"
	"github.com/henrylee2cn/pholcus/node/task"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/spider"
)

func WTaskConf1() {
	cache.Task.RunMode = Input.RunMode // 节点角色
	cache.Task.Port = Input.Port       // 主节点端口
	cache.Task.Master = Input.Master   //服务器(主节点)地址，不含端口
}

func WTaskConf2() {
	// 纠正协程数
	if Input.ThreadNum == 0 {
		Input.ThreadNum = 1
	}
	cache.Task.ThreadNum = Input.ThreadNum
	cache.Task.BaseSleeptime = Input.BaseSleeptime
	cache.Task.RandomSleepPeriod = Input.RandomSleepPeriod //随机暂停最大增益时长
	cache.Task.OutType = Input.OutType
	cache.Task.DockerCap = Input.DockerCap //分段转储容器容量
	// 选填项
	cache.Task.MaxPage = Input.MaxPage
	cache.AutoDockerQueueCap()
}

// 根据GUI提交信息生成蜘蛛列表
func InitSpiders() int {
	Pholcus.Spiders.Reset()

	// 遍历任务
	for _, sps := range Input.Spiders {
		sps.Spider.SetPausetime(Input.BaseSleeptime, Input.RandomSleepPeriod)
		sps.Spider.SetMaxPage(Input.MaxPage)
		Pholcus.Spiders.Add(sps.Spider)
	}

	// 遍历关键词
	Pholcus.Spiders.AddKeywords(Input.Keywords)

	return Pholcus.Spiders.Len()
}

// 从task准备运行条件
func TaskToReady(t *task.Task) {
	// 清空历史任务
	Pholcus.Spiders.Reset()

	// 更改全局配置
	cache.Task.OutType = t.OutType
	cache.Task.ThreadNum = t.ThreadNum
	cache.Task.DockerCap = t.DockerCap
	cache.Task.DockerQueueCap = t.DockerQueueCap

	// 初始化蜘蛛队列
	for _, n := range t.Spiders {
		if sp := spider.Menu.GetByName(n); sp != nil {
			sp.SetPausetime(t.BaseSleeptime, t.RandomSleepPeriod)
			sp.SetMaxPage(t.MaxPage)
			Pholcus.Spiders.Add(sp)
		}
	}
	Pholcus.Spiders.AddKeywords(t.Keywords)
}
