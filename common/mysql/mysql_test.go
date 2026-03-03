package mysql

import (
	"database/sql"
	"regexp"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestNew(t *testing.T) {
	r := New()
	if r.IsErr() {
		t.Fatalf("New() = %v, want Ok", r)
	}
	tbl := r.Unwrap()
	if tbl == nil {
		t.Fatal("New() returned nil Table")
	}
}

func TestTable_Clone(t *testing.T) {
	tbl := New().Unwrap().
		SetTableName("users").
		AddColumn("name VARCHAR(255)", "age INT")
	cloned := tbl.Clone()
	if cloned == tbl {
		t.Error("Clone() should return a new instance")
	}
	if cloned.tableName != tbl.tableName {
		t.Errorf("Clone().tableName = %q, want %q", cloned.tableName, tbl.tableName)
	}
	if len(cloned.columnNames) != len(tbl.columnNames) {
		t.Errorf("Clone().columnNames len = %d, want %d", len(cloned.columnNames), len(tbl.columnNames))
	}
	if cloned.customPrimaryKey != tbl.customPrimaryKey {
		t.Errorf("Clone().customPrimaryKey = %v, want %v", cloned.customPrimaryKey, tbl.customPrimaryKey)
	}
}

func TestTable_SetTableName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple", "users", "`users`"},
		{"with_backtick", "`users`", "`users`"},
		{"empty", "", "``"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := New().Unwrap().SetTableName(tt.in)
			if tbl.tableName != tt.want {
				t.Errorf("SetTableName(%q) tableName = %q, want %q", tt.in, tbl.tableName, tt.want)
			}
		})
	}
}

func TestTable_AddColumn(t *testing.T) {
	tbl := New().Unwrap().AddColumn("id INT", "name VARCHAR(255)")
	if len(tbl.columnNames) != 2 {
		t.Fatalf("AddColumn len = %d, want 2", len(tbl.columnNames))
	}
	if tbl.columnNames[0][0] != "`id`" || tbl.columnNames[0][1] != "INT" {
		t.Errorf("column[0] = %q %q, want `id` INT", tbl.columnNames[0][0], tbl.columnNames[0][1])
	}
	if tbl.columnNames[1][0] != "`name`" || tbl.columnNames[1][1] != "VARCHAR(255)" {
		t.Errorf("column[1] = %q %q, want `name` VARCHAR(255)", tbl.columnNames[1][0], tbl.columnNames[1][1])
	}
	tbl.AddColumn("  email TEXT  ")
	if tbl.columnNames[2][0] != "`email`" || tbl.columnNames[2][1] != "TEXT" {
		t.Errorf("column[2] = %q %q, want `email` TEXT", tbl.columnNames[2][0], tbl.columnNames[2][1])
	}
}

func TestTable_CustomPrimaryKey(t *testing.T) {
	tbl := New().Unwrap().CustomPrimaryKey("pk BIGINT PRIMARY KEY")
	if !tbl.customPrimaryKey {
		t.Error("CustomPrimaryKey() should set customPrimaryKey=true")
	}
	if len(tbl.columnNames) != 1 {
		t.Fatalf("CustomPrimaryKey len = %d, want 1", len(tbl.columnNames))
	}
	if tbl.columnNames[0][0] != "`pk`" || tbl.columnNames[0][1] != "BIGINT PRIMARY KEY" {
		t.Errorf("column = %q %q", tbl.columnNames[0][0], tbl.columnNames[0][1])
	}
}

func TestDB(t *testing.T) {
	d, e := DB()
	if d != nil {
		t.Errorf("DB() before Refresh = %v, want nil", d)
	}
	if e != nil {
		t.Errorf("DB() err = %v, want nil", e)
	}
}

func TestTable_Create_EmptyColumns(t *testing.T) {
	tbl := New().Unwrap().SetTableName("t")
	r := tbl.Create()
	if r.IsOk() {
		t.Error("Create() with empty columns should return Err")
	}
	if !strings.Contains(r.UnwrapErr().Error(), "Column can not be empty") {
		t.Errorf("Create() err = %v", r.UnwrapErr())
	}
}

