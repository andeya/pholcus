package collector

import (
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
)

/************************ Mysql 输出 ***************************/

func init() {
	Output["mysql"] = func(self *Collector, dataIndex int) {
		db, ok := mysql.MysqlPool.GetOne().(*mysql.MysqlSrc)
		if !ok || db == nil {
			logs.Log.Error("链接Mysql数据库超时，无法输出！")
			return
		}
		defer mysql.MysqlPool.Free(db)

		var mysqls = make(map[string]*mysql.MyTable)
		var namespace = util.FileNameReplace(self.namespace())

		for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
			subNamespace := util.FileNameReplace(self.subNamespace(datacell))
			if _, ok := mysqls[subNamespace]; !ok {
				mysqls[subNamespace] = mysql.New(db.DB)
				mysqls[subNamespace].SetTableName(namespace + "__" + subNamespace)
				for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
					mysqls[subNamespace].AddColumn(title + ` MEDIUMTEXT`)
				}

				mysqls[subNamespace].
					AddColumn(`Url VARCHAR(255)`, `ParentUrl VARCHAR(255)`, `DownloadTime VARCHAR(50)`).
					Create()
			}

			for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					mysqls[subNamespace].AddRow(v)
				} else {
					mysqls[subNamespace].AddRow(util.JsonString(vd[title]))
				}
			}

			err := mysqls[subNamespace].
				AddRow(datacell["Url"].(string), datacell["ParentUrl"].(string), datacell["DownloadTime"].(string)).
				Update()
			util.CheckErr(err)
		}
	}
}
