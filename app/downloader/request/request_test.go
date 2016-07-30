package request

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
	c, _ := json.Marshal(&a)

	var b = Request{}
	json.Unmarshal(c, &b)

	b.SetTemp("1", map[string]int{"11": 11})
	b.SetTemp("2", []int{22})
	b.SetTemp("4", 44)
	b.SetTemp("5", "55")
	b.SetTemp("x", x{"henry"})

	t.Logf("%#v", b.TempIsJson)
	t.Logf("%#v", b.Temp)

	t.Logf("1：%#v\n", b.GetTemp("1", map[string]int{}))

	t.Logf("2：%#v\n", b.GetTemp("2", []int{}))

	t.Logf("3：%#v\n", b.GetTemp("3", map[string]int{}))

	t.Logf("4：%v\n", b.GetTemp("4", 0))

	t.Logf("5：%#v\n", b.GetTemp("5", ""))

	t.Logf("6：%v\n", b.GetTemp("6", 0))

	t.Logf("x：%v\n", b.GetTemp("x", x{}))

	_b := b.Copy()
	_b.SetTemp("6", 666)
	t.Logf("%#v", _b.TempIsJson)
	t.Logf("%#v", _b.Temp)

	t.Logf("5：%#v\n", _b.GetTemp("5", 1.0))
	t.Logf("5：%#v\n", _b.GetTemp("5", ""))

	t.Logf("6：%#v\n", _b.GetTemp("6", 0))

	t.Logf("x：%v\n", b.GetTemp("x", &x{}))

	t.Logf("10000：%#v\n", _b.GetTemp("10000", 999))
}

type x struct {
	Name string
}
