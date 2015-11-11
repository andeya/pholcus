package pinyin_test

import (
	"github.com/henrylee2cn/pholcus/common/pinyin"
	"testing"
)

func TestExamplePinyin_default(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	t.Log("default:", pinyin.Pinyin(hans, a))
	// Output: default: [[zhong] [guo] [ren]]
}

func TestExamplePinyin_normal(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	t.Log("Normal:", pinyin.Pinyin(hans, a))
	// Output: Normal: [[zhong] [guo] [ren]]
}

func TestExamplePinyin_tone(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	a.Style = pinyin.Tone
	t.Log("Tone:", pinyin.Pinyin(hans, a))
	// Output: Tone: [[zhōng] [guó] [rén]]
}

func TestExamplePinyin_tone2(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	a.Style = pinyin.Tone2
	t.Log("Tone2:", pinyin.Pinyin(hans, a))
	// Output: Tone2: [[zho1ng] [guo2] [re2n]]
}

func TestExamplePinyin_initials(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	a.Style = pinyin.Initials
	t.Log("Initials:", pinyin.Pinyin(hans, a))
	// Output: Initials: [[zh] [g] [r]]
}

func TestExamplePinyin_firstLetter(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	a.Style = pinyin.FirstLetter
	t.Log(pinyin.Pinyin(hans, a))
	// Output: [[z] [g] [r]]
}

func TestExamplePinyin_finals(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	a.Style = pinyin.Finals
	t.Log(pinyin.Pinyin(hans, a))
	// Output: [[ong] [uo] [en]]
}

func TestExamplePinyin_finalsTone(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	a.Style = pinyin.FinalsTone
	t.Log(pinyin.Pinyin(hans, a))
	// Output: [[ōng] [uó] [én]]
}

func TestExamplePinyin_finalsTone2(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	a.Style = pinyin.FinalsTone2
	t.Log(pinyin.Pinyin(hans, a))
	// Output: [[o1ng] [uo2] [e2n]]
}

func TestExamplePinyin_heteronym(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	a.Heteronym = true
	a.Style = pinyin.Tone2
	t.Log(pinyin.Pinyin(hans, a))
	// Output: [[zho1ng zho4ng] [guo2] [re2n]]
}

func TestExampleLazyPinyin(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	t.Log(pinyin.LazyPinyin(hans, a))
	// Output: [zhong guo ren]
}

func TestExampleSlug(t *testing.T) {
	hans := "中国人"
	a := pinyin.NewArgs()
	t.Log(pinyin.Slug(hans, a))
	// Output: zhong-guo-ren
}
