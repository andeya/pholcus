package goutil

import (
	"strings"
)

// SnakeString converts the accepted string to a snake string (XxYy to xx_yy)
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	for _, d := range StringToBytes(s) {
		if d >= 'A' && d <= 'Z' {
			if j {
				data = append(data, '_')
				j = false
			}
		} else if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(BytesToString(data))
}

// CamelString converts the accepted string to a camel string (xx_yy to XxYy)
func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

var spaceReplacer = strings.NewReplacer(
	"  ", " ",
	"\n\n", "\n",
	"\r\r", "\r",
	"\t\t", "\t",
	"\r\n\r\n", "\r\n",
	" \n", "\n",
	"\t\n", "\n",
	" \t", "\t",
	"\t ", "\t",
	"\v\v", "\v",
	"\f\f", "\f",
	string(0x85)+string(0x85),
	string(0x85),
	string(0xA0)+string(0xA0),
	string(0xA0),
)

// SpaceInOne combines multiple consecutive space characters into one.
func SpaceInOne(s string) string {
	var old string
	for old != s {
		old = s
		s = spaceReplacer.Replace(s)
	}
	return s
}
