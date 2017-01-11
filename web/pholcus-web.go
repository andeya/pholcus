// [spider frame (golang)] Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能；
//（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）。
// Web 界面版。
package web

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

var (
	ip         *string
	port       *int
	addr       string
	spiderMenu []map[string]string
)

// 获取外部参数
func Flag() {
	flag.String("b ******************************************** only for web ******************************************** -b", "", "")
	// web服务器IP与端口号
	ip = flag.String("b_ip", "0.0.0.0", "   <Web Server IP>")
	port = flag.Int("b_port", 9090, "   <Web Server Port>")
}

// 执行入口
func Run() {
	appInit()

	// web服务器地址
	addr = *ip + ":" + strconv.Itoa(*port)

	// 预绑定路由
	Router()

	log.Printf("[pholcus] Server running on %v\n", addr)

	// 自动打开web浏览器
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "http://localhost:"+strconv.Itoa(*port))
	case "darwin":
		cmd = exec.Command("open", "http://localhost:"+strconv.Itoa(*port))
	}
	if cmd != nil {
		go func() {
			log.Println("[pholcus] Open the default browser after two seconds...")
			time.Sleep(time.Second * 2)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}()
	}

	// 监听端口
	err := http.ListenAndServe(addr, nil) //设置监听的端口
	if err != nil {
		logs.Log.Emergency("ListenAndServe: %v", err)
	}
}

func appInit() {
	app.LogicApp.SetLog(Lsc).SetAppConf("Mode", cache.Task.Mode)

	spiderMenu = func() (spmenu []map[string]string) {
		// 获取蜘蛛家族
		for _, sp := range app.LogicApp.GetSpiderLib() {
			spmenu = append(spmenu, map[string]string{"name": sp.GetName(), "description": sp.GetDescription()})
		}
		return spmenu
	}()
}
