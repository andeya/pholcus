package main

import (
	// 按界面需求选择相应版本
	// "github.com/henrylee2cn/pholcus/web" // web版
	// "github.com/henrylee2cn/pholcus/cmd" // cmd版
	"github.com/henrylee2cn/pholcus/gui" // gui版

	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

// 导入自己的规则库（须保证最后声明，即最先导入）
import (
	_ "github.com/pholcus/spider_lib" // 此为公开维护的spider规则库
	// _ "path/myrule_lib" // 同样你也可以自由添加自己的规则库
)

// 自定义相关配置，将覆盖默认值
func setConf() {
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

func main() {
	// 开启错误日志调试功能（打印行号及Debug信息）
	logs.Debug(true)

	defer func() {
		if err := recover(); err != nil {
			logs.Log.Emergency("%v", err)
		}
	}()

	setConf() // 不调用则为默认值

	// 开始运行
	// web.Run() // web版
	// cmd.Run() // cmd版
	gui.Run() // gui版
}
