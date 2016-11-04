package collector

import (
	"fmt"
	"sync"

	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
)

/************************ Mysql 输出 ***************************/

func init() {
	var (
		mysqlTable     = map[string]*mysql.MyTable{}
		mysqlTableLock sync.RWMutex
	)

	var getMysqlTable = func(name string) (*mysql.MyTable, bool) {
		mysqlTableLock.RLock()
		defer mysqlTableLock.RUnlock()
		tab, ok := mysqlTable[name]
		if ok {
			return tab.Clone(), true
		}
		return nil, false
	}

	var setMysqlTable = func(name string, tab *mysql.MyTable) {
		mysqlTableLock.Lock()
		mysqlTable[name] = tab
		mysqlTableLock.Unlock()
	}

	DataOutput["mysql"] = func(self *Collector) error {
		_, err := mysql.DB()
		if err != nil {
			return fmt.Errorf("Mysql数据库链接失败: %v", err)
		}
		var (
			mysqls    = make(map[string]*mysql.MyTable)
			namespace = util.FileNameReplace(self.namespace())
		)
		for _, datacell := range self.dataDocker {
			subNamespace := util.FileNameReplace(self.subNamespace(datacell))
			tName := joinNamespaces(namespace, subNamespace)
			table, ok := mysqls[tName]
			if !ok {
				table, ok = getMysqlTable(tName)
				if ok {
					mysqls[tName] = table
				} else {
					table = mysql.New()
					table.SetTableName(tName)
					for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
						table.AddColumn(title + ` MEDIUMTEXT`)
					}
					if self.Spider.OutDefaultField() {
						table.AddColumn(`Url VARCHAR(255)`, `ParentUrl VARCHAR(255)`, `DownloadTime VARCHAR(50)`)
					}
					if err := table.Create(); err != nil {
						logs.Log.Error("%v", err)
						continue
					} else {
						setMysqlTable(tName, table)
						mysqls[tName] = table
					}
				}
			}
			data := []string{}
			for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					data = append(data, v)
				} else {
					data = append(data, util.JsonString(vd[title]))
				}
			}
			if self.Spider.OutDefaultField() {
				data = append(data, datacell["Url"].(string), datacell["ParentUrl"].(string), datacell["DownloadTime"].(string))
			}
			table.AutoInsert(data)
		}
		for _, tab := range mysqls {
			util.CheckErr(tab.FlushInsert())
		}
		mysqls = nil
		return nil
	}
}
