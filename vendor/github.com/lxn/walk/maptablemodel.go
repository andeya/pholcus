// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"sort"
)

type mapTableModel struct {
	TableModelBase
	SorterBase
	dataMembers []string
	dataSource  interface{}
	items       []map[string]interface{}
}

func newMapTableModel(dataSource interface{}) (TableModel, error) {
	items, ok := dataSource.([]map[string]interface{})
	if !ok {
		return nil, newError("dataSource must be assignable to []map[string]interface{}")
	}

	return &mapTableModel{dataSource: dataSource, items: items}, nil
}

func (m *mapTableModel) setDataMembers(dataMembers []string) {
	m.dataMembers = dataMembers
}

func (m *mapTableModel) RowCount() int {
	return len(m.items)
}

func (m *mapTableModel) Value(row, col int) interface{} {
	if m.items[row] == nil {
		if populator, ok := m.dataSource.(Populator); ok {
			if err := populator.Populate(row); err != nil {
				return err
			}
		}

		if m.items[row] == nil {
			return nil
		}
	}

	return m.items[row][m.dataMembers[col]]
}

func (m *mapTableModel) Sort(col int, order SortOrder) error {
	m.col, m.order = col, order

	sort.Stable(m)

	m.changedPublisher.Publish()

	return nil
}

func (m *mapTableModel) Len() int {
	return m.RowCount()
}

func (m *mapTableModel) Less(i, j int) bool {
	col := m.SortedColumn()

	return less(m.Value(i, col), m.Value(j, col), m.SortOrder())
}

func (m *mapTableModel) Swap(i, j int) {
	m.items[i], m.items[j] = m.items[j], m.items[i]
}
