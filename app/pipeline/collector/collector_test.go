package collector

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/pipeline/collector/data"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/runtime/cache"
)

func TestNewCollector(t *testing.T) {
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
	}
	tests := []struct {
		name     string
		batchCap int
		wantCap  int
	}{
		{"normal", 10, 10},
		{"one", 1, 1},
		{"zero", 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector(sp, "csv", tt.batchCap)
			if c == nil {
				t.Fatal("NewCollector returned nil")
			}
			if cap(c.DataChan) != tt.wantCap {
				t.Errorf("DataChan cap = %d, want %d", cap(c.DataChan), tt.wantCap)
			}
			if cap(c.FileChan) != tt.wantCap {
				t.Errorf("FileChan cap = %d, want %d", cap(c.FileChan), tt.wantCap)
			}
		})
	}
}

func TestCollector_CollectData(t *testing.T) {
	sp := &spider.Spider{
		Name:     "S",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
	}
	c := NewCollector(sp, "csv", 10)
	c.Start()
	defer c.Stop()

	cell := data.GetDataCell("r1", map[string]interface{}{"a": "b"}, "u", "pu", "dt")
	r := c.CollectData(cell)
	if r.IsErr() {
		t.Errorf("CollectData: %v", r.UnwrapErr())
	}
}

func TestCollector_CollectFile(t *testing.T) {
	sp := &spider.Spider{
		Name:     "S",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
	}
	c := NewCollector(sp, "csv", 10)
	c.Start()
	defer c.Stop()

	cell := data.GetFileCell("r1", "test.txt", []byte("hello"))
	r := c.CollectFile(cell)
	if r.IsErr() {
		t.Errorf("CollectFile: %v", r.UnwrapErr())
	}
}

func TestCollector_Stop(t *testing.T) {
	sp := &spider.Spider{
		Name:     "S",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
	}
	c := NewCollector(sp, "csv", 10)
	c.Start()
	c.Stop()
}

func TestCollector_OutputData_EmptyBuf(t *testing.T) {
	sp := &spider.Spider{
		Name:     "S",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
	}
	c := NewCollector(sp, "csv", 1)
	c.outputData()
}

func TestCollector_OutputData_PanicRecovery(t *testing.T) {
	oldFn := DataOutput["csv"]
	DataOutput["csv"] = func(*Collector) result.VoidResult { panic("test panic") }
	defer func() { DataOutput["csv"] = oldFn }()

	sp := &spider.Spider{
		Name:     "S",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{"r1": {ItemFields: []string{"f1"}}}},
	}
	c := NewCollector(sp, "csv", 1)
	c.dataBuf = []data.DataCell{
		data.GetDataCell("r1", map[string]interface{}{"f1": "v1"}, "u", "pu", "dt"),
	}
	c.dataBatch = 1
	c.addDataSum(1)
	c.outputData()
}

func TestCollector_OutputData_ErrorResult(t *testing.T) {
	oldFn := DataOutput["csv"]
	DataOutput["csv"] = func(*Collector) result.VoidResult { return result.FmtErrVoid("test error") }
	defer func() { DataOutput["csv"] = oldFn }()

	sp := &spider.Spider{
		Name:     "S",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{"r1": {ItemFields: []string{"f1"}}}},
	}
	c := NewCollector(sp, "csv", 1)
	c.dataBuf = []data.DataCell{
		data.GetDataCell("r1", map[string]interface{}{"f1": "v1"}, "u", "pu", "dt"),
	}
	c.dataBatch = 1
	c.addDataSum(1)
	c.outputData()
}


func TestCollector_OutputCSV(t *testing.T) {
	tmp := t.TempDir()
	_ = config.Conf()
	conf := config.Conf()
	conf.TextDir = tmp
	cache.StartTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if cache.Task == nil {
		cache.Task = &cache.AppConf{}
	}
	cache.Task.OutType = "csv"
	cache.Task.Mode = 0
	cache.Task.SuccessInherit = false

	go func() {
		for range cache.ReportChan {
		}
	}()

	sp := &spider.Spider{
		Name:     "CSVSpider",
		Keyin:    "",
		RuleTree: &spider.RuleTree{
			Trunk: map[string]*spider.Rule{
				"list": {ItemFields: []string{"title", "url"}},
			},
		},
	}
	sp.ReqmatrixInit()

	c := NewCollector(sp, "csv", 2)
	c.dataBuf = []data.DataCell{
		data.GetDataCell("list", map[string]interface{}{"title": "t1", "url": "u1"}, "http://a.com", "http://p.com", "2024-01-15"),
		data.GetDataCell("list", map[string]interface{}{"title": "t2", "url": "u2"}, "http://b.com", "http://p.com", "2024-01-15"),
	}
	c.dataBatch = 1
	c.addDataSum(2)

	DataOutput["csv"](c)

	var matches []string
	filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".csv" {
			matches = append(matches, path)
		}
		return nil
	})
	if len(matches) == 0 {
		t.Fatal("no CSV file created")
	}
	content, err := readFile(matches[0])
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}
	if len(content) < 10 {
		t.Errorf("CSV content too short: %q", content)
	}
}

