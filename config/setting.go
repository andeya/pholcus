package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/config"
	"github.com/andeya/pholcus/runtime/status"
)

// Default configuration values from the config file.
const (
	crawlcap              int    = 50                          // Max spider pool capacity
	logcap                int64  = 10000                       // Log buffer capacity
	loglevel              string = "debug"                     // Global log level (also file output level)
	logconsolelevel       string = "info"                      // Console log level
	logfeedbacklevel      string = "error"                     // Client-to-server feedback log level
	loglineinfo           bool   = false                       // Whether to print line info in logs
	logsave               bool   = true                        // Whether to save all logs to local file
	phantomjs             string = WORK_ROOT + "/phantomjs"    // PhantomJS binary path
	proxylib              string = WORK_ROOT + "/proxy.lib"    // Proxy IP file path
	spiderdir             string = WORK_ROOT + "/spiders"      // Dynamic rule directory
	fileoutdir            string = WORK_ROOT + "/file_out"     // Output dir for files (images, HTML, etc.)
	textoutdir            string = WORK_ROOT + "/text_out"     // Output dir for text (excel/csv)
	dbname                string = TAG                         // Database name
	mgoconnstring         string = "127.0.0.1:27017"           // MongoDB connection string
	mgoconncap            int    = 1024                        // MongoDB connection pool size
	mgoconngcsecond       int64  = 600                         // MongoDB connection pool GC interval (seconds)
	mysqlconnstring       string = "root:@tcp(127.0.0.1:3306)" // MySQL connection string
	mysqlconncap          int    = 2048                        // MySQL connection pool size
	mysqlmaxallowedpacket int    = 1048576                     // MySQL max allowed packet (bytes, default 1MB)
	beanstalkHost         string = "localhost:11300"           // Beanstalkd default host (with port)
	beanstalkTube         string = "pholcus"                   // Beanstalkd default tube
	kafkabrokers          string = "127.0.0.1:9092"            // Kafka brokers (comma-separated)

	mode        int    = status.UNSET // Node role
	port        int    = 2015         // Master node port
	master      string = "127.0.0.1"  // Master node address (no port)
	thread      int    = 20           // Global max concurrency
	pause       int64  = 300          // Pause duration reference (ms, random: Pausetime/2 ~ Pausetime*2)
	outtype     string = "csv"        // Output type
	dockercap   int    = 10000        // Segment dump container capacity
	limit       int64  = 0            // Crawl limit; 0=unlimited; custom if rule sets initial LIMIT
	proxyminute int64  = 0            // Proxy IP rotation interval (minutes)
	success     bool   = true         // Inherit success history
	failure     bool   = true         // Inherit failure history
)

var setting = func() config.Configer {
	mustMkdirAll(HISTORY_DIR)
	mustMkdirAll(CACHE_DIR)
	mustMkdirAll(PHANTOMJS_TEMP)

	iniconfResult := config.NewConfig("ini", CONFIG)
	if iniconfResult.IsErr() {
		file := result.Ret(os.Create(CONFIG)).Unwrap()
		if r := result.RetVoid(file.Close()); r.IsErr() {
			log.Printf("[W] close config file: %v", r.UnwrapErr())
		}
		iniconfResult = config.NewConfig("ini", CONFIG)
		if iniconfResult.IsErr() {
			panic(iniconfResult.UnwrapErr())
		}
		defaultConfig(iniconfResult.Unwrap())
		iniconfResult.Unwrap().SaveConfigFile(CONFIG)
	} else {
		trySet(iniconfResult.Unwrap())
	}

	iniconf := iniconfResult.Unwrap()
	mustMkdirAll(iniconf.String("spiderdir").UnwrapOr(""))
	mustMkdirAll(iniconf.String("fileoutdir").UnwrapOr(""))
	mustMkdirAll(iniconf.String("textoutdir").UnwrapOr(""))

	return iniconf
}()

func mustMkdirAll(dir string) {
	if err := os.MkdirAll(filepath.Clean(dir), 0777); err != nil {
		log.Fatalf("[F] create directory %q: %v", dir, err)
	}
}

