package spider

import (
	"testing"
)

func TestSpiderSpecies_Add(t *testing.T) {
	ss := &SpiderSpecies{
		list: []*Spider{},
		hash: map[string]*Spider{},
	}
	sp1 := &Spider{Name: "TestSpider", RuleTree: &RuleTree{Trunk: map[string]*Rule{}}}
	got := ss.Add(sp1)
	if got != sp1 {
		t.Error("Add should return the spider")
	}
	if sp1.Name != "TestSpider" {
		t.Errorf("Name = %q, want TestSpider", sp1.Name)
	}
	if len(ss.list) != 1 || ss.hash["TestSpider"] != sp1 {
		t.Error("spider not registered")
	}

	sp2 := &Spider{Name: "TestSpider", RuleTree: &RuleTree{Trunk: map[string]*Rule{}}}
	ss.Add(sp2)
	if sp2.Name != "TestSpider(2)" {
		t.Errorf("duplicate Name = %q, want TestSpider(2)", sp2.Name)
	}
	if ss.hash["TestSpider(2)"] != sp2 {
		t.Error("duplicate spider not registered")
	}
}

func TestSpiderSpecies_Get(t *testing.T) {
	ss := &SpiderSpecies{
		list:   []*Spider{},
		hash:   map[string]*Spider{},
		sorted: false,
	}
	sp1 := &Spider{Name: "BSpider", RuleTree: &RuleTree{Trunk: map[string]*Rule{}}}
	sp2 := &Spider{Name: "ASpider", RuleTree: &RuleTree{Trunk: map[string]*Rule{}}}
	ss.Add(sp1)
	ss.Add(sp2)
	got := ss.Get()
	if len(got) != 2 {
		t.Fatalf("Get() len = %d, want 2", len(got))
	}
	if got[0].GetName() != "ASpider" || got[1].GetName() != "BSpider" {
		t.Errorf("Get() should be sorted by pinyin: %q, %q", got[0].GetName(), got[1].GetName())
	}
}

func TestSpiderSpecies_GetByNameOpt(t *testing.T) {
	ss := &SpiderSpecies{
		list: []*Spider{},
		hash: map[string]*Spider{},
	}
	sp := &Spider{Name: "X", RuleTree: &RuleTree{Trunk: map[string]*Rule{}}}
	ss.Add(sp)

	opt := ss.GetByNameOpt("X")
	if !opt.IsSome() {
		t.Fatal("GetByNameOpt(X) should be Some")
	}
	if opt.Unwrap() != sp {
		t.Error("GetByNameOpt returned wrong spider")
	}

	opt2 := ss.GetByNameOpt("Nonexistent")
	if opt2.IsSome() {
		t.Error("GetByNameOpt(Nonexistent) should be None")
	}
}

func TestSpider_GetName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"A", "A"},
		{"", ""},
	}
	for _, tt := range tests {
		sp := &Spider{Name: tt.name}
		if got := sp.GetName(); got != tt.want {
			t.Errorf("GetName() = %q, want %q", got, tt.want)
		}
	}
}

func TestSpider_GetSubName(t *testing.T) {
	sp := &Spider{Keyin: "test", RuleTree: &RuleTree{Trunk: map[string]*Rule{}}}
	got := sp.GetSubName()
	if got == "" {
		t.Error("GetSubName() should not be empty")
	}
	if sp.GetSubName() != got {
		t.Error("GetSubName() should be deterministic")
	}
}

func TestSpider_GetRule(t *testing.T) {
	r1 := &Rule{}
	trunk := map[string]*Rule{"r1": r1}
	sp := &Spider{RuleTree: &RuleTree{Trunk: trunk}}

	if got := sp.GetRule("r1"); got != r1 {
		t.Error("GetRule(r1) mismatch")
	}
	if got := sp.GetRule("missing"); got != nil {
		t.Errorf("GetRule(missing) = %v, want nil", got)
	}
}

func TestSpider_MustGetRule(t *testing.T) {
	r1 := &Rule{}
	sp := &Spider{RuleTree: &RuleTree{Trunk: map[string]*Rule{"r1": r1}}}

	if got := sp.MustGetRule("r1"); got != r1 {
		t.Error("MustGetRule(r1) mismatch")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetRule(missing) should panic")
		}
	}()
	sp.MustGetRule("missing")
}

func TestSpider_GetRules(t *testing.T) {
	trunk := map[string]*Rule{"a": {}, "b": {}}
	sp := &Spider{RuleTree: &RuleTree{Trunk: trunk}}
	got := sp.GetRules()
	if len(got) != 2 || got["a"] == nil || got["b"] == nil {
		t.Errorf("GetRules() = %v", got)
	}
}

func TestSpider_GetItemFields(t *testing.T) {
	rule := &Rule{ItemFields: []string{"a", "b", "c"}}
	sp := &Spider{}
	got := sp.GetItemFields(rule)
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Errorf("GetItemFields() = %v", got)
	}
}

