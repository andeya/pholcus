// [spider frame (golang)] Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能；
//（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）。
// Web 界面版。
package web

import (
	"flag"
	"net/http"
	"runtime"
	"strconv"

	"github.com/henrylee2cn/pholcus/reporter"
)

var (
	ip   string
	port string

	addr      string
	wsAddr    string
	wslogAddr string
)

func init() {
	// web服务器端口号
	ip := flag.String("ip", "0.0.0.0", "   <Web Server IP>\n")
	port := flag.Int("port", 9090, "   <Web Server Port>\n")
	flag.Parse()

	addr = *ip + ":" + strconv.Itoa(*port)
	wsAddr = *ip + ":" + strconv.Itoa(*port+1)
	wslogAddr = *ip + ":" + strconv.Itoa(*port+2)
}

func Run() {
	// 开启最大核心数运行
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 预绑定路由
	Router()
	// 开启websocket
	go func() {
		reporter.Println("[pholcus] websocket server Running on ", wsAddr)
		if err := http.ListenAndServe(wsAddr, nil); err != nil {
			reporter.Fatal("Websocket ListenAndServe: ", err)
		}
	}()
	// 开启websocket log
	go func() {
		reporter.Println("[pholcus] websocket log server Running on ", wslogAddr)
		if err := http.ListenAndServe(wslogAddr, nil); err != nil {
			reporter.Fatal("Websocket Log ListenAndServe: ", err)
		}
	}()
	// 开启http
	reporter.Println("[pholcus] http server Running on ", addr)
	err := http.ListenAndServe(addr, nil) //设置监听的端口
	if err != nil {
		reporter.Fatal("Http ListenAndServe: ", err)
	}
}
