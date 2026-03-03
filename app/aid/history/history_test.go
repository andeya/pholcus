package history

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/common/mysql"
	"github.com/andeya/pholcus/config"
)

func setupHistoryDir(t *testing.T) (cleanup func()) {
	tmp := t.TempDir()
	historyDir := filepath.Join(tmp, config.WorkRoot, config.HistoryTag)
	if err := os.MkdirAll(historyDir, 0777); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	return func() { os.Chdir(orig) }
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		subName string
	}{
		{"spider1", ""},
		{"spider2", "sub"},
	}
	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.subName, func(t *testing.T) {
			cleanup := setupHistoryDir(t)
			defer cleanup()
			_ = config.Conf()
			h := New(tt.name, tt.subName)
			if h == nil {
				t.Fatal("New returned nil")
			}
			if got := h.UpsertSuccess("id1"); !got {
				t.Error("UpsertSuccess want true")
			}
			if got := h.UpsertSuccess("id1"); got {
				t.Error("UpsertSuccess duplicate want false")
			}
		})
	}
}

func TestHistory_ReadSuccess_File(t *testing.T) {
	tests := []struct {
		name     string
		inherit  bool
		fileData string
		checkOld bool
	}{
		{"no inherit", false, "", false},
		{"inherit no file", true, "", false},
		{"inherit with data", true, `,"id1":true,"id2":true`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupHistoryDir(t)
			defer cleanup()
			_ = config.Conf()
			h := New("test", "").(*History)
			if tt.fileData != "" {
				if err := os.WriteFile(h.Success.fileName, []byte(tt.fileData), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			}
			r := h.ReadSuccess("file", tt.inherit)
			if r.IsErr() {
				t.Errorf("ReadSuccess: %v", r.UnwrapErr())
			}
			if tt.checkOld {
				if len(h.Success.old) != 2 || !h.Success.HasSuccess("id1") || !h.Success.HasSuccess("id2") {
					t.Errorf("expected ids in old, got len=%d old=%v", len(h.Success.old), h.Success.old)
				}
			}
		})
	}
}

