package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andeya/pholcus/logs/logs"
	"github.com/andeya/pholcus/runtime/status"
	"gopkg.in/ini.v1"
)

const testINI = `crawlcap   = 50
phantomjs  = pholcus_pkg/phantomjs
proxylib   = pholcus_pkg/proxy.lib
spiderdir  = dyn_rules
fileoutdir = pholcus_pkg/file_out
textoutdir = pholcus_pkg/text_out
dbname     = pholcus

[mgo]
connstring   = 127.0.0.1:27017
conncap      = 1024
conngcsecond = 600

[mysql]
connstring       = root:@tcp(127.0.0.1:3306)
conncap          = 2048
maxallowedpacket = 1048576

[beanstalkd]
host = localhost:11300
tube = pholcus

[kafka]
brokers = 127.0.0.1:9092

[log]
cap           = 10000
level         = debug
consolelevel  = info
feedbacklevel = error
lineinfo      = false
save          = true

[run]
mode        = -1
port        = 2015
master      = 127.0.0.1
thread      = 20
pause       = 300
outtype     = csv
batchcap    = 10000
limit       = 0
proxyminute = 0
success     = true
failure     = true
`

func TestMapTo(t *testing.T) {
	cfg, err := ini.Load([]byte(testINI))
	if err != nil {
		t.Fatalf("ini.Load: %v", err)
	}

	var c Config
	if err := cfg.MapTo(&c); err != nil {
		t.Fatalf("MapTo: %v", err)
	}

	if c.CrawlsCap != 50 {
		t.Errorf("CrawlsCap = %d, want 50", c.CrawlsCap)
	}
	if c.SpiderDir != "dyn_rules" {
		t.Errorf("SpiderDir = %q, want %q", c.SpiderDir, "dyn_rules")
	}
	if c.DBName != "pholcus" {
		t.Errorf("DBName = %q, want %q", c.DBName, "pholcus")
	}
	if c.Mgo.ConnStr != "127.0.0.1:27017" {
		t.Errorf("Mgo.ConnStr = %q, want %q", c.Mgo.ConnStr, "127.0.0.1:27017")
	}
	if c.Mgo.ConnCap != 1024 {
		t.Errorf("Mgo.ConnCap = %d, want 1024", c.Mgo.ConnCap)
	}
	if c.Log.LevelStr != "debug" {
		t.Errorf("Log.LevelStr = %q, want %q", c.Log.LevelStr, "debug")
	}
	if c.Log.Save != true {
		t.Errorf("Log.Save = %v, want true", c.Log.Save)
	}
	if c.Run.Mode != -1 {
		t.Errorf("Run.Mode = %d, want -1", c.Run.Mode)
	}
	if c.Run.Port != 2015 {
		t.Errorf("Run.Port = %d, want 2015", c.Run.Port)
	}
	if c.Run.SuccessInherit != true {
		t.Errorf("Run.SuccessInherit = %v, want true", c.Run.SuccessInherit)
	}
}

