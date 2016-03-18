package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/henrylee2cn/pholcus/common/config"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

// 配置文件涉及的默认配置。
const (
	crawlcap        int    = 50                          // 蜘蛛池最大容量
	datachancap     int    = 2 << 14                     // 收集器容量(默认65536)
	logcap          int64  = 10000                       // 日志缓存的容量
	phantomjs       string = WORK_ROOT + "/phantomjs"    // phantomjs文件路径
	proxylib        string = WORK_ROOT + "/proxy.lib"    // 代理ip文件路径
	spiderdir       string = WORK_ROOT + "/spiders"      // 动态规则目录
	dbname          string = TAG                         // 数据库名称
	mgoconnstring   string = "127.0.0.1:27017"           // mongodb连接字符串
	mgoconncap      int    = 1024                        // mongodb连接池容量
	mysqlconnstring string = "root:@tcp(127.0.0.1:3306)" // mysql连接字符串
	mysqlconncap    int    = 2048                        // mysql连接池容量
	mode            int    = status.UNSET                // 节点角色
	port            int    = 2015                        // 主节点端口
	master          string = "127.0.0.1"                 // 服务器(主节点)地址，不含端口
	thread          int    = 20                          // 全局最大并发量
	pause           int64  = 300                         // 暂停时长参考/ms(随机: Pausetime/2 ~ Pausetime*2)
	outtype         string = "csv"                       // 输出方式
	dockercap       int    = 10000                       // 分段转储容器容量
	limit           int64  = 0                           // 采集上限，0为不限，若在规则中设置初始值为LIMIT则为自定义限制，否则默认限制请求数
	proxyminute     int64  = 0                           // 代理IP更换的间隔分钟数
	success         bool   = true                        // 继承历史成功记录
	failure         bool   = true                        // 继承历史失败记录
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

func trySet(iniconf config.ConfigContainer) {
	if v, e := iniconf.Int("crawlcap"); v <= 0 || e != nil {
		iniconf.Set("crawlcap", strconv.Itoa(crawlcap))
	}
	if v, e := iniconf.Int("datachancap"); v <= 0 || e != nil {
		iniconf.Set("datachancap", strconv.Itoa(datachancap))
	}
	if v, e := iniconf.Int64("logcap"); v <= 0 || e != nil {
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
	if v, e := iniconf.Int("mgo::conncap"); v <= 0 || e != nil {
		iniconf.Set("mgo::conncap", strconv.Itoa(mgoconncap))
	}
	if v := iniconf.String("mysql::connstring"); v == "" {
		iniconf.Set("mysql::connstring", mysqlconnstring)
	}
	if v, e := iniconf.Int("mysql::conncap"); v <= 0 || e != nil {
		iniconf.Set("mysql::conncap", strconv.Itoa(mysqlconncap))
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
