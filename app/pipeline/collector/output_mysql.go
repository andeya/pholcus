package collector

import (
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/common/util"
)

/************************ Mysql 输出 ***************************/

func init() {
	Output["mysql"] = func(self *Collector, dataIndex int) {
		db := mysql.MysqlPool.GetOne().(*mysql.MysqlSrc)
		defer mysql.MysqlPool.Free(db)

		var mysqls = make(map[string]*mysql.MyTable)
		var namespace = util.FileNameReplace(self.namespace())

		for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
			subNamespace := util.FileNameReplace(self.subNamespace(datacell))
			if _, ok := mysqls[subNamespace]; !ok {
				mysqls[subNamespace] = mysql.New(db.DB)
				mysqls[subNamespace].SetTableName("`" + namespace + "__" + subNamespace + "`")
				for _, title := range self.GetRule(datacell["RuleName"].(string)).GetOutFeild() {
					mysqls[subNamespace].AddColumn(title)
				}

				mysqls[subNamespace].
					AddColumn("Url", "ParentUrl", "DownloadTime").
					Create()
			}

			for _, title := range self.GetRule(datacell["RuleName"].(string)).GetOutFeild() {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					mysqls[subNamespace].AddRow(v)
				} else {
					mysqls[subNamespace].AddRow(util.JsonString(vd[title]))
				}
			}

			mysqls[subNamespace].
				AddRow(datacell["Url"].(string), datacell["ParentUrl"].(string), datacell["DownloadTime"].(string)).
				Update()
		}
	}
}
