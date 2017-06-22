package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/henrylee2cn/pholcus/common/config"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

// 配置文件涉及的默认配置。
const (
	crawlcap int = 50 // 蜘蛛池最大容量
	// datachancap             int    = 2 << 14                     // 收集器容量(默认65536)
	logcap                int64  = 10000                       // 日志缓存的容量
	loglevel              string = "debug"                     // 全局日志打印级别（亦是日志文件输出级别）
	logconsolelevel       string = "info"                      // 日志在控制台的显示级别
	logfeedbacklevel      string = "error"                     // 客户端反馈至服务端的日志级别
	loglineinfo           bool   = false                       // 日志是否打印行信息
	logsave               bool   = true                        // 是否保存所有日志到本地文件
	phantomjs             string = WORK_ROOT + "/phantomjs"    // phantomjs文件路径
	proxylib              string = WORK_ROOT + "/proxy.lib"    // 代理ip文件路径
	spiderdir             string = WORK_ROOT + "/spiders"      // 动态规则目录
	fileoutdir            string = WORK_ROOT + "/file_out"     // 文件（图片、HTML等）结果的输出目录
	textoutdir            string = WORK_ROOT + "/text_out"     // excel或csv输出方式下，文本结果的输出目录
	dbname                string = TAG                         // 数据库名称
	mgoconnstring         string = "127.0.0.1:27017"           // mongodb连接字符串
	mgoconncap            int    = 1024                        // mongodb连接池容量
	mgoconngcsecond       int64  = 600                         // mongodb连接池GC时间，单位秒
	mysqlconnstring       string = "root:@tcp(127.0.0.1:3306)" // mysql连接字符串
	mysqlconncap          int    = 2048                        // mysql连接池容量
	mysqlmaxallowedpacket int    = 1048576                     //mysql通信缓冲区的最大长度，单位B，默认1MB
	kafkabrokers          string = "127.0.0.1:9092"            //kafka broker字符串,逗号分割

	mode        int    = status.UNSET // 节点角色
	port        int    = 2015         // 主节点端口
	master      string = "127.0.0.1"  // 服务器(主节点)地址，不含端口
	thread      int    = 20           // 全局最大并发量
	pause       int64  = 300          // 暂停时长参考/ms(随机: Pausetime/2 ~ Pausetime*2)
	outtype     string = "csv"        // 输出方式
	dockercap   int    = 10000        // 分段转储容器容量
	limit       int64  = 0            // 采集上限，0为不限，若在规则中设置初始值为LIMIT则为自定义限制，否则默认限制请求数
	proxyminute int64  = 0            // 代理IP更换的间隔分钟数
	success     bool   = true         // 继承历史成功记录
	failure     bool   = true         // 继承历史失败记录
)

var setting = func() config.Configer {
	os.MkdirAll(filepath.Clean(HISTORY_DIR), 0777)
	os.MkdirAll(filepath.Clean(CACHE_DIR), 0777)
	os.MkdirAll(filepath.Clean(PHANTOMJS_TEMP), 0777)

	iniconf, err := config.NewConfig("ini", CONFIG)
	if err != nil {
		file, err := os.Create(CONFIG)
		file.Close()
		iniconf, err = config.NewConfig("ini", CONFIG)
		if err != nil {
			panic(err)
		}
		defaultConfig(iniconf)
		iniconf.SaveConfigFile(CONFIG)
	} else {
		trySet(iniconf)
	}

	os.MkdirAll(filepath.Clean(iniconf.String("spiderdir")), 0777)
	os.MkdirAll(filepath.Clean(iniconf.String("fileoutdir")), 0777)
	os.MkdirAll(filepath.Clean(iniconf.String("textoutdir")), 0777)

	return iniconf
}()

func defaultConfig(iniconf config.Configer) {
	iniconf.Set("crawlcap", strconv.Itoa(crawlcap))
	// iniconf.Set("datachancap", strconv.Itoa(datachancap))
	iniconf.Set("log::cap", strconv.FormatInt(logcap, 10))
	iniconf.Set("log::level", loglevel)
	iniconf.Set("log::consolelevel", logconsolelevel)
	iniconf.Set("log::feedbacklevel", logfeedbacklevel)
	iniconf.Set("log::lineinfo", fmt.Sprint(loglineinfo))
	iniconf.Set("log::save", fmt.Sprint(logsave))
	iniconf.Set("phantomjs", phantomjs)
	iniconf.Set("proxylib", proxylib)
	iniconf.Set("spiderdir", spiderdir)
	iniconf.Set("fileoutdir", fileoutdir)
	iniconf.Set("textoutdir", textoutdir)
	iniconf.Set("dbname", dbname)
	iniconf.Set("mgo::connstring", mgoconnstring)
	iniconf.Set("mgo::conncap", strconv.Itoa(mgoconncap))
	iniconf.Set("mgo::conngcsecond", strconv.FormatInt(mgoconngcsecond, 10))
	iniconf.Set("mysql::connstring", mysqlconnstring)
	iniconf.Set("mysql::conncap", strconv.Itoa(mysqlconncap))
	iniconf.Set("mysql::maxallowedpacket", strconv.Itoa(mysqlmaxallowedpacket))
	iniconf.Set("kafka::brokers", kafkabrokers)
	iniconf.Set("run::mode", strconv.Itoa(mode))
	iniconf.Set("run::port", strconv.Itoa(port))
	iniconf.Set("run::master", master)
	iniconf.Set("run::thread", strconv.Itoa(thread))
	iniconf.Set("run::pause", strconv.FormatInt(pause, 10))
	iniconf.Set("run::outtype", outtype)
	iniconf.Set("run::dockercap", strconv.Itoa(dockercap))
	iniconf.Set("run::limit", strconv.FormatInt(limit, 10))
	iniconf.Set("run::proxyminute", strconv.FormatInt(proxyminute, 10))
	iniconf.Set("run::success", fmt.Sprint(success))
	iniconf.Set("run::failure", fmt.Sprint(failure))
}

