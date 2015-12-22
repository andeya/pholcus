package context

import (
	"encoding/json"
	"testing"
)

func TestTemp(t *testing.T) {
	var a = Temp{}
	a.Set("3", map[string]int{"33": 33})
	a.Set("6", 66)
	x, _ := json.Marshal(&a)
	t.Logf("%v", string(x))
	var b = Temp{}
	json.Unmarshal(x, &b)

	b.Set("1", map[string]int{"11": 11})
	b.Set("2", []int{22})
	b.Set("4", 44)
	b.Set("5", "55")

	c := map[string]int{}
	b.Get("1", &c)
	t.Logf("%#v\n", c)

	d := []int{}
	b.Get("2", &d)
	t.Logf("%#v\n", d)

	e := map[string]int{}
	b.Get("3", &e)
	t.Logf("%#v\n", e)

	f := 0
	b.Get("4", &f)
	t.Logf("%v\n", f)

	g := ""
	b.Get("5", &g)
	t.Logf("%#v\n", g)

	h := 0
	b.Get("6", &h)
	t.Logf("%v\n", h)

	t.Logf("%#v", b)
}
