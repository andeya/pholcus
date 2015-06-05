package gui

import (
	. "github.com/henrylee2cn/pholcus/spiders"
)

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
	OutType:           "excel",
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

var (
	// 任务选项
	SpiderModel = NewGUISpiderModel([]*GUISpiderCore{
		&GUISpiderCore{
			Spider:      BaiduSearch,
			Description: "百度搜索结果 [www.baidu.com]",
		},
		&GUISpiderCore{
			Spider:      GoogleSearch,
			Description: "谷歌搜索结果 [www.google.com镜像]",
		},
		&GUISpiderCore{
			Spider:      TaobaoSearch,
			Description: "淘宝宝贝搜索结果 [s.taobao.com]",
		},
		&GUISpiderCore{
			Spider:      JDSearch,
			Description: "京东搜索结果 [search.jd.com]",
		},
		&GUISpiderCore{
			Spider:      AlibabaProduct,
			Description: "阿里巴巴产品搜索 [s.1688.com/selloffer/offer_search.htm]",
		},
		&GUISpiderCore{
			Spider:      Wangyi,
			Description: "网易排行榜新闻，含点击/跟帖排名 [Auto Page] [news.163.com/rank]",
		},
		&GUISpiderCore{
			Spider:      BaiduNews,
			Description: "百度RSS新闻，实现轮询更新 [Auto Page] [news.baidu.com]",
		},
		&GUISpiderCore{
			Spider:      Kaola,
			Description: "考拉海淘商品数据 [Auto Page] [www.kaola.com]",
		},
		&GUISpiderCore{
			Spider:      Shunfenghaitao,
			Description: "顺丰海淘商品数据 [Auto Page] [www.sfht.com]",
		},
		&GUISpiderCore{
			Spider:      Miyabaobei,
			Description: "蜜芽宝贝商品数据 [Auto Page] [www.miyabaobei.com]",
		},
		&GUISpiderCore{
			Spider:      Hollandandbarrett,
			Description: "Hollandand&Barrett商品数据 [Auto Page] [www.Hollandandbarrett.com]",
		},
	})

	// 暂停时间选项及输出类型选项
	GUIOpt = struct {
		OutType   []*KV
		SleepTime []*KV
	}{
		OutType: []*KV{
			{Key: "excel", String: "excel"},
			{Key: "csv", String: "csv"},
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
)
