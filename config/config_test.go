package config

import (
	"os"
	"path/filepath"
	"testing"

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
