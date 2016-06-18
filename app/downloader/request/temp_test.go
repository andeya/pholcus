package request

import (
	"encoding/json"
	"testing"
)

func TestTemp(t *testing.T) {
	var a = Temp{}
	a.set("3", map[string]int{"33": 33})
	a.set("6", 66)
	x, _ := json.Marshal(&a)
	t.Logf("%v", string(x))
	var b = Temp{}
	json.Unmarshal(x, &b)

	b.set("1", map[string]int{"11": 11})
	b.set("2", []int{22})
	b.set("4", 44)
	b.set("5", "55")

	c := map[string]int{}
	b.get("1", &c)
	t.Logf("%#v\n", c)

	d := []int{}
	b.get("2", &d)
	t.Logf("%#v\n", d)

	e := map[string]int{}
	b.get("3", &e)
	t.Logf("%#v\n", e)

	f := 0
	b.get("4", &f)
	t.Logf("%v\n", f)

	g := ""
	b.get("5", &g)
	t.Logf("%#v\n", g)

	h := 0
	b.get("6", &h)
	t.Logf("%v\n", h)

	t.Logf("%#v", b)
}