func defaultConfig(iniconf config.Configer) {
	_ = iniconf.Set("crawlcap", strconv.Itoa(crawlcap))
	// iniconf.Set("datachancap", strconv.Itoa(datachancap))
	_ = iniconf.Set("log::cap", strconv.FormatInt(logcap, 10))
	_ = iniconf.Set("log::level", loglevel)
	_ = iniconf.Set("log::consolelevel", logconsolelevel)
	_ = iniconf.Set("log::feedbacklevel", logfeedbacklevel)
	_ = iniconf.Set("log::lineinfo", fmt.Sprint(loglineinfo))
	_ = iniconf.Set("log::save", fmt.Sprint(logsave))
	_ = iniconf.Set("phantomjs", phantomjs)
	_ = iniconf.Set("proxylib", proxylib)
	_ = iniconf.Set("spiderdir", spiderdir)
	_ = iniconf.Set("fileoutdir", fileoutdir)
	_ = iniconf.Set("textoutdir", textoutdir)
	_ = iniconf.Set("dbname", dbname)
	_ = iniconf.Set("mgo::connstring", mgoconnstring)
	_ = iniconf.Set("mgo::conncap", strconv.Itoa(mgoconncap))
	_ = iniconf.Set("mgo::conngcsecond", strconv.FormatInt(mgoconngcsecond, 10))
	_ = iniconf.Set("mysql::connstring", mysqlconnstring)
	_ = iniconf.Set("mysql::conncap", strconv.Itoa(mysqlconncap))
	_ = iniconf.Set("mysql::maxallowedpacket", strconv.Itoa(mysqlmaxallowedpacket))
	_ = iniconf.Set("kafka::brokers", kafkabrokers)
	_ = iniconf.Set("run::mode", strconv.Itoa(mode))
	_ = iniconf.Set("run::port", strconv.Itoa(port))
	_ = iniconf.Set("run::master", master)
	_ = iniconf.Set("run::thread", strconv.Itoa(thread))
	_ = iniconf.Set("run::pause", strconv.FormatInt(pause, 10))
	_ = iniconf.Set("run::outtype", outtype)
	_ = iniconf.Set("run::dockercap", strconv.Itoa(dockercap))
	_ = iniconf.Set("run::limit", strconv.FormatInt(limit, 10))
	_ = iniconf.Set("run::proxyminute", strconv.FormatInt(proxyminute, 10))
	_ = iniconf.Set("run::success", fmt.Sprint(success))
	_ = iniconf.Set("run::failure", fmt.Sprint(failure))
}

