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
	columnNames      [][2]string   // 标题字段
	rowsCount        int           // 当前缓存的待插入数据的行数
	args             []interface{} // 数据
	sqlCode          string
	customPrimaryKey bool
	size             int // 内容大小的近似值
}

var (
	err                error
	db                 *sql.DB
	once               sync.Once
	max_allowed_packet = config.MYSQL_MAX_ALLOWED_PACKET - 1024
	maxConnChan        = make(chan bool, config.MYSQL_CONN_CAP) //最大执行数限制
)

func DB() (*sql.DB, error) {
	return db, err
}

func Refresh() {
	once.Do(func() {
		db, err = sql.Open("mysql", config.MYSQL_CONN_STR+"/"+config.DB_NAME+"?charset=utf8")
		if err != nil {
			logs.Log.Error("Mysql：%v\n", err)
			return
		}
		db.SetMaxOpenConns(config.MYSQL_CONN_CAP)
		db.SetMaxIdleConns(config.MYSQL_CONN_CAP)
	})
	if err = db.Ping(); err != nil {
		logs.Log.Error("Mysql：%v\n", err)
	}
}

func New() *MyTable {
	return &MyTable{}
}

func (m *MyTable) Clone() *MyTable {
	return &MyTable{
		tableName:        m.tableName,
		columnNames:      m.columnNames,
		customPrimaryKey: m.customPrimaryKey,
	}
}

//设置表名
func (self *MyTable) SetTableName(name string) *MyTable {
	self.tableName = wrapSqlKey(name)
	return self
}

//设置表单列
func (self *MyTable) AddColumn(names ...string) *MyTable {
	for _, name := range names {
		name = strings.Trim(name, " ")
		idx := strings.Index(name, " ")
		self.columnNames = append(self.columnNames, [2]string{wrapSqlKey(name[:idx]), name[idx+1:]})
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
	self.sqlCode = `CREATE TABLE IF NOT EXISTS ` + self.tableName + " ("
	if !self.customPrimaryKey {
		self.sqlCode += `id INT(12) NOT NULL PRIMARY KEY AUTO_INCREMENT,`
	}
	for _, title := range self.columnNames {
		self.sqlCode += title[0] + ` ` + title[1] + `,`
	}
	self.sqlCode = self.sqlCode[:len(self.sqlCode)-1] + `) ENGINE=MyISAM DEFAULT CHARSET=utf8;`

	maxConnChan <- true
	defer func() {
		self.sqlCode = ""
		<-maxConnChan
	}()

	// debug
	// println("Create():", self.sqlCode)

	_, err := db.Exec(self.sqlCode)
	return err
}

//清空表单，执行前须保证SetTableName()已经执行
func (self *MyTable) Truncate() error {
	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	_, err := db.Exec(`TRUNCATE TABLE ` + self.tableName)
	return err
}

//设置插入的1行数据
func (self *MyTable) addRow(value []string) *MyTable {
	for i, count := 0, len(value); i < count; i++ {
		self.args = append(self.args, value[i])
	}
	self.rowsCount++
	return self
}

//智能插入数据，每次1行
func (self *MyTable) AutoInsert(value []string) *MyTable {
	if self.rowsCount > 100 {
		util.CheckErr(self.FlushInsert())
		return self.AutoInsert(value)
	}
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
func (self *MyTable) FlushInsert() error {
	if self.rowsCount == 0 {
		return nil
	}

	colCount := len(self.columnNames)
	if colCount == 0 {
		return nil
	}

	self.sqlCode = `INSERT INTO ` + self.tableName + `(`

	for _, v := range self.columnNames {
		self.sqlCode += v[0] + ","
	}

	self.sqlCode = self.sqlCode[:len(self.sqlCode)-1] + `) VALUES `

	blank := ",(" + strings.Repeat(",?", colCount)[1:] + ")"
	self.sqlCode += strings.Repeat(blank, self.rowsCount)[1:] + `;`

	defer func() {
		// 清空临时数据
		self.args = []interface{}{}
		self.rowsCount = 0
		self.size = 0
		self.sqlCode = ""
	}()

	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()

	// debug
	// println("FlushInsert():", self.sqlCode)

	_, err := db.Exec(self.sqlCode, self.args...)
	return err
}

// 获取全部数据
func (self *MyTable) SelectAll() (*sql.Rows, error) {
	if self.tableName == "" {
		return nil, errors.New("表名不能为空")
	}
	self.sqlCode = `SELECT * FROM ` + self.tableName + `;`

	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	return db.Query(self.sqlCode)
}

func wrapSqlKey(s string) string {
	return "`" + strings.Replace(s, "`", "", -1) + "`"
}
