// [spider frame (golang)] Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能；
//（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）。
// 命令行界面版。
package cmd

import (
	"flag"
	// "bufio"
	// "os"
	// "fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

var LogicApp = app.New().Init(status.OFFLINE, 0, "")

func Run() {
	// 开启最大核心数运行
	runtime.GOMAXPROCS(runtime.NumCPU())

	// //运行模式
	// modeflag := flag.Int("运行模式", 0, "*运行模式: [0] 单机    [1] 服务端    [2] 客户端\r\n")

	// //端口号，非单机模式填写
	// portflag := flag.Int("端口号", 0, "端口号: 只填写数字即可，不含冒号\r\n")

	// //主节点ip，客户端模式填写
	// masterflag := flag.String("服务端IP", "127.0.0.1", "主节点IP: 服务端IP地址，不含端口\r\n")

	// 蜘蛛列表
	var spiderlist string
	for k, v := range LogicApp.GetAllSpiders() {
		spiderlist += "    {" + strconv.Itoa(k) + "} " + v.GetName() + "  " + v.GetDescription() + "\r\n"
	}
	spiderlist = "   【蜘蛛列表】   (选择多蜘蛛以\",\"间隔)\r\n\r\n" + spiderlist
	spiderflag := flag.String("spider", "", spiderlist+"\r\n")

	// 输出方式
	var outputlib string
	for _, v := range LogicApp.GetOutputLib() {
		outputlib += "{" + v + "} " + v + "    "
	}
	outputlib = strings.TrimRight(outputlib, "    ") + "\r\n"
	outputlib = "   【输出方式】   " + outputlib
	outputflag := flag.String("output", LogicApp.GetOutputLib()[0], outputlib)

	// 并发协程数
	goroutineflag := flag.Uint("go", 20, "   【并发协程】   {1~99999}\r\n")

	// 分批输出
	dockerflag := flag.Uint("docker", 10000, "   【分批输出】   每 {1~5000000} 条数据输出一次\r\n")

	// 暂停时间
	pasetimeflag := flag.String("pase", "1000,3000", "   【暂停时间】   格式如 {基准时间,随机增益} (单位ms)\r\n")

	// 自定义输入
	keywordflag := flag.String("kw", "", "   【自定义输入<选填>】   多关键词以\",\"隔开\r\n")

	// 采集页数
	maxpageflag := flag.Int("page", 0, "   【采集页数<选填>】\r\n")

	// 备注说明
	flag.String("z", "", "   【说明<非参数>】   各项参数值请参考{}中内容，同一参数包含多个值时以\",\"隔开\r\n\r\n  example：pholcus-cmd.exe -spider=3,8 -output=csv -go=500 -docker=5000 -pase=1000,3000 -kw=pholcus,golang -page=100\r\n")

	flag.Parse()

	// 转换关键词
	keyword := strings.Replace(*keywordflag, ",", "|", -1)

	// 获得暂停时间设置
	var pase [2]uint64
	ptf := strings.Split(*pasetimeflag, ",")
	pase[0], _ = strconv.ParseUint(ptf[0], 10, 64)
	pase[1], _ = strconv.ParseUint(ptf[1], 10, 64)

	// 创建蜘蛛队列
	sps := []*spider.Spider{}
	if *spiderflag == "" {
		logs.Log.Warning(" *     —— 亲，任务列表不能为空哦~")
		return
	}
	for _, idx := range strings.Split(*spiderflag, ",") {
		i, _ := strconv.Atoi(idx)
		sps = append(sps, LogicApp.GetAllSpiders()[i])
	}
	// fmt.Println("输入配置", *outputflag, *spiderflag, *goroutineflag, *dockerflag, *pasetimeflag, *keywordflag, *maxpageflag)

	// 配置运行参数
	LogicApp.SetThreadNum(*goroutineflag).
		SetDockerCap(*dockerflag).
		SetOutType(*outputflag).
		SetMaxPage(*maxpageflag).
		SetPausetime([2]uint{uint(pase[0]), uint(pase[1])}).
		SpiderPrepare(sps, keyword).
		Run()
}
