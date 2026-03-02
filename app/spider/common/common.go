package common

import (
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/goquery"
	"github.com/andeya/pholcus/common/mahonia"
	"github.com/andeya/pholcus/common/ping"
)

// CleanHtml strips HTML tags at increasing levels of aggressiveness based on depth.
func CleanHtml(str string, depth int) string {
	if depth > 0 {
		re, _ := regexp.Compile("<[\\S\\s]+?>")
		str = re.ReplaceAllStringFunc(str, strings.ToLower)
	}
	if depth > 1 {
		re, _ := regexp.Compile("<style[\\S\\s]+?</style>")
		str = re.ReplaceAllString(str, "")
	}
	if depth > 2 {
		re, _ := regexp.Compile("<script[\\S\\s]+?</script>")
		str = re.ReplaceAllString(str, "")
	}
	if depth > 3 {
		re, _ := regexp.Compile("<[\\S\\s]+?>")
		str = re.ReplaceAllString(str, "\n")
	}
	if depth > 4 {
		re, _ := regexp.Compile("\\s{2,}")
		str = re.ReplaceAllString(str, "\n")
	}
	return str
}

// ExtractArticle extracts the main article body from an HTML page.
// Heuristic: the parent of the tag with the longest text node is treated as the article body.
func ExtractArticle(html string) string {
	re := regexp.MustCompile("<[\\S\\s]+?>")
	html = re.ReplaceAllStringFunc(html, strings.ToLower)
	re = regexp.MustCompile("<head[\\S\\s]+?</head>")
	html = re.ReplaceAllString(html, "")
	re = regexp.MustCompile("<style[\\S\\s]+?</style>")
	html = re.ReplaceAllString(html, "")
	re = regexp.MustCompile("<script[\\S\\s]+?</script>")
	html = re.ReplaceAllString(html, "")
	re = regexp.MustCompile("<![\\S\\s]+?>")
	html = re.ReplaceAllString(html, "")

	re = regexp.MustCompile("<[A-Za-z]+[^<]*>([^<>]+)</[A-Za-z]+>")
	ss := re.FindAllStringSubmatch(html, -1)

	var maxLen int
	var idx int
	for k, v := range ss {
		l := len([]rune(v[1]))
		if l > maxLen {
			maxLen = l
			idx = k
		}
	}

	html = strings.ReplaceAll(html, ss[idx][0], `<pholcus id="pholcus">`+ss[idx][1]+`</pholcus>`)
	r := strings.NewReader(html)
	docResult := goquery.NewDocumentFromReader(r)
	if docResult.IsErr() {
		return ""
	}
	return docResult.Unwrap().Find("pholcus#pholcus").Parent().Text()
}

// Deprive removes common whitespace escape characters.
func Deprive(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, ` `, "")
	return s
}

// Deprive2 removes both actual and literal whitespace escape sequences.
func Deprive2(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, `\n`, "")
	s = strings.ReplaceAll(s, `\r`, "")
	s = strings.ReplaceAll(s, `\t`, "")
	s = strings.ReplaceAll(s, ` `, "")
	return s
}

// Floor truncates f to n decimal places.
func Floor(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f)*pow10_n) / pow10_n
}

// SplitCookies parses a cookie string (e.g. "mt=ci%3D-1_0; thw=cn; v=0;") into []*http.Cookie.
func SplitCookies(cookieStr string) (cookies []*http.Cookie) {
	slice := strings.Split(cookieStr, ";")
	for _, v := range slice {
		oneCookie := &http.Cookie{}
		s := strings.Split(v, "=")
		if len(s) == 2 {
			oneCookie.Name = strings.Trim(s[0], " ")
			oneCookie.Value = strings.Trim(s[1], " ")
			cookies = append(cookies, oneCookie)
		}
	}
	return
}

func DecodeString(src, charset string) string {
	return mahonia.NewDecoder(charset).ConvertString(src)
}

func EncodeString(src, charset string) string {
	return mahonia.NewEncoder(charset).ConvertString(src)
}

func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

// func GBKToUTF8(src string) string {
// 	return DecodeString(EncodeString(src, "ISO-8859-1"), "GBK")
// }

func GBKToUTF8(src string) string {
	return DecodeString(src, "GB18030")
}

// UnicodeToUTF8 converts HTML numeric character references (e.g. "&#21654;&#21857;") to UTF-8.
func UnicodeToUTF8(str string) string {
	str = strings.TrimLeft(str, "&#")
	str = strings.TrimRight(str, ";")
	strSlice := strings.Split(str, ";&#")

	for k, s := range strSlice {
		if i, err := strconv.Atoi(s); err == nil {
			strSlice[k] = string(rune(i))
		}
	}
	return strings.Join(strSlice, "")
}

// Unicode16ToUTF8 converts \uXXXX escape sequences in a string to UTF-8 characters.
func Unicode16ToUTF8(str string) string {
	i := 0
	if strings.Index(str, `\u`) > 0 {
		i = 1
	}
	strSlice := strings.Split(str, `\u`)
	last := len(strSlice) - 1
	if len(strSlice[last]) > 4 {
		strSlice = append(strSlice, string(strSlice[last][4:]))
		strSlice[last] = string(strSlice[last][:4])
	}
	for ; i <= last; i++ {
		if x, err := strconv.ParseInt(strSlice[i], 16, 32); err == nil {
			strSlice[i] = string(rune(x))
		}
	}
	return strings.Join(strSlice, "")
}

