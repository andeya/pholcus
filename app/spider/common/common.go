package common

import (
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/henrylee2cn/pholcus/common/mahonia"
	"github.com/henrylee2cn/pholcus/common/ping"
)

// 清除标签
func CleanHtml(str string, depth int) string {
	if depth > 0 {
		//将HTML标签全转换成小写
		re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
		str = re.ReplaceAllStringFunc(str, strings.ToLower)
	}
	if depth > 1 {
		//去除STYLE
		re, _ := regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
		str = re.ReplaceAllString(str, "")
	}
	if depth > 2 {
		//去除SCRIPT
		re, _ := regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
		str = re.ReplaceAllString(str, "")
	}
	if depth > 3 {
		//去除所有尖括号内的HTML代码，并换成换行符
		re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
		str = re.ReplaceAllString(str, "\n")
	}
	if depth > 4 {
		//去除连续的换行符
		re, _ := regexp.Compile("\\s{2,}")
		str = re.ReplaceAllString(str, "\n")
	}
	return str
}

// 提取文章页的正文
// 思路：认为文本节点最长的标签的父标签为文章正文
func ExtractArticle(html string) string {
	//将HTML标签全转换成小写
	re := regexp.MustCompile("<[\\S\\s]+?>")
	html = re.ReplaceAllStringFunc(html, strings.ToLower)
	//去除head
	re = regexp.MustCompile("<head[\\S\\s]+?</head>")
	html = re.ReplaceAllString(html, "")
	//去除STYLE
	re = regexp.MustCompile("<style[\\S\\s]+?</style>")
	html = re.ReplaceAllString(html, "")
	//去除SCRIPT
	re = regexp.MustCompile("<script[\\S\\s]+?</script>")
	html = re.ReplaceAllString(html, "")
	//去除注释
	re = regexp.MustCompile("<![\\S\\s]+?>")
	html = re.ReplaceAllString(html, "")
	// fmt.Println(html)

	// 获取每个子标签
	re = regexp.MustCompile("<[A-Za-z]+[^<]*>([^<>]+)</[A-Za-z]+>")
	ss := re.FindAllStringSubmatch(html, -1)
	// fmt.Printf("所有子标签：\n%#v\n", ss)

	var maxLen int
	var idx int
	for k, v := range ss {
		l := len([]rune(v[1]))
		if l > maxLen {
			maxLen = l
			idx = k
		}
	}
	// fmt.Println("最长段落：", ss[idx][0])

	html = strings.Replace(html, ss[idx][0], `<pholcus id="pholcus">`+ss[idx][1]+`</pholcus>`, -1)
	r := strings.NewReader(html)
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return ""
	}
	return dom.Find("pholcus#pholcus").Parent().Text()
}

// 去除常见转义字符
func Deprive(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, ` `, "", -1)
	return s
}

// 去除常见转义字符
func Deprive2(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, `\n`, "", -1)
	s = strings.Replace(s, `\r`, "", -1)
	s = strings.Replace(s, `\t`, "", -1)
	s = strings.Replace(s, ` `, "", -1)
	return s
}

// 舍去尾数
func Floor(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f)*pow10_n) / pow10_n
}

// cookies字符串转[]*http.Cookie，（如"mt=ci%3D-1_0; thw=cn; sec=5572dc7c40ce07d4e8c67e4879a; v=0;"）
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

func GBKToUTF8(src string) string {
	return DecodeString(EncodeString(src, "ISO-8859-1"), "GBK")
}

//将"&#21654;&#21857;&#33394;&#124;&#32511;&#33394;"转为"咖啡色|绿色"
func UnicodeToUTF8(str string) string {
	str = strings.TrimLeft(str, "&#")
	str = strings.TrimRight(str, ";")
	strSlice := strings.Split(str, ";&#")

	for k, s := range strSlice {
		if i, err := strconv.Atoi(s); err == nil {
			strSlice[k] = string(i)
		}
	}
	return strings.Join(strSlice, "")
}

//将`{"area":[["quanguo","\u5168\u56fd\u8054\u9500"]]｝`转为`{"area":[["quanguo","全国联销"]]｝`
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
			strSlice[i] = string(x)
		}
	}
	return strings.Join(strSlice, "")
}

//@SchemeAndHost https://www.baidu.com
//@path /search?w=x
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

func Pinger(address string, timeoutSecond int) error {
	return ping.Pinger(address, timeoutSecond)
}

func Ping(address string, timeoutSecond int) (alive bool, err error, timedelay time.Duration) {
	return ping.Ping(address, timeoutSecond)
}
