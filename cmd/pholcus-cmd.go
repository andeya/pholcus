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

	// 创建蜘蛛队列
	sps := []*spider.Spider{}

	for _, idx := range strings.Split(*spiderflag, ",") {
		idx = strings.TrimSpace(idx)
		if idx == "" {
			continue
		}
		i, _ := strconv.Atoi(idx)
		sps = append(sps, app.LogicApp.GetSpiderLib()[i])
	}

	// 运行
	app.LogicApp.SpiderPrepare(sps).Run()

	if app.LogicApp.GetAppConf("Mode") == status.OFFLINE {
		return
	}

	select {}
}