func TestDefaultConf(t *testing.T) {
	c := defaultConf()
	if c.CrawlsCap != 50 {
		t.Errorf("CrawlsCap = %d, want 50", c.CrawlsCap)
	}
	if c.PhantomJS != WorkRoot+"/phantomjs" {
		t.Errorf("PhantomJS = %q, want %q", c.PhantomJS, WorkRoot+"/phantomjs")
	}
	if c.ProxyFile != WorkRoot+"/proxy.lib" {
		t.Errorf("ProxyFile = %q, want %q", c.ProxyFile, WorkRoot+"/proxy.lib")
	}
	if c.SpiderDir != WorkRoot+"/spiders" {
		t.Errorf("SpiderDir = %q, want %q", c.SpiderDir, WorkRoot+"/spiders")
	}
	if c.FileDir != WorkRoot+"/file_out" {
		t.Errorf("FileDir = %q, want %q", c.FileDir, WorkRoot+"/file_out")
	}
	if c.TextDir != WorkRoot+"/text_out" {
		t.Errorf("TextDir = %q, want %q", c.TextDir, WorkRoot+"/text_out")
	}
	if c.DBName != Tag {
		t.Errorf("DBName = %q, want %q", c.DBName, Tag)
	}
	if c.Mgo.ConnStr != "127.0.0.1:27017" || c.Mgo.ConnCap != 1024 || c.Mgo.ConnGCSeconds != 600 {
		t.Errorf("Mgo = %+v", c.Mgo)
	}
	if c.MySQL.ConnStr != "root:@tcp(127.0.0.1:3306)" || c.MySQL.ConnCap != 2048 || c.MySQL.MaxAllowedPacket != 1048576 {
		t.Errorf("MySQL = %+v", c.MySQL)
	}
	if c.Beanstalkd.Host != "localhost:11300" || c.Beanstalkd.Tube != "pholcus" {
		t.Errorf("Beanstalkd = %+v", c.Beanstalkd)
	}
	if c.Kafka.Brokers != "127.0.0.1:9092" {
		t.Errorf("Kafka.Brokers = %q, want 127.0.0.1:9092", c.Kafka.Brokers)
	}
	if c.Log.Cap != 10000 || c.Log.LevelStr != "debug" || c.Log.ConsoleLevelStr != "info" || c.Log.FeedbackLevelStr != "error" || !c.Log.Save {
		t.Errorf("Log = %+v", c.Log)
	}
	if c.Run.Mode != status.UNSET || c.Run.Port != 2015 || c.Run.Master != "127.0.0.1" || c.Run.ThreadNum != 20 ||
		c.Run.Pausetime != 300 || c.Run.OutType != "csv" || c.Run.BatchCap != 10000 || !c.Run.SuccessInherit || !c.Run.FailureInherit {
		t.Errorf("Run = %+v", c.Run)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"app", logs.LevelApp},
		{"emergency", logs.LevelEmergency},
		{"alert", logs.LevelAlert},
		{"critical", logs.LevelCritical},
		{"error", logs.LevelError},
		{"warning", logs.LevelWarning},
		{"notice", logs.LevelNotice},
		{"informational", logs.LevelInformational},
		{"info", logs.LevelInformational},
		{"debug", logs.LevelDebug},
		{"DEBUG", logs.LevelDebug},
		{"INFO", logs.LevelInformational},
		{"unknown", logs.LevelDebug},
		{"", logs.LevelDebug},
	}
	for _, tt := range tests {
		if got := parseLogLevel(tt.in); got != tt.want {
			t.Errorf("parseLogLevel(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestLogConfigLevels(t *testing.T) {
	tests := []struct {
		name              string
		level             string
		consoleLevel      string
		feedbackLevel     string
		wantLevel         int
		wantConsoleLevel  int
		wantFeedbackLevel int
	}{
		{"debug-info-error", "debug", "info", "error", logs.LevelDebug, logs.LevelDebug, logs.LevelDebug},
		{"info-warning-error", "info", "warning", "error", logs.LevelInformational, logs.LevelInformational, logs.LevelInformational},
		{"console-above-global", "info", "debug", "error", logs.LevelInformational, logs.LevelDebug, logs.LevelInformational},
		{"feedback-above-global", "error", "error", "debug", logs.LevelError, logs.LevelError, logs.LevelDebug},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &LogConfig{LevelStr: tt.level, ConsoleLevelStr: tt.consoleLevel, FeedbackLevelStr: tt.feedbackLevel}
			if got := c.Level(); got != tt.wantLevel {
				t.Errorf("Level() = %d, want %d", got, tt.wantLevel)
			}
			if got := c.ConsoleLevel(); got != tt.wantConsoleLevel {
				t.Errorf("ConsoleLevel() = %d, want %d", got, tt.wantConsoleLevel)
			}
			if got := c.FeedbackLevel(); got != tt.wantFeedbackLevel {
				t.Errorf("FeedbackLevel() = %d, want %d", got, tt.wantFeedbackLevel)
			}
		})
	}
}

func TestConf(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, WorkRoot)
	if err := os.MkdirAll(configDir, 0777); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	configPath := filepath.Join(configDir, "config.ini")
	if err := os.WriteFile(configPath, []byte(testINI), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(orig)

	c := Conf()
	if c == nil {
		t.Fatal("Conf() returned nil")
	}
	if c.CrawlsCap != 50 {
		t.Errorf("CrawlsCap = %d, want 50", c.CrawlsCap)
	}
	if c.SpiderDir != "dyn_rules" {
		t.Errorf("SpiderDir = %q, want dyn_rules", c.SpiderDir)
	}
	if c.DBName != "pholcus" {
		t.Errorf("DBName = %q, want pholcus", c.DBName)
	}
	if c.Mgo.ConnStr != "127.0.0.1:27017" {
		t.Errorf("Mgo.ConnStr = %q, want 127.0.0.1:27017", c.Mgo.ConnStr)
	}
	if c.Run.Mode != -1 {
		t.Errorf("Run.Mode = %d, want -1", c.Run.Mode)
	}
	if c2 := Conf(); c2 != c {
		t.Error("Conf() should return same pointer on subsequent calls")
	}
}