// @SchemeAndHost https://www.baidu.com
// @path /search?w=x
func MakeUrl(path string, schemeAndHost ...string) (string, bool) {
	if string(path[0]) != "/" && strings.ToLower(string(path[0])) != "h" {
		path = "/" + path
	}
	u := path
	idx := strings.Index(path, "://")
	if idx < 0 {
		if len(schemeAndHost) > 0 {
			u = schemeAndHost[0] + u
		} else {
			return u, false
		}
	}
	_, err := url.Parse(u)
	if err != nil {
		return u, false
	}
	return u, true
}

func Pinger(address string, timeoutSecond int) result.VoidResult {
	return ping.Pinger(address, timeoutSecond)
}

func Ping(address string, timeoutSecond int) result.Result[ping.PingResult] {
	return ping.Ping(address, timeoutSecond)
}

// htmlReg matches comment blocks and blank lines for HTML filtering.
var htmlReg = regexp.MustCompile(`(\*{1,2}[\s\S]*?\*)|(<!-[\s\S]*?-->)|(^\s*\n)`)

// ProcessHtml removes comments from an HTML string.
func ProcessHtml(html string) string {
	html = htmlReg.ReplaceAllString(html, "")
	return html
}

// DepriveBreak removes all line-break characters (both actual and literal escape sequences).
func DepriveBreak(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, `\n`, "")
	s = strings.ReplaceAll(s, `\r`, "")
	s = strings.ReplaceAll(s, `\t`, "")
	return s
}

// DepriveMutiBreak collapses consecutive blank lines into a single newline.
func DepriveMutiBreak(s string) string {
	re, _ := regexp.Compile(`([^\n\f\r\t 　 ]*)([ 　 ]*[\n\f\r\t]+[ 　 ]*)+`)
	return re.ReplaceAllString(s, "${1}\n")

}

// HrefSub appends query parameters to an existing URL.
func HrefSub(src string, sub string) string {
	if len(sub) > 0 {
		if strings.Index(src, "?") > -1 {
			src += "&" + sub
		} else {
			src += "?" + sub
		}
	}
	return src
}

var domainReg = regexp.MustCompile(`([a-zA-Z0-9]+://([a-zA-Z0-9\:\_\-\.])+(/)?)(.)*`)

// GetHerf resolves a relative or absolute href against a base URL and current page URL.
func GetHerf(baseurl string, url string, herf string, mustBase bool) string {
	if strings.HasPrefix(herf, `javascript:`) {
		return ``
	}
	result := ""
	herf = Deprive2(herf)
	if !strings.HasSuffix(baseurl, "/") {
		baseurl += "/"
	}

	if !mustBase && !strings.HasPrefix(url, baseurl) {
		baseurl = domainReg.ReplaceAllString(url, "$1")
	}

	refIndex := strings.LastIndex(url, "/") + 1
	/*sub := url[refIndex:]
	if !strings.HasSuffix(url, "/") {
		if strings.Index(sub, ".") == -1 &&
			strings.Index(sub, "?") == -1 &&
			strings.Index(sub, "#") == -1 {
			url = url + `/`
		} else {
			url = url[:refIndex]
		}
	}*/
	url = url[:refIndex]

	/*refIndex = strings.LastIndex(herf, "/") + 1
	sub = herf[refIndex:]
	if len(sub) > 0 &&
		strings.Index(sub, ".") == -1 &&
		strings.Index(sub, "?") == -1 &&
		strings.Index(sub, "#") == -1 {
		herf = herf + `/`
	}*/

	if strings.HasPrefix(herf, "./../") {
		herf = strings.Replace(herf, "./", "", 1)
	}

	if len(herf) == 0 {
		result = ""
	} else if herf == "/" {
		result = baseurl
	} else if strings.HasPrefix(herf, "./") {
		/*reg := regexp.MustCompile(`^(./)(.*)`)
		result = url + strings.Trim(reg.ReplaceAllString(herf, "$2"), " ")*/
		result = url + strings.Replace(herf, "./", "", 1)
	} else if strings.HasPrefix(herf, "/") {
		//reg = regexp.MustCompile(`^(http)(s)?(://)([0-9A-Za-z.\-_]+)(/)(.*)`)
		result = strings.Trim(baseurl, " ") + herf[1:]
	} else if mustBase && !strings.HasPrefix(herf, baseurl) &&
		(strings.Index(herf, "://") > -1 ||
			(strings.Index(herf, "/") == -1 &&
				strings.Count(herf, ".") > 3)) { //IP

		result = ""
	} else if strings.Index(herf, "://") > -1 ||
		(strings.Index(herf, "/") == -1 && strings.Count(herf, ".") > 3) { //IP
		result = herf
	} else {
		count := strings.Count(herf, "../")
		if count > 0 {
			urlArr := strings.SplitAfter(url, "/")
			len := cap(urlArr) - count - 1
			if len > 2 {
				preUrl := ""
				for i, str := range urlArr {
					if len > i {
						preUrl += str
					}
				}
				result = preUrl + strings.ReplaceAll(herf, "../", "")
			}
		} else {
			result = url + herf
		}
	}

	/*if strings.Count(result, "://") > 1 {
		result = strings.SplitN(result, "://", 2)[1]
	}*/

	return result
}
