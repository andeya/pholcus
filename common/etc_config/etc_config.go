// Package etc_config implements config initialization of one spider.
package etc_config

import (
	"github.com/henrylee2cn/pholcus/common/config"
	"github.com/henrylee2cn/pholcus/common/util"
	"os"
)

// Configpath gets default config path like "WD/etc/main.conf".
func configpath() string {
	//wd, _ := os.Getwd()
	wd := os.Getenv("GOPATH")
	if wd == "" {
		panic("GOPATH is not setted in env.")
	}
	logpath := wd + "/etc/"
	filename := "main.conf"
	err := os.MkdirAll(logpath, 0755)
	if err != nil {
		panic("logpath error : " + logpath + "\n")
	}
	return logpath + filename
}

// Config is a config singleton object for one spider.
var conf *config.Config
var path string

// StartConf is used in Spider for initialization at first time.
func StartConf(configFilePath string) *config.Config {
	if configFilePath != "" && !util.IsFileExists(configFilePath) {
		panic("config path is not valiad:" + configFilePath)
	}

	path = configFilePath
	return Conf()
}

// Conf gets singleton instance of Config.
func Conf() *config.Config {
	if conf == nil {
		if path == "" {
			path = configpath()
		}
		conf = config.NewConfig().Load(path)
	}
	return conf
}
