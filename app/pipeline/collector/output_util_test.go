package collector

import (
	"strings"
	"testing"

	"github.com/andeya/pholcus/app/spider"
)

func TestJoinNamespaces(t *testing.T) {
	tests := []struct {
		namespace   string
		subNamespace string
		want        string
	}{
		{"", "", ""},
		{"", "sub", "sub"},
		{"ns", "", "ns"},
		{"ns", "sub", "ns__sub"},
		{"a", "b", "a__b"},
	}
	for _, tt := range tests {
		got := joinNamespaces(tt.namespace, tt.subNamespace)
		if got != tt.want {
			t.Errorf("joinNamespaces(%q, %q) = %q, want %q", tt.namespace, tt.subNamespace, got, tt.want)
		}
	}
}

func TestCollector_Namespace(t *testing.T) {
	tests := []struct {
		name      string
		keyin     string
		namespace func(*spider.Spider) string
		want      string
	}{
		{"name_only", "", nil, "Spider"},
		{"name_with_keyin", "kw", nil, ""},
		{"custom", "", func(sp *spider.Spider) string { return "custom_ns" }, "custom_ns"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &spider.Spider{
				Name:     "Spider",
				Keyin:    tt.keyin,
				RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
			}
			if tt.namespace != nil {
				sp.Namespace = tt.namespace
			}
			c := NewCollector(sp, "csv", 1)
			got := c.namespace()
			if tt.namespace != nil {
				if got != tt.want {
					t.Errorf("namespace() = %q, want %q", got, tt.want)
				}
			} else if tt.keyin == "" {
				if got != "Spider" && !strings.HasPrefix(got, "Spider__") {
					t.Errorf("namespace() = %q, want Spider or Spider__<sub>", got)
				}
			} else {
				sub := sp.GetSubName()
				if len(sub) == 0 || got != "Spider__"+sub {
					t.Errorf("namespace() = %q, want Spider__%s", got, sub)
				}
			}
		})
	}
}

func TestCollector_SubNamespace(t *testing.T) {
	tests := []struct {
		name     string
		subNs    func(*spider.Spider, map[string]interface{}) string
		dataCell map[string]interface{}
		wantRule string
	}{
		{"default", nil, map[string]interface{}{"RuleName": "r1"}, "r1"},
		{"custom", func(sp *spider.Spider, dc map[string]interface{}) string { return "custom" }, map[string]interface{}{"RuleName": "r1"}, "custom"},
		{"panic_recovered", func(sp *spider.Spider, dc map[string]interface{}) string { panic("test") }, map[string]interface{}{"RuleName": "r1"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &spider.Spider{
				Name:     "S",
				RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}},
			}
			if tt.subNs != nil {
				sp.SubNamespace = tt.subNs
			}
			c := NewCollector(sp, "csv", 1)
			got := c.subNamespace(tt.dataCell)
			if got != tt.wantRule {
				t.Errorf("subNamespace() = %q, want %q", got, tt.wantRule)
			}
		})
	}
}