func TestReflectFromAndReload(t *testing.T) {
	tmpFile := t.TempDir() + "/test_config.ini"

	orig := Config{
		CrawlsCap: 99,
		SpiderDir: "custom_rules",
		DBName:    "testdb",
		Mgo:       MgoConfig{ConnStr: "10.0.0.1:27017", ConnCap: 512, ConnGCSeconds: 300},
		Log:       LogConfig{Cap: 5000, LevelStr: "info", ConsoleLevelStr: "warning", FeedbackLevelStr: "error", Save: true},
		Run:       RunConfig{Mode: 0, Port: 8080, Master: "10.0.0.1", ThreadNum: 10, Pausetime: 500, OutType: "mysql", BatchCap: 5000, SuccessInherit: true, FailureInherit: false},
	}

	iniFile := ini.Empty()
	if err := ini.ReflectFrom(iniFile, &orig); err != nil {
		t.Fatalf("ReflectFrom: %v", err)
	}
	if err := iniFile.SaveTo(tmpFile); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	cfg, err := ini.Load(tmpFile)
	if err != nil {
		t.Fatalf("ini.Load: %v", err)
	}
	var loaded Config
	if err := cfg.MapTo(&loaded); err != nil {
		t.Fatalf("MapTo: %v", err)
	}

	if loaded.CrawlsCap != 99 {
		t.Errorf("CrawlsCap = %d, want 99", loaded.CrawlsCap)
	}
	if loaded.SpiderDir != "custom_rules" {
		t.Errorf("SpiderDir = %q, want %q", loaded.SpiderDir, "custom_rules")
	}
	if loaded.Mgo.ConnStr != "10.0.0.1:27017" {
		t.Errorf("Mgo.ConnStr = %q, want %q", loaded.Mgo.ConnStr, "10.0.0.1:27017")
	}
	if loaded.Log.LevelStr != "info" {
		t.Errorf("Log.LevelStr = %q, want %q", loaded.Log.LevelStr, "info")
	}
	if loaded.Run.Port != 8080 {
		t.Errorf("Run.Port = %d, want 8080", loaded.Run.Port)
	}
	if loaded.Run.FailureInherit != false {
		t.Errorf("Run.FailureInherit = %v, want false", loaded.Run.FailureInherit)
	}
}

func TestLoadSampleConfig(t *testing.T) {
	samplePath := filepath.Join("..", "sample", "pholcus_pkg", "config.ini")
	if _, err := os.Stat(samplePath); err != nil {
		t.Skipf("sample config not found: %v", err)
	}

	cfg, err := ini.Load(samplePath)
	if err != nil {
		t.Fatalf("ini.Load(%q): %v", samplePath, err)
	}

	var c Config
	if err := cfg.MapTo(&c); err != nil {
		t.Fatalf("MapTo: %v", err)
	}

	if c.CrawlsCap != 50 {
		t.Errorf("CrawlsCap = %d, want 50", c.CrawlsCap)
	}
	if c.SpiderDir != "dyn_rules" {
		t.Errorf("SpiderDir = %q, want %q", c.SpiderDir, "dyn_rules")
	}
	if c.Mgo.ConnCap != 1024 {
		t.Errorf("Mgo.ConnCap = %d, want 1024", c.Mgo.ConnCap)
	}
	if c.MySQL.MaxAllowedPacket != 1048576 {
		t.Errorf("MySQL.MaxAllowedPacket = %d, want 1048576", c.MySQL.MaxAllowedPacket)
	}
	if c.Beanstalkd.Host != "localhost:11300" {
		t.Errorf("Beanstalkd.Host = %q, want %q", c.Beanstalkd.Host, "localhost:11300")
	}
	if c.Kafka.Brokers != "127.0.0.1:9092" {
		t.Errorf("Kafka.Brokers = %q, want %q", c.Kafka.Brokers, "127.0.0.1:9092")
	}
	if c.Log.LevelStr != "debug" {
		t.Errorf("Log.LevelStr = %q, want %q", c.Log.LevelStr, "debug")
	}
	if c.Run.Mode != -1 {
		t.Errorf("Run.Mode = %d, want -1", c.Run.Mode)
	}
	if c.Run.SuccessInherit != true {
		t.Errorf("Run.SuccessInherit = %v, want true", c.Run.SuccessInherit)
	}
}
