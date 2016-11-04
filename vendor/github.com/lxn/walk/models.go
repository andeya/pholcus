// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

// BindingValueProvider is the interface that a model must implement to support
// data binding with widgets like ComboBox.
type BindingValueProvider interface {
	BindingValue(index int) interface{}
}

// ListModel is the interface that a model must implement to support widgets
// like ComboBox.
type ListModel interface {
	// ItemCount returns the number of items in the model.
	ItemCount() int

	// Value returns the value that should be displayed for the given index.
	Value(index int) interface{}

	// ItemsReset returns the event that the model should publish when the
	// number of its items changes.
	ItemsReset() *Event

	// ItemChanged returns the event that the model should publish when an item
	// was changed.
	ItemChanged() *IntEvent
}

// ListModelBase implements the ItemsReset and ItemChanged methods of the
// ListModel interface.
type ListModelBase struct {
	itemsResetPublisher  EventPublisher
	itemChangedPublisher IntEventPublisher
}

func (lmb *ListModelBase) ItemsReset() *Event {
	return lmb.itemsResetPublisher.Event()
}

func (lmb *ListModelBase) ItemChanged() *IntEvent {
	return lmb.itemChangedPublisher.Event()
}

func (lmb *ListModelBase) PublishItemsReset() {
	lmb.itemsResetPublisher.Publish()
}

func (lmb *ListModelBase) PublishItemChanged(index int) {
	lmb.itemChangedPublisher.Publish(index)
}

// ReflectListModel provides an alternative to the ListModel interface. It
// uses reflection to obtain data.
type ReflectListModel interface {
	// Items returns the model data, which must be a slice of pointer to struct.
	Items() interface{}

	// ItemsReset returns the event that the model should publish when the
	// number of its items changes.
	ItemsReset() *Event

	// ItemChanged returns the event that the model should publish when an item
	// was changed.
	ItemChanged() *IntEvent

	setValueFunc(value func(index int) interface{})
}

// ReflectListModelBase implements the ItemsReset and ItemChanged methods of
// the ReflectListModel interface.
type ReflectListModelBase struct {
	ListModelBase
	value func(index int) interface{}
}

func (rlmb *ReflectListModelBase) setValueFunc(value func(index int) interface{}) {
	rlmb.value = value
}

func (rlmb *ReflectListModelBase) Value(index int) interface{} {
	return rlmb.value(index)
}

// TableModel is the interface that a model must implement to support widgets
// like TableView.
type TableModel interface {
	// RowCount returns the number of rows in the model.
	RowCount() int

	// Value returns the value that should be displayed for the given cell.
	Value(row, col int) interface{}

	// RowsReset returns the event that the model should publish when the number
	// of its rows changes.
	RowsReset() *Event

	// RowChanged returns the event that the model should publish when a row was
	// changed.
	RowChanged() *IntEvent
}

// TableModelBase implements the RowsReset and RowChanged methods of the
// TableModel interface.
type TableModelBase struct {
	rowsResetPublisher  EventPublisher
	rowChangedPublisher IntEventPublisher
}

func (tmb *TableModelBase) RowsReset() *Event {
	return tmb.rowsResetPublisher.Event()
}

func (tmb *TableModelBase) RowChanged() *IntEvent {
	return tmb.rowChangedPublisher.Event()
}

func (tmb *TableModelBase) PublishRowsReset() {
	tmb.rowsResetPublisher.Publish()
}

func (tmb *TableModelBase) PublishRowChanged(row int) {
	tmb.rowChangedPublisher.Publish(row)
}

// ReflectTableModel provides an alternative to the TableModel interface. It
// uses reflection to obtain data.
type ReflectTableModel interface {
	// Items returns the model data, which must be a slice of pointer to struct.
	Items() interface{}

	// RowsReset returns the event that the model should publish when the
	// number of its items changes.
	RowsReset() *Event

	// RowChanged returns the event that the model should publish when an item
	// was changed.
	RowChanged() *IntEvent

	setValueFunc(value func(row, col int) interface{})
}

// ReflectTableModelBase implements the ItemsReset and ItemChanged methods of
// the ReflectTableModel interface.
type ReflectTableModelBase struct {
	TableModelBase
	value func(row, col int) interface{}
}

func (rtmb *ReflectTableModelBase) setValueFunc(value func(row, col int) interface{}) {
	rtmb.value = value
}

func (rtmb *ReflectTableModelBase) Value(row, col int) interface{} {
	return rtmb.value(row, col)
}

type interceptedSorter interface {
	sorterBase() *SorterBase
	setSortFunc(sort func(col int, order SortOrder) error)
}

// SortedReflectTableModelBase implements the RowsReset and RowChanged methods
// of the ReflectTableModel interface as well as the Sorter interface for
// pre-implemented in-memory sorting.
type SortedReflectTableModelBase struct {
	ReflectTableModelBase
	SorterBase
	sort func(col int, order SortOrder) error
}

func (srtmb *SortedReflectTableModelBase) setSortFunc(sort func(col int, order SortOrder) error) {
	srtmb.sort = sort
}

func (srtmb *SortedReflectTableModelBase) sorterBase() *SorterBase {
	return &srtmb.SorterBase
}

func (srtmb *SortedReflectTableModelBase) Sort(col int, order SortOrder) error {
	if srtmb.sort != nil {
		return srtmb.sort(col, order)
	}

	return srtmb.SorterBase.Sort(col, order)
}

