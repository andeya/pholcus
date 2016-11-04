// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type TableViewColumnList struct {
	tv    *TableView
	items []*TableViewColumn
}

func newTableViewColumnList(tv *TableView) *TableViewColumnList {
	return &TableViewColumnList{tv: tv}
}

// Add adds a TableViewColumn to the end of the list.
func (l *TableViewColumnList) Add(item *TableViewColumn) error {
	return l.Insert(len(l.items), item)
}

// At returns the TableViewColumn as the specified index.
//
// Bounds are not checked.
func (l *TableViewColumnList) At(index int) *TableViewColumn {
	return l.items[index]
}

func (l *TableViewColumnList) atInListView(index int) *TableViewColumn {
	var idx int

	for _, item := range l.items {
		if item.visible {
			if idx == index {
				return item
			}

			idx++
		}
	}

	return nil
}

// Clear removes all TableViewColumns from the list.
func (l *TableViewColumnList) Clear() error {
	for _ = range l.items {
		if err := l.RemoveAt(0); err != nil {
			return err
		}
	}

	return nil
}

// Index returns the index of the specified TableViewColumn or -1 if it is not
// found.
func (l *TableViewColumnList) Index(item *TableViewColumn) int {
	for i, lvi := range l.items {
		if lvi == item {
			return i
		}
	}

	return -1
}

// Contains returns whether the specified TableViewColumn is found in the list.
func (l *TableViewColumnList) Contains(item *TableViewColumn) bool {
	return l.Index(item) > -1
}

// Insert inserts TableViewColumn item at position index.
//
// A TableViewColumn cannot be contained in multiple TableViewColumnLists at the
// same time.
func (l *TableViewColumnList) Insert(index int, item *TableViewColumn) error {
	if item.tv != nil {
		return newError("duplicate insert")
	}

	item.tv = l.tv

	if item.visible {
		if err := item.create(); err != nil {
			item.tv = nil
			return err
		}
	}

	l.items = append(l.items, nil)
	copy(l.items[index+1:], l.items[index:])
	l.items[index] = item

	return nil
}

// Len returns the number of TableViewColumns in  the list.
func (l *TableViewColumnList) Len() int {
	return len(l.items)
}

// Remove removes the specified TableViewColumn from the list.
func (l *TableViewColumnList) Remove(item *TableViewColumn) error {
	index := l.Index(item)
	if index == -1 {
		return nil
	}

	return l.RemoveAt(index)
}

// RemoveAt removes the TableViewColumn at position index.
func (l *TableViewColumnList) RemoveAt(index int) error {
	tvc := l.items[index]

	if err := tvc.destroy(); err != nil {
		return err
	}

	tvc.tv = nil

	l.items = append(l.items[:index], l.items[index+1:]...)

	return nil
}

func (l *TableViewColumnList) unsetColumnsTV() {
	for _, tvc := range l.items {
		tvc.tv = nil
	}
}