func trySet(iniconf config.Configer) {
	if v, e := iniconf.Int("crawlcap"); v <= 0 || e != nil {
		iniconf.Set("crawlcap", strconv.Itoa(crawlcap))
	}

	// if v, e := iniconf.Int("datachancap"); v <= 0 || e != nil {
	// 	iniconf.Set("datachancap", strconv.Itoa(datachancap))
	// }

	if v, e := iniconf.Int64("log::cap"); v <= 0 || e != nil {
		iniconf.Set("log::cap", strconv.FormatInt(logcap, 10))
	}

	level := iniconf.String("log::level")
	if logLevel(level) == -10 {
		level = loglevel
	}
	iniconf.Set("log::level", level)

	consolelevel := iniconf.String("log::consolelevel")
	if logLevel(consolelevel) == -10 {
		consolelevel = logconsolelevel
	}
	iniconf.Set("log::consolelevel", logLevel2(consolelevel, level))

	feedbacklevel := iniconf.String("log::feedbacklevel")
	if logLevel(feedbacklevel) == -10 {
		feedbacklevel = logfeedbacklevel
	}
	iniconf.Set("log::feedbacklevel", logLevel2(feedbacklevel, level))

	if _, e := iniconf.Bool("log::lineinfo"); e != nil {
		iniconf.Set("log::lineinfo", fmt.Sprint(loglineinfo))
	}

	if _, e := iniconf.Bool("log::save"); e != nil {
		iniconf.Set("log::save", fmt.Sprint(logsave))
	}

	if v := iniconf.String("phantomjs"); v == "" {
		iniconf.Set("phantomjs", phantomjs)
	}

	if v := iniconf.String("proxylib"); v == "" {
		iniconf.Set("proxylib", proxylib)
	}

	if v := iniconf.String("spiderdir"); v == "" {
		iniconf.Set("spiderdir", spiderdir)
	}

	if v := iniconf.String("fileoutdir"); v == "" {
		iniconf.Set("fileoutdir", fileoutdir)
	}

	if v := iniconf.String("textoutdir"); v == "" {
		iniconf.Set("textoutdir", textoutdir)
	}

	if v := iniconf.String("dbname"); v == "" {
		iniconf.Set("dbname", dbname)
	}

	if v := iniconf.String("mgo::connstring"); v == "" {
		iniconf.Set("mgo::connstring", mgoconnstring)
	}

	if v, e := iniconf.Int("mgo::conncap"); v <= 0 || e != nil {
		iniconf.Set("mgo::conncap", strconv.Itoa(mgoconncap))
	}

	if v, e := iniconf.Int64("mgo::conngcsecond"); v <= 0 || e != nil {
		iniconf.Set("mgo::conngcsecond", strconv.FormatInt(mgoconngcsecond, 10))
	}

	if v := iniconf.String("mysql::connstring"); v == "" {
		iniconf.Set("mysql::connstring", mysqlconnstring)
	}

	if v, e := iniconf.Int("mysql::conncap"); v <= 0 || e != nil {
		iniconf.Set("mysql::conncap", strconv.Itoa(mysqlconncap))
	}

	if v, e := iniconf.Int("mysql::maxallowedpacket"); v <= 0 || e != nil {
		iniconf.Set("mysql::maxallowedpacket", strconv.Itoa(mysqlmaxallowedpacket))
	}

	if v := iniconf.String("kafka::brokers"); v == "" {
		iniconf.Set("kafka::brokers", kafkabrokers)
	}

	if v, e := iniconf.Int("run::mode"); v < status.UNSET || v > status.CLIENT || e != nil {
		iniconf.Set("run::mode", strconv.Itoa(mode))
	}

	if v, e := iniconf.Int("run::port"); v <= 0 || e != nil {
		iniconf.Set("run::port", strconv.Itoa(port))
	}

	if v := iniconf.String("run::master"); v == "" {
		iniconf.Set("run::master", master)
	}

	if v, e := iniconf.Int("run::thread"); v <= 0 || e != nil {
		iniconf.Set("run::thread", strconv.Itoa(thread))
	}

	if v, e := iniconf.Int64("run::pause"); v < 0 || e != nil {
		iniconf.Set("run::pause", strconv.FormatInt(pause, 10))
	}

	if v := iniconf.String("run::outtype"); v == "" {
		iniconf.Set("run::outtype", outtype)
	}

	if v, e := iniconf.Int("run::dockercap"); v <= 0 || e != nil {
		iniconf.Set("run::dockercap", strconv.Itoa(dockercap))
	}

	if v, e := iniconf.Int64("run::limit"); v < 0 || e != nil {
		iniconf.Set("run::limit", strconv.FormatInt(limit, 10))
	}

	if v, e := iniconf.Int64("run::proxyminute"); v <= 0 || e != nil {
		iniconf.Set("run::proxyminute", strconv.FormatInt(proxyminute, 10))
	}

	if _, e := iniconf.Bool("run::success"); e != nil {
		iniconf.Set("run::success", fmt.Sprint(success))
	}

	if _, e := iniconf.Bool("run::failure"); e != nil {
		iniconf.Set("run::failure", fmt.Sprint(failure))
	}

	iniconf.SaveConfigFile(CONFIG)
}

func logLevel2(l string, g string) string {
	a, b := logLevel(l), logLevel(g)
	if a < b {
		return l
	}
	return g
}
