// Package config 提供了软件配置、路径和运行参数的加载与管理功能。
package config

import (
	"strings"

	"github.com/andeya/pholcus/logs/logs"
	"github.com/andeya/pholcus/runtime/status"
)

// Software information.
const (
	Version  string = "v1.3.5"                                      // Version number
	Author   string = "andeya"                                      // Author
	Name     string = "Pholcus幽灵蛛数据采集"                              // Software name
	FullName string = Name + "_" + Version + " （by " + Author + "）" // Full name
	Tag      string = "pholcus"                                     // Identifier
)

// Path constants derived from Tag.
const (
	WorkRoot      string = Tag + "_pkg"                   // Runtime directory name
	ConfigFile    string = WorkRoot + "/config.ini"       // Config file path
	CacheDir      string = WorkRoot + "/cache"            // Cache directory
	LogPath       string = WorkRoot + "/logs/pholcus.log" // Log file path
	LogAsync      bool   = true                           // Whether to output logs asynchronously
	PhantomJSTemp string = CacheDir                       // Surfer-Phantom: temp dir for JS files
	HistoryTag    string = "history"                      // History record identifier
	HistoryDir    string = WorkRoot + "/" + HistoryTag    // History dir for excel/csv output
	SpiderExt     string = ".pholcus.xml"                 // Dynamic rule extension (recommended)
	SpiderExtOld  string = ".pholcus.html"                // Dynamic rule extension (legacy)
)

// Config holds all runtime-configurable values, initialized with defaults.
// Fields are overwritten by LoadConfig() from the INI file via struct tags.
type Config struct {
	CrawlsCap int    `ini:"crawlcap"`
	PhantomJS string `ini:"phantomjs"`
	ProxyFile string `ini:"proxylib"`
	SpiderDir string `ini:"spiderdir"`
	FileDir   string `ini:"fileoutdir"`
	TextDir   string `ini:"textoutdir"`
	DBName    string `ini:"dbname"`

	Mgo        MgoConfig        `ini:"mgo"`
	MySQL      MySQLConfig      `ini:"mysql"`
	Beanstalkd BeanstalkdConfig `ini:"beanstalkd"`
	Kafka      KafkaConfig      `ini:"kafka"`
	Log        LogConfig        `ini:"log"`
	Run        RunConfig        `ini:"run"`
}

type MgoConfig struct {
	ConnStr       string `ini:"connstring"`
	ConnCap       int    `ini:"conncap"`
	ConnGCSeconds int64  `ini:"conngcsecond"`
}

type MySQLConfig struct {
	ConnStr          string `ini:"connstring"`
	ConnCap          int    `ini:"conncap"`
	MaxAllowedPacket int    `ini:"maxallowedpacket"`
}

type BeanstalkdConfig struct {
	Host string `ini:"host"`
	Tube string `ini:"tube"`
}

type KafkaConfig struct {
	Brokers string `ini:"brokers"`
}

type LogConfig struct {
	Cap              int64  `ini:"cap"`
	LevelStr         string `ini:"level"`
	ConsoleLevelStr  string `ini:"consolelevel"`
	FeedbackLevelStr string `ini:"feedbacklevel"`
	LineInfo         bool   `ini:"lineinfo"`
	Save             bool   `ini:"save"`
}

// Level returns the global log level as int.
func (c *LogConfig) Level() int {
	return parseLogLevel(c.LevelStr)
}

// ConsoleLevel returns the console log level, clamped to at least the global level.
func (c *LogConfig) ConsoleLevel() int {
	if l := parseLogLevel(c.ConsoleLevelStr); l >= c.Level() {
		return l
	}
	return c.Level()
}

// FeedbackLevel returns the feedback log level, clamped to at least the global level.
func (c *LogConfig) FeedbackLevel() int {
	if l := parseLogLevel(c.FeedbackLevelStr); l >= c.Level() {
		return l
	}
	return c.Level()
}

type RunConfig struct {
	Mode           int    `ini:"mode"`
	Port           int    `ini:"port"`
	Master         string `ini:"master"`
	ThreadNum      int    `ini:"thread"`
	Pausetime      int64  `ini:"pause"`
	OutType        string `ini:"outtype"`
	BatchCap       int    `ini:"batchcap"`
	Limit          int64  `ini:"limit"`
	ProxyMinute    int64  `ini:"proxyminute"`
	SuccessInherit bool   `ini:"success"`
	FailureInherit bool   `ini:"failure"`
}

// defaultConf returns a Config populated with built-in defaults.
func defaultConf() Config {
	return Config{
		CrawlsCap: 50,
		PhantomJS: WorkRoot + "/phantomjs",
		ProxyFile: WorkRoot + "/proxy.lib",
		SpiderDir: WorkRoot + "/spiders",
		FileDir:   WorkRoot + "/file_out",
		TextDir:   WorkRoot + "/text_out",
		DBName:    Tag,
		Mgo: MgoConfig{
			ConnStr:       "127.0.0.1:27017",
			ConnCap:       1024,
			ConnGCSeconds: 600,
		},
		MySQL: MySQLConfig{
			ConnStr:          "root:@tcp(127.0.0.1:3306)",
			ConnCap:          2048,
			MaxAllowedPacket: 1048576,
		},
		Beanstalkd: BeanstalkdConfig{
			Host: "localhost:11300",
			Tube: "pholcus",
		},
		Kafka: KafkaConfig{
			Brokers: "127.0.0.1:9092",
		},
		Log: LogConfig{
			Cap:              10000,
			LevelStr:         "debug",
			ConsoleLevelStr:  "info",
			FeedbackLevelStr: "error",
			Save:             true,
		},
		Run: RunConfig{
			Mode:           status.UNSET,
			Port:           2015,
			Master:         "127.0.0.1",
			ThreadNum:      20,
			Pausetime:      300,
			OutType:        "csv",
			BatchCap:       10000,
			SuccessInherit: true,
			FailureInherit: true,
		},
	}
}

func parseLogLevel(l string) int {
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
	case "informational", "info":
		return logs.LevelInformational
	case "debug":
		return logs.LevelDebug
	}
	return logs.LevelDebug
}
