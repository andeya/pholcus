// 命令行模式，目前仅实现简单的单机模式。
// 作为破砖引玉，更复杂的功能如服务端、客户端模式、暂停等可以自己模仿实现
package command

import (
	"flag"
	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/spider"
	// "bufio"
	"log"
	// "os"
	// "fmt"
	"strconv"
	"strings"
)

var LogicApp = app.New()

func Run() {

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
	sps := []spider.Spider{}
	if *spiderflag == "" {
		log.Println(" *     —— 亲，任务列表不能为空哦~")
		return
	}
	for _, idx := range strings.Split(*spiderflag, ",") {
		i, _ := strconv.Atoi(idx)
		sps = append(sps, *LogicApp.GetAllSpiders()[i])
	}
	// fmt.Println("输入配置", *outputflag, *spiderflag, *goroutineflag, *dockerflag, *pasetimeflag, *keywordflag, *maxpageflag)

	// 配置运行参数
	LogicApp.SetRunMode(0).
		SetThreadNum(*goroutineflag).
		SetDockerCap(*dockerflag).
		SetOutType(*outputflag).
		SetMaxPage(*maxpageflag).
		SetBaseSleeptime(uint(pase[0])).
		SetRandomSleepPeriod(uint(pase[1])).
		SetSpiderQueue(sps, keyword).
		Run()
}
