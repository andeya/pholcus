package spider

import (
	// "bytes"
	"code.google.com/p/mahonia"
	// "golang.org/x/text/encoding/simplifiedchinese"
	// "golang.org/x/text/transform"
	// "io/ioutil"
	// "github.com/henrylee2cn/pholcus/downloader/context"
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

// func Encode(src string) (dst string) {
// 	data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(src)), simplifiedchinese.GBK.NewEncoder()))
// 	if err == nil {
// 		dst = string(data)
// 	}
// 	return
// }
// func Decode(src string) (dst string) {
// 	data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(src)), simplifiedchinese.GBK.NewDecoder()))
// 	if err == nil {
// 		dst = string(data)
// 	}
// 	return
// }

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
