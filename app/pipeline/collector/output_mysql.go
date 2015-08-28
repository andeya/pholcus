package collector

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/henrylee2cn/pholcus/common/pool"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"strings"
)

/************************ Mysql 输出 ***************************/
var mysqlPool = pool.NewPool(new(mysqlFish), 1024)

type mysqlFish struct {
	*sql.DB
}

func (self *mysqlFish) New() pool.Fish {
	db, err := sql.Open("mysql", config.MYSQL_OUTPUT.User+":"+config.MYSQL_OUTPUT.Password+"@tcp("+config.MYSQL_OUTPUT.Host+")/"+config.MYSQL_OUTPUT.DefaultDB+"?charset=utf8")
	if err != nil {
		panic(err)
	}
	return &mysqlFish{DB: db}
}

// 判断连接有效性
func (self *mysqlFish) Usable() bool {
	if self.DB.Ping() != nil {
		return false
	}
	return true
}

// 自毁方法，在被资源池删除时调用
func (self *mysqlFish) Close() {
	self.DB.Close()
}

func (*mysqlFish) Clean() {}

func init() {
	Output["mysql"] = func(self *Collector, dataIndex int) {
		db := mysqlPool.GetOne().(*mysqlFish)
		defer mysqlPool.Free(db)

		var newMysql = new(myTable)

		for Name, Rule := range self.GetRules() {
			//跳过不输出的数据
			if len(Rule.GetOutFeild()) == 0 {
				continue
			}

			newMysql.setTableName("`" + tabName(self, Name) + "`")

			for _, title := range Rule.GetOutFeild() {
				newMysql.addColumn(title)
			}

			newMysql.addColumn("当前连接", "上级链接", "下载时间").
				create(db.DB)

			num := 0 //小计

			for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
				if datacell["RuleName"].(string) == Name {
					for _, title := range Rule.GetOutFeild() {
						vd := datacell["Data"].(map[string]interface{})
						if v, ok := vd[title].(string); ok || vd[title] == nil {
							newMysql.addRow(v)
						} else {
							newMysql.addRow(util.JsonString(vd[title]))
						}
					}
					newMysql.addRow(datacell["Url"].(string), datacell["ParentUrl"].(string), datacell["DownloadTime"].(string)).
						update(db.DB)

					num++
				}
			}
			newMysql = new(myTable)
		}
	}
}

//sql转换结构体
type myTable struct {
	tableName   string
	columnNames []string
	rowValues   []string
	sqlCode     string
}

//设置表名
func (self *myTable) setTableName(name string) *myTable {
	self.tableName = name
	return self
}

//设置表单列
func (self *myTable) addColumn(name ...string) *myTable {
	self.columnNames = append(self.columnNames, name...)
	return self
}

//生成"创建表单"的语句，执行前须保证setTableName()、addColumn()已经执行
func (self *myTable) create(db *sql.DB) {
	if self.tableName != "" {
		self.sqlCode = `create table if not exists ` + self.tableName + `(`
		self.sqlCode += ` id int(8) not null primary key auto_increment`

		if self.columnNames != nil {
			for _, rowValues := range self.columnNames {
				self.sqlCode += `,` + rowValues + ` varchar(255) not null`
			}
		}
		self.sqlCode += `);`
	}
	stmt, err := db.Prepare(self.sqlCode)
	util.CheckErr(err)

	_, err = stmt.Exec()
	util.CheckErr(err)
}

//设置插入的1行数据
func (self *myTable) addRow(value ...string) *myTable {
	self.rowValues = append(self.rowValues, value...)
	return self
}

//向sqlCode添加"插入1行数据"的语句，执行前须保证create()、addRow()已经执行
//insert into table1(field1,field2) values(rowValues[0],rowValues[1])
func (self *myTable) update(db *sql.DB) {
	if self.tableName != "" {
		self.sqlCode = `insert into ` + self.tableName + `(`
		if self.columnNames != nil {
			for _, v1 := range self.columnNames {
				self.sqlCode += "`" + v1 + "`" + `,`
			}
			self.sqlCode = string(self.sqlCode[:len(self.sqlCode)-1])
			self.sqlCode += `)values(`
		}
		if self.rowValues != nil {
			for _, v2 := range self.rowValues {
				v2 = strings.Replace(v2, `"`, `\"`, -1)
				self.sqlCode += `"` + v2 + `"` + `,`
			}
			self.sqlCode = string(self.sqlCode[:len(self.sqlCode)-1])
			self.sqlCode += `);`
		}
	}

	stmt, err := db.Prepare(self.sqlCode)
	util.CheckErr(err)

	_, err = stmt.Exec()
	util.CheckErr(err)

	// 清空临时数据
	self.rowValues = []string{}
}
