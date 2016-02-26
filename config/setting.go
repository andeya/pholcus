package config

import (
	"os"

	"github.com/henrylee2cn/pholcus/common/config"
)

var setting config.ConfigContainer = initConfig()

func initConfig() config.ConfigContainer {
	mkdir()

	name := APP_TAG + "/pholcus.ini"
	iniconf, err := config.NewConfig("ini", name)
	if err != nil {
		file, err := os.Create(name)
		file.Close()
		iniconf, err = config.NewConfig("ini", name)
		if err != nil {
			panic(err)
		}
		defaultConfig(iniconf)
		iniconf.SaveConfigFile(name)

	}
	return iniconf
}

func defaultConfig(iniconf config.ConfigContainer) {
	iniconf.Set("phantomjs", phantomjs)
	iniconf.Set("history", history)
	iniconf.Set("proxylib", proxylib)
	iniconf.Set("spiderlib", spiderlib)
	iniconf.Set("mgo::dbname", mgodbname)
	iniconf.Set("mgo::connstring", mgoconnstring)
	iniconf.Set("mysql::dbname", mysqldbname)
	iniconf.Set("mysql::connstring", mysqlconnstring)
}

func mkdir() {
	os.MkdirAll(spiderlib, 0777)
	os.MkdirAll(history, 0777)
}
