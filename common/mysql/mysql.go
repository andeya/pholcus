package mysql

import (
	"database/sql"
	"errors"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

/************************ Mysql 输出 ***************************/
//sql转换结构体
type MyTable struct {
	tableName        string
	columnNames      [][2]string // 标题字段
	rows             [][]string  // 多行数据
	sqlCode          string
	customPrimaryKey bool
	size             int //内容大小的近似值
}

var (
	db                 *sql.DB
	err                error
	maxConnChan        = make(chan bool, config.MYSQL_CONN_CAP) //最大执行数限制
	max_allowed_packet = config.MYSQL_MAX_ALLOWED_PACKET - 1024
	lock               sync.RWMutex
)

func DB() (*sql.DB, error) {
	return db, err
}

func Refresh() {
	lock.Lock()
	defer lock.Unlock()
	db, err = sql.Open("mysql", config.MYSQL_CONN_STR+"/"+config.DB_NAME+"?charset=utf8")
	if err != nil {
		logs.Log.Error("Mysql：%v\n", err)
		return
	}
	db.SetMaxOpenConns(config.MYSQL_CONN_CAP)
	db.SetMaxIdleConns(config.MYSQL_CONN_CAP)
	if err = db.Ping(); err != nil {
		logs.Log.Error("Mysql：%v\n", err)
	}
}

func New() *MyTable {
	return &MyTable{}
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
func (self *MyTable) Create() error {
	if len(self.columnNames) == 0 {
		return errors.New("Column can not be empty")
	}
	self.sqlCode = `create table if not exists ` + "`" + self.tableName + "` ("
	if !self.customPrimaryKey {
		self.sqlCode += `id int(12) not null primary key auto_increment,`
	}
	for _, title := range self.columnNames {
		self.sqlCode += "`" + title[0] + "`" + ` ` + title[1] + `,`
	}
	self.sqlCode = string(self.sqlCode[:len(self.sqlCode)-1])
	self.sqlCode += `) default charset=utf8;`

	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	lock.RLock()
	stmt, err := db.Prepare(self.sqlCode)
	lock.RUnlock()
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

//清空表单，执行前须保证SetTableName()已经执行
func (self *MyTable) Truncate() error {
	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	lock.RLock()
	stmt, err := db.Prepare(`TRUNCATE TABLE ` + "`" + self.tableName + "`")
	lock.RUnlock()
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

//设置插入的1行数据
func (self *MyTable) addRow(value []string) *MyTable {
	self.rows = append(self.rows, value)
	return self
}

//智能插入数据，每次1行
func (self *MyTable) AutoInsert(value []string) *MyTable {
	var nsize int
	for _, v := range value {
		nsize += len(v)
	}
	if nsize > max_allowed_packet {
		logs.Log.Error("%v", "packet for query is too large. Try adjusting the 'maxallowedpacket'variable in the 'config.ini'")
		return self
	}
	self.size += nsize
	if self.size > max_allowed_packet {
		util.CheckErr(self.FlushInsert())
		return self.AutoInsert(value)
	}
	return self.addRow(value)
}

//向sqlCode添加"插入数据"的语句，执行前须保证Create()、AutoInsert()已经执行
//insert into table1(field1,field2) values(rows[0]),(rows[1])...
func (self *MyTable) FlushInsert() error {
	if len(self.rows) == 0 {
		return nil
	}

	self.sqlCode = `insert into ` + "`" + self.tableName + "`" + `(`
	if len(self.columnNames) != 0 {
		for _, v := range self.columnNames {
			self.sqlCode += "`" + v[0] + "`,"
		}
		self.sqlCode = self.sqlCode[:len(self.sqlCode)-1] + `)values`
	}
	for _, row := range self.rows {
		self.sqlCode += `(`
		for _, v := range row {
			v = strings.Replace(v, `\`, `\\`, -1)
			v = strings.Replace(v, `"`, `\"`, -1)
			v = strings.Replace(v, `'`, `\'`, -1)
			self.sqlCode += `"` + v + `",`
		}
		self.sqlCode = self.sqlCode[:len(self.sqlCode)-1] + `),`
	}
	self.sqlCode = self.sqlCode[:len(self.sqlCode)-1] + `;`

	defer func() {
		// 清空临时数据
		self.rows = [][]string{}
		self.size = 0
		self.sqlCode = ""
	}()

	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	lock.RLock()
	stmt, err := db.Prepare(self.sqlCode)
	lock.RUnlock()
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

// 获取全部数据
func (self *MyTable) SelectAll() (*sql.Rows, error) {
	if self.tableName == "" {
		return nil, errors.New("表名不能为空")
	}
	self.sqlCode = `select * from ` + "`" + self.tableName + "`" + `;`

	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	lock.RLock()
	defer lock.RUnlock()
	return db.Query(self.sqlCode)
}
