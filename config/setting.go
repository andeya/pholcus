package config

import (
	"os"
	"strconv"

	"github.com/henrylee2cn/pholcus/common/config"
)

// 配置文件涉及的默认配置。
const (
	crawlcap        int    = 50                          // 蜘蛛池最大容量
	datachancap     int    = 2 << 14                     // 收集器容量(默认65536)
	logcap          int64  = 10000                       // 日志缓存的容量
	phantomjs       string = WORK_ROOT + "/phantomjs"    // phantomjs文件路径
	proxylib        string = WORK_ROOT + "/proxy.lib"    // 代理ip文件路径
	spiderdir       string = WORK_ROOT + "/spiders"      // 动态规则目录
	dbname          string = WORK_ROOT                   // 数据库名称
	mgoconnstring   string = "127.0.0.1:27017"           // mongodb连接字符串
	mgoconncap      int    = 1024                        // mongodb连接池容量
	mysqlconnstring string = "root:@tcp(127.0.0.1:3306)" // mysql连接字符串
	mysqlconncap    int    = 1024                        // mysql连接池容量
	port            int    = 2015                        // 主节点端口
	master          string = "127.0.0.1"                 // 服务器(主节点)地址，不含端口
)

var setting = func() config.ConfigContainer {
	os.MkdirAll(HISTORY_DIR, 0777)
	os.MkdirAll(CACHE_DIR, 0777)
	os.MkdirAll(PHANTOMJS_TEMP, 0777)
	os.MkdirAll(FILE_DIR, 0777)
	os.MkdirAll(TEXT_DIR, 0777)
	os.MkdirAll(TEXT_DIR, 0777)

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

	os.MkdirAll(iniconf.String("spiderdir"), 0777)

	return iniconf
}()

func defaultConfig(iniconf config.ConfigContainer) {
	iniconf.Set("crawlcap", strconv.Itoa(crawlcap))
	iniconf.Set("datachancap", strconv.Itoa(datachancap))
	iniconf.Set("logcap", strconv.FormatInt(logcap, 10))
	iniconf.Set("phantomjs", phantomjs)
	iniconf.Set("proxylib", proxylib)
	iniconf.Set("spiderdir", spiderdir)
	iniconf.Set("dbname", dbname)
	iniconf.Set("mgo::connstring", mgoconnstring)
	iniconf.Set("mgo::conncap", strconv.Itoa(mgoconncap))
	iniconf.Set("mysql::connstring", mysqlconnstring)
	iniconf.Set("mysql::conncap", strconv.Itoa(mysqlconncap))
	iniconf.Set("port", strconv.Itoa(port))
	iniconf.Set("master", master)
}

func trySet(iniconf config.ConfigContainer) {
	if v, e := iniconf.Int("crawlcap"); v == 0 || e != nil {
		iniconf.Set("crawlcap", strconv.Itoa(crawlcap))
	}
	if v, e := iniconf.Int("datachancap"); v == 0 || e != nil {
		iniconf.Set("datachancap", strconv.Itoa(datachancap))
	}
	if v, e := iniconf.Int64("logcap"); v == 0 || e != nil {
		iniconf.Set("logcap", strconv.FormatInt(logcap, 10))
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
	if v := iniconf.String("dbname"); v == "" {
		iniconf.Set("dbname", dbname)
	}
	if v := iniconf.String("mgo::connstring"); v == "" {
		iniconf.Set("mgo::connstring", mgoconnstring)
	}
	if v, e := iniconf.Int("mgo::conncap"); v == 0 || e != nil {
		iniconf.Set("mgo::conncap", strconv.Itoa(mgoconncap))
	}
	if v := iniconf.String("mysql::connstring"); v == "" {
		iniconf.Set("mysql::connstring", mysqlconnstring)
	}
	if v, e := iniconf.Int("mysql::conncap"); v == 0 || e != nil {
		iniconf.Set("mysql::conncap", strconv.Itoa(mysqlconncap))
	}
	if v, e := iniconf.Int("port"); v == 0 || e != nil {
		iniconf.Set("port", strconv.Itoa(port))
	}
	if v := iniconf.String("master"); v == "" {
		iniconf.Set("master", master)
	}
	iniconf.SaveConfigFile(CONFIG)
}
