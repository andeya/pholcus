package main

import (
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/exec"
	"github.com/henrylee2cn/pholcus/logs"

	_ "github.com/pholcus/spider_lib" // 此为公开维护的spider规则库
	// _ "github.com/my_lib" // 同样你也可以自由添加自己的规则库
)

func main() {
	// 允许日志打印行号
	logs.ShowLineNum()

	// 初始化配置，不调用则为默认值
	SetConf()

	// 开始运行，参数："web"/"gui"/"cmd"，默认为web版
	// 其中gui版仅支持Windows系统
	exec.Run("web")
}

// 自定义相关配置，将覆盖默认值
func SetConf() {
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

	// Surfer-Phantom下载器配置
	config.SURFER_PHANTOM.FULL_APP_NAME = "phantomjs" //phantomjs软件相对路径与名称
}
