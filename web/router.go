package web

import (
	"mime"
	"net/http"

	ws "github.com/henrylee2cn/pholcus/common/websocket"
)

func init() {
	mime.AddExtensionType(".css", "text/css")
}

// 路由
func Router() {
	// 设置websocket请求路由
	http.Handle("/ws", ws.Handler(wsHandle))
	// 设置websocket报告打印专用路由
	http.Handle("/ws/log", ws.Handler(wsLogHandle))
	//设置http访问的路由
	http.HandleFunc("/", web)
	//static file server

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(assetFS())))
	// http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("web/static/"))))
}
