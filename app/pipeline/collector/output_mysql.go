package collector

import (
	"fmt"

	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/common/util"
)

/************************ Mysql 输出 ***************************/

func init() {
	Output["mysql"] = func(self *Collector, dataIndex int) error {
		db, err := mysql.DB()
		if err != nil {
			return fmt.Errorf("Mysql数据库链接失败: %v", err)
		}

		var (
			mysqls    = make(map[string]*mysql.MyTable)
			namespace = util.FileNameReplace(self.namespace())
		)

		for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
			subNamespace := util.FileNameReplace(self.subNamespace(datacell))
			var tName = namespace
			if subNamespace != "" {
				tName += "__" + subNamespace
			}
			if _, ok := mysqls[subNamespace]; !ok {
				mysqls[subNamespace] = mysql.New(db)
				mysqls[subNamespace].SetTableName(tName)
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
		return nil
	}
}
