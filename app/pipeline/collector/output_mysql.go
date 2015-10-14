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

		var newMysql = new(mysql.MyTable)

		for Name, Rule := range self.GetRules() {
			//跳过不输出的数据
			if len(Rule.GetOutFeild()) == 0 {
				continue
			}

			newMysql.SetTableName("`" + tabName(self, Name) + "`")

			for _, title := range Rule.GetOutFeild() {
				newMysql.AddColumn(title)
			}

			newMysql.AddColumn("当前连接", "上级链接", "下载时间").
				Create(db.DB)

			num := 0 //小计

			for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
				if datacell["RuleName"].(string) == Name {
					for _, title := range Rule.GetOutFeild() {
						vd := datacell["Data"].(map[string]interface{})
						if v, ok := vd[title].(string); ok || vd[title] == nil {
							newMysql.AddRow(v)
						} else {
							newMysql.AddRow(util.JsonString(vd[title]))
						}
					}
					newMysql.AddRow(datacell["Url"].(string), datacell["ParentUrl"].(string), datacell["DownloadTime"].(string)).
						Update(db.DB)

					num++
				}
			}
			newMysql = new(mysql.MyTable)
		}
	}
}
