package pinyin

import (
	"sort"
)

// 按首字母排序
func SortInitials(strs []string) {
	a := NewArgs()
	l := len(strs)
	initials := make([]string, l)
	newStrs := map[string]string{}

	for i := 0; i < l; i++ {
		for ii, py := range Pinyin(strs[i], a) {
			if len(py) == 0 {
				initials[i] += string([]rune(strs[i])[ii])
			} else {
				initials[i] += py[0]
			}
		}
		newStrs[initials[i]] = strs[i]
	}

	sort.Strings(initials)

	for i := 0; i < l; i++ {
		strs[i] = newStrs[initials[i]]
	}
}
