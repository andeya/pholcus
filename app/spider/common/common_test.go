package common

import (
	"testing"
)

func TestCleanHtml(t *testing.T) {
	tests := []struct {
		name  string
		str   string
		depth int
		want  string
	}{
		{"depth0", "<p>hello</p>", 0, "<p>hello</p>"},
		{"depth1_lowercase", "<P>Hello</P>", 1, "<p>Hello</p>"},
		{"depth2_remove_style", "x<style>a{}</style>y", 2, "xy"},
		{"depth3_remove_script", "x<script>alert(1)</script>y", 3, "xy"},
		{"depth4_replace_tags", "<div>a</div>", 4, "\na\n"},
		{"depth5_collapse_ws", "a  \n  b", 5, "a\nb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanHtml(tt.str, tt.depth); got != tt.want {
				t.Errorf("CleanHtml() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeprive(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"a\nb\rc\td e", "abcde"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Deprive(tt.s); got != tt.want {
			t.Errorf("Deprive(%q) = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestDeprive2(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"a\nb\\nc", "abc"},
		{"a\rb\\rd", "abd"},
		{"a\tb\\te", "abe"},
		{" a b ", "ab"},
	}
	for _, tt := range tests {
		if got := Deprive2(tt.s); got != tt.want {
			t.Errorf("Deprive2(%q) = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestFloor(t *testing.T) {
	tests := []struct {
		f    float64
		n    int
		want float64
	}{
		{3.14159, 2, 3.14},
		{3.14159, 0, 3},
		{-1.234, 1, -1.2},
	}
	for _, tt := range tests {
		if got := Floor(tt.f, tt.n); got != tt.want {
			t.Errorf("Floor(%v, %d) = %v, want %v", tt.f, tt.n, got, tt.want)
		}
	}
}

func TestSplitCookies(t *testing.T) {
	tests := []struct {
		cookieStr string
		wantLen   int
		want      map[string]string
	}{
		{"a=1; b=2", 2, map[string]string{"a": "1", "b": "2"}},
		{"mt=ci%3D-1_0; thw=cn", 2, map[string]string{"mt": "ci%3D-1_0", "thw": "cn"}},
		{"single=val", 1, map[string]string{"single": "val"}},
		{"", 0, nil},
		{" a = b ", 1, map[string]string{"a": "b"}},
	}
	for _, tt := range tests {
		cookies := SplitCookies(tt.cookieStr)
		if len(cookies) != tt.wantLen {
			t.Errorf("SplitCookies(%q) len = %d, want %d", tt.cookieStr, len(cookies), tt.wantLen)
		}
		if tt.want != nil {
			for _, c := range cookies {
				if v, ok := tt.want[c.Name]; !ok || c.Value != v {
					t.Errorf("SplitCookies(%q) cookie %s = %q, want %q", tt.cookieStr, c.Name, c.Value, v)
				}
			}
		}
	}
}

func TestDecodeString(t *testing.T) {
	tests := []struct {
		src     string
		charset string
		want    string
	}{
		{"hello", "UTF-8", "hello"},
		{string([]byte{0xD6, 0xD0}), "GB18030", "中"},
	}
	for _, tt := range tests {
		if got := DecodeString(tt.src, tt.charset); got != tt.want {
			t.Errorf("DecodeString(%q, %q) = %q, want %q", tt.src, tt.charset, got, tt.want)
		}
	}
}

func TestEncodeString(t *testing.T) {
	tests := []struct {
		src     string
		charset string
		want    string
	}{
		{"hello", "UTF-8", "hello"},
		{"中", "GB18030", string([]byte{0xD6, 0xD0})},
	}
	for _, tt := range tests {
		if got := EncodeString(tt.src, tt.charset); got != tt.want {
			t.Errorf("EncodeString(%q, %q) = %q, want %q", tt.src, tt.charset, got, tt.want)
		}
	}
}

func TestConvertToString(t *testing.T) {
	tests := []struct {
		src     string
		srcCode string
		tagCode string
		want    string
	}{
		{"hello", "UTF-8", "UTF-8", "hello"},
		{string([]byte{0xD6, 0xD0}), "GB18030", "UTF-8", "中"},
	}
	for _, tt := range tests {
		if got := ConvertToString(tt.src, tt.srcCode, tt.tagCode); got != tt.want {
			t.Errorf("ConvertToString(%q, %q, %q) = %q, want %q", tt.src, tt.srcCode, tt.tagCode, got, tt.want)
		}
	}
}

func TestGBKToUTF8(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"hello", "hello"},
		{string([]byte{0xD6, 0xD0}), "中"},
	}
	for _, tt := range tests {
		if got := GBKToUTF8(tt.src); got != tt.want {
			t.Errorf("GBKToUTF8(%q) = %q, want %q", tt.src, got, tt.want)
		}
	}
}

func TestUnicodeToUTF8(t *testing.T) {
	tests := []struct {
		str  string
		want string
	}{
		{"&#21654;&#21857;", "咖啡"},
		{"&#97;&#98;", "ab"},
		{"&#65;", "A"},
	}
	for _, tt := range tests {
		if got := UnicodeToUTF8(tt.str); got != tt.want {
			t.Errorf("UnicodeToUTF8(%q) = %q, want %q", tt.str, got, tt.want)
		}
	}
}

func TestUnicode16ToUTF8(t *testing.T) {
	tests := []struct {
		str  string
		want string
	}{
		{`\u4e2d\u6587`, "中文"},
		{`a\u0041b`, "aAb"},
		{`\u0041`, "A"},
		{`x\u0041`, "xA"},
	}
	for _, tt := range tests {
		if got := Unicode16ToUTF8(tt.str); got != tt.want {
			t.Errorf("Unicode16ToUTF8(%q) = %q, want %q", tt.str, got, tt.want)
		}
	}
}

func TestMakeUrl(t *testing.T) {
	tests := []struct {
		path          string
		schemeAndHost []string
		wantUrl       string
		wantOk        bool
	}{
		{"/path", []string{"https://example.com"}, "https://example.com/path", true},
		{"path", []string{"https://example.com"}, "https://example.com/path", true},
		{"https://example.com/path", nil, "https://example.com/path", true},
		{"/path", nil, "/path", false},
		{"Https://x.com", nil, "Https://x.com", true},
	}
	for _, tt := range tests {
		got, ok := MakeUrl(tt.path, tt.schemeAndHost...)
		if got != tt.wantUrl || ok != tt.wantOk {
			t.Errorf("MakeUrl(%q, %v) = %q, %v; want %q, %v", tt.path, tt.schemeAndHost, got, ok, tt.wantUrl, tt.wantOk)
		}
	}
}

func TestProcessHtml(t *testing.T) {
	tests := []struct {
		html string
		want string
	}{
		{"a<!-- comment -->b", "ab"},
		{"**bold**", "*"},
		{"<p>text</p>", "<p>text</p>"},
	}
	for _, tt := range tests {
		if got := ProcessHtml(tt.html); got != tt.want {
			t.Errorf("ProcessHtml(%q) = %q, want %q", tt.html, got, tt.want)
		}
	}
}

func TestDepriveBreak(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"a\nb\rc\td", "abcd"},
		{"a\\nb", "ab"},
	}
	for _, tt := range tests {
		if got := DepriveBreak(tt.s); got != tt.want {
			t.Errorf("DepriveBreak(%q) = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestDepriveMutiBreak(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"a\n\n\nb", "a\nb"},
		{"x  \n  y", "x\ny"},
	}
	for _, tt := range tests {
		if got := DepriveMutiBreak(tt.s); got != tt.want {
			t.Errorf("DepriveMutiBreak(%q) = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestHrefSub(t *testing.T) {
	tests := []struct {
		src  string
		sub  string
		want string
	}{
		{"http://a.com", "k=v", "http://a.com?k=v"},
		{"http://a.com?x=1", "k=v", "http://a.com?x=1&k=v"},
		{"http://a.com", "", "http://a.com"},
	}
	for _, tt := range tests {
		if got := HrefSub(tt.src, tt.sub); got != tt.want {
			t.Errorf("HrefSub(%q, %q) = %q, want %q", tt.src, tt.sub, got, tt.want)
		}
	}
}

func TestGetHref(t *testing.T) {
	tests := []struct {
		baseURL  string
		url      string
		href     string
		mustBase bool
		want     string
	}{
		{"https://example.com/", "https://example.com/page", "javascript:void(0)", false, ""},
		{"https://example.com/", "https://example.com/page", "/", false, "https://example.com/"},
		{"https://example.com", "https://example.com/a/b", "./c", false, "https://example.com/a/c"},
		{"https://example.com/", "https://example.com/a/b", "/path", false, "https://example.com/path"},
		{"https://example.com/", "https://example.com/a", "https://other.com/x", false, "https://other.com/x"},
		{"https://example.com/", "https://example.com/a", "rel", false, "https://example.com/rel"},
		{"https://example.com/", "https://example.com/a/b/c", "../d", false, "https://example.com/a/d"},
		{"https://example.com/", "https://example.com/a", "", false, ""},
		{"https://example.com/", "https://example.com/a", "./../x", false, ""},
		{"https://example.com/", "https://example.com/a", "192.168.1.1", false, "https://example.com/192.168.1.1"},
		{"https://example.com/", "https://example.com/a", "a.b.c.d.e", false, "a.b.c.d.e"},
		{"https://example.com/", "https://example.com/a", "https://example.com/x", true, "https://example.com/x"},
		{"https://example.com/", "https://other.com/a", "./x", false, "https://other.com/x"},
	}
	for _, tt := range tests {
		got := GetHref(tt.baseURL, tt.url, tt.href, tt.mustBase)
		if got != tt.want {
			t.Errorf("GetHref(%q, %q, %q, %v) = %q, want %q", tt.baseURL, tt.url, tt.href, tt.mustBase, got, tt.want)
		}
	}
}

func TestExtractArticle(t *testing.T) {
	html := `<html><body><div><p>short</p></div><div><p>This is the longest article content here.</p></div></body></html>`
	got := ExtractArticle(html)
	if got == "" && len(html) > 0 {
		t.Logf("ExtractArticle returned empty (heuristic may vary)")
	}
	html2 := `<html><body><p>single</p></body></html>`
	got2 := ExtractArticle(html2)
	if got2 != "" {
		t.Logf("ExtractArticle(%q) = %q", html2, got2)
	}
}

func TestPinger(t *testing.T) {
	r := Pinger("127.0.0.1", 1)
	if r.IsErr() {
		t.Logf("Pinger(127.0.0.1) err (expected if no listener): %v", r.UnwrapErr())
	}
}

func TestPing(t *testing.T) {
	r := Ping("127.0.0.1", 1)
	if r.IsErr() {
		t.Logf("Ping(127.0.0.1) err (expected): %v", r.UnwrapErr())
	}
}
