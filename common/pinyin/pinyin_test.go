package pinyin

import (
	"reflect"
	"testing"
)

func TestSortInitials(t *testing.T) {
	strs := []string{"北京", "上海", "杭州", "广州"}
	SortInitials(strs)
	expected := []string{"北京", "广州", "杭州", "上海"}
	if !reflect.DeepEqual(strs, expected) {
		t.Errorf("SortInitials got %v, want %v", strs, expected)
	}
}

func TestFinalEmptyInitial(t *testing.T) {
	a := NewArgs()
	a.Style = Finals
	result := Pinyin("鹅恩", a)
	expected := [][]string{{"e"}, {"en"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Pinyin(Finals) got %v, want %v", result, expected)
	}
}
