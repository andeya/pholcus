// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"reflect"
	"sort"
)

type reflectModel interface {
	Items() interface{}
}

type bindingAndDisplayMemberSetter interface {
	setBindingMember(member string)
	setDisplayMember(member string)
}

type reflectListModel struct {
	ListModelBase
	bindingMember string
	displayMember string
	dataSource    interface{}
	items         interface{}
	value         reflect.Value
}

func newReflectListModel(dataSource interface{}) (ListModel, error) {
	items, err := itemsFromReflectModelDataSource(dataSource, "ReflectListModel")
	if err != nil {
		return nil, err
	}

	m := &reflectListModel{
		dataSource: dataSource,
		items:      items,
		value:      reflect.ValueOf(items),
	}

	if rlm, ok := dataSource.(ReflectListModel); ok {
		rlm.setValueFunc(func(index int) interface{} {
			return m.Value(index)
		})

		rlm.ItemChanged().Attach(func(index int) {
			m.PublishItemChanged(index)
		})

		rlm.ItemsReset().Attach(func() {
			m.items = rlm.Items()
			m.value = reflect.ValueOf(m.items)

			m.PublishItemsReset()
		})
	}

	return m, nil
}

func (m *reflectListModel) setBindingMember(member string) {
	m.bindingMember = member
}

func (m *reflectListModel) setDisplayMember(member string) {
	m.displayMember = member
}

func (m *reflectListModel) ItemCount() int {
	return m.value.Len()
}

func (m *reflectListModel) BindingValue(index int) interface{} {
	return valueFromSlice(m.dataSource, m.value, m.bindingMember, index)
}

func (m *reflectListModel) Value(index int) interface{} {
	return valueFromSlice(m.dataSource, m.value, m.displayMember, index)
}

type dataMembersSetter interface {
	setDataMembers(dataMembers []string)
}

type reflectTableModel struct {
	TableModelBase
	sorterBase  *SorterBase
	dataMembers []string
	dataSource  interface{}
	items       interface{}
	value       reflect.Value
}

func newReflectTableModel(dataSource interface{}) (TableModel, error) {
	items, err := itemsFromReflectModelDataSource(dataSource, "ReflectTableModel")
	if err != nil {
		return nil, err
	}

	m := &reflectTableModel{
		dataSource: dataSource,
		items:      items,
		value:      reflect.ValueOf(items),
	}

	if rtm, ok := dataSource.(ReflectTableModel); ok {
		rtm.setValueFunc(func(row, col int) interface{} {
			return m.Value(row, col)
		})

		rtm.RowChanged().Attach(func(index int) {
			m.PublishRowChanged(index)
		})

		rtm.RowsReset().Attach(func() {
			m.items = rtm.Items()
			m.value = reflect.ValueOf(m.items)

			m.PublishRowsReset()

			if is, ok := dataSource.(interceptedSorter); ok {
				sb := is.sorterBase()
				m.sort(sb.SortedColumn(), sb.SortOrder())
			}
		})
	} else {
		m.sorterBase = new(SorterBase)
	}

	if is, ok := dataSource.(interceptedSorter); ok {
		m.sorterBase = is.sorterBase()
		is.setSortFunc(func(col int, order SortOrder) error {
			return m.sort(col, order)
		})
	}

	_, isImageProvider := dataSource.(ImageProvider)
	_, isSortable := dataSource.(Sorter)
	if !isSortable {
		isSortable = m.sorterBase != nil
	}
	if isImageProvider {
		if isSortable {
			return &sortedImageReflectTableModel{reflectTableModel: m}, nil
		} else {
			return &imageReflectTableModel{reflectTableModel: m}, nil
		}
	} else if isSortable {
		return &sortedReflectTableModel{reflectTableModel: m}, nil
	}

	return m, nil
}

func (m *reflectTableModel) setDataMembers(dataMembers []string) {
	m.dataMembers = dataMembers
}

func (m *reflectTableModel) RowCount() int {
	return m.value.Len()
}

func (m *reflectTableModel) Value(row, col int) interface{} {
	return valueFromSlice(m.dataSource, m.value, m.dataMembers[col], row)
}

func (m *reflectTableModel) Checked(row int) bool {
	if m.value.Index(row).IsNil() {
		return false
	}

	if checker, ok := m.dataSource.(ItemChecker); ok {
		return checker.Checked(row)
	}

	return false
}

