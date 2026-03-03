package request

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
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

	t.Logf("%#v", b.TempIsJSON)
	t.Logf("%#v", b.Temp)

	t.Logf("1：%#v\n", b.GetTemp("1", map[string]int{}))

	t.Logf("2：%#v\n", b.GetTemp("2", []int{}))

	t.Logf("3：%#v\n", b.GetTemp("3", map[string]int{}))

	t.Logf("4：%v\n", b.GetTemp("4", 0))

	t.Logf("5：%#v\n", b.GetTemp("5", ""))

	t.Logf("6：%v\n", b.GetTemp("6", 0))

	t.Logf("x：%v\n", b.GetTemp("x", x{}))

	_b := b.Copy().Unwrap()
	_b.SetTemp("6", 666)
	t.Logf("%#v", _b.TempIsJSON)
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

func TestPrepare(t *testing.T) {
	t.Run("invalid URL", func(t *testing.T) {
		r := &Request{URL: "://invalid"}
		res := r.Prepare()
		if res.IsOk() {
			t.Error("expected Prepare to fail for invalid URL")
		}
	})

	t.Run("edge cases", func(t *testing.T) {
		tests := []struct {
			name string
			req  *Request
			chk  func(*Request)
		}{
			{
				name: "negative DialTimeout",
				req:  &Request{URL: "http://a.com", Rule: "r", DialTimeout: -1},
				chk:  func(r *Request) { r.Prepare(); if r.DialTimeout != 0 { t.Errorf("DialTimeout=%v", r.DialTimeout) } },
			},
			{
				name: "negative ConnTimeout",
				req:  &Request{URL: "http://a.com", Rule: "r", ConnTimeout: -1},
				chk:  func(r *Request) { r.Prepare(); if r.ConnTimeout != 0 { t.Errorf("ConnTimeout=%v", r.ConnTimeout) } },
			},
			{
				name: "negative Priority",
				req:  &Request{URL: "http://a.com", Rule: "r", Priority: -5},
				chk:  func(r *Request) { r.Prepare(); if r.Priority != 0 { t.Errorf("Priority=%v", r.Priority) } },
			},
			{
				name: "DownloaderID out of range low",
				req:  &Request{URL: "http://a.com", Rule: "r", DownloaderID: -1},
				chk:  func(r *Request) { r.Prepare(); if r.DownloaderID != SurfID { t.Errorf("DownloaderID=%v", r.DownloaderID) } },
			},
			{
				name: "DownloaderID out of range high",
				req:  &Request{URL: "http://a.com", Rule: "r", DownloaderID: 99},
				chk:  func(r *Request) { r.Prepare(); if r.DownloaderID != SurfID { t.Errorf("DownloaderID=%v", r.DownloaderID) } },
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.chk(tt.req)
			})
		}
	})
}

func TestSerializeUnSerialize(t *testing.T) {
	r := &Request{
		Spider: "s", URL: "http://example.com", Rule: "r",
		Method: "POST", PostData: "a=1",
		Header:       http.Header{"X-Custom": {"v"}},
		EnableCookie: true,
		Temp:         Temp{"k": "v"},
	}
	r.Prepare()

	res := r.Serialize()
	if res.IsErr() {
		t.Fatalf("Serialize: %v", res.Err())
	}
	s := res.Unwrap()
	if s == "" {
		t.Error("Serialize returned empty string")
	}

	ures := UnSerialize(s)
	if ures.IsErr() {
		t.Fatalf("UnSerialize: %v", ures.Err())
	}
	req := ures.Unwrap()
	if req.URL != r.URL || req.Method != r.Method || req.Spider != r.Spider {
		t.Errorf("UnSerialize mismatch: got %+v", req)
	}
}

func TestUnSerializeInvalid(t *testing.T) {
	res := UnSerialize("invalid json {{{")
	if res.IsOk() {
		t.Error("expected UnSerialize to fail")
	}
}

func TestUnique(t *testing.T) {
	r := &Request{Spider: "s", Rule: "r", URL: "http://a.com", Method: "GET"}
	r.Prepare()
	u1 := r.Unique()
	u2 := r.Unique()
	if u1 != u2 || len(u1) != 32 {
		t.Errorf("Unique: %q vs %q", u1, u2)
	}
}

func TestCopy(t *testing.T) {
	r := &Request{Spider: "s", URL: "http://a.com", Rule: "r"}
	r.Prepare()
	r.SetTemp("x", 1)
	cres := r.Copy()
	if cres.IsErr() {
		t.Fatal(cres.Err())
	}
	c := cres.Unwrap()
	if c.URL != r.URL || c.Spider != r.Spider {
		t.Errorf("Copy mismatch")
	}
	if v, ok := c.GetTemp("x", 0).(float64); !ok || v != 1 {
		t.Errorf("Copy Temp mismatch: got %v", c.GetTemp("x", 0))
	}
}

