package main

import (
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/exec"
	// "github.com/henrylee2cn/pholcus/logs"

	_ "github.com/pholcus/spider_lib" // 此为公开维护的spider规则库
	// _ "spider_lib_pte" // 同样你也可以自由添加自己的规则库
)

func main() {
	// 设置运行时默认操作界面，并开始运行
	// 运行软件前，可设置 -a_ui 参数为"web"、"gui"或"cmd"，指定本次运行的操作界面
	// 其中"gui"仅支持Windows系统
	exec.DefaultRun("gui")
}

// 自定义相关配置，将覆盖默认值
func init() {
	// 允许日志打印行号
	// logs.ShowLineNum()

	//mongodb链接字符串
	config.MGO.CONN_STR = "127.0.0.1:27017"
	//mongodb数据库
	config.MGO.DB = "pholcus"
	//mongodb连接池容量
	config.MGO.MAX_CONNS = 1024

	//mysql服务器地址
	config.MYSQL.CONN_STR = "root:@tcp(127.0.0.1:3306)"
	//msyql数据库
	config.MYSQL.DB = "pholcus"
	//mysql连接池容量
	config.MYSQL.MAX_CONNS = 1024

	// 历史记录文件名前缀
	config.HISTORY.FILE_NAME_PREFIX = "history"

	// 代理IP完整文件名
	config.PROXY_FULL_FILE_NAME = "proxy.pholcus"

	// Surfer-Phantom下载器配置
	config.SURFER_PHANTOM.FULL_APP_NAME = "phantomjs" //phantomjs软件相对路径与名称
}
