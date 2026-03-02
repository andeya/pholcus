package pinyin

import (
	"regexp"
	"strings"

	"github.com/andeya/gust/option"
)

// Package metadata.
const (
	Version   = "0.2.1"
	Author    = "mozillazg, 闲耘"
	License   = "MIT"
	Copyright = "Copyright (c) 2014 mozillazg, 闲耘"
)

// Pinyin style constants (recommended).
const (
	Normal      = 0 // Plain style, no tone marks. e.g. pin yin
	Tone        = 1 // Tone on first vowel of final. e.g. pīn yīn
	Tone2       = 2 // Tone as digit [0-4] after syllable. e.g. pi1n yi1n
	Initials    = 3 // Initial consonants only. e.g. "zh g"
	FirstLetter = 4 // First letter of each syllable only. e.g. p y
	Finals      = 5 // Finals only, no tone. e.g. ong uo
	FinalsTone  = 6 // Finals with tone on first vowel. e.g. ōng uó
	FinalsTone2 = 7 // Finals with tone as digit [0-4]. e.g. o1ng uo2
)

// Pinyin style constants (backward compatible aliases).
const (
	NORMAL       = 0
	TONE         = 1
	TONE2        = 2
	INITIALS     = 3
	FIRST_LETTER = 4
	FINALS       = 5
	FINALS_TONE  = 6
	FINALS_TONE2 = 7
)

// Initial consonants table.
var initials = strings.Split(
	"b,p,m,f,d,t,n,l,g,k,h,j,q,x,r,zh,ch,sh,z,c,s",
	",",
)

var rePhoneticSymbolSource = func(m map[string]string) string {
	s := ""
	for k := range m {
		s = s + k
	}
	return s
}(phoneticSymbol)

var rePhoneticSymbol = regexp.MustCompile("[" + rePhoneticSymbolSource + "]")
var reTone2 = regexp.MustCompile("([aeoiuvnm])([0-4])$")

// Args holds pinyin conversion options.
type Args struct {
	Style     int    // Pinyin style (default: Normal)
	Heteronym bool   // Enable heteronym mode for multi-reading characters (default: false)
	Separator string // Separator used in Slug (default: "-")
}

// Style is the default pinyin style.
var Style = Normal

// Heteronym enables multi-reading character mode when true.
var Heteronym = false

// Separator is the default join separator for Slug.
var Separator = "-"

// NewArgs returns Args with default configuration.
func NewArgs() Args {
	return Args{Style, Heteronym, Separator}
}

// initial extracts the initial consonant from a single pinyin syllable.
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

// final extracts the final (rhyme) from a single pinyin syllable.
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

	py := rePhoneticSymbol.ReplaceAllStringFunc(p, func(m string) string {
		symbol, _ := phoneticSymbol[m]
		switch a.Style {
		case Normal, FirstLetter, Finals:
			m = reTone2.ReplaceAllString(symbol, "$1")
		case Tone2, FinalsTone2:
			m = symbol
		default:
			// Tone on vowel (Tone, FinalsTone)
		}
		return m
	})

	switch a.Style {
	case FirstLetter:
		py = string([]byte(py)[0])
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

// SinglePinyin converts a single Chinese character (rune) to pinyin.
func SinglePinyin(r rune, a Args) []string {
	value, ok := PinyinDict[int(r)]
	opt := option.BoolOpt(value, ok)
	pys := []string{}
	if opt.IsSome() {
		value := opt.Unwrap()
		pys = strings.Split(value, ",")
		if !a.Heteronym {
			pys = pys[:1]
		}
	}
	return applyStyle(pys, a)
}

// Pinyin converts Chinese characters to pinyin, with heteronym support.
func Pinyin(s string, a Args) [][]string {
	hans := []rune(s)
	pys := [][]string{}
	for _, r := range hans {
		pys = append(pys, SinglePinyin(r, a))
	}
	return pys
}

// LazyPinyin converts Chinese to pinyin, returning a flat slice.
// Unlike Pinyin, it does not support heteronyms and uses only the first reading per character.
func LazyPinyin(s string, a Args) []string {
	a.Heteronym = false
	pys := []string{}
	for _, v := range Pinyin(s, a) {
		pys = append(pys, v[0])
	}
	return pys
}

// Slug joins LazyPinyin results with the configured separator.
func Slug(s string, a Args) string {
	separator := a.Separator
	return strings.Join(LazyPinyin(s, a), separator)
}
