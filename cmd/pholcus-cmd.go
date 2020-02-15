// [spider frame (golang)] Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能；
//（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）。
// 命令行界面版。
package cmd

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

var (
	spiderflag *string
)

// 获取外部参数
func Flag() {
	// 分类说明
	flag.String("c ******************************************** only for cmd ******************************************** -c", "", "")

	// 蜘蛛列表
	spiderflag = flag.String(
		"c_spider",
		"",
		func() string {
			var spiderlist string
			for k, v := range app.LogicApp.GetSpiderLib() {
				spiderlist += "   [" + strconv.Itoa(k) + "] " + v.GetName() + "  " + v.GetDescription() + "\r\n"
			}
			return "   <蜘蛛列表: 选择多蜘蛛以 \",\" 间隔>\r\n" + spiderlist
		}())

	// 备注说明
	flag.String(
		"c_z",
		"",
		"CMD-EXAMPLE: $ pholcus -_ui=cmd -a_mode="+strconv.Itoa(status.OFFLINE)+" -c_spider=3,8 -a_outtype=csv -a_thread=20 -a_dockercap=5000 -a_pause=300 -a_proxyminute=0 -a_keyins=\"<pholcus><golang>\" -a_limit=10 -a_success=true -a_failure=true\n",
	)
}

// 执行入口
func Run() {
	app.LogicApp.Init(cache.Task.Mode, cache.Task.Port, cache.Task.Master)
	if cache.Task.Mode == status.UNSET {
		return
	}
	switch app.LogicApp.GetAppConf("Mode").(int) {
	case status.SERVER:
		for {
			parseInput()
			run()
		}
	case status.CLIENT:
		run()
		select {}
	default:
		run()
	}
}

// 运行
func run() {
	// 创建蜘蛛队列
	sps := []*spider.Spider{}
	*spiderflag = strings.TrimSpace(*spiderflag)
	if *spiderflag == "*" {
		sps = app.LogicApp.GetSpiderLib()

	} else {
		for _, idx := range strings.Split(*spiderflag, ",") {
			idx = strings.TrimSpace(idx)
			if idx == "" {
				continue
			}
			i, _ := strconv.Atoi(idx)
			sps = append(sps, app.LogicApp.GetSpiderLib()[i])
		}
	}

	app.LogicApp.SpiderPrepare(sps).Run()
}

// 服务器模式下接收添加任务的参数
func parseInput() {
	logs.Log.Informational("\n添加任务参数——必填：%v\n添加任务参数——必填可选：%v\n", "-c_spider", []string{
		"-a_keyins",
		"-a_limit",
		"-a_outtype",
		"-a_thread",
		"-a_pause",
		"-a_proxyminute",
		"-a_dockercap",
		"-a_success",
		"-a_failure"})
	logs.Log.Informational("\n添加任务：\n")
retry:
	*spiderflag = ""
	input := [12]string{}
	fmt.Scanln(&input[0], &input[1], &input[2], &input[3], &input[4], &input[5], &input[6], &input[7], &input[8], &input[9])
	if strings.Index(input[0], "=") < 4 {
		logs.Log.Informational("\n添加任务的参数不正确，请重新输入：")
		goto retry
	}
	for _, v := range input {
		i := strings.Index(v, "=")
		if i < 4 {
			continue
		}
		key, value := v[:i], v[i+1:]
		switch key {
		case "-a_keyins":
			cache.Task.Keyins = value
		case "-a_limit":
			limit, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				break
			}
			cache.Task.Limit = limit
		case "-a_outtype":
			cache.Task.OutType = value
		case "-a_thread":
			thread, err := strconv.Atoi(value)
			if err != nil {
				break
			}
			cache.Task.ThreadNum = thread
		case "-a_pause":
			pause, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				break
			}
			cache.Task.Pausetime = pause
		case "-a_proxyminute":
			proxyminute, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				break
			}
			cache.Task.ProxyMinute = proxyminute
		case "-a_dockercap":
			dockercap, err := strconv.Atoi(value)
			if err != nil {
				break
			}
			if dockercap < 1 {
				dockercap = 1
			}
			cache.Task.DockerCap = dockercap
		case "-a_success":
			if value == "true" {
				cache.Task.SuccessInherit = true
			} else if value == "false" {
				cache.Task.SuccessInherit = false
			}
		case "-a_failure":
			if value == "true" {
				cache.Task.FailureInherit = true
			} else if value == "false" {
				cache.Task.FailureInherit = false
			}
		case "-c_spider":
			*spiderflag = value
		default:
			logs.Log.Informational("\n不可含有未知参数，必填参数：%v\n可选参数：%v\n", "-c_spider", []string{
				"-a_keyins",
				"-a_limit",
				"-a_outtype",
				"-a_thread",
				"-a_pause",
				"-a_proxyminute",
				"-a_dockercap",
				"-a_success",
				"-a_failure"})
			goto retry
		}
	}
}