func TestHistory_ReadSuccess_EmptyFile(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	if err := os.WriteFile(h.Success.fileName, []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	r := h.ReadSuccess("file", true)
	if r.IsErr() {
		t.Errorf("ReadSuccess: %v", r.UnwrapErr())
	}
}

func TestHistory_ReadSuccess_InheritPaths(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	h.Success.inheritable = true
	r := h.ReadSuccess("file", true)
	if r.IsErr() {
		t.Errorf("ReadSuccess inheritable: %v", r.UnwrapErr())
	}

	h.Success.inheritable = false
	h.Success.old["x"] = true
	r = h.ReadSuccess("file", true)
	if r.IsErr() {
		t.Errorf("ReadSuccess: %v", r.UnwrapErr())
	}
	if len(h.Success.old) != 0 {
		t.Error("expected old cleared when switching to inherit")
	}
}

func TestHistory_ReadFailure_File(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	req := &request.Request{Spider: "s", URL: "http://a.com", Rule: "r", Method: "GET", Header: make(http.Header)}
	req.Prepare()
	ser := req.Serialize().Unwrap()
	fileData, _ := json.Marshal(map[string]string{req.Unique(): ser})

	tests := []struct {
		name     string
		inherit  bool
		fileData []byte
	}{
		{"no inherit", false, nil},
		{"inherit no file", true, nil},
		{"inherit with data", true, fileData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fileData != nil {
				if err := os.WriteFile(h.Failure.fileName, tt.fileData, 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			}
			r := h.ReadFailure("file", tt.inherit)
			if r.IsErr() {
				t.Errorf("ReadFailure: %v", r.UnwrapErr())
			}
		})
	}
}

func TestHistory_ReadFailure_EmptyFile(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	if err := os.WriteFile(h.Failure.fileName, []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	r := h.ReadFailure("file", true)
	if r.IsErr() {
		t.Errorf("ReadFailure: %v", r.UnwrapErr())
	}
}

func TestHistory_Empty(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	h.UpsertSuccess("id1")
	req := &request.Request{Spider: "s", URL: "http://a.com", Rule: "r", Method: "GET", Header: make(http.Header)}
	req.Prepare()
	h.UpsertFailure(req)

	h.Empty()

	if h.HasSuccess("id1") {
		t.Error("Empty should clear success")
	}
	pulled := h.PullFailure()
	if len(pulled) != 0 {
		t.Error("Empty should clear failure")
	}
}

func TestHistory_FlushSuccess_File(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	h.UpsertSuccess("id1")
	h.UpsertSuccess("id2")

	r := h.FlushSuccess("file")
	if r.IsErr() {
		t.Errorf("FlushSuccess: %v", r.UnwrapErr())
	}
	if _, err := os.Stat(h.Success.fileName); err != nil {
		t.Errorf("FlushSuccess file: %v", err)
	}
}

func TestHistory_FlushFailure_File(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	req := &request.Request{Spider: "s", URL: "http://a.com", Rule: "r", Method: "GET", Header: make(http.Header)}
	req.Prepare()
	h.UpsertFailure(req)

	r := h.FlushFailure("file")
	if r.IsErr() {
		t.Errorf("FlushFailure: %v", r.UnwrapErr())
	}
	if _, err := os.Stat(h.Failure.fileName); err != nil {
		t.Errorf("FlushFailure file: %v", err)
	}
}

func TestHistory_FlushSuccess_Empty(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	r := h.FlushSuccess("file")
	if r.IsErr() {
		t.Errorf("FlushSuccess empty: %v", r.UnwrapErr())
	}
}

func TestHistory_FlushFailure_Empty(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	r := h.FlushFailure("file")
	if r.IsErr() {
		t.Errorf("FlushFailure empty: %v", r.UnwrapErr())
	}
}

func TestHistory_ReadSuccess_FileNotFound(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	r := h.ReadSuccess("file", true)
	if r.IsErr() {
		t.Errorf("ReadSuccess file not found: %v", r.UnwrapErr())
	}
}

func TestHistory_ReadSuccess_ReadFailure_MysqlMock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()
	cleanup := mysql.SetDBForTest(sqlDB)
	defer cleanup()

	cleanupDir := setupHistoryDir(t)
	defer cleanupDir()
	_ = config.Conf()

	h := New("test", "").(*History)

	rows := sqlmock.NewRows([]string{"id"}).AddRow("id1").AddRow("id2")
	mock.ExpectQuery("SELECT \\* FROM").WillReturnRows(rows)
	r := h.ReadSuccess("mysql", true)
	if r.IsErr() {
		t.Errorf("ReadSuccess mysql: %v", r.UnwrapErr())
	}
	if len(h.Success.old) != 2 {
		t.Errorf("ReadSuccess mysql: want 2 old, got %d", len(h.Success.old))
	}

	req := &request.Request{Spider: "s", URL: "http://a.com", Rule: "r", Method: "GET", Header: make(http.Header)}
	req.Prepare()
	ser := req.Serialize().Unwrap()
	rows2 := sqlmock.NewRows([]string{"id", "failure"}).AddRow(req.Unique(), ser)
	mock.ExpectQuery("SELECT \\* FROM").WillReturnRows(rows2)
	r = h.ReadFailure("mysql", true)
	if r.IsErr() {
		t.Errorf("ReadFailure mysql: %v", r.UnwrapErr())
	}
	if len(h.Failure.list) != 1 {
		t.Errorf("ReadFailure mysql: want 1, got %d", len(h.Failure.list))
	}
}

func TestHistory_ReadSuccess_MysqlDBError(t *testing.T) {
	cleanup := mysql.SetDBForTest(nil)
	defer cleanup()

	cleanupDir := setupHistoryDir(t)
	defer cleanupDir()
	_ = config.Conf()

	h := New("test", "").(*History)
	r := h.ReadSuccess("mysql", true)
	if r.IsErr() {
		t.Errorf("ReadSuccess mysql no db: %v", r.UnwrapErr())
	}
}

func TestHistory_ReadFailure_MysqlDBError(t *testing.T) {
	cleanup := mysql.SetDBForTest(nil)
	defer cleanup()

	cleanupDir := setupHistoryDir(t)
	defer cleanupDir()
	_ = config.Conf()

	h := New("test", "").(*History)
	r := h.ReadFailure("mysql", true)
	if r.IsErr() {
		t.Errorf("ReadFailure mysql no db: %v", r.UnwrapErr())
	}
}

func TestHistory_ReadSuccess_MysqlSelectError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()
	cleanup := mysql.SetDBForTest(sqlDB)
	defer cleanup()

	cleanupDir := setupHistoryDir(t)
	defer cleanupDir()
	_ = config.Conf()

	h := New("test", "").(*History)
	mock.ExpectQuery("SELECT \\* FROM").WillReturnError(sql.ErrConnDone)
	r := h.ReadSuccess("mysql", true)
	if r.IsErr() {
		t.Errorf("ReadSuccess mysql select err: %v", r.UnwrapErr())
	}
}

func TestHistory_FlushSuccess_FlushFailure_MysqlMock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()
	cleanup := mysql.SetDBForTest(sqlDB)
	defer cleanup()

	cleanupDir := setupHistoryDir(t)
	defer cleanupDir()
	_ = config.Conf()

	h := New("test", "").(*History)
	h.UpsertSuccess("id1")
	h.UpsertSuccess("id2")

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO").WithArgs("id1", "id2").WillReturnResult(sqlmock.NewResult(2, 2))
	r := h.FlushSuccess("mysql")
	if r.IsErr() {
		t.Errorf("FlushSuccess mysql: %v", r.UnwrapErr())
	}

	req := &request.Request{Spider: "s", URL: "http://a.com", Rule: "r", Method: "GET", Header: make(http.Header)}
	req.Prepare()
	h.UpsertFailure(req)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))
	r = h.FlushFailure("mysql")
	if r.IsErr() {
		t.Errorf("FlushFailure mysql: %v", r.UnwrapErr())
	}
}

func TestHistory_ReadFailure_InvalidData(t *testing.T) {
	cleanup := setupHistoryDir(t)
	defer cleanup()
	_ = config.Conf()

	h := New("test", "").(*History)
	req := &request.Request{Spider: "s", URL: "http://a.com", Rule: "r", Method: "GET", Header: make(http.Header)}
	req.Prepare()
	ser := req.Serialize().Unwrap()
	fileData := map[string]string{req.Unique(): ser, "badkey": "{invalid}"}
	data, _ := json.Marshal(fileData)
	if err := os.WriteFile(h.Failure.fileName, data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	r := h.ReadFailure("file", true)
	if r.IsErr() {
		t.Errorf("ReadFailure: %v", r.UnwrapErr())
	}
	if len(h.Failure.list) != 1 {
		t.Errorf("expected 1 valid record, got %d", len(h.Failure.list))
	}
}
