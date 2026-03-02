package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/andeya/gust/result"
	"github.com/andeya/gust/syncutil"
	"github.com/andeya/pholcus/runtime/cache"
	"gopkg.in/ini.v1"
)

var lazyConf = syncutil.NewLazyValueWithFunc(doLoadConfig)

// Conf returns the global configuration pointer, loading from INI on first access.
func Conf() *Config {
	return lazyConf.GetPtr()
}

func doLoadConfig() result.Result[Config] {
	conf := defaultConf()

	for _, dir := range []string{HistoryDir, CacheDir, PhantomJSTemp} {
		if err := os.MkdirAll(filepath.Clean(dir), 0777); err != nil {
			log.Printf("[W] create dir %q: %v", dir, err)
		}
	}

	if _, err := os.Stat(ConfigFile); err == nil {
		if cfg, err := ini.Load(ConfigFile); err != nil {
			log.Printf("[W] load config %q: %v", ConfigFile, err)
		} else if err := cfg.MapTo(&conf); err != nil {
			log.Printf("[W] map config: %v", err)
		}
	}

	iniFile := ini.Empty()
	if err := ini.ReflectFrom(iniFile, &conf); err == nil {
		if err := iniFile.SaveTo(ConfigFile); err != nil {
			log.Printf("[W] save config file: %v", err)
		}
	}

	for _, dir := range []string{conf.SpiderDir, conf.FileDir, conf.TextDir} {
		if err := os.MkdirAll(filepath.Clean(dir), 0777); err != nil {
			log.Printf("[W] create dir %q: %v", dir, err)
		}
	}

	cache.Task = &cache.AppConf{
		Mode:           conf.Run.Mode,
		Port:           conf.Run.Port,
		Master:         conf.Run.Master,
		ThreadNum:      conf.Run.ThreadNum,
		Pausetime:      conf.Run.Pausetime,
		OutType:        conf.Run.OutType,
		BatchCap:       conf.Run.BatchCap,
		Limit:          conf.Run.Limit,
		ProxyMinute:    conf.Run.ProxyMinute,
		SuccessInherit: conf.Run.SuccessInherit,
		FailureInherit: conf.Run.FailureInherit,
	}
	return result.Ok(conf)
}
