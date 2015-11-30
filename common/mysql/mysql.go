package mysql

import (
	"database/sql"
	"errors"
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
	tableName        string
	columnNames      [][2]string
	rowValues        []string
	sqlCode          string
	customPrimaryKey bool
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
func (self *MyTable) AddColumn(names ...string) *MyTable {
	for _, name := range names {
		name = strings.Trim(name, " ")
		idx := strings.Index(name, " ")
		self.columnNames = append(self.columnNames, [2]string{string(name[:idx]), string(name[idx+1:])})
	}
	return self
}

//设置主键的语句（可选）
func (self *MyTable) CustomPrimaryKey(primaryKeyCode string) *MyTable {
	self.AddColumn(primaryKeyCode)
	self.customPrimaryKey = true
	return self
}

//生成"创建表单"的语句，执行前须保证SetTableName()、AddColumn()已经执行
func (self *MyTable) Create() *MyTable {
	if len(self.columnNames) == 0 {
		return self
	}
	self.sqlCode = `create table if not exists ` + self.tableName + `(`
	if !self.customPrimaryKey {
		self.sqlCode += `id int(12) not null primary key auto_increment,`
	}
	for _, rowValues := range self.columnNames {
		self.sqlCode += rowValues[0] + ` ` + rowValues[1] + `,`
	}
	self.sqlCode = string(self.sqlCode[:len(self.sqlCode)-1])
	self.sqlCode += `);`
	stmt, err := self.DB.Prepare(self.sqlCode)
	util.CheckErr(err)

	_, err = stmt.Exec()
	util.CheckErr(err)
	return self
}

//设置插入的1行数据
func (self *MyTable) AddRow(value ...string) *MyTable {
	self.rowValues = append(self.rowValues, value...)
	return self
}

//向sqlCode添加"插入1行数据"的语句，执行前须保证Create()、AddRow()已经执行
//insert into table1(field1,field2) values(rowValues[0],rowValues[1])
func (self *MyTable) Update() error {
	if len(self.rowValues) == 0 {
		return errors.New("Mysql更新内容为空")
	}

	self.sqlCode = `insert into ` + self.tableName + `(`
	if len(self.columnNames) != 0 {
		for _, v := range self.columnNames {
			self.sqlCode += "`" + v[0] + "`" + `,`
		}
		self.sqlCode = string(self.sqlCode[:len(self.sqlCode)-1])
		self.sqlCode += `)values(`
	}
	for _, v := range self.rowValues {
		v = strings.Replace(v, `"`, `\"`, -1)
		self.sqlCode += `"` + v + `"` + `,`
	}
	self.sqlCode = string(self.sqlCode[:len(self.sqlCode)-1])
	self.sqlCode += `);`

	stmt, err := self.DB.Prepare(self.sqlCode)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	// 清空临时数据
	self.rowValues = []string{}

	return nil
}

// 获取全部数据
func (self *MyTable) SelectAll() (*sql.Rows, error) {
	if self.tableName == "" {
		return nil, errors.New("表名不能为空")
	}
	self.sqlCode = `select * from ` + self.tableName + `;`
	return self.DB.Query(self.sqlCode)
}
