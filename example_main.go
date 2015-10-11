package main

import (
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/exec"
	"github.com/henrylee2cn/pholcus/logs"

	_ "github.com/pholcus/spider_lib" // 此为公开维护的spider规则库
	// _ "path/myrule_lib"               // 同样你也可以自由添加自己的规则库
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
	//mongodb服务器地址
	config.MGO_OUTPUT.Host = "127.0.0.1:27017"
	// mongodb输出时的内容分类
	// key:蜘蛛规则清单
	// value:数据库名
	config.MGO_OUTPUT.DBClass = map[string]string{
		"百度RSS新闻": "1_1",
	}
	// mongodb输出时非默认数据库时以当前时间为集合名
	// h: 精确到小时 (格式 2015-08-28-09)
	// d: 精确到天 (格式 2015-08-28)
	config.MGO_OUTPUT.TableFmt = "d"

	//mysql服务器地址
	config.MYSQL_OUTPUT.Host = "127.0.0.1:3306"
	//msyql数据库
	config.MYSQL_OUTPUT.DefaultDB = "pholcus"
	//mysql用户
	config.MYSQL_OUTPUT.User = "root"
	//mysql密码
	config.MYSQL_OUTPUT.Password = ""
}
