// Package mysql provides MySQL database connection and operation wrapper.
package mysql

import (
	"database/sql"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/andeya/gust/result"
	"github.com/andeya/gust/syncutil"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
)

// Table holds SQL conversion state for MySQL output.
type Table struct {
	tableName        string
	columnNames      [][2]string   // column header fields
	rowsCount        int           // number of rows pending insert
	args             []interface{} // data
	sqlCode          string
	customPrimaryKey bool
	size             int // approximate content size
}

type mysqlConst struct {
	maxPkt      int
	maxConnChan chan bool
}

var (
	err  error
	db   *sql.DB
	once sync.Once
)

var lazyConst = syncutil.NewLazyValueWithFunc(func() result.Result[mysqlConst] {
	return result.Ok(mysqlConst{
		maxPkt:      config.Conf().MySQL.MaxAllowedPacket - 1024,
		maxConnChan: make(chan bool, config.Conf().MySQL.ConnCap),
	})
})

func getMysqlConst() *mysqlConst {
	return lazyConst.GetPtr()
}

// DB returns the MySQL database connection and any initialization error.
func DB() (*sql.DB, error) {
	return db, err
}

// SetDBForTest injects db for testing. Returns a cleanup that restores the previous db and err.
func SetDBForTest(d *sql.DB) func() {
	origDB, origErr := db, err
	db = d
	if d == nil {
		err = sql.ErrConnDone
	} else {
		err = nil
	}
	return func() {
		db, err = origDB, origErr
	}
}

// Refresh initializes or reconnects the MySQL database.
func Refresh() {
	once.Do(func() {
		db, err = sql.Open("mysql", config.Conf().MySQL.ConnStr+"/"+config.Conf().DBName+"?charset=utf8")
		if err != nil {
			logs.Log().Error("Mysql: %v\n", err)
			return
		}
		db.SetMaxOpenConns(config.Conf().MySQL.ConnCap)
		db.SetMaxIdleConns(config.Conf().MySQL.ConnCap)
	})
	if err = db.Ping(); err != nil {
		logs.Log().Error("Mysql: %v\n", err)
	}
}

// New creates a new Table instance.
func New() result.Result[*Table] {
	return result.Ok(&Table{})
}

// Clone returns a copy of the Table with the same table name, columns, and primary key settings.
func (m *Table) Clone() *Table {
	return &Table{
		tableName:        m.tableName,
		columnNames:      m.columnNames,
		customPrimaryKey: m.customPrimaryKey,
	}
}

// SetTableName sets the table name for the Table.
func (t *Table) SetTableName(name string) *Table {
	t.tableName = wrapSQLKey(name)
	return t
}

// AddColumn adds one or more column definitions to the table.
func (t *Table) AddColumn(names ...string) *Table {
	for _, name := range names {
		name = strings.Trim(name, " ")
		idx := strings.Index(name, " ")
		t.columnNames = append(t.columnNames, [2]string{wrapSQLKey(name[:idx]), name[idx+1:]})
	}
	return t
}

// CustomPrimaryKey sets a custom primary key definition (optional).
func (t *Table) CustomPrimaryKey(primaryKeyCode string) *Table {
	t.AddColumn(primaryKeyCode)
	t.customPrimaryKey = true
	return t
}

// Create generates and executes a CREATE TABLE statement. Requires prior SetTableName() and AddColumn().
func (t *Table) Create() (r result.VoidResult) {
	defer r.Catch()
	if len(t.columnNames) == 0 {
		return result.FmtErrVoid("Column can not be empty")
	}
	t.sqlCode = `CREATE TABLE IF NOT EXISTS ` + t.tableName + " ("
	if !t.customPrimaryKey {
		t.sqlCode += `id INT(12) NOT NULL PRIMARY KEY AUTO_INCREMENT,`
	}
	for _, title := range t.columnNames {
		t.sqlCode += title[0] + ` ` + title[1] + `,`
	}
	t.sqlCode = t.sqlCode[:len(t.sqlCode)-1] + `) ENGINE=MyISAM DEFAULT CHARSET=utf8;`

	mc := getMysqlConst()
	mc.maxConnChan <- true
	defer func() {
		t.sqlCode = ""
		<-mc.maxConnChan
	}()

	_, err := db.Exec(t.sqlCode)
	result.RetVoid(err).Unwrap()
	return result.OkVoid()
}

// Truncate empties the table. Requires prior SetTableName().
func (t *Table) Truncate() (r result.VoidResult) {
	defer r.Catch()
	mc := getMysqlConst()
	mc.maxConnChan <- true
	defer func() {
		<-mc.maxConnChan
	}()
	_, err := db.Exec(`TRUNCATE TABLE ` + t.tableName)
	result.RetVoid(err).Unwrap()
	return result.OkVoid()
}

func (t *Table) addRow(value []string) *Table {
	for i, count := 0, len(value); i < count; i++ {
		t.args = append(t.args, value[i])
	}
	t.rowsCount++
	return t
}

// AutoInsert adds a row for insert, flushing automatically when buffer is full or size limit is reached.
func (t *Table) AutoInsert(value []string) *Table {
	mc := getMysqlConst()
	if t.rowsCount > 100 {
		t.FlushInsert().Unwrap()
		return t.AutoInsert(value)
	}
	var nsize int
	for _, v := range value {
		nsize += len(v)
	}
	if nsize > mc.maxPkt {
		logs.Log().Error("%v", "packet for query is too large. Try adjusting the 'maxallowedpacket'variable in the 'config.ini'")
		return t
	}
	t.size += nsize
	if t.size > mc.maxPkt {
		t.FlushInsert().Unwrap()
		return t.AutoInsert(value)
	}
	return t.addRow(value)
}

// FlushInsert executes the buffered INSERT. Create and AutoInsert must be called first.
func (t *Table) FlushInsert() (r result.VoidResult) {
	defer r.Catch()
	if t.rowsCount == 0 {
		return result.OkVoid()
	}

	colCount := len(t.columnNames)
	if colCount == 0 {
		return result.OkVoid()
	}

	t.sqlCode = `INSERT INTO ` + t.tableName + `(`

	for _, v := range t.columnNames {
		t.sqlCode += v[0] + ","
	}

	t.sqlCode = t.sqlCode[:len(t.sqlCode)-1] + `) VALUES `

	blank := ",(" + strings.Repeat(",?", colCount)[1:] + ")"
	t.sqlCode += strings.Repeat(blank, t.rowsCount)[1:] + `;`

	defer func() {
		t.args = []interface{}{}
		t.rowsCount = 0
		t.size = 0
		t.sqlCode = ""
	}()

	mc := getMysqlConst()
	mc.maxConnChan <- true
	defer func() {
		<-mc.maxConnChan
	}()

	_, err := db.Exec(t.sqlCode, t.args...)
	result.RetVoid(err).Unwrap()
	return result.OkVoid()
}

// SelectAll returns all rows from the table. SetTableName must be called first.
func (t *Table) SelectAll() result.Result[*sql.Rows] {
	if t.tableName == "" {
		return result.FmtErr[*sql.Rows]("table name cannot be empty")
	}
	t.sqlCode = `SELECT * FROM ` + t.tableName + `;`

	mc := getMysqlConst()
	mc.maxConnChan <- true
	defer func() {
		<-mc.maxConnChan
	}()
	return result.Ret(db.Query(t.sqlCode))
}

func wrapSQLKey(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "") + "`"
}
