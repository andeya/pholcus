// [spider frame (golang)] Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能；
//（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）。
// 命令行界面版。
package cmd

import (
	"flag"
	"strconv"
	"strings"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

var (
	spiderflag         *string
	outputflag         *string
	threadflag         *int
	dockerflag         *int
	pauseflag          *int64
	proxyflag          *int64
	keywordflag        *string
	maxpageflag        *int64
	successInheritflag *bool
	failureInheritflag *bool
)

// 获取外部参数
func Flag() {
	// 分类说明
	flag.String("c . . . . . . . . . . . . .. . . . . . . . . . . only for cmd . . . . . . . . . . . . . .. . . . . . . . . . c", "cmd", "\r\n")

	// 自定义输入
	keywordflag = flag.String("c_keyword", "", "   <自定义输入 选填 多关键词以 \",\" 隔开>")

	// 蜘蛛列表
	spiderflag = flag.String("c_spider", "", func() string {
		var spiderlist string
		for k, v := range app.LogicApp.GetSpiderLib() {
			spiderlist += "   {" + strconv.Itoa(k) + "} " + v.GetName() + "  " + v.GetDescription() + "\r\n"
		}
		return "   <蜘蛛列表 选择多蜘蛛以 \",\" 间隔>\r\n" + spiderlist
	}())

	// 输出方式
	outputflag = flag.String("c_output", app.LogicApp.GetOutputLib()[0], func() string {
		var outputlib string
		for _, v := range app.LogicApp.GetOutputLib() {
			outputlib += "{" + v + "} "
		}
		return "   <输出方式> " + strings.TrimRight(outputlib, " ")
	}())

	// 并发协程数
	threadflag = flag.Int("c_thread", cache.Task.ThreadNum, "   <并发协程> {1~99999}\n")

	// 平均暂停时间
	pauseflag = flag.Int64("c_pause", cache.Task.Pausetime, "   <平均暂停时间/ms> {>=100} ")

	// 代理IP更换频率
	proxyflag = flag.Int64("c_proxy", cache.Task.ProxyMinute, "   <代理IP更换频率/m 为0时不使用代理> {>=0} ")

	// 分批输出
	dockerflag = flag.Int("c_docker", cache.Task.DockerCap, "   <分批输出> {1~5000000}")

	// 采集页数
	maxpageflag = flag.Int64("c_maxpage", 0, "   <采集页数> {>=0}")

	// 继承之前的去重记录
	successInheritflag = flag.Bool("c_inherit_y", true, "   <继承并保存成功记录 {true/false}>")
	failureInheritflag = flag.Bool("c_inherit_n", true, "   <继承并保存失败记录 {true/false}>")

	// 备注说明
	flag.String(
		"c_z",
		"cmd-example",
		" pholcus -a_ui=cmd -c_spider=3,8 -c_output=csv -c_thread=20 -c_docker=5000 -c_pause=300 -c_proxy=0 -c_keyword=pholcus,golang -c_maxpage=10 -c_inherit_y=true -c_inherit_n=true\r\n",
	)
}

// 执行入口
func Run() {
	app.LogicApp.Init(status.OFFLINE, 0, "")

	// //运行模式
	// modeflag := flag.Int("运行模式", 0, "*运行模式: [0] 单机    [1] 服务端    [2] 客户端\r\n")

	// //端口号，非单机模式填写
	// portflag := flag.Int("端口号", 0, "端口号: 只填写数字即可，不含冒号\r\n")

	// //主节点ip，客户端模式填写
	// masterflag := flag.String("服务端IP", "127.0.0.1", "主节点IP: 服务端IP地址，不含端口\r\n")

	// 转换关键词
	keyword := strings.Replace(*keywordflag, ",", "|", -1)

	// 创建蜘蛛队列
	sps := []*spider.Spider{}
	if *spiderflag == "" {
		logs.Log.Warning(" *     —— 亲，任务列表不能为空哦~")
		return
	}
	for _, idx := range strings.Split(*spiderflag, ",") {
		i, _ := strconv.Atoi(idx)
		sps = append(sps, app.LogicApp.GetSpiderLib()[i])
	}

	// 配置运行参数
	app.LogicApp.SetAppConf("ThreadNum", *threadflag).
		SetAppConf("DockerCap", *dockerflag).
		SetAppConf("OutType", *outputflag).
		SetAppConf("MaxPage", *maxpageflag).
		SetAppConf("Pausetime", *pauseflag).
		SetAppConf("ProxyMinute", *proxyflag).
		SetAppConf("Keywords", keyword).
		SetAppConf("SuccessInherit", *successInheritflag).
		SetAppConf("FailureInherit", *failureInheritflag).
		SpiderPrepare(sps).
		Run()
}