func (m *reflectTableModel) SetChecked(row int, checked bool) error {
	if m.value.Index(row).IsNil() {
		return nil
	}

	if checker, ok := m.dataSource.(ItemChecker); ok {
		return checker.SetChecked(row, checked)
	}

	return nil
}

func (m *reflectTableModel) ColumnSortable(col int) bool {
	if sorter, ok := m.dataSource.(Sorter); ok {
		return sorter.ColumnSortable(col)
	}

	return true
}

func (m *reflectTableModel) SortChanged() *Event {
	if sorter, ok := m.dataSource.(Sorter); ok {
		return sorter.SortChanged()
	}

	if m.sorterBase != nil {
		return m.sorterBase.SortChanged()
	}

	return nil
}

func (m *reflectTableModel) SortedColumn() int {
	if sorter, ok := m.dataSource.(Sorter); ok {
		return sorter.SortedColumn()
	}

	if m.sorterBase != nil {
		return m.sorterBase.SortedColumn()
	}

	return -1
}

func (m *reflectTableModel) SortOrder() SortOrder {
	if sorter, ok := m.dataSource.(Sorter); ok {
		return sorter.SortOrder()
	}

	if m.sorterBase != nil {
		return m.sorterBase.SortOrder()
	}

	return SortAscending
}

func (m *reflectTableModel) sort(col int, order SortOrder) error {
	if sb := m.sorterBase; sb != nil {
		sb.col, sb.order = col, order

		sort.Stable(m)

		sb.changedPublisher.Publish()

		return nil
	}

	if sorter, ok := m.dataSource.(Sorter); ok {
		return sorter.Sort(col, order)
	}

	return nil
}

func (m *reflectTableModel) Len() int {
	return m.RowCount()
}

func (m *reflectTableModel) Less(i, j int) bool {
	col := m.SortedColumn()

	return less(m.Value(i, col), m.Value(j, col), m.SortOrder())
}

func (m *reflectTableModel) Swap(i, j int) {
	vi := m.value.Index(i)
	vj := m.value.Index(j)

	viv := vi.Interface()
	vjv := vj.Interface()

	vi.Set(reflect.ValueOf(vjv))
	vj.Set(reflect.ValueOf(viv))
}

type imageReflectTableModel struct {
	*reflectTableModel
}

func (m *imageReflectTableModel) Image(index int) interface{} {
	if m.value.Index(index).IsNil() {
		return nil
	}

	return m.dataSource.(ImageProvider).Image(index)
}

type sortedReflectTableModel struct {
	*reflectTableModel
}

func (m *sortedReflectTableModel) Sort(col int, order SortOrder) error {
	return m.reflectTableModel.sort(col, order)
}

type sortedImageReflectTableModel struct {
	*reflectTableModel
}

func (m *sortedImageReflectTableModel) Sort(col int, order SortOrder) error {
	return m.reflectTableModel.sort(col, order)
}

func (m *sortedImageReflectTableModel) Image(index int) interface{} {
	if m.value.Index(index).IsNil() {
		return nil
	}

	return m.dataSource.(ImageProvider).Image(index)
}

func itemsFromReflectModelDataSource(dataSource interface{}, requiredInterfaceName string) (interface{}, error) {
	var items interface{}
	if rm, ok := dataSource.(reflectModel); ok {
		items = rm.Items()
	} else {
		items = dataSource
	}

	if requiredInterfaceName == "ReflectListModel" {
		if _, ok := dataSource.([]string); ok {
			return items, nil
		}
	}

	if t := reflect.TypeOf(items); t == nil ||
		t.Kind() != reflect.Slice ||
		t.Elem().Kind() != reflect.Ptr ||
		t.Elem().Elem().Kind() != reflect.Struct {

		return nil, newError(fmt.Sprintf("dataSource must be a slice of pointers to struct or must implement %s.", requiredInterfaceName))
	}

	return items, nil
}

func valueFromSlice(dataSource interface{}, itemsValue reflect.Value, member string, index int) interface{} {
	if member == "" {
		if strs, ok := dataSource.([]string); ok {
			return strs[index]
		}

		return ""
	}

	v := itemsValue.Index(index)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		if populator, ok := dataSource.(Populator); ok {
			if err := populator.Populate(index); err != nil {
				return err
			}
		}

		if v.IsNil() {
			return nil
		}
	}

	vv, err := reflectValueFromPath(v, member)
	if err != nil {
		return err
	}

	return vv.Interface()
}