func trySet(iniconf config.Configer) {
	if v := iniconf.Int("crawlcap").UnwrapOr(0); v <= 0 {
		_ = iniconf.Set("crawlcap", strconv.Itoa(crawlcap))
	}

	// if v := iniconf.Int("datachancap").UnwrapOr(0); v <= 0 {
	// 	iniconf.Set("datachancap", strconv.Itoa(datachancap))
	// }

	if v := iniconf.Int64("log::cap").UnwrapOr(int64(0)); v <= 0 {
		_ = iniconf.Set("log::cap", strconv.FormatInt(logcap, 10))
	}

	level := iniconf.String("log::level").UnwrapOr("")
	if logLevel(level) == -10 {
		level = loglevel
	}
	_ = iniconf.Set("log::level", level)

	consolelevel := iniconf.String("log::consolelevel").UnwrapOr("")
	if logLevel(consolelevel) == -10 {
		consolelevel = logconsolelevel
	}
	_ = iniconf.Set("log::consolelevel", logLevel2(consolelevel, level))

	feedbacklevel := iniconf.String("log::feedbacklevel").UnwrapOr("")
	if logLevel(feedbacklevel) == -10 {
		feedbacklevel = logfeedbacklevel
	}
	_ = iniconf.Set("log::feedbacklevel", logLevel2(feedbacklevel, level))

	if iniconf.Bool("log::lineinfo").IsErr() {
		_ = iniconf.Set("log::lineinfo", fmt.Sprint(loglineinfo))
	}

	if iniconf.Bool("log::save").IsErr() {
		_ = iniconf.Set("log::save", fmt.Sprint(logsave))
	}

	if iniconf.String("phantomjs").IsNone() || iniconf.String("phantomjs").UnwrapOr("") == "" {
		_ = iniconf.Set("phantomjs", phantomjs)
	}

	if iniconf.String("proxylib").IsNone() || iniconf.String("proxylib").UnwrapOr("") == "" {
		_ = iniconf.Set("proxylib", proxylib)
	}

	if iniconf.String("spiderdir").IsNone() || iniconf.String("spiderdir").UnwrapOr("") == "" {
		_ = iniconf.Set("spiderdir", spiderdir)
	}

	if iniconf.String("fileoutdir").IsNone() || iniconf.String("fileoutdir").UnwrapOr("") == "" {
		_ = iniconf.Set("fileoutdir", fileoutdir)
	}

	if iniconf.String("textoutdir").IsNone() || iniconf.String("textoutdir").UnwrapOr("") == "" {
		_ = iniconf.Set("textoutdir", textoutdir)
	}

	if iniconf.String("dbname").IsNone() || iniconf.String("dbname").UnwrapOr("") == "" {
		_ = iniconf.Set("dbname", dbname)
	}

	if iniconf.String("mgo::connstring").IsNone() || iniconf.String("mgo::connstring").UnwrapOr("") == "" {
		_ = iniconf.Set("mgo::connstring", mgoconnstring)
	}

	if v := iniconf.Int("mgo::conncap").UnwrapOr(0); v <= 0 {
		_ = iniconf.Set("mgo::conncap", strconv.Itoa(mgoconncap))
	}

	if v := iniconf.Int64("mgo::conngcsecond").UnwrapOr(int64(0)); v <= 0 {
		_ = iniconf.Set("mgo::conngcsecond", strconv.FormatInt(mgoconngcsecond, 10))
	}

	if iniconf.String("mysql::connstring").IsNone() || iniconf.String("mysql::connstring").UnwrapOr("") == "" {
		_ = iniconf.Set("mysql::connstring", mysqlconnstring)
	}

	if v := iniconf.Int("mysql::conncap").UnwrapOr(0); v <= 0 {
		_ = iniconf.Set("mysql::conncap", strconv.Itoa(mysqlconncap))
	}

	if v := iniconf.Int("mysql::maxallowedpacket").UnwrapOr(0); v <= 0 {
		_ = iniconf.Set("mysql::maxallowedpacket", strconv.Itoa(mysqlmaxallowedpacket))
	}

	if iniconf.String("kafka::brokers").IsNone() || iniconf.String("kafka::brokers").UnwrapOr("") == "" {
		_ = iniconf.Set("kafka::brokers", kafkabrokers)
	}

	if v := iniconf.Int("run::mode").UnwrapOr(-1); v < status.UNSET || v > status.CLIENT {
		_ = iniconf.Set("run::mode", strconv.Itoa(mode))
	}

	if v := iniconf.Int("run::port").UnwrapOr(0); v <= 0 {
		_ = iniconf.Set("run::port", strconv.Itoa(port))
	}

	if iniconf.String("run::master").IsNone() || iniconf.String("run::master").UnwrapOr("") == "" {
		_ = iniconf.Set("run::master", master)
	}

	if v := iniconf.Int("run::thread").UnwrapOr(0); v <= 0 {
		_ = iniconf.Set("run::thread", strconv.Itoa(thread))
	}

	if v := iniconf.Int64("run::pause").UnwrapOr(-1); v < 0 {
		_ = iniconf.Set("run::pause", strconv.FormatInt(pause, 10))
	}

	if iniconf.String("run::outtype").IsNone() || iniconf.String("run::outtype").UnwrapOr("") == "" {
		_ = iniconf.Set("run::outtype", outtype)
	}

	if v := iniconf.Int("run::dockercap").UnwrapOr(0); v <= 0 {
		_ = iniconf.Set("run::dockercap", strconv.Itoa(dockercap))
	}

	if v := iniconf.Int64("run::limit").UnwrapOr(-1); v < 0 {
		_ = iniconf.Set("run::limit", strconv.FormatInt(limit, 10))
	}

	if v := iniconf.Int64("run::proxyminute").UnwrapOr(int64(0)); v <= 0 {
		_ = iniconf.Set("run::proxyminute", strconv.FormatInt(proxyminute, 10))
	}

	if iniconf.Bool("run::success").IsErr() {
		_ = iniconf.Set("run::success", fmt.Sprint(success))
	}

	if iniconf.Bool("run::failure").IsErr() {
		_ = iniconf.Set("run::failure", fmt.Sprint(failure))
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
