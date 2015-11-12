package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/henrylee2cn/pholcus/common/pool"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"strings"
)

/************************ Mysql 输出 ***************************/
var MysqlPool = pool.NewPool(new(MysqlSrc), config.MYSQL.MAX_CONNS, 10)

type MysqlSrc struct {
	*sql.DB
}

func (self *MysqlSrc) New() pool.Src {
	db, err := sql.Open("mysql", config.MYSQL.CONN_STR+"/"+config.MYSQL.DB+"?charset=utf8")
	if err != nil {
		logs.Log.Error("Mysql：%v", err)
		return nil
	}
	return &MysqlSrc{DB: db}
}

// 判断连接是否失效
func (self *MysqlSrc) Expired() bool {
	if self.DB == nil || self.DB.Ping() != nil {
		return true
	}
	return false
}

// 自毁方法，在被资源池删除时调用
func (self *MysqlSrc) Close() {
	if self.DB == nil {
		return
	}
	self.DB.Close()
}

func (*MysqlSrc) Clean() {}

//sql转换结构体
type MyTable struct {
	tableName   string
	columnNames []string
	rowValues   []string
	sqlCode     string
	*sql.DB
}

func New(db *sql.DB) *MyTable {
	return &MyTable{
		DB: db,
	}
}

//设置表名
func (self *MyTable) SetTableName(name string) *MyTable {
	self.tableName = name
	return self
}

//设置表单列
func (self *MyTable) AddColumn(name ...string) *MyTable {
	self.columnNames = append(self.columnNames, name...)
	return self
}

//生成"创建表单"的语句，执行前须保证SetTableName()、AddColumn()已经执行
func (self *MyTable) Create() {
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
	stmt, err := self.DB.Prepare(self.sqlCode)
	util.CheckErr(err)

	_, err = stmt.Exec()
	util.CheckErr(err)
}

//设置插入的1行数据
func (self *MyTable) AddRow(value ...string) *MyTable {
	self.rowValues = append(self.rowValues, value...)
	return self
}

//向sqlCode添加"插入1行数据"的语句，执行前须保证Create()、AddRow()已经执行
//insert into table1(field1,field2) values(rowValues[0],rowValues[1])
func (self *MyTable) Update() {
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

	stmt, err := self.DB.Prepare(self.sqlCode)
	util.CheckErr(err)

	_, err = stmt.Exec()
	util.CheckErr(err)

	// 清空临时数据
	self.rowValues = []string{}
}
