package config

//****************************************全局配置*******************************************\\

const (
	//软件名
	APP_NAME = "Pholcus幽灵蛛数据采集_v0.5.2 （by henrylee2cn）"
	// 蜘蛛池容量
	CRAWLS_CAP = 50

	// 收集器容量
	DATA_CAP = 2 << 14 //65536

	// mongodb数据库服务器
	DB_URL = "127.0.0.1:27017"

	//mongodb数据库名称
	DB_NAME = "temp-collection-tentinet"

	//mongodb数据库集合
	DB_COLLECTION = "news"

	//mysql地址
	MYSQL_HOST = "127.0.0.1:3306"
	//msyql数据库
	MYSQL_DB = "pholcus"
	//mysql用户
	MYSQL_USER = "root"
	//mysql密码
	MYSQL_PW = ""
)
