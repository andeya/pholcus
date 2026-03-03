//go:build !coverage

package collector

import (
	"sync"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/mysql"
	"github.com/andeya/pholcus/common/util"
)

// --- MySQL Output ---

func init() {
	var (
		mysqlTable     = map[string]*mysql.Table{}
		mysqlTableLock sync.RWMutex
	)

	var getMysqlTable = func(name string) (*mysql.Table, bool) {
		mysqlTableLock.RLock()
		defer mysqlTableLock.RUnlock()
		tab, ok := mysqlTable[name]
		if ok {
			return tab.Clone(), true
		}
		return nil, false
	}

	var setMysqlTable = func(name string, tab *mysql.Table) {
		mysqlTableLock.Lock()
		mysqlTable[name] = tab
		mysqlTableLock.Unlock()
	}

	DataOutput["mysql"] = func(col *Collector) (r result.VoidResult) {
		defer r.Catch()
		_, err := mysql.DB()
		result.RetVoid(err).Unwrap()
		var (
			mysqls    = make(map[string]*mysql.Table)
			namespace = util.FileNameReplace(col.namespace())
		)
		for _, datacell := range col.dataBuf {
			subNamespace := util.FileNameReplace(col.subNamespace(datacell))
			tName := joinNamespaces(namespace, subNamespace)
			table, ok := mysqls[tName]
			if !ok {
				table, ok = getMysqlTable(tName)
				if ok {
					mysqls[tName] = table
				} else {
					table = mysql.New().Unwrap()
					table.SetTableName(tName)
					for _, title := range col.MustGetRule(datacell["RuleName"].(string)).ItemFields {
						table.AddColumn(title + ` MEDIUMTEXT`)
					}
					if col.Spider.OutDefaultField() {
						table.AddColumn(`Url VARCHAR(255)`, `ParentUrl VARCHAR(255)`, `DownloadTime VARCHAR(50)`)
					}
					table.Create().Unwrap()
					setMysqlTable(tName, table)
					mysqls[tName] = table
				}
			}
			data := []string{}
			for _, title := range col.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					data = append(data, v)
				} else {
					data = append(data, util.JSONString(vd[title]))
				}
			}
			if col.Spider.OutDefaultField() {
				data = append(data, datacell["Url"].(string), datacell["ParentUrl"].(string), datacell["DownloadTime"].(string))
			}
			table.AutoInsert(data)
		}
		for _, tab := range mysqls {
			tab.FlushInsert().Unwrap()
		}
		mysqls = nil
		return result.OkVoid()
	}
}