func TestSpider_GetItemField(t *testing.T) {
	rule := &Rule{ItemFields: []string{"a", "b"}}
	sp := &Spider{}
	tests := []struct {
		index int
		want  string
	}{
		{0, "a"},
		{1, "b"},
		{-1, ""},
		{2, ""},
	}
	for _, tt := range tests {
		if got := sp.GetItemField(rule, tt.index); got != tt.want {
			t.Errorf("GetItemField(rule, %d) = %q, want %q", tt.index, got, tt.want)
		}
	}
}

func TestSpider_GetItemFieldIndex(t *testing.T) {
	rule := &Rule{ItemFields: []string{"x", "y", "z"}}
	sp := &Spider{}
	tests := []struct {
		field string
		want  int
	}{
		{"x", 0},
		{"y", 1},
		{"z", 2},
		{"missing", -1},
	}
	for _, tt := range tests {
		if got := sp.GetItemFieldIndex(rule, tt.field); got != tt.want {
			t.Errorf("GetItemFieldIndex(rule, %q) = %d, want %d", tt.field, got, tt.want)
		}
	}
}

func TestSpider_UpsertItemField(t *testing.T) {
	rule := &Rule{ItemFields: []string{"a", "b"}}
	sp := &Spider{}
	if got := sp.UpsertItemField(rule, "c"); got != 2 {
		t.Errorf("UpsertItemField(c) = %d, want 2", got)
	}
	if got := sp.UpsertItemField(rule, "a"); got != 0 {
		t.Errorf("UpsertItemField(a) existing = %d, want 0", got)
	}
	if len(rule.ItemFields) != 3 {
		t.Errorf("ItemFields len = %d, want 3", len(rule.ItemFields))
	}
}

func TestSpider_GetID_SetID(t *testing.T) {
	sp := &Spider{}
	sp.SetID(42)
	if sp.GetID() != 42 {
		t.Errorf("GetID() = %d, want 42", sp.GetID())
	}
}

func TestSpider_GetKeyin_SetKeyin(t *testing.T) {
	sp := &Spider{}
	sp.SetKeyin("kw")
	if sp.GetKeyin() != "kw" {
		t.Errorf("GetKeyin() = %q, want kw", sp.GetKeyin())
	}
}

func TestSpider_GetLimit_SetLimit(t *testing.T) {
	sp := &Spider{}
	sp.SetLimit(100)
	if sp.GetLimit() != 100 {
		t.Errorf("GetLimit() = %d, want 100", sp.GetLimit())
	}
}

func TestSpider_GetEnableCookie(t *testing.T) {
	sp := &Spider{EnableCookie: true}
	if !sp.GetEnableCookie() {
		t.Error("GetEnableCookie() = false")
	}
}

func TestSpider_GetDescription(t *testing.T) {
	sp := &Spider{Description: "desc"}
	if sp.GetDescription() != "desc" {
		t.Errorf("GetDescription() = %q", sp.GetDescription())
	}
}

func TestSpider_SetPausetime(t *testing.T) {
	sp := &Spider{}
	sp.SetPausetime(1000)
	if sp.Pausetime != 1000 {
		t.Errorf("SetPausetime = %d, want 1000", sp.Pausetime)
	}
	sp.SetPausetime(500)
	if sp.Pausetime != 1000 {
		t.Errorf("SetPausetime without runtime should not overwrite: %d", sp.Pausetime)
	}
	sp.SetPausetime(200, true)
	if sp.Pausetime != 200 {
		t.Errorf("SetPausetime(runtime=true) = %d, want 200", sp.Pausetime)
	}
}

func TestSpider_OutDefaultField(t *testing.T) {
	sp := &Spider{NotDefaultField: false}
	if !sp.OutDefaultField() {
		t.Error("OutDefaultField() = false")
	}
	sp.NotDefaultField = true
	if sp.OutDefaultField() {
		t.Error("OutDefaultField() = true when NotDefaultField")
	}
}

func TestSpider_Copy(t *testing.T) {
	sp := &Spider{
		Name:        "S",
		Description: "D",
		RuleTree: &RuleTree{
			Trunk: map[string]*Rule{
				"r1": {ItemFields: []string{"f1"}},
			},
		},
	}
	cp := sp.Copy()
	if cp.Name != sp.Name || cp.Description != sp.Description {
		t.Error("Copy name/description mismatch")
	}
	if cp.RuleTree.Trunk["r1"] == sp.RuleTree.Trunk["r1"] {
		t.Error("Copy should share Rule")
	}
	if len(cp.RuleTree.Trunk["r1"].ItemFields) != 1 || cp.RuleTree.Trunk["r1"].ItemFields[0] != "f1" {
		t.Error("Copy ItemFields mismatch")
	}
}
