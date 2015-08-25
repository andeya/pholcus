package common

import (
	// "bytes"
	"github.com/henrylee2cn/mahonia"
	// "golang.org/x/text/encoding/simplifiedchinese"
	// "golang.org/x/text/transform"
	// "io/ioutil"
	// "github.com/henrylee2cn/pholcus/downloader/context"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

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
			// r := rune(i)
			strSlice[k] = string(i)
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
