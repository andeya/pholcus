package crawler

import (
	"testing"

	spider "github.com/andeya/pholcus/app/spider"
)

func makeSpider(name string, keyin string) *spider.Spider {
	return &spider.Spider{
		Name:  name,
		Keyin: keyin,
		RuleTree: &spider.RuleTree{
			Trunk: map[string]*spider.Rule{},
		},
	}
}

func TestNewSpiderQueue(t *testing.T) {
	q := NewSpiderQueue()
	if q == nil {
		t.Fatal("NewSpiderQueue returned nil")
	}
	if q.Len() != 0 {
		t.Errorf("Len() = %d, want 0", q.Len())
	}
}

func TestSpiderQueue_Add_Len_Reset(t *testing.T) {
	tests := []struct {
		name   string
		adds   []*spider.Spider
		wantLen int
	}{
		{"empty", nil, 0},
		{"one", []*spider.Spider{makeSpider("a", "")}, 1},
		{"two", []*spider.Spider{makeSpider("a", ""), makeSpider("b", "")}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewSpiderQueue()
			for _, sp := range tt.adds {
				q.Add(sp)
			}
			if got := q.Len(); got != tt.wantLen {
				t.Errorf("Len() = %d, want %d", got, tt.wantLen)
			}
			q.Reset()
			if q.Len() != 0 {
				t.Errorf("after Reset Len() = %d, want 0", q.Len())
			}
		})
	}
}

func TestSpiderQueue_AddAll(t *testing.T) {
	list := []*spider.Spider{
		makeSpider("a", ""),
		makeSpider("b", ""),
		makeSpider("c", ""),
	}
	q := NewSpiderQueue()
	q.AddAll(list)
	if got := q.Len(); got != 3 {
		t.Errorf("AddAll Len() = %d, want 3", got)
	}
	all := q.GetAll()
	for i := range list {
		if all[i].GetName() != list[i].GetName() {
			t.Errorf("GetAll()[%d].GetName() = %q, want %q", i, all[i].GetName(), list[i].GetName())
		}
	}
}

func TestSpiderQueue_GetByIndex_GetByIndexOpt(t *testing.T) {
	sp1 := makeSpider("s1", "")
	sp2 := makeSpider("s2", "")
	q := NewSpiderQueue()
	q.Add(sp1)
	q.Add(sp2)

	tests := []struct {
		idx    int
		want   *spider.Spider
		optSome bool
	}{
		{0, sp1, true},
		{1, sp2, true},
		{-1, nil, false},
		{2, nil, false},
		{10, nil, false},
	}
	for _, tt := range tests {
		got := q.GetByIndex(tt.idx)
		if got != tt.want {
			t.Errorf("GetByIndex(%d) = %v, want %v", tt.idx, got, tt.want)
		}
		opt := q.GetByIndexOpt(tt.idx)
		if opt.IsSome() != tt.optSome {
			t.Errorf("GetByIndexOpt(%d).IsSome() = %v, want %v", tt.idx, opt.IsSome(), tt.optSome)
		}
		if opt.IsSome() && opt.Unwrap() != tt.want {
			t.Errorf("GetByIndexOpt(%d).Unwrap() = %v, want %v", tt.idx, opt.Unwrap(), tt.want)
		}
	}
}

func TestSpiderQueue_GetByName_GetByNameOpt(t *testing.T) {
	sp1 := makeSpider("alpha", "")
	sp2 := makeSpider("beta", "")
	q := NewSpiderQueue()
	q.Add(sp1)
	q.Add(sp2)

	tests := []struct {
		name    string
		want    *spider.Spider
		optSome bool
	}{
		{"alpha", sp1, true},
		{"beta", sp2, true},
		{"nonexistent", nil, false},
		{"", nil, false},
	}
	for _, tt := range tests {
		got := q.GetByName(tt.name)
		if got != tt.want {
			t.Errorf("GetByName(%q) = %v, want %v", tt.name, got, tt.want)
		}
		opt := q.GetByNameOpt(tt.name)
		if opt.IsSome() != tt.optSome {
			t.Errorf("GetByNameOpt(%q).IsSome() = %v, want %v", tt.name, opt.IsSome(), tt.optSome)
		}
		if opt.IsSome() && opt.Unwrap() != tt.want {
			t.Errorf("GetByNameOpt(%q).Unwrap() = %v, want %v", tt.name, opt.Unwrap(), tt.want)
		}
	}
}

func TestSpiderQueue_AddKeyins(t *testing.T) {
	tests := []struct {
		name    string
		spiders []*spider.Spider
		keyins  string
		wantLen int
	}{
		{"empty_keyins", []*spider.Spider{makeSpider("a", "")}, "", 1},
		{"no_keyin_spiders", []*spider.Spider{makeSpider("a", "x")}, "<k1><k2>", 1},
		{"with_keyin_spiders", []*spider.Spider{makeSpider("a", spider.KEYIN)}, "<k1><k2>", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewSpiderQueue()
			for _, sp := range tt.spiders {
				q.Add(sp)
			}
			q.AddKeyins(tt.keyins)
			if got := q.Len(); got != tt.wantLen {
				t.Errorf("AddKeyins Len() = %d, want %d", got, tt.wantLen)
			}
		})
	}
}

func TestSpiderQueue_Add_setsID(t *testing.T) {
	q := NewSpiderQueue()
	sp1 := makeSpider("a", "")
	sp2 := makeSpider("b", "")
	q.Add(sp1)
	q.Add(sp2)
	if sp1.GetID() != 0 {
		t.Errorf("first Add ID = %d, want 0", sp1.GetID())
	}
	if sp2.GetID() != 1 {
		t.Errorf("second Add ID = %d, want 1", sp2.GetID())
	}
}

func TestSpiderQueue_GetByIndexOpt(t *testing.T) {
	q := NewSpiderQueue()
	opt := q.GetByIndexOpt(0)
	if opt.IsSome() {
		t.Error("GetByIndexOpt(0) on empty queue should be None")
	}
}

func TestSpiderQueue_AddKeyins_emptyUnit2(t *testing.T) {
	q := NewSpiderQueue()
	q.Add(makeSpider("a", "fixed"))
	q.AddKeyins("<x><y>")
	if q.Len() != 1 {
		t.Errorf("AddKeyins with no KEYIN spiders Len() = %d, want 1", q.Len())
	}
}
