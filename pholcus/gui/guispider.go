// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gui

import (
	"github.com/henrylee2cn/pholcus/spiders/spider"
	"github.com/lxn/walk"
	"sort"
	// . "github.com/lxn/walk/declarative"
)

type GUISpiderCore struct {
	Spider      *spider.Spider
	Description string
}

type GUISpider struct {
	*GUISpiderCore
	Index   int
	Title   string
	checked bool
}

type GUISpiderModel struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn int
	sortOrder  walk.SortOrder
	// evenBitmap *walk.Bitmap
	// oddIcon    *walk.Icon
	items []*GUISpider
}

func NewGUISpiderModel(list []*GUISpiderCore) *GUISpiderModel {
	m := new(GUISpiderModel)
	// m.evenBitmap, _ = walk.NewBitmapFromFile("")
	// m.oddIcon, _ = walk.NewIconFromFile("img/x.ico")
	for i, t := range list {
		m.items = append(m.items, &GUISpider{
			Index: i + 1,
			Title: t.Spider.GetName(),
			GUISpiderCore: &GUISpiderCore{
				Description: t.Description,
				Spider:      t.Spider,
			},
		})
	}

	return m
}

// Called by the TableView from SetModel and every time the model publishes a
// RowsReset event.
func (m *GUISpiderModel) RowCount() int {
	return len(m.items)
}

// Called by the TableView when it needs the text to display for a given cell.
func (m *GUISpiderModel) Value(row, col int) interface{} {
	item := m.items[row]

	switch col {
	case 0:
		return item.Index

	case 1:
		return item.Title

	case 2:
		return item.Description

	case 3:
		return item.Spider
	}
	panic("unexpected col")
}

// Called by the TableView to retrieve if a given row is checked.
func (m *GUISpiderModel) Checked(row int) bool {
	return m.items[row].checked
}

// Called by the TableView when the user toggled the check box of a given row.
func (m *GUISpiderModel) SetChecked(row int, checked bool) error {
	m.items[row].checked = checked

	return nil
}

//获取被选中的结果
func (m *GUISpiderModel) GetChecked() []*GUISpider {
	rc := []*GUISpider{}
	for idx, item := range m.items {
		if m.Checked(idx) {
			rc = append(rc, item)
		}
	}
	return rc
}

// Called by the TableView to sort the model.
func (m *GUISpiderModel) Sort(col int, order walk.SortOrder) error {
	m.sortColumn, m.sortOrder = col, order

	sort.Sort(m)

	return m.SorterBase.Sort(col, order)
}

func (m *GUISpiderModel) Len() int {
	return len(m.items)
}

func (m *GUISpiderModel) Less(i, j int) bool {
	a, b := m.items[i], m.items[j]

	c := func(ls bool) bool {
		if m.sortOrder == walk.SortAscending {
			return ls
		}

		return !ls
	}

	switch m.sortColumn {
	case 0:
		return c(a.Index < b.Index)

	case 1:
		return c(a.Title < b.Title)

	case 2:
		return c(a.Description < b.Description)

		// case 3:
		// 	return c(a.Spider < b.Spider)
	}

	panic("unreachable")
}

func (m *GUISpiderModel) Swap(i, j int) {
	m.items[i], m.items[j] = m.items[j], m.items[i]
}

// Called by the TableView to retrieve an item image.
// func (m *GUISpiderModel) Image(row int) interface{} {
// 	// if m.items[row].Index%2 == 0 {
// 	// 	return m.oddIcon
// 	// }
// 	return m.evenBitmap
// }
