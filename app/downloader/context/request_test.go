package context

import (
	"encoding/json"
	"testing"
)

func TestReqTemp(t *testing.T) {
	var a = &Request{
		Temp: Temp{"3": map[string]int{"33": 33}},
	}
	a.Prepare()
	a.SetTemp("6", 66)
	x, _ := json.Marshal(&a)
	t.Logf("%v", string(x))
	var b = Request{}
	json.Unmarshal(x, &b)

	b.SetTemp("1", map[string]int{"11": 11})
	b.SetTemp("2", []int{22})
	b.SetTemp("4", 44)
	b.SetTemp("5", "55")

	c := map[string]int{}
	b.GetTemp("1", &c)
	t.Logf("%#v\n", c)

	d := []int{}
	b.GetTemp("2", &d)
	t.Logf("%#v\n", d)

	e := map[string]int{}
	b.GetTemp("3", &e)
	t.Logf("%#v\n", e)

	f := 0
	b.GetTemp("4", &f)
	t.Logf("%v\n", f)

	g := ""
	b.GetTemp("5", &g)
	t.Logf("%#v\n", g)

	h := 0
	b.GetTemp("6", &h)
	t.Logf("%v\n", h)

	_b := b.Copy()
	_b.SetTemp("6", 666)

	i := ""
	_b.GetTemp("5", &i)
	t.Logf("%#v\n", i)

	t.Logf("%#v\n", _b.GetTemp("5", ""))

	t.Logf("%#v", _b)
}
