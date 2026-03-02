package mysql

import (
	"database/sql"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
)

// MyTable holds SQL conversion state for MySQL output.
type MyTable struct {
	tableName        string
	columnNames      [][2]string   // column header fields
	rowsCount        int           // number of rows pending insert
	args             []interface{} // data
	sqlCode          string
	customPrimaryKey bool
	size             int // approximate content size
}

var (
	err                error
	db                 *sql.DB
	once               sync.Once
	max_allowed_packet = config.MYSQL_MAX_ALLOWED_PACKET - 1024
	maxConnChan        = make(chan bool, config.MYSQL_CONN_CAP) // max concurrent execution limit
)

// DB returns the MySQL database connection and any initialization error.
func DB() (*sql.DB, error) {
	return db, err
}

// Refresh initializes or reconnects the MySQL database.
func Refresh() {
	once.Do(func() {
		db, err = sql.Open("mysql", config.MYSQL_CONN_STR+"/"+config.DB_NAME+"?charset=utf8")
		if err != nil {
			logs.Log.Error("Mysql: %v\n", err)
			return
		}
		db.SetMaxOpenConns(config.MYSQL_CONN_CAP)
		db.SetMaxIdleConns(config.MYSQL_CONN_CAP)
	})
	if err = db.Ping(); err != nil {
		logs.Log.Error("Mysql: %v\n", err)
	}
}

// New creates a new MyTable instance.
func New() result.Result[*MyTable] {
	return result.Ok(&MyTable{})
}

// Clone returns a copy of the MyTable with the same table name, columns, and primary key settings.
func (m *MyTable) Clone() *MyTable {
	return &MyTable{
		tableName:        m.tableName,
		columnNames:      m.columnNames,
		customPrimaryKey: m.customPrimaryKey,
	}
}

// SetTableName sets the table name for the MyTable.
func (self *MyTable) SetTableName(name string) *MyTable {
	self.tableName = wrapSqlKey(name)
	return self
}

// AddColumn adds one or more column definitions to the table.
func (self *MyTable) AddColumn(names ...string) *MyTable {
	for _, name := range names {
		name = strings.Trim(name, " ")
		idx := strings.Index(name, " ")
		self.columnNames = append(self.columnNames, [2]string{wrapSqlKey(name[:idx]), name[idx+1:]})
	}
	return self
}

// CustomPrimaryKey sets a custom primary key definition (optional).
func (self *MyTable) CustomPrimaryKey(primaryKeyCode string) *MyTable {
	self.AddColumn(primaryKeyCode)
	self.customPrimaryKey = true
	return self
}

// Create generates and executes a CREATE TABLE statement. Requires prior SetTableName() and AddColumn().
func (self *MyTable) Create() (r result.VoidResult) {
	defer r.Catch()
	if len(self.columnNames) == 0 {
		return result.FmtErrVoid("Column can not be empty")
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

	_, err := db.Exec(self.sqlCode)
	result.RetVoid(err).Unwrap()
	return result.OkVoid()
}

// Truncate empties the table. Requires prior SetTableName().
func (self *MyTable) Truncate() (r result.VoidResult) {
	defer r.Catch()
	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	_, err := db.Exec(`TRUNCATE TABLE ` + self.tableName)
	result.RetVoid(err).Unwrap()
	return result.OkVoid()
}

func (self *MyTable) addRow(value []string) *MyTable {
	for i, count := 0, len(value); i < count; i++ {
		self.args = append(self.args, value[i])
	}
	self.rowsCount++
	return self
}

// AutoInsert adds a row for insert, flushing automatically when buffer is full or size limit is reached.
func (self *MyTable) AutoInsert(value []string) *MyTable {
	if self.rowsCount > 100 {
		self.FlushInsert().Unwrap()
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
		self.FlushInsert().Unwrap()
		return self.AutoInsert(value)
	}
	return self.addRow(value)
}

// FlushInsert executes the buffered INSERT. Create and AutoInsert must be called first.
func (self *MyTable) FlushInsert() (r result.VoidResult) {
	defer r.Catch()
	if self.rowsCount == 0 {
		return result.OkVoid()
	}

	colCount := len(self.columnNames)
	if colCount == 0 {
		return result.OkVoid()
	}

	self.sqlCode = `INSERT INTO ` + self.tableName + `(`

	for _, v := range self.columnNames {
		self.sqlCode += v[0] + ","
	}

	self.sqlCode = self.sqlCode[:len(self.sqlCode)-1] + `) VALUES `

	blank := ",(" + strings.Repeat(",?", colCount)[1:] + ")"
	self.sqlCode += strings.Repeat(blank, self.rowsCount)[1:] + `;`

	defer func() {
		self.args = []interface{}{}
		self.rowsCount = 0
		self.size = 0
		self.sqlCode = ""
	}()

	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()

	_, err := db.Exec(self.sqlCode, self.args...)
	result.RetVoid(err).Unwrap()
	return result.OkVoid()
}

// SelectAll returns all rows from the table. SetTableName must be called first.
func (self *MyTable) SelectAll() result.Result[*sql.Rows] {
	if self.tableName == "" {
		return result.FmtErr[*sql.Rows]("表名不能为空")
	}
	self.sqlCode = `SELECT * FROM ` + self.tableName + `;`

	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	return result.Ret(db.Query(self.sqlCode))
}

func wrapSqlKey(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "") + "`"
}