func TestGettersSetters(t *testing.T) {
	r := &Request{URL: "http://a.com", Rule: "r"}
	r.Prepare()

	tests := []struct {
		name string
		fn   func()
	}{
		{"GetURL", func() { r.SetURL("http://u.com"); if r.GetURL() != "http://u.com" { t.Error("GetURL") } }},
		{"GetMethod", func() { r.SetMethod("post"); if r.GetMethod() != "POST" { t.Error("GetMethod") } }},
		{"GetReferer", func() { r.SetReferer("http://ref"); if r.GetReferer() != "http://ref" { t.Error("GetReferer") } }},
		{"GetPostData", func() { r.PostData = "p=1"; if r.GetPostData() != "p=1" { t.Error("GetPostData") } }},
		{"GetHeader", func() { r.SetHeader("A", "1"); if r.GetHeader().Get("A") != "1" { t.Error("GetHeader") } }},
		{"AddHeader", func() { r.AddHeader("B", "2"); if r.GetHeader().Get("B") != "2" { t.Error("AddHeader") } }},
		{"GetEnableCookie", func() { r.SetEnableCookie(true); if !r.GetEnableCookie() { t.Error("GetEnableCookie") } }},
		{"GetCookies", func() { r.SetCookies("c=1"); if r.GetCookies() != "c=1" { t.Error("GetCookies") } }},
		{"GetDialTimeout", func() { r.DialTimeout = 5 * time.Second; if r.GetDialTimeout() != 5*time.Second { t.Error("GetDialTimeout") } }},
		{"GetConnTimeout", func() { r.ConnTimeout = 10 * time.Second; if r.GetConnTimeout() != 10*time.Second { t.Error("GetConnTimeout") } }},
		{"GetTryTimes", func() { r.TryTimes = 5; if r.GetTryTimes() != 5 { t.Error("GetTryTimes") } }},
		{"GetRetryPause", func() { r.RetryPause = 3 * time.Second; if r.GetRetryPause() != 3*time.Second { t.Error("GetRetryPause") } }},
		{"GetProxy", func() { r.SetProxy("http://p"); if r.GetProxy() != "http://p" { t.Error("GetProxy") } }},
		{"GetRedirectTimes", func() { r.RedirectTimes = 2; if r.GetRedirectTimes() != 2 { t.Error("GetRedirectTimes") } }},
		{"GetRuleName", func() { r.SetRuleName("r1"); if r.GetRuleName() != "r1" { t.Error("GetRuleName") } }},
		{"GetSpiderName", func() { r.SetSpiderName("sp"); if r.GetSpiderName() != "sp" { t.Error("GetSpiderName") } }},
		{"IsReloadable", func() { r.SetReloadable(true); if !r.IsReloadable() { t.Error("IsReloadable") } }},
		{"GetPriority", func() { r.SetPriority(3); if r.GetPriority() != 3 { t.Error("GetPriority") } }},
		{"GetDownloaderID", func() { r.SetDownloaderID(PhantomID); if r.GetDownloaderID() != PhantomID { t.Error("GetDownloaderID") } }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn()
		})
	}
}

func TestGetTempOpt(t *testing.T) {
	r := &Request{URL: "http://a.com", Rule: "r", Temp: Temp{"a": 1}}
	r.Prepare()

	if opt := r.GetTempOpt("missing"); opt.IsSome() {
		t.Error("expected None for missing key")
	}
	if opt := r.GetTempOpt("a"); !opt.IsSome() || opt.Unwrap() != 1 {
		t.Errorf("GetTempOpt(a)=%v", opt)
	}

	r.SetTemp("j", map[string]int{"x": 1})
	sres := r.Serialize()
	if sres.IsErr() {
		t.Fatal(sres.Err())
	}
	ures := UnSerialize(sres.Unwrap())
	if ures.IsErr() {
		t.Fatal(ures.Err())
	}
	req := ures.Unwrap()
	if opt := req.GetTempOpt("j"); !opt.IsSome() {
		t.Error("GetTempOpt(j) expected Some")
	}
}

func TestGetTemps(t *testing.T) {
	r := &Request{URL: "http://a.com", Rule: "r", Temp: Temp{"k": "v"}}
	r.Prepare()
	temps := r.GetTemps()
	if temps["k"] != "v" {
		t.Errorf("GetTemps=%v", temps)
	}
}

func TestSetTemps(t *testing.T) {
	r := &Request{URL: "http://a.com", Rule: "r"}
	r.Prepare()
	r.SetTemps(map[string]interface{}{"x": 1, "y": "2"})
	if r.Temp["x"] != 1 || r.Temp["y"] != "2" {
		t.Errorf("SetTemps=%v", r.Temp)
	}
}

func TestGetTempPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for nil defaultValue")
		}
	}()
	r := &Request{URL: "http://a.com", Rule: "r"}
	r.Prepare()
	r.GetTemp("k", nil)
}
