package web

import (
	"flag"
	"github.com/henrylee2cn/pholcus/reporter"
	"net/http"
	"strconv"
)

var webport string = ":"
var wsport string = ":"
var wslogport string = ":"

func init() {
	// web服务器端口号
	port := flag.Int("webport", 9090, "   <Web Server Port>\n")
	flag.Parse()
	webport += strconv.Itoa(*port)
	wsport += strconv.Itoa(*port + 1)
	wslogport += strconv.Itoa(*port + 2)
}

func Run() {
	// 预绑定路由
	Router()
	// 开启websocket
	go func() {
		reporter.Println("[pholcus] websocket server Running on ", wsport)
		if err := http.ListenAndServe(wsport, nil); err != nil {
			reporter.Fatal("Websocket ListenAndServe: ", err)
		}
	}()
	// 开启websocket log
	go func() {
		reporter.Println("[pholcus] websocket log server Running on ", wslogport)
		if err := http.ListenAndServe(wslogport, nil); err != nil {
			reporter.Fatal("Websocket Log ListenAndServe: ", err)
		}
	}()
	// 开启http
	reporter.Println("[pholcus] http server Running on ", webport)
	err := http.ListenAndServe(webport, nil) //设置监听的端口
	if err != nil {
		reporter.Fatal("Http ListenAndServe: ", err)
	}
}
