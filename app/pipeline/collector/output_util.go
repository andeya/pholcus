package collector

import (
	"github.com/henrylee2cn/pholcus/config"
	"time"
)

/************************ for database output ***************************/

// 返回数据库及集合名称
func dbOrTabName(c *Collector) (dbName, tableName string) {
	if v, ok := config.MGO_OUTPUT.DBClass[c.Spider.GetName()]; ok {
		switch config.MGO_OUTPUT.TableFmt {
		case "h":
			return v, time.Now().Format("2006-01-02-15")
		case "d":
			fallthrough
		default:
			return v, time.Now().Format("2006-01-02")
		}
	}
	return config.MGO_OUTPUT.DefaultDB, ""
}

// 当输出数据库为config.MGO_OUTPUT.DefaultDB时，使用tabName获取table名
func tabName(c *Collector, ruleName string) string {
	var k = c.Spider.GetKeyword()
	if k != "" {
		k = "-" + k
	}
	return c.Spider.GetName() + "-" + ruleName + k
}
