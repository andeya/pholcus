// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type Foo struct {
	Index   int
	Bar     string
	Baz     float64
	Quux    time.Time
	checked bool
}

type FooModel struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn int
	sortOrder  walk.SortOrder
	evenBitmap *walk.Bitmap
	oddIcon    *walk.Icon
	items      []*Foo
}

func NewFooModel() *FooModel {
	m := new(FooModel)
	m.evenBitmap, _ = walk.NewBitmapFromFile("../img/open.png")
	m.oddIcon, _ = walk.NewIconFromFile("../img/x.ico")
	m.ResetRows()
	return m
}

// Called by the TableView from SetModel and every time the model publishes a
// RowsReset event.
func (m *FooModel) RowCount() int {
	return len(m.items)
}

// Called by the TableView when it needs the text to display for a given cell.
func (m *FooModel) Value(row, col int) interface{} {
	item := m.items[row]

	switch col {
	case 0:
		return item.Index

	case 1:
		return item.Bar

	case 2:
		return item.Baz

	case 3:
		return item.Quux
	}

	panic("unexpected col")
}

// Called by the TableView to retrieve if a given row is checked.
func (m *FooModel) Checked(row int) bool {
	return m.items[row].checked
}

// Called by the TableView when the user toggled the check box of a given row.
func (m *FooModel) SetChecked(row int, checked bool) error {
	m.items[row].checked = checked

	return nil
}

// Called by the TableView to sort the model.
func (m *FooModel) Sort(col int, order walk.SortOrder) error {
	m.sortColumn, m.sortOrder = col, order

	sort.Stable(m)

	return m.SorterBase.Sort(col, order)
}

func (m *FooModel) Len() int {
	return len(m.items)
}

func (m *FooModel) Less(i, j int) bool {
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
		return c(a.Bar < b.Bar)

	case 2:
		return c(a.Baz < b.Baz)

	case 3:
		return c(a.Quux.Before(b.Quux))
	}

	panic("unreachable")
}

func (m *FooModel) Swap(i, j int) {
	m.items[i], m.items[j] = m.items[j], m.items[i]
}

// Called by the TableView to retrieve an item image.
func (m *FooModel) Image(row int) interface{} {
	if m.items[row].Index%2 == 0 {
		return m.evenBitmap
	}

	return m.oddIcon
}

func (m *FooModel) ResetRows() {
	// Create some random data.
	m.items = make([]*Foo, rand.Intn(50000))

	now := time.Now()

	for i := range m.items {
		m.items[i] = &Foo{
			Index: i,
			Bar:   strings.Repeat("*", rand.Intn(5)+1),
			Baz:   rand.Float64() * 1000,
			Quux:  time.Unix(rand.Int63n(now.Unix()), 0),
		}
	}

	// Notify TableView and other interested parties about the reset.
	m.PublishRowsReset()

	m.Sort(m.sortColumn, m.sortOrder)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	model := NewFooModel()

	var tv *walk.TableView

	MainWindow{
		Title:  "Walk TableView Example",
		Size:   Size{800, 600},
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			PushButton{
				Text:      "Reset Rows",
				OnClicked: model.ResetRows,
			},
			PushButton{
				Text: "Select first 5 even Rows",
				OnClicked: func() {
					tv.SetSelectedIndexes([]int{0, 2, 4, 6, 8})
				},
			},
			TableView{
				AssignTo:              &tv,
				AlternatingRowBGColor: walk.RGB(255, 255, 224),
				CheckBoxes:            true,
				ColumnsOrderable:      true,
				MultiSelection:        true,
				Columns: []TableViewColumn{
					{Title: "#"},
					{Title: "Bar"},
					{Title: "Baz", Format: "%.2f", Alignment: AlignFar},
					{Title: "Quux", Format: "2006-01-02 15:04:05", Width: 150},
				},
				Model: model,
				OnSelectedIndexesChanged: func() {
					fmt.Printf("SelectedIndexes: %v\n", tv.SelectedIndexes())
				},
			},
		},
	}.Run()
}