func TestCollector_OutputCSV_NonStringData(t *testing.T) {
	tmp := t.TempDir()
	conf := config.Conf()
	conf.TextDir = tmp
	cache.StartTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if cache.Task == nil {
		cache.Task = &cache.AppConf{}
	}

	sp := &spider.Spider{
		Name:     "CSVSpider2",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{"r1": {ItemFields: []string{"n", "v"}}}},
	}

	c := NewCollector(sp, "csv", 1)
	c.dataBuf = []data.DataCell{
		data.GetDataCell("r1", map[string]interface{}{"n": 123, "v": 3.14}, "u", "pu", "dt"),
	}
	c.dataBatch = 1
	c.addDataSum(1)

	DataOutput["csv"](c)

	var matches []string
	filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".csv" {
			matches = append(matches, path)
		}
		return nil
	})
	if len(matches) == 0 {
		t.Fatal("no CSV file created")
	}
}

func TestCollector_OutputCSV_NotDefaultField(t *testing.T) {
	tmp := t.TempDir()
	conf := config.Conf()
	conf.TextDir = tmp
	cache.StartTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if cache.Task == nil {
		cache.Task = &cache.AppConf{}
	}

	sp := &spider.Spider{
		Name:            "CSVSpider3",
		NotDefaultField: true,
		RuleTree:        &spider.RuleTree{Trunk: map[string]*spider.Rule{"r1": {ItemFields: []string{"x"}}}},
	}

	c := NewCollector(sp, "csv", 1)
	c.dataBuf = []data.DataCell{
		data.GetDataCell("r1", map[string]interface{}{"x": "y"}, "u", "pu", "dt"),
	}
	c.dataBatch = 1
	c.addDataSum(1)

	DataOutput["csv"](c)

	var matches []string
	filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".csv" {
			matches = append(matches, path)
		}
		return nil
	})
	if len(matches) == 0 {
		t.Fatal("no CSV file created")
	}
}

func TestCollector_OutputExcel(t *testing.T) {
	tmp := t.TempDir()
	_ = config.Conf()
	conf := config.Conf()
	conf.TextDir = tmp
	cache.StartTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if cache.Task == nil {
		cache.Task = &cache.AppConf{}
	}
	cache.Task.OutType = "excel"

	sp := &spider.Spider{
		Name:     "ExcelSpider",
		RuleTree: &spider.RuleTree{
			Trunk: map[string]*spider.Rule{
				"sheet1": {ItemFields: []string{"col1", "col2"}},
			},
		},
	}

	c := NewCollector(sp, "excel", 2)
	c.dataBuf = []data.DataCell{
		data.GetDataCell("sheet1", map[string]interface{}{"col1": "v1", "col2": "v2"}, "u", "pu", "dt"),
		data.GetDataCell("sheet1", map[string]interface{}{"col1": 99, "col2": 1.5}, "u2", "pu2", "dt2"),
	}
	c.dataBatch = 1
	c.addDataSum(2)

	DataOutput["excel"](c)

	var excelMatches []string
	filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".xlsx" {
			excelMatches = append(excelMatches, path)
		}
		return nil
	})
	if len(excelMatches) == 0 {
		t.Fatal("no Excel file created")
	}
}

func TestCollector_OutputFile(t *testing.T) {
	tmp := t.TempDir()
	_ = config.Conf()
	conf := config.Conf()
	conf.FileDir = tmp

	sp := &spider.Spider{
		Name:     "FileSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
	}

	c := NewCollector(sp, "csv", 1)
	c.wait.Add(1)
	fc := data.GetFileCell("r1", "subdir/file.txt", []byte("file content"))
	c.outputFile(fc)

	var filePath string
	filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Base(path) == "file.txt" {
			filePath = path
		}
		return nil
	})
	if filePath == "" {
		t.Fatal("no file.txt created")
	}
	content, err := readFile(filePath)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if content != "file content" {
		t.Errorf("content = %q, want %q", content, "file content")
	}
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func TestCollector_OutputFile_MkdirFail(t *testing.T) {
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	conf := config.Conf()
	oldFileDir := conf.FileDir
	conf.FileDir = blocker
	defer func() { conf.FileDir = oldFileDir }()

	sp := &spider.Spider{
		Name:     "S",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
	}
	c := NewCollector(sp, "csv", 1)
	c.wait.Add(1)
	fc := data.GetFileCell("r1", "x/file.txt", []byte("content"))
	c.outputFile(fc)
}
