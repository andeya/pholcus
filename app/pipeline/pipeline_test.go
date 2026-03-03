package pipeline

import (
	"testing"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/pipeline/collector"
	"github.com/andeya/pholcus/app/pipeline/collector/data"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/runtime/cache"
)

func TestNew(t *testing.T) {
	sp := &spider.Spider{
		Name:     "TestSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
	}
	tests := []struct {
		name     string
		outType  string
		batchCap int
	}{
		{"csv", "csv", 10},
		{"excel", "excel", 5},
		{"batch_one", "csv", 1},
		{"batch_zero", "csv", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(sp, tt.outType, tt.batchCap)
			if p == nil {
				t.Fatal("New returned nil")
			}
			col, ok := p.(*collector.Collector)
			if !ok {
				t.Fatalf("New returned %T, want *collector.Collector", p)
			}
			_ = tt.outType
			wantCap := tt.batchCap
			if wantCap < 1 {
				wantCap = 1
			}
			if cap(col.DataChan) != wantCap {
				t.Errorf("DataChan cap = %d, want %d", cap(col.DataChan), wantCap)
			}
		})
	}
}

func TestGetOutputLib(t *testing.T) {
	lib := GetOutputLib()
	if len(lib) == 0 {
		t.Fatal("GetOutputLib returned empty")
	}
	for i := 1; i < len(lib); i++ {
		if lib[i] < lib[i-1] {
			t.Errorf("GetOutputLib not sorted: %q >= %q", lib[i-1], lib[i])
		}
	}
}

func TestRegisterOutput(t *testing.T) {
	origLen := len(GetOutputLib())
	RegisterOutput("_test_output_", func(*collector.Collector) result.VoidResult { return result.OkVoid() })
	lib := GetOutputLib()
	if len(lib) != origLen+1 {
		t.Errorf("after RegisterOutput len = %d, want %d", len(lib), origLen+1)
	}
	found := false
	for _, name := range lib {
		if name == "_test_output_" {
			found = true
			break
		}
	}
	if !found {
		t.Error("_test_output_ not in GetOutputLib")
	}
	for i := 1; i < len(lib); i++ {
		if lib[i] < lib[i-1] {
			t.Errorf("GetOutputLib not sorted after RegisterOutput")
		}
	}
}

func TestRefreshOutput(t *testing.T) {
	if cache.Task == nil {
		cache.Task = &cache.AppConf{}
	}
	oldOutType := cache.Task.OutType
	cache.Task.OutType = "csv"
	defer func() { cache.Task.OutType = oldOutType }()
	RefreshOutput()
}

func TestRefresherFunc(t *testing.T) {
	called := false
	f := refresherFunc(func() { called = true })
	f.Refresh()
	if !called {
		t.Error("Refresh should have been called")
	}
}

func TestPipeline_StartStopCollect(t *testing.T) {
	sp := &spider.Spider{
		Name:     "PipeSpider",
		RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{"r1": {ItemFields: []string{"f1"}}}},
	}
	p := New(sp, "csv", 2)
	p.Start()
	defer p.Stop()

	cell := data.GetDataCell("r1", map[string]interface{}{"f1": "v1"}, "u", "pu", "dt")
	r := p.CollectData(cell)
	if r.IsErr() {
		t.Errorf("CollectData: %v", r.UnwrapErr())
	}
}
