// [spider frame (golang)] Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能；
//（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）。
// 命令行界面版。
package cmd

import (
	"flag"
	// "bufio"
	// "os"
	// "fmt"
	"strconv"
	"strings"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

var (
	spiderflag             *string
	outputflag             *string
	goroutineflag          *uint
	dockerflag             *uint
	pauseflag              *string
	keywordflag            *string
	maxpageflag            *int
	inheritDeduplicateflag *bool
)

// 获取外部参数
func Flag() {
	// 分类说明
	flag.String("c . . . . . . . . . . . . .. . . . . . . . . . . only for cmd . . . . . . . . . . . . . .. . . . . . . . . . c", "cmd", "\r\n")

	// 蜘蛛列表
	var spiderlist string
	for k, v := range app.LogicApp.GetSpiderLib() {
		spiderlist += "   {" + strconv.Itoa(k) + "} " + v.GetName() + "  " + v.GetDescription() + "\r\n"
	}
	spiderlist = "   <蜘蛛列表 选择多蜘蛛以 \",\" 间隔>\r\n" + spiderlist
	spiderflag = flag.String("c_spider", "", spiderlist)

	// 输出方式
	var outputlib string
	for _, v := range app.LogicApp.GetOutputLib() {
		outputlib += "{" + v + "} "
	}
	outputlib = "   <输出方式> " + strings.TrimRight(outputlib, " ")
	outputflag = flag.String("c_output", app.LogicApp.GetOutputLib()[0], outputlib)

	// 并发协程数
	goroutineflag = flag.Uint("c_goroutine", 20, "   <并发协程> {1~99999}")

	// 分批输出
	dockerflag = flag.Uint("c_docker", 10000, "   <分批输出> {1~5000000}")

	// 暂停时间
	pauseflag = flag.String("c_pause", "1000,3000", "   <暂停时间/ms> {基准时间,随机增益} ")

	// 自定义输入
	keywordflag = flag.String("c_keyword", "", "   <自定义输入 选填 多关键词以 \",\" 隔开>")

	// 采集页数
	maxpageflag = flag.Int("c_maxpage", 0, "   <采集页数 选填>")

	// 继承之前的去重记录
	inheritDeduplicateflag = flag.Bool("c_inheritDeduplicate", true, "   <继承历史去重样本>")

	// 备注说明
	flag.String("c_z", "cmd-example", " pholcus -a_ui=web -c_spider=3,8 -c_output=csv -c_goroutine=500 -c_docker=5000 -c_pause=1000,3000 -c_keyword=pholcus,golang -c_maxpage=100 -c_inheritDeduplicate=true\r\n")
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

	// 获得暂停时间设置
	var pause [2]uint64
	ptf := strings.Split(*pauseflag, ",")
	pause[0], _ = strconv.ParseUint(ptf[0], 10, 64)
	pause[1], _ = strconv.ParseUint(ptf[1], 10, 64)

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
	app.LogicApp.SetAppConf("ThreadNum", *goroutineflag).
		SetAppConf("DockerCap", *dockerflag).
		SetAppConf("OutType", *outputflag).
		SetAppConf("MaxPage", *maxpageflag).
		SetAppConf("Pausetime", [2]uint{uint(pause[0]), uint(pause[1])}).
		SetAppConf("Keywords", keyword).
		SetAppConf("InheritDeduplication", *inheritDeduplicateflag).
		SpiderPrepare(sps).
		Run()
}
