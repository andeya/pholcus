package config

import (
	"strings"

	"github.com/henrylee2cn/pholcus/logs/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

// 软件信息。
const (
	VERSION   string = "v1.2.0"                                      // 软件版本号
	AUTHOR    string = "henrylee2cn"                                 // 软件作者
	NAME      string = "Pholcus幽灵蛛数据采集"                              // 软件名
	FULL_NAME string = NAME + "_" + VERSION + " （by " + AUTHOR + "）" // 软件全称
	TAG       string = "pholcus"                                     // 软件标识符
	ICON_PNG  string = `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`
)

// 默认配置。
const (
	WORK_ROOT      string = TAG + "_pkg"                    // 运行时的目录名称
	CONFIG         string = WORK_ROOT + "/config.ini"       // 配置文件路径
	CACHE_DIR      string = WORK_ROOT + "/cache"            // 缓存文件目录
	LOG            string = WORK_ROOT + "/logs/pholcus.log" // 日志文件路径
	LOG_ASYNC      bool   = true                            // 是否异步输出日志
	PHANTOMJS_TEMP string = CACHE_DIR                       // Surfer-Phantom下载器：js文件临时目录
	HISTORY_TAG    string = "history"                       // 历史记录的标识符
	HISTORY_DIR    string = WORK_ROOT + "/" + HISTORY_TAG   // excel或csv输出方式下，历史记录目录
	SPIDER_EXT     string = ".pholcus.html"                 // 动态规则扩展名
)

// 来自配置文件的配置项。
var (
	CRAWLS_CAP int = setting.DefaultInt("crawlcap", crawlcap) // 蜘蛛池最大容量
	// DATA_CHAN_CAP            int    = setting.DefaultInt("datachancap", datachancap)                               // 收集器容量
	PHANTOMJS                string = setting.String("phantomjs")                                          // Surfer-Phantom下载器：phantomjs程序路径
	PROXY                    string = setting.String("proxylib")                                           // 代理IP文件路径
	SPIDER_DIR               string = setting.String("spiderdir")                                          // 动态规则目录
	FILE_DIR                 string = setting.String("fileoutdir")                                         // 文件（图片、HTML等）结果的输出目录
	TEXT_DIR                 string = setting.String("textoutdir")                                         // excel或csv输出方式下，文本结果的输出目录
	DB_NAME                  string = setting.String("dbname")                                             // 数据库名称
	MGO_CONN_STR             string = setting.String("mgo::connstring")                                    // mongodb连接字符串
	MGO_CONN_CAP             int    = setting.DefaultInt("mgo::conncap", mgoconncap)                       // mongodb连接池容量
	MGO_CONN_GC_SECOND       int64  = setting.DefaultInt64("mgo::conngcsecond", mgoconngcsecond)           // mongodb连接池GC时间，单位秒
	MYSQL_CONN_STR           string = setting.String("mysql::connstring")                                  // mysql连接字符串
	MYSQL_CONN_CAP           int    = setting.DefaultInt("mysql::conncap", mysqlconncap)                   // mysql连接池容量
	MYSQL_MAX_ALLOWED_PACKET int    = setting.DefaultInt("mysql::maxallowedpacket", mysqlmaxallowedpacket) // mysql通信缓冲区的最大长度

	KAFKA_BORKERS string = setting.DefaultString("kafka::brokers", kafkabrokers) //kafka brokers

	LOG_CAP            int64 = setting.DefaultInt64("log::cap", logcap)          // 日志缓存的容量
	LOG_LEVEL          int   = logLevel(setting.String("log::level"))            // 全局日志打印级别（亦是日志文件输出级别）
	LOG_CONSOLE_LEVEL  int   = logLevel(setting.String("log::consolelevel"))     // 日志在控制台的显示级别
	LOG_FEEDBACK_LEVEL int   = logLevel(setting.String("log::feedbacklevel"))    // 客户端反馈至服务端的日志级别
	LOG_LINEINFO       bool  = setting.DefaultBool("log::lineinfo", loglineinfo) // 日志是否打印行信息                                  // 客户端反馈至服务端的日志级别
	LOG_SAVE           bool  = setting.DefaultBool("log::save", logsave)         // 是否保存所有日志到本地文件
)

func init() {
	// 主要运行时参数的初始化
	cache.Task = &cache.AppConf{
		Mode:           setting.DefaultInt("run::mode", mode),                 // 节点角色
		Port:           setting.DefaultInt("run::port", port),                 // 主节点端口
		Master:         setting.String("run::master"),                         // 服务器(主节点)地址，不含端口
		ThreadNum:      setting.DefaultInt("run::thread", thread),             // 全局最大并发量
		Pausetime:      setting.DefaultInt64("run::pause", pause),             // 暂停时长参考/ms(随机: Pausetime/2 ~ Pausetime*2)
		OutType:        setting.String("run::outtype"),                        // 输出方式
		DockerCap:      setting.DefaultInt("run::dockercap", dockercap),       // 分段转储容器容量
		Limit:          setting.DefaultInt64("run::limit", limit),             // 采集上限，0为不限，若在规则中设置初始值为LIMIT则为自定义限制，否则默认限制请求数
		ProxyMinute:    setting.DefaultInt64("run::proxyminute", proxyminute), // 代理IP更换的间隔分钟数
		SuccessInherit: setting.DefaultBool("run::success", success),          // 继承历史成功记录
		FailureInherit: setting.DefaultBool("run::failure", failure),          // 继承历史失败记录
	}
}

func logLevel(l string) int {
	switch strings.ToLower(l) {
	case "app":
		return logs.LevelApp
	case "emergency":
		return logs.LevelEmergency
	case "alert":
		return logs.LevelAlert
	case "critical":
		return logs.LevelCritical
	case "error":
		return logs.LevelError
	case "warning":
		return logs.LevelWarning
	case "notice":
		return logs.LevelNotice
	case "informational":
		return logs.LevelInformational
	case "info":
		return logs.LevelInformational
	case "debug":
		return logs.LevelDebug
	}
	return -10
}