func TestTable_FlushInsert_NoRows(t *testing.T) {
	tbl := New().Unwrap().SetTableName("t").AddColumn("a INT")
	r := tbl.FlushInsert()
	if r.IsErr() {
		t.Errorf("FlushInsert() with rowsCount=0 = %v", r.UnwrapErr())
	}
}

func TestTable_FlushInsert_NoColumns(t *testing.T) {
	tbl := New().Unwrap().SetTableName("t")
	tbl = tbl.AutoInsert([]string{"x"})
	r := tbl.FlushInsert()
	if r.IsErr() {
		t.Errorf("FlushInsert() with colCount=0 = %v", r.UnwrapErr())
	}
}

func TestTable_SelectAll_EmptyTableName(t *testing.T) {
	tbl := New().Unwrap()
	r := tbl.SelectAll()
	if r.IsOk() {
		t.Error("SelectAll() with empty tableName should return Err")
	}
	if !strings.Contains(r.UnwrapErr().Error(), "表名不能为空") {
		t.Errorf("SelectAll() err = %v", r.UnwrapErr())
	}
}

func TestTable_AutoInsert_OversizedPacket(t *testing.T) {
	mc := getMysqlConst()
	oversized := strings.Repeat("x", mc.maxPkt+1)
	tbl := New().Unwrap().SetTableName("t").AddColumn("a VARCHAR(1)")
	before := tbl.rowsCount
	tbl = tbl.AutoInsert([]string{oversized})
	if tbl.rowsCount != before {
		t.Errorf("AutoInsert(oversized) should not add row, rowsCount = %d", tbl.rowsCount)
	}
}

func TestTable_AutoInsert_SmallValue(t *testing.T) {
	tbl := New().Unwrap().SetTableName("t").AddColumn("a VARCHAR(10)")
	tbl = tbl.AutoInsert([]string{"hello"})
	if tbl.rowsCount != 1 {
		t.Errorf("AutoInsert() rowsCount = %d, want 1", tbl.rowsCount)
	}
	if len(tbl.args) != 1 || tbl.args[0] != "hello" {
		t.Errorf("AutoInsert() args = %v", tbl.args)
	}
}

func TestTable_AutoInsert_MultipleColumns(t *testing.T) {
	tbl := New().Unwrap().SetTableName("t").AddColumn("a INT", "b VARCHAR(10)")
	tbl = tbl.AutoInsert([]string{"1", "foo"})
	if tbl.rowsCount != 1 {
		t.Errorf("rowsCount = %d, want 1", tbl.rowsCount)
	}
	if len(tbl.args) != 2 {
		t.Errorf("args len = %d, want 2", tbl.args)
	}
}

func TestNew_Unwrap(t *testing.T) {
	r := New()
	tbl := r.Unwrap()
	_ = tbl.SetTableName("x").AddColumn("id INT")
	r2 := New()
	tbl2 := r2.Unwrap()
	if tbl == tbl2 {
		t.Error("New() should return distinct Table instances")
	}
}

func TestTable_Clone_PreservesCustomPrimaryKey(t *testing.T) {
	tbl := New().Unwrap().CustomPrimaryKey("id BIGINT")
	cloned := tbl.Clone()
	if !cloned.customPrimaryKey {
		t.Error("Clone() should preserve customPrimaryKey")
	}
}

func TestTable_SetTableName_Chain(t *testing.T) {
	tbl := New().Unwrap().SetTableName("a").SetTableName("b")
	if tbl.tableName != "`b`" {
		t.Errorf("SetTableName chain = %q, want `b`", tbl.tableName)
	}
}

func TestTable_AddColumn_EmptyTrimmed(t *testing.T) {
	tbl := New().Unwrap()
	tbl.AddColumn("x INT")
	if len(tbl.columnNames) != 1 {
		t.Fatalf("len = %d", len(tbl.columnNames))
	}
	if tbl.columnNames[0][0] != "`x`" {
		t.Errorf("column[0][0] = %q", tbl.columnNames[0][0])
	}
}

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	origDB := db
	db = sqlDB
	return sqlDB, mock, func() {
		db = origDB
		sqlDB.Close()
	}
}

