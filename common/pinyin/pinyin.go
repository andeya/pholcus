package pinyin

import (
	"regexp"
	"strings"
)

// Meta
const (
	Version   = "0.2.1"
	Author    = "mozillazg, 闲耘"
	License   = "MIT"
	Copyright = "Copyright (c) 2014 mozillazg, 闲耘"
)

// 拼音风格(推荐)
const (
	Normal      = 0 // 普通风格，不带声调（默认风格）。如： pin yin
	Tone        = 1 // 声调风格1，拼音声调在韵母第一个字母上。如： pīn yīn
	Tone2       = 2 // 声调风格2，即拼音声调在各个拼音之后，用数字 [0-4] 进行表示。如： pi1n yi1n
	Initials    = 3 // 声母风格，只返回各个拼音的声母部分。如： 中国 的拼音 zh g
	FirstLetter = 4 // 首字母风格，只返回拼音的首字母部分。如： p y
	Finals      = 5 // 韵母风格1，只返回各个拼音的韵母部分，不带声调。如： ong uo
	FinalsTone  = 6 // 韵母风格2，带声调，声调在韵母第一个字母上。如： ōng uó
	FinalsTone2 = 7 // 韵母风格2，带声调，声调在各个拼音之后，用数字 [0-4] 进行表示。如： o1ng uo2
)

// 拼音风格(兼容之前的版本)
const (
	NORMAL       = 0 // 普通风格，不带声调（默认风格）。如： pin yin
	TONE         = 1 // 声调风格1，拼音声调在韵母第一个字母上。如： pīn yīn
	TONE2        = 2 // 声调风格2，即拼音声调在各个拼音之后，用数字 [0-4] 进行表示。如： pi1n yi1n
	INITIALS     = 3 // 声母风格，只返回各个拼音的声母部分。如： 中国 的拼音 zh g
	FIRST_LETTER = 4 // 首字母风格，只返回拼音的首字母部分。如： p y
	FINALS       = 5 // 韵母风格1，只返回各个拼音的韵母部分，不带声调。如： ong uo
	FINALS_TONE  = 6 // 韵母风格2，带声调，声调在韵母第一个字母上。如： ōng uó
	FINALS_TONE2 = 7 // 韵母风格2，带声调，声调在各个拼音之后，用数字 [0-4] 进行表示。如： o1ng uo2
)

// 声母表
var initials = strings.Split(
	"b,p,m,f,d,t,n,l,g,k,h,j,q,x,r,zh,ch,sh,z,c,s",
	",",
)

// 所有带声调的字符
var rePhoneticSymbolSource = func(m map[string]string) string {
	s := ""
	for k := range m {
		s = s + k
	}
	return s
}(phoneticSymbol)

// 匹配带声调字符的正则表达式
var rePhoneticSymbol = regexp.MustCompile("[" + rePhoneticSymbolSource + "]")

// 匹配使用数字标识声调的字符的正则表达式
var reTone2 = regexp.MustCompile("([aeoiuvnm])([0-4])$")

// Args 配置信息
type Args struct {
	Style     int    // 拼音风格（默认： Normal)
	Heteronym bool   // 是否启用多音字模式（默认：禁用）
	Separator string // Slug 中使用的分隔符（默认：-)
}

// 默认配置：风格
var Style = Normal

// 默认配置：是否启用多音字模式
var Heteronym = false

// 默认配置： `Slug` 中 Join 所用的分隔符
var Separator = "-"

// NewArgs 返回包含默认配置的 `Args`
func NewArgs() Args {
	return Args{Style, Heteronym, Separator}
}

// 获取单个拼音中的声母
func initial(p string) string {
	s := ""
	for _, v := range initials {
		if strings.HasPrefix(p, v) {
			s = v
			break
		}
	}
	return s
}

// 获取单个拼音中的韵母
func final(p string) string {
	i := initial(p)
	if i == "" {
		return p
	}
	return strings.Join(strings.SplitN(p, i, 2), "")
}

func toFixed(p string, a Args) string {
	if a.Style == Initials {
		return initial(p)
	}

	// 替换拼音中的带声调字符
	py := rePhoneticSymbol.ReplaceAllStringFunc(p, func(m string) string {
		symbol, _ := phoneticSymbol[m]
		switch a.Style {
		// 不包含声调
		case Normal, FirstLetter, Finals:
			// 去掉声调: a1 -> a
			m = reTone2.ReplaceAllString(symbol, "$1")
		case Tone2, FinalsTone2:
			// 返回使用数字标识声调的字符
			m = symbol
		default:
			// 	// 声调在头上
		}
		return m
	})

	switch a.Style {
	// 首字母
	case FirstLetter:
		py = string([]byte(py)[0])
	// 韵母
	case Finals, FinalsTone, FinalsTone2:
		py = final(py)
	}
	return py
}

func applyStyle(p []string, a Args) []string {
	newP := []string{}
	for _, v := range p {
		newP = append(newP, toFixed(v, a))
	}
	return newP
}

// SinglePinyin 把单个 `rune` 类型的汉字转换为拼音.
func SinglePinyin(r rune, a Args) []string {
	value, ok := PinyinDict[int(r)]
	pys := []string{}
	if ok {
		pys = strings.Split(value, ",")
		if !a.Heteronym {
			pys = strings.Split(value, ",")[:1]
		}
	}
	return applyStyle(pys, a)
}

// Pinyin 汉字转拼音，支持多音字模式.
func Pinyin(s string, a Args) [][]string {
	hans := []rune(s)
	pys := [][]string{}
	for _, r := range hans {
		pys = append(pys, SinglePinyin(r, a))
	}
	return pys
}

// LazyPinyin 汉字转拼音，与 `Pinyin` 的区别是：
// 返回值类型不同，并且不支持多音字模式，每个汉字只取第一个音.
func LazyPinyin(s string, a Args) []string {
	a.Heteronym = false
	pys := []string{}
	for _, v := range Pinyin(s, a) {
		pys = append(pys, v[0])
	}
	return pys
}

// Slug join `LazyPinyin` 的返回值.
func Slug(s string, a Args) string {
	separator := a.Separator
	return strings.Join(LazyPinyin(s, a), separator)
}
