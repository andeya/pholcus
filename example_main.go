package main

import (
	"github.com/henrylee2cn/pholcus/config"
	// 按界面需求选择相应版本
	// "github.com/henrylee2cn/pholcus/web" // web版
	// "github.com/henrylee2cn/pholcus/cmd" // cmd版
	"github.com/henrylee2cn/pholcus/gui" // gui版
)

// 导入自己的规则库（须保证最后声明，即最先导入）
import (
	_ "github.com/henrylee2cn/pholcus/spider_lib" // 此为公开维护的spider规则库
	// _ "path/myrule_lib" // 同样你也可以自由添加自己的规则库
)

// 自定义相关配置，将覆盖默认值
func setConf() {
	//mongodb数据库服务器
	config.MGO_URL = "127.0.0.1:27017"
	//mongodb数据库名称
	config.MGO_NAME = "temp-collection-tentinet"
	//mongodb数据库集合
	config.MGO_COLLECTION = "news"
	//mysql地址
	config.MYSQL_HOST = "127.0.0.1:3306"
	//msyql数据库
	config.MYSQL_DB = "pholcus"
	//mysql用户
	config.MYSQL_USER = "root"
	//mysql密码
	config.MYSQL_PW = ""
}

func main() {
	// setConf() // 不调用则为默认值

	// 开始运行
	// web.Run() // web版
	// cmd.Run() // cmd版
	gui.Run() // gui版
}