func TestTable_Create_WithMock(t *testing.T) {
	_, mock, teardown := setupMockDB(t)
	defer teardown()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `users` \\(id INT\\(12\\) NOT NULL PRIMARY KEY AUTO_INCREMENT,`name` VARCHAR\\(255\\),`age` INT\\) ENGINE=MyISAM DEFAULT CHARSET=utf8").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tbl := New().Unwrap().SetTableName("users").AddColumn("name VARCHAR(255)", "age INT")
	r := tbl.Create()
	if r.IsErr() {
		t.Errorf("Create() = %v", r.UnwrapErr())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestTable_Create_CustomPrimaryKey_WithMock(t *testing.T) {
	_, mock, teardown := setupMockDB(t)
	defer teardown()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `t` \\(`pk` BIGINT PRIMARY KEY\\) ENGINE=MyISAM DEFAULT CHARSET=utf8").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tbl := New().Unwrap().SetTableName("t").CustomPrimaryKey("pk BIGINT PRIMARY KEY")
	r := tbl.Create()
	if r.IsErr() {
		t.Errorf("Create() = %v", r.UnwrapErr())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestTable_Truncate_WithMock(t *testing.T) {
	_, mock, teardown := setupMockDB(t)
	defer teardown()

	mock.ExpectExec("TRUNCATE TABLE `users`").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tbl := New().Unwrap().SetTableName("users")
	r := tbl.Truncate()
	if r.IsErr() {
		t.Errorf("Truncate() = %v", r.UnwrapErr())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestTable_FlushInsert_WithMock(t *testing.T) {
	_, mock, teardown := setupMockDB(t)
	defer teardown()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `t`(`a`) VALUES (?);")).
		WithArgs("hello").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tbl := New().Unwrap().SetTableName("t").AddColumn("a VARCHAR(10)")
	tbl = tbl.AutoInsert([]string{"hello"})
	r := tbl.FlushInsert()
	if r.IsErr() {
		t.Errorf("FlushInsert() = %v", r.UnwrapErr())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestTable_SelectAll_WithMock(t *testing.T) {
	_, mock, teardown := setupMockDB(t)
	defer teardown()

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "a")
	mock.ExpectQuery("SELECT \\* FROM `users`").
		WillReturnRows(rows)

	tbl := New().Unwrap().SetTableName("users")
	r := tbl.SelectAll()
	if r.IsErr() {
		t.Errorf("SelectAll() = %v", r.UnwrapErr())
		return
	}
	rs := r.Unwrap()
	defer rs.Close()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
	_ = rs
}

func TestTable_Create_ExecError(t *testing.T) {
	_, mock, teardown := setupMockDB(t)
	defer teardown()

	mock.ExpectExec("CREATE TABLE").
		WillReturnError(sql.ErrConnDone)

	tbl := New().Unwrap().SetTableName("t").AddColumn("a INT")
	r := tbl.Create()
	if r.IsOk() {
		t.Error("Create() should return Err on Exec failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestTable_Truncate_ExecError(t *testing.T) {
	_, mock, teardown := setupMockDB(t)
	defer teardown()

	mock.ExpectExec("TRUNCATE TABLE").
		WillReturnError(sql.ErrConnDone)

	tbl := New().Unwrap().SetTableName("t")
	r := tbl.Truncate()
	if r.IsOk() {
		t.Error("Truncate() should return Err on Exec failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestTable_FlushInsert_ExecError(t *testing.T) {
	_, mock, teardown := setupMockDB(t)
	defer teardown()

	mock.ExpectExec("INSERT INTO").
		WillReturnError(sql.ErrConnDone)

	tbl := New().Unwrap().SetTableName("t").AddColumn("a INT")
	tbl = tbl.AutoInsert([]string{"x"})
	r := tbl.FlushInsert()
	if r.IsOk() {
		t.Error("FlushInsert() should return Err on Exec failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestTable_SelectAll_QueryError(t *testing.T) {
	_, mock, teardown := setupMockDB(t)
	defer teardown()

	mock.ExpectQuery("SELECT \\* FROM `t`").
		WillReturnError(sql.ErrConnDone)

	tbl := New().Unwrap().SetTableName("t")
	r := tbl.SelectAll()
	if r.IsOk() {
		t.Error("SelectAll() should return Err on Query failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