// Populator is an interface that can be implemented by Reflect*Models and slice
// types to populate themselves on demand.
//
// Widgets like TableView, ListBox and ComboBox support lazy population of a
// Reflect*Model or slice, if it implements this interface.
type Populator interface {
	// Populate initializes the slot specified by index.
	//
	// For best performance it is probably a good idea to populate more than a
	// single slot of the slice at once.
	Populate(index int) error
}

// ImageProvider is the interface that a model must implement to support
// displaying an item image.
type ImageProvider interface {
	// Image returns the image to display for the item at index index.
	//
	// Supported types are *walk.Bitmap, *walk.Icon and string. A string will be
	// interpreted as a file path and the icon associated with the file will be
	// used. It is not supported to use strings together with the other options
	// in the same model instance.
	Image(index int) interface{}
}

// ItemChecker is the interface that a model must implement to support check
// boxes in a widget like TableView.
type ItemChecker interface {
	// Checked returns if the specified item is checked.
	Checked(index int) bool

	// SetChecked sets if the specified item is checked.
	SetChecked(index int, checked bool) error
}

// SortOrder specifies the order by which items are sorted.
type SortOrder int

const (
	// SortAscending specifies ascending sort order.
	SortAscending SortOrder = iota

	// SortDescending specifies descending sort order.
	SortDescending
)

// Sorter is the interface that a model must implement to support sorting with a
// widget like TableView.
type Sorter interface {
	// ColumnSortable returns whether column col is sortable.
	ColumnSortable(col int) bool

	// Sort sorts column col in order order.
	//
	// If col is -1 then no column is to be sorted. Sort must publish the event
	// returned from SortChanged() after sorting.
	Sort(col int, order SortOrder) error

	// SortChanged returns an event that is published after sorting.
	SortChanged() *Event

	// SortedColumn returns the index of the currently sorted column, or -1 if
	// no column is currently sorted.
	SortedColumn() int

	// SortOrder returns the current sort order.
	SortOrder() SortOrder
}

// SorterBase implements the Sorter interface.
//
// You still need to provide your own implementation of at least the Sort method
// to actually sort and reset the model. Your Sort method should call the
// SorterBase implementation so the SortChanged event, that e.g. a TableView
// widget depends on, is published.
type SorterBase struct {
	changedPublisher EventPublisher
	col              int
	order            SortOrder
}

func (sb *SorterBase) ColumnSortable(col int) bool {
	return true
}

func (sb *SorterBase) Sort(col int, order SortOrder) error {
	sb.col, sb.order = col, order

	sb.changedPublisher.Publish()

	return nil
}

func (sb *SorterBase) SortChanged() *Event {
	return sb.changedPublisher.Event()
}

func (sb *SorterBase) SortedColumn() int {
	return sb.col
}

func (sb *SorterBase) SortOrder() SortOrder {
	return sb.order
}

// Imager provides access to an image of objects like tree items.
type Imager interface {
	// Image returns the image to display for an item.
	//
	// Supported types are *walk.Bitmap, *walk.Icon and string. A string will be
	// interpreted as a file path and the icon associated with the file will be
	// used. It is not supported to use strings together with the other options
	// in the same model instance.
	Image() interface{}
}

// TreeItem represents an item in a TreeView widget.
type TreeItem interface {
	// Text returns the text of the item.
	Text() string

	// Parent returns the parent of the item.
	Parent() TreeItem

	// ChildCount returns the number of children of the item.
	ChildCount() int

	// ChildAt returns the child at the specified index.
	ChildAt(index int) TreeItem
}

// TreeModel provides widgets like TreeView with item data.
type TreeModel interface {
	// LazyPopulation returns if the model prefers on-demand population.
	//
	// This is useful for models that potentially contain huge amounts of items,
	// e.g. a model that represents a file system.
	LazyPopulation() bool

	// RootCount returns the number of root items.
	RootCount() int

	// RootAt returns the root item at the specified index.
	RootAt(index int) TreeItem

	// ItemsReset returns the event that the model should publish when the
	// descendants of the specified item, or all items if no item is specified,
	// are reset.
	ItemsReset() *TreeItemEvent

	// ItemChanged returns the event that the model should publish when an item
	// was changed.
	ItemChanged() *TreeItemEvent
}

// TreeModelBase partially implements the TreeModel interface.
//
// You still need to provide your own implementation of at least the
// RootCount and RootAt methods. If your model needs lazy population,
// you will also have to implement LazyPopulation.
type TreeModelBase struct {
	itemsResetPublisher  TreeItemEventPublisher
	itemChangedPublisher TreeItemEventPublisher
}

func (tmb *TreeModelBase) LazyPopulation() bool {
	return false
}

func (tmb *TreeModelBase) ItemsReset() *TreeItemEvent {
	return tmb.itemsResetPublisher.Event()
}

func (tmb *TreeModelBase) ItemChanged() *TreeItemEvent {
	return tmb.itemChangedPublisher.Event()
}

func (tmb *TreeModelBase) PublishItemsReset(parent TreeItem) {
	tmb.itemsResetPublisher.Publish(parent)
}

func (tmb *TreeModelBase) PublishItemChanged(item TreeItem) {
	tmb.itemChangedPublisher.Publish(item)
}
