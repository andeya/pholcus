// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

var (
	defaultTVRowBGColor Color = Color(win.GetSysColor(win.COLOR_WINDOW))
	white                     = win.COLORREF(RGB(255, 255, 255))
)

const (
	tableViewCurrentIndexChangedTimerId = 1 + iota
	tableViewSelectedIndexesChangedTimerId
)

// TableView is a model based widget for record centric, tabular data.
//
// TableView is implemented as a virtual mode list view to support quite large
// amounts of data.
type TableView struct {
	WidgetBase
	columns                            *TableViewColumnList
	model                              TableModel
	providedModel                      interface{}
	itemChecker                        ItemChecker
	imageProvider                      ImageProvider
	hIml                               win.HIMAGELIST
	usingSysIml                        bool
	imageUintptr2Index                 map[uintptr]int32
	filePath2IconIndex                 map[string]int32
	rowsResetHandlerHandle             int
	rowChangedHandlerHandle            int
	sortChangedHandlerHandle           int
	selectedIndexes                    []int
	prevIndex                          int
	currentIndex                       int
	currentIndexChangedPublisher       EventPublisher
	selectedIndexesChangedPublisher    EventPublisher
	itemActivatedPublisher             EventPublisher
	columnClickedPublisher             IntEventPublisher
	columnsOrderableChangedPublisher   EventPublisher
	columnsSizableChangedPublisher     EventPublisher
	publishNextSelClear                bool
	inSetSelectedIndexes               bool
	lastColumnStretched                bool
	inEraseBkgnd                       bool
	persistent                         bool
	itemStateChangedEventDelay         int
	alternatingRowBGColor              Color
	hasDarkAltBGColor                  bool
	delayedCurrentIndexChangedCanceled bool
	sortedColumnIndex                  int
	sortOrder                          SortOrder
}

// NewTableView creates and returns a *TableView as child of the specified
// Container.
func NewTableView(parent Container) (*TableView, error) {
	return NewTableViewWithStyle(parent, win.LVS_SHOWSELALWAYS)
}

// NewTableViewWithStyle creates and returns a *TableView as child of the specified
// Container and with the provided additional style bits set.
func NewTableViewWithStyle(parent Container, style uint32) (*TableView, error) {
	tv := &TableView{
		alternatingRowBGColor: defaultTVRowBGColor,
		imageUintptr2Index:    make(map[uintptr]int32),
		filePath2IconIndex:    make(map[string]int32),
	}

	tv.columns = newTableViewColumnList(tv)

	if err := InitWidget(
		tv,
		parent,
		"SysListView32",
		win.WS_TABSTOP|win.WS_VISIBLE|win.LVS_OWNERDATA|win.LVS_REPORT|style,
		win.WS_EX_CLIENTEDGE); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			tv.Dispose()
		}
	}()

	tv.SetPersistent(true)

	exStyle := tv.SendMessage(win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	exStyle |= win.LVS_EX_DOUBLEBUFFER | win.LVS_EX_FULLROWSELECT | win.LVS_EX_LABELTIP
	tv.SendMessage(win.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)

	if err := tv.setTheme("Explorer"); err != nil {
		return nil, err
	}

	tv.currentIndex = -1

	tv.MustRegisterProperty("ColumnsOrderable", NewBoolProperty(
		func() bool {
			return tv.ColumnsOrderable()
		},
		func(b bool) error {
			tv.SetColumnsOrderable(b)
			return nil
		},
		tv.columnsOrderableChangedPublisher.Event()))

	tv.MustRegisterProperty("ColumnsSizable", NewBoolProperty(
		func() bool {
			return tv.ColumnsSizable()
		},
		func(b bool) error {
			return tv.SetColumnsSizable(b)
		},
		tv.columnsSizableChangedPublisher.Event()))

	tv.MustRegisterProperty("HasCurrentItem", NewReadOnlyBoolProperty(
		func() bool {
			return tv.CurrentIndex() != -1
		},
		tv.CurrentIndexChanged()))

	succeeded = true

	return tv, nil
}

// Dispose releases the operating system resources, associated with the
// *TableView.
func (tv *TableView) Dispose() {
	tv.columns.unsetColumnsTV()

	tv.disposeImageListAndCaches()

	if tv.hWnd != 0 {
		if !win.KillTimer(tv.hWnd, tableViewCurrentIndexChangedTimerId) {
			lastError("KillTimer")
		}
		if !win.KillTimer(tv.hWnd, tableViewSelectedIndexesChangedTimerId) {
			lastError("KillTimer")
		}
	}

	tv.WidgetBase.Dispose()
}

// LayoutFlags returns a combination of LayoutFlags that specify how the
// *TableView wants to be treated by Layout implementations.
func (*TableView) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

// MinSizeHint returns the minimum outer Size, including decorations, that
// makes sense for the *TableView.
func (tv *TableView) MinSizeHint() Size {
	return Size{10, 10}
}

// SizeHint returns a sensible Size for a *TableView.
func (tv *TableView) SizeHint() Size {
	return Size{100, 100}
}

// ColumnsOrderable returns if the user can reorder columns by dragging and
// dropping column headers.
func (tv *TableView) ColumnsOrderable() bool {
	exStyle := tv.SendMessage(win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	return exStyle&win.LVS_EX_HEADERDRAGDROP > 0
}

// SetColumnsOrderable sets if the user can reorder columns by dragging and
// dropping column headers.
func (tv *TableView) SetColumnsOrderable(enabled bool) {
	exStyle := tv.SendMessage(win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	if enabled {
		exStyle |= win.LVS_EX_HEADERDRAGDROP
	} else {
		exStyle &^= win.LVS_EX_HEADERDRAGDROP
	}
	tv.SendMessage(win.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)

	tv.columnsOrderableChangedPublisher.Publish()
}

// ColumnsSizable returns if the user can change column widths by dragging
// dividers in the header.
func (tv *TableView) ColumnsSizable() bool {
	headerHWnd := win.HWND(tv.SendMessage(win.LVM_GETHEADER, 0, 0))

	style := win.GetWindowLong(headerHWnd, win.GWL_STYLE)

	return style&win.HDS_NOSIZING == 0
}

// SetColumnsSizable sets if the user can change column widths by dragging
// dividers in the header.
func (tv *TableView) SetColumnsSizable(b bool) error {
	headerHWnd := win.HWND(tv.SendMessage(win.LVM_GETHEADER, 0, 0))

	style := win.GetWindowLong(headerHWnd, win.GWL_STYLE)

	if b {
		style &^= win.HDS_NOSIZING
	} else {
		style |= win.HDS_NOSIZING
	}

	if 0 == win.SetWindowLong(headerHWnd, win.GWL_STYLE, style) {
		return lastError("SetWindowLong(GWL_STYLE)")
	}

	tv.columnsSizableChangedPublisher.Publish()

	return nil
}

// SortableByHeaderClick returns if the user can change sorting by clicking the header.
func (tv *TableView) SortableByHeaderClick() bool {
	return !tv.hasStyleBits(win.LVS_NOSORTHEADER)
}

// AlternatingRowBGColor returns the alternating row background color.
func (tv *TableView) AlternatingRowBGColor() Color {
	return tv.alternatingRowBGColor
}

// SetAlternatingRowBGColor sets the alternating row background color.
func (tv *TableView) SetAlternatingRowBGColor(c Color) {
	tv.alternatingRowBGColor = c

	tv.hasDarkAltBGColor = int(c.R())+int(c.G())+int(c.B()) < 128*3

	tv.Invalidate()
}

// Columns returns the list of columns.
func (tv *TableView) Columns() *TableViewColumnList {
	return tv.columns
}

// VisibleColumnsInDisplayOrder returns a slice of visible columns in display
// order.
func (tv *TableView) VisibleColumnsInDisplayOrder() []*TableViewColumn {
	visibleCols := tv.visibleColumns()
	indices := make([]int32, len(visibleCols))

	if win.FALSE == tv.SendMessage(win.LVM_GETCOLUMNORDERARRAY, uintptr(len(indices)), uintptr(unsafe.Pointer(&indices[0]))) {
		newError("LVM_GETCOLUMNORDERARRAY")
		return nil
	}

	orderedCols := make([]*TableViewColumn, len(visibleCols))

	for i, j := range indices {
		orderedCols[i] = visibleCols[j]
	}

	return orderedCols
}

// RowsPerPage returns the number of fully visible rows.
func (tv *TableView) RowsPerPage() int {
	return int(tv.SendMessage(win.LVM_GETCOUNTPERPAGE, 0, 0))
}

// UpdateItem ensures the item at index will be redrawn.
//
// If the model supports sorting, it will be resorted.
func (tv *TableView) UpdateItem(index int) error {
	if s, ok := tv.model.(Sorter); ok {
		if err := s.Sort(s.SortedColumn(), s.SortOrder()); err != nil {
			return err
		}

		return tv.Invalidate()
	} else {
		if win.FALSE == tv.SendMessage(win.LVM_UPDATE, uintptr(index), 0) {
			return newError("LVM_UPDATE")
		}
	}

	return nil
}

func (tv *TableView) attachModel() {
	tv.rowsResetHandlerHandle = tv.model.RowsReset().Attach(func() {
		tv.setItemCount()

		tv.SetCurrentIndex(-1)
	})

	tv.rowChangedHandlerHandle = tv.model.RowChanged().Attach(func(row int) {
		tv.UpdateItem(row)
	})

	if sorter, ok := tv.model.(Sorter); ok {
		tv.sortChangedHandlerHandle = sorter.SortChanged().Attach(func() {
			col := sorter.SortedColumn()
			tv.setSelectedColumnIndex(col)
			tv.setSortIcon(col, sorter.SortOrder())
			tv.Invalidate()
		})
	}
}

func (tv *TableView) detachModel() {
	tv.model.RowsReset().Detach(tv.rowsResetHandlerHandle)
	tv.model.RowChanged().Detach(tv.rowChangedHandlerHandle)
	if sorter, ok := tv.model.(Sorter); ok {
		sorter.SortChanged().Detach(tv.sortChangedHandlerHandle)
	}
}

// Model returns the model of the TableView.
func (tv *TableView) Model() interface{} {
	return tv.providedModel
}

// SetModel sets the model of the TableView.
//
// It is required that mdl either implements walk.TableModel,
// walk.ReflectTableModel or be a slice of pointers to struct or a
// []map[string]interface{}. A walk.TableModel implementation must also
// implement walk.Sorter to support sorting, all other options get sorting for
// free. To support item check boxes and icons, mdl must implement
// walk.ItemChecker and walk.ImageProvider, respectively. On-demand model
// population for a walk.ReflectTableModel or slice requires mdl to implement
// walk.Populator.
func (tv *TableView) SetModel(mdl interface{}) error {
	model, ok := mdl.(TableModel)
	if !ok && mdl != nil {
		var err error
		if model, err = newReflectTableModel(mdl); err != nil {
			if model, err = newMapTableModel(mdl); err != nil {
				return err
			}
		}
	}
	tv.providedModel = mdl

	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	if tv.model != nil {
		tv.detachModel()

		tv.disposeImageListAndCaches()
	}

	tv.model = model

	tv.itemChecker, _ = model.(ItemChecker)
	tv.imageProvider, _ = model.(ImageProvider)

	if model != nil {
		tv.attachModel()

		if dms, ok := model.(dataMembersSetter); ok {
			// FIXME: This depends on columns to be initialized before
			// calling this method.
			dataMembers := make([]string, len(tv.columns.items))

			for i, col := range tv.columns.items {
				dataMembers[i] = col.DataMemberEffective()
			}

			dms.setDataMembers(dataMembers)
		}

		if sorter, ok := tv.model.(Sorter); ok {
			sorter.Sort(tv.sortedColumnIndex, tv.sortOrder)
		}
	}

	tv.SetCurrentIndex(-1)

	return tv.setItemCount()
}

// TableModel returns the TableModel of the TableView.
func (tv *TableView) TableModel() TableModel {
	return tv.model
}

func (tv *TableView) setItemCount() error {
	var count int

	if tv.model != nil {
		count = tv.model.RowCount()
	}

	if 0 == tv.SendMessage(win.LVM_SETITEMCOUNT, uintptr(count), win.LVSICF_NOSCROLL) {
		return newError("SendMessage(LVM_SETITEMCOUNT)")
	}

	return nil
}

// CheckBoxes returns if the *TableView has check boxes.
func (tv *TableView) CheckBoxes() bool {
	return tv.SendMessage(win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)&win.LVS_EX_CHECKBOXES > 0
}

// SetCheckBoxes sets if the *TableView has check boxes.
func (tv *TableView) SetCheckBoxes(value bool) {
	exStyle := tv.SendMessage(win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	oldStyle := exStyle
	if value {
		exStyle |= win.LVS_EX_CHECKBOXES
	} else {
		exStyle &^= win.LVS_EX_CHECKBOXES
	}
	if exStyle != oldStyle {
		tv.SendMessage(win.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
	}

	mask := tv.SendMessage(win.LVM_GETCALLBACKMASK, 0, 0)

	if value {
		mask |= win.LVIS_STATEIMAGEMASK
	} else {
		mask &^= win.LVIS_STATEIMAGEMASK
	}

	if win.FALSE == tv.SendMessage(win.LVM_SETCALLBACKMASK, mask, 0) {
		newError("SendMessage(LVM_SETCALLBACKMASK)")
	}
}

func (tv *TableView) fromLVColIdx(index int32) int {
	var idx int32

	for i, tvc := range tv.columns.items {
		if tvc.visible {
			if idx == index {
				return i
			}

			idx++
		}
	}

	return -1
}

func (tv *TableView) toLVColIdx(index int) int32 {
	var idx int32

	for i, tvc := range tv.columns.items {
		if tvc.visible {
			if i == index {
				return idx
			}

			idx++
		}
	}

	return -1
}

func (tv *TableView) visibleColumnCount() int {
	var count int

	for _, tvc := range tv.columns.items {
		if tvc.visible {
			count++
		}
	}

	return count
}

func (tv *TableView) visibleColumns() []*TableViewColumn {
	var cols []*TableViewColumn

	for _, tvc := range tv.columns.items {
		if tvc.visible {
			cols = append(cols, tvc)
		}
	}

	return cols
}

/*func (tv *TableView) selectedColumnIndex() int {
	return tv.fromLVColIdx(tv.SendMessage(LVM_GETSELECTEDCOLUMN, 0, 0))
}*/

func (tv *TableView) setSelectedColumnIndex(index int) {
	tv.SendMessage(win.LVM_SETSELECTEDCOLUMN, uintptr(tv.toLVColIdx(index)), 0)
}

func (tv *TableView) setSortIcon(index int, order SortOrder) error {
	headerHwnd := win.HWND(tv.SendMessage(win.LVM_GETHEADER, 0, 0))

	idx := int(tv.toLVColIdx(index))

	for i := range tv.visibleColumns() {
		item := win.HDITEM{
			Mask: win.HDI_FORMAT,
		}

		iPtr := uintptr(i)
		itemPtr := uintptr(unsafe.Pointer(&item))

		if win.SendMessage(headerHwnd, win.HDM_GETITEM, iPtr, itemPtr) == 0 {
			return newError("SendMessage(HDM_GETITEM)")
		}

		if i == idx {
			switch order {
			case SortAscending:
				item.Fmt &^= win.HDF_SORTDOWN
				item.Fmt |= win.HDF_SORTUP

			case SortDescending:
				item.Fmt &^= win.HDF_SORTUP
				item.Fmt |= win.HDF_SORTDOWN
			}
		} else {
			item.Fmt &^= win.HDF_SORTDOWN | win.HDF_SORTUP
		}

		if win.SendMessage(headerHwnd, win.HDM_SETITEM, iPtr, itemPtr) == 0 {
			return newError("SendMessage(HDM_SETITEM)")
		}
	}

	return nil
}

// ColumnClicked returns the event that is published after a column header was
// clicked.
func (tv *TableView) ColumnClicked() *IntEvent {
	return tv.columnClickedPublisher.Event()
}

// ItemActivated returns the event that is published after an item was
// activated.
//
// An item is activated when it is double clicked or the enter key is pressed
// when the item is selected.
func (tv *TableView) ItemActivated() *Event {
	return tv.itemActivatedPublisher.Event()
}

// CurrentIndex returns the index of the current item, or -1 if there is no
// current item.
func (tv *TableView) CurrentIndex() int {
	return tv.currentIndex
}

// SetCurrentIndex sets the index of the current item.
//
// Call this with a value of -1 to have no current item.
func (tv *TableView) SetCurrentIndex(value int) error {
	var lvi win.LVITEM

	lvi.StateMask = win.LVIS_FOCUSED | win.LVIS_SELECTED
	if value > -1 {
		lvi.State = win.LVIS_FOCUSED | win.LVIS_SELECTED
	}

	if win.FALSE == tv.SendMessage(win.LVM_SETITEMSTATE, uintptr(value), uintptr(unsafe.Pointer(&lvi))) {
		return newError("SendMessage(LVM_SETITEMSTATE)")
	}

	if value != -1 {
		if win.FALSE == tv.SendMessage(win.LVM_ENSUREVISIBLE, uintptr(value), uintptr(0)) {
			return newError("SendMessage(LVM_ENSUREVISIBLE)")
		}
		// Windows bug? Sometimes a second LVM_ENSUREVISIBLE is required.
		if win.FALSE == tv.SendMessage(win.LVM_ENSUREVISIBLE, uintptr(value), uintptr(0)) {
			return newError("SendMessage(LVM_ENSUREVISIBLE)")
		}
	}

	tv.currentIndex = value

	if value == -1 {
		tv.currentIndexChangedPublisher.Publish()
	}

	return nil
}

// CurrentIndexChanged is the event that is published after CurrentIndex has
// changed.
func (tv *TableView) CurrentIndexChanged() *Event {
	return tv.currentIndexChangedPublisher.Event()
}

// MultiSelection returns whether multiple items can be selected at once.
//
// By default only a single item can be selected at once.
func (tv *TableView) MultiSelection() bool {
	style := uint(win.GetWindowLong(tv.hWnd, win.GWL_STYLE))
	if style == 0 {
		lastError("GetWindowLong")
		return false
	}

	return style&win.LVS_SINGLESEL == 0
}

// SetMultiSelection sets whether multiple items can be selected at once.
func (tv *TableView) SetMultiSelection(multiSel bool) error {
	return tv.ensureStyleBits(win.LVS_SINGLESEL, !multiSel)
}

// SelectedIndexes returns the indexes of the currently selected items.
func (tv *TableView) SelectedIndexes() []int {
	indexes := make([]int, len(tv.selectedIndexes))

	for i, j := range tv.selectedIndexes {
		indexes[i] = j
	}

	return indexes
}

// SetSelectedIndexes sets the indexes of the currently selected items.
func (tv *TableView) SetSelectedIndexes(indexes []int) error {
	tv.inSetSelectedIndexes = true
	defer func() {
		tv.inSetSelectedIndexes = false
		tv.publishSelectedIndexesChanged()
	}()

	lvi := &win.LVITEM{StateMask: win.LVIS_FOCUSED | win.LVIS_SELECTED}
	lp := uintptr(unsafe.Pointer(lvi))

	if win.FALSE == tv.SendMessage(win.LVM_SETITEMSTATE, ^uintptr(0), lp) {
		return newError("SendMessage(LVM_SETITEMSTATE)")
	}

	lvi.State = win.LVIS_FOCUSED | win.LVIS_SELECTED
	for _, i := range indexes {
		if win.FALSE == tv.SendMessage(win.LVM_SETITEMSTATE, uintptr(i), lp) {
			return newError("SendMessage(LVM_SETITEMSTATE)")
		}
	}

	idxs := make([]int, len(indexes))

	for i, j := range indexes {
		idxs[i] = j
	}

	tv.selectedIndexes = idxs

	return nil
}

func (tv *TableView) updateSelectedIndexes() {
	count := int(tv.SendMessage(win.LVM_GETSELECTEDCOUNT, 0, 0))
	indexes := make([]int, count)

	j := -1
	for i := 0; i < count; i++ {
		j = int(tv.SendMessage(win.LVM_GETNEXTITEM, uintptr(j), win.LVNI_SELECTED))
		indexes[i] = j
	}

	changed := len(indexes) != len(tv.selectedIndexes)
	if !changed {
		for i := 0; i < len(indexes); i++ {
			if indexes[i] != tv.selectedIndexes[i] {
				changed = true
				break
			}
		}
	}

	if changed {
		tv.selectedIndexes = indexes
		tv.publishSelectedIndexesChanged()
	}
}

// ItemStateChangedEventDelay returns the delay in milliseconds, between the
// moment the state of an item in the *TableView changes and the moment the
// associated event is published.
//
// By default there is no delay.
func (tv *TableView) ItemStateChangedEventDelay() int {
	return tv.itemStateChangedEventDelay
}

// SetItemStateChangedEventDelay sets the delay in milliseconds, between the
// moment the state of an item in the *TableView changes and the moment the
// associated event is published.
//
// An example where this may be useful is a master-details scenario. If the
// master TableView is configured to delay the event, you can avoid pointless
// updates of the details TableView, if the user uses arrow keys to rapidly
// navigate the master view.
func (tv *TableView) SetItemStateChangedEventDelay(delay int) {
	tv.itemStateChangedEventDelay = delay
}

// SelectedIndexesChanged returns the event that is published when the list of
// selected item indexes changed.
func (tv *TableView) SelectedIndexesChanged() *Event {
	return tv.selectedIndexesChangedPublisher.Event()
}

func (tv *TableView) publishSelectedIndexesChanged() {
	if tv.itemStateChangedEventDelay > 0 {
		if 0 == win.SetTimer(
			tv.hWnd,
			tableViewSelectedIndexesChangedTimerId,
			uint32(tv.itemStateChangedEventDelay),
			0) {

			lastError("SetTimer")
		}
	} else {
		tv.selectedIndexesChangedPublisher.Publish()
	}
}

// LastColumnStretched returns if the last column should take up all remaining
// horizontal space of the *TableView.
func (tv *TableView) LastColumnStretched() bool {
	return tv.lastColumnStretched
}

// SetLastColumnStretched sets if the last column should take up all remaining
// horizontal space of the *TableView.
//
// The effect of setting this is persistent.
func (tv *TableView) SetLastColumnStretched(value bool) error {
	if value {
		if err := tv.StretchLastColumn(); err != nil {
			return err
		}
	}

	tv.lastColumnStretched = value

	return nil
}

// StretchLastColumn makes the last column take up all remaining horizontal
// space of the *TableView.
//
// The effect of this is not persistent.
func (tv *TableView) StretchLastColumn() error {
	colCount := tv.visibleColumnCount()
	if colCount == 0 {
		return nil
	}

	if 0 == tv.SendMessage(win.LVM_SETCOLUMNWIDTH, uintptr(colCount-1), win.LVSCW_AUTOSIZE_USEHEADER) {
		return newError("LVM_SETCOLUMNWIDTH failed")
	}

	return nil
}

// Persistent returns if the *TableView should persist its UI state, like column
// widths. See *App.Settings for details.
func (tv *TableView) Persistent() bool {
	return tv.persistent
}

// SetPersistent sets if the *TableView should persist its UI state, like column
// widths. See *App.Settings for details.
func (tv *TableView) SetPersistent(value bool) {
	tv.persistent = value
}

type tableViewState struct {
	SortColumnName     string
	SortOrder          SortOrder
	ColumnDisplayOrder []string // Also indicates visibility
	Columns            []tableViewColumnState
}

type tableViewColumnState struct {
	Name  string
	Title string
	Width int
}

// SaveState writes the UI state of the *TableView to the settings.
func (tv *TableView) SaveState() error {
	var tvs tableViewState

	tvs.SortColumnName = tv.columns.items[tv.sortedColumnIndex].name
	tvs.SortOrder = tv.sortOrder

	tvs.Columns = make([]tableViewColumnState, tv.columns.Len())

	for i, tvc := range tv.columns.items {
		tvcs := &tvs.Columns[i]

		tvcs.Name = tvc.name
		tvcs.Title = tvc.titleOverride
		tvcs.Width = tvc.Width()
	}

	visibleCols := tv.visibleColumns()
	indices := make([]int32, len(visibleCols))
	var lParam uintptr
	if len(visibleCols) > 0 {
		lParam = uintptr(unsafe.Pointer(&indices[0]))
	}
	if 0 == tv.SendMessage(win.LVM_GETCOLUMNORDERARRAY, uintptr(len(visibleCols)), lParam) {
		return newError("LVM_GETCOLUMNORDERARRAY")
	}

	tvs.ColumnDisplayOrder = make([]string, len(visibleCols))
	for i, j := range indices {
		tvs.ColumnDisplayOrder[i] = visibleCols[j].name
	}

	state, err := json.Marshal(tvs)
	if err != nil {
		return err
	}

	return tv.putState(string(state))
}

// RestoreState restores the UI state of the *TableView from the settings.
func (tv *TableView) RestoreState() error {
	state, err := tv.getState()
	if err != nil {
		return err
	}
	if state == "" {
		return nil
	}

	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	var tvs tableViewState

	if err := json.Unmarshal(([]byte)(state), &tvs); err != nil {
		return tv.restoreStateOldStyle(state)
	}

	name2tvc := make(map[string]*TableViewColumn)

	for _, tvc := range tv.columns.items {
		name2tvc[tvc.name] = tvc
	}

	name2tvcs := make(map[string]*tableViewColumnState)

	for i, tvcs := range tvs.Columns {
		name2tvcs[tvcs.Name] = &tvs.Columns[i]

		if tvc := name2tvc[tvcs.Name]; tvc != nil {
			if err := tvc.SetTitleOverride(tvcs.Title); err != nil {
				return err
			}
			if err := tvc.SetWidth(tvcs.Width); err != nil {
				return err
			}
			var visible bool
			for _, name := range tvs.ColumnDisplayOrder {
				if name == tvc.name {
					visible = true
					break
				}
			}
			if err := tvc.SetVisible(visible); err != nil {
				return err
			}
		}
	}

	visibleCount := tv.visibleColumnCount()

	indices := make([]int32, visibleCount)

	knownNames := make(map[string]struct{})

	displayOrder := make([]string, 0, visibleCount)
	for _, name := range tvs.ColumnDisplayOrder {
		knownNames[name] = struct{}{}
		if _, ok := name2tvc[name]; ok {
			displayOrder = append(displayOrder, name)
		}
	}
	for _, tvc := range tv.visibleColumns() {
		if _, ok := knownNames[tvc.name]; !ok {
			displayOrder = append(displayOrder, tvc.name)
		}
	}

	for i, tvc := range tv.visibleColumns() {
		for j, name := range displayOrder {
			if tvc.name == name {
				indices[j] = int32(i)
				break
			}
		}
	}

	wParam := uintptr(len(indices))
	var lParam uintptr
	if len(indices) > 0 {
		lParam = uintptr(unsafe.Pointer(&indices[0]))
	}
	if 0 == tv.SendMessage(win.LVM_SETCOLUMNORDERARRAY, wParam, lParam) {
		return newError("LVM_SETCOLUMNORDERARRAY")
	}

	for i, c := range tvs.Columns {
		if c.Name == tvs.SortColumnName {
			tv.sortedColumnIndex = i
			tv.sortOrder = tvs.SortOrder
			break
		}
	}

	if sorter, ok := tv.model.(Sorter); ok {
		if !sorter.ColumnSortable(tv.sortedColumnIndex) {
			for i := range tvs.Columns {
				if sorter.ColumnSortable(i) {
					tv.sortedColumnIndex = i
				}
			}
		}

		sorter.Sort(tv.sortedColumnIndex, tvs.SortOrder)
	}

	//if sorter, ok := tv.model.(Sorter); ok {
	//	index := -1
	//	order := SortAscending

	//	for i, c := range tvs.Columns {
	//		if tv.columns.items[i].visible && sorter.ColumnSortable(i) {
	//			if c.Name == tvs.SortColumnName {
	//				index = i
	//				order = tvs.SortOrder
	//				break
	//			} else if index == -1 {
	//				index = i
	//			}
	//		}
	//	}

	//	if index > -1 {
	//		sorter.Sort(index, order)
	//	}
	//}

	return nil
}

//// SaveState writes the UI state of the *TableView to the settings.
//func (tv *TableView) SaveState() error {
//	buf := new(bytes.Buffer)

//	count := tv.columns.Len()
//	if count > 0 {
//		for i := 0; i < count; i++ {
//			if i > 0 {
//				buf.WriteString(" ")
//			}

//			width := tv.Columns().At(i).Width()
//			if width == 0 {
//				width = 100
//			}

//			buf.WriteString(strconv.Itoa(int(width)))
//		}

//		buf.WriteString(";")

//		visibleCount := tv.visibleColumnCount()

//		indices := make([]int32, visibleCount)
//		lParam := uintptr(unsafe.Pointer(&indices[0]))

//		if 0 == tv.SendMessage(win.LVM_GETCOLUMNORDERARRAY, uintptr(visibleCount), lParam) {
//			return newError("LVM_GETCOLUMNORDERARRAY")
//		}

//		for i, idx := range indices {
//			if i > 0 {
//				buf.WriteString(" ")
//			}

//			buf.WriteString(strconv.Itoa(int(idx)))
//		}

//		buf.WriteString(";")

//		for i, tvc := range tv.columns.items {
//			if i > 0 {
//				buf.WriteString("|")
//			}

//			buf.WriteString(tvc.TitleOverride())
//		}

//		buf.WriteString(";")

//		for i, tvc := range tv.columns.items {
//			if i > 0 {
//				buf.WriteString(" ")
//			}

//			if tvc.Visible() {
//				buf.WriteString("1")
//			} else {
//				buf.WriteString("0")
//			}
//		}

//		buf.WriteString(";")

//		if sorter, ok := tv.model.(Sorter); ok {
//			buf.WriteString(strconv.Itoa(sorter.SortedColumn()))
//			buf.WriteString(" ")
//			buf.WriteString(strconv.Itoa(int(sorter.SortOrder())))
//		} else {
//			buf.WriteString("- -")
//		}
//	}

//	return tv.putState(buf.String())
//}

func (tv *TableView) restoreStateOldStyle(state string) error {
	parts := strings.Split(state, ";")

	widthStrs := strings.Split(parts[0], " ")

	// FIXME: Solve this in a better way.
	if len(widthStrs) > tv.columns.Len() {
		log.Print("*TableView.RestoreState: failed due to unexpected column count (FIXME!)")
		return nil
	}

	// Do visibility stuff first.
	if len(parts) > 3 {
		visible := strings.Split(parts[3], " ")

		for i, v := range visible {
			if err := tv.columns.At(i).SetVisible(v == "1"); err != nil {
				return err
			}
		}
	}

	for i, str := range widthStrs {
		width, err := strconv.Atoi(str)
		if err != nil {
			return err
		}

		if err := tv.Columns().At(i).SetWidth(width); err != nil {
			return err
		}
	}

	if len(parts) > 1 {
		indexStrs := strings.Split(parts[1], " ")

		indices := make([]int32, len(indexStrs))

		var failed bool
		for i, s := range indexStrs {
			idx, err := strconv.Atoi(s)
			if err != nil {
				failed = true
				break
			}
			indices[i] = int32(idx)
		}

		if !failed {
			wParam := uintptr(len(indices))
			lParam := uintptr(unsafe.Pointer(&indices[0]))
			if 0 == tv.SendMessage(win.LVM_SETCOLUMNORDERARRAY, wParam, lParam) {
				return newError("LVM_SETCOLUMNORDERARRAY")
			}
		}
	}

	if len(parts) > 2 {
		titleOverrides := strings.Split(parts[2], "|")

		for i, to := range titleOverrides {
			if err := tv.columns.At(i).SetTitleOverride(to); err != nil {
				return err
			}
		}
	}

	if sorter, ok := tv.model.(Sorter); ok && len(parts) > 4 {
		sortParts := strings.Split(parts[4], " ")
		if colStr := sortParts[0]; colStr != "-" {
			col, err := strconv.Atoi(colStr)
			if err != nil {
				return err
			}
			if sorter.ColumnSortable(col) {
				ord, err := strconv.Atoi(sortParts[1])
				if err != nil {
					return err
				}
				if err := sorter.Sort(col, SortOrder(ord)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (tv *TableView) toggleItemChecked(index int) error {
	checked := tv.itemChecker.Checked(index)

	if err := tv.itemChecker.SetChecked(index, !checked); err != nil {
		return wrapError(err)
	}

	if win.FALSE == tv.SendMessage(win.LVM_UPDATE, uintptr(index), 0) {
		return newError("SendMessage(LVM_UPDATE)")
	}

	return nil
}

func (tv *TableView) applyImageListForImage(image interface{}) {
	tv.hIml, tv.usingSysIml, _ = imageListForImage(image)

	tv.SendMessage(win.LVM_SETIMAGELIST, win.LVSIL_SMALL, uintptr(tv.hIml))

	tv.imageUintptr2Index = make(map[uintptr]int32)
	tv.filePath2IconIndex = make(map[string]int32)
}

func (tv *TableView) disposeImageListAndCaches() {
	if tv.hIml != 0 && !tv.usingSysIml {
		tv.SendMessage(win.LVM_SETIMAGELIST, win.LVSIL_SMALL, 0)

		win.ImageList_Destroy(tv.hIml)
	}
	tv.hIml = 0

	tv.imageUintptr2Index = nil
	tv.filePath2IconIndex = nil
}

func (tv *TableView) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_ERASEBKGND:
		if tv.lastColumnStretched && !tv.inEraseBkgnd {
			tv.inEraseBkgnd = true
			defer func() {
				tv.inEraseBkgnd = false
			}()
			tv.StretchLastColumn()
		}
		return 1

	case win.WM_GETDLGCODE:
		if wParam == win.VK_RETURN {
			return win.DLGC_WANTALLKEYS
		}

	case win.WM_LBUTTONDOWN, win.WM_RBUTTONDOWN, win.WM_LBUTTONDBLCLK, win.WM_RBUTTONDBLCLK:
		var hti win.LVHITTESTINFO
		hti.Pt = win.POINT{win.GET_X_LPARAM(lParam), win.GET_Y_LPARAM(lParam)}
		tv.SendMessage(win.LVM_HITTEST, 0, uintptr(unsafe.Pointer(&hti)))

		if hti.Flags == win.LVHT_NOWHERE {
			if tv.MultiSelection() {
				tv.publishNextSelClear = true
			} else {
				// We keep the current item, if in single item selection mode.
				tv.SetFocus()
				return 0
			}
		}

		switch msg {
		case win.WM_LBUTTONDOWN, win.WM_RBUTTONDOWN:
			if hti.Flags == win.LVHT_ONITEMSTATEICON &&
				tv.itemChecker != nil &&
				tv.CheckBoxes() {

				tv.toggleItemChecked(int(hti.IItem))
			}

		case win.WM_LBUTTONDBLCLK, win.WM_RBUTTONDBLCLK:
			if tv.currentIndex != tv.prevIndex && tv.itemStateChangedEventDelay > 0 {
				tv.prevIndex = tv.currentIndex
				tv.currentIndexChangedPublisher.Publish()
			}
		}

	case win.WM_KEYDOWN:
		if wParam == win.VK_SPACE &&
			tv.currentIndex > -1 &&
			tv.itemChecker != nil &&
			tv.CheckBoxes() {

			tv.toggleItemChecked(tv.currentIndex)
		}

	case win.WM_NOTIFY:
		switch ((*win.NMHDR)(unsafe.Pointer(lParam))).Code {
		case win.LVN_GETDISPINFO:
			di := (*win.NMLVDISPINFO)(unsafe.Pointer(lParam))

			row := int(di.Item.IItem)
			col := tv.fromLVColIdx(di.Item.ISubItem)

			if di.Item.Mask&win.LVIF_TEXT > 0 {
				var text string
				switch val := tv.model.Value(row, col).(type) {
				case string:
					text = val

				case float32:
					prec := tv.columns.items[col].precision
					if prec == 0 {
						prec = 2
					}
					text = FormatFloatGrouped(float64(val), prec)

				case float64:
					prec := tv.columns.items[col].precision
					if prec == 0 {
						prec = 2
					}
					text = FormatFloatGrouped(val, prec)

				case time.Time:
					if val.Year() > 1601 {
						text = val.Format(tv.columns.items[col].format)
					}

				case *big.Rat:
					prec := tv.columns.items[col].precision
					if prec == 0 {
						prec = 2
					}
					text = formatBigRatGrouped(val, prec)

				default:
					text = fmt.Sprintf(tv.columns.items[col].format, val)
				}

				utf16 := syscall.StringToUTF16(text)
				buf := (*[264]uint16)(unsafe.Pointer(di.Item.PszText))
				max := mini(len(utf16), int(di.Item.CchTextMax))
				copy((*buf)[:], utf16[:max])
				(*buf)[max-1] = 0
			}

			if tv.imageProvider != nil && di.Item.Mask&win.LVIF_IMAGE > 0 {
				if image := tv.imageProvider.Image(row); image != nil {
					if tv.hIml == 0 {
						tv.applyImageListForImage(image)
					}

					di.Item.IImage = imageIndexMaybeAdd(
						image,
						tv.hIml,
						tv.usingSysIml,
						tv.imageUintptr2Index,
						tv.filePath2IconIndex)
				}
			}

			if di.Item.StateMask&win.LVIS_STATEIMAGEMASK > 0 &&
				tv.itemChecker != nil {
				checked := tv.itemChecker.Checked(row)

				if checked {
					di.Item.State = 0x2000
				} else {
					di.Item.State = 0x1000
				}
			}

		case win.NM_CUSTOMDRAW:
			if tv.alternatingRowBGColor != defaultTVRowBGColor {
				nmlvcd := (*win.NMLVCUSTOMDRAW)(unsafe.Pointer(lParam))

				switch nmlvcd.Nmcd.DwDrawStage {
				case win.CDDS_PREPAINT:
					return win.CDRF_NOTIFYITEMDRAW

				case win.CDDS_ITEMPREPAINT:
					if nmlvcd.Nmcd.DwItemSpec%2 == 1 {
						/*if tv.hasDarkAltBGColor &&
							nmlvcd.Nmcd.UItemState&win.CDIS_HOT == 0 &&
							tv.SendMessage(win.LVM_GETITEMSTATE, nmlvcd.Nmcd.DwItemSpec, win.LVIS_SELECTED) == 0 &&
							int32(tv.SendMessage(win.LVM_GETSELECTEDCOLUMN, 0, 0)) != nmlvcd.ISubItem {
							fmt.Printf("selcol: %d, subitem: %d\n", int32(tv.SendMessage(win.LVM_GETSELECTEDCOLUMN, 0, 0)), nmlvcd.ISubItem)
							nmlvcd.ClrText = white
						}*/
						nmlvcd.ClrTextBk = win.COLORREF(tv.alternatingRowBGColor)
					}

					return win.CDRF_NOTIFYSUBITEMDRAW

				case win.CDDS_ITEMPREPAINT | win.CDDS_SUBITEM:
					if nmlvcd.Nmcd.DwItemSpec%2 == 1 &&
						tv.hasDarkAltBGColor &&
						nmlvcd.Nmcd.UItemState&win.CDIS_HOT == 0 &&
						tv.SendMessage(win.LVM_GETITEMSTATE, nmlvcd.Nmcd.DwItemSpec, win.LVIS_SELECTED) == 0 &&
						int32(tv.SendMessage(win.LVM_GETSELECTEDCOLUMN, 0, 0)) != nmlvcd.ISubItem {

						nmlvcd.ClrText = white
					}

					return win.CDRF_NEWFONT
				}
			}

			return win.CDRF_DODEFAULT

		case win.LVN_COLUMNCLICK:
			nmlv := (*win.NMLISTVIEW)(unsafe.Pointer(lParam))

			col := tv.fromLVColIdx(nmlv.ISubItem)

			if sorter, ok := tv.model.(Sorter); ok && sorter.ColumnSortable(col) {
				prevCol := sorter.SortedColumn()
				var order SortOrder
				if col != prevCol || sorter.SortOrder() == SortDescending {
					order = SortAscending
				} else {
					order = SortDescending
				}
				tv.sortedColumnIndex = col
				tv.sortOrder = order
				sorter.Sort(col, order)
			}

			tv.columnClickedPublisher.Publish(col)

		case win.LVN_ITEMCHANGED:
			nmlv := (*win.NMLISTVIEW)(unsafe.Pointer(lParam))
			if nmlv.IItem == -1 && !tv.publishNextSelClear {
				break
			}
			tv.publishNextSelClear = false

			selectedNow := nmlv.UNewState&win.LVIS_SELECTED > 0
			selectedBefore := nmlv.UOldState&win.LVIS_SELECTED > 0
			if selectedNow && !selectedBefore {
				tv.prevIndex = tv.currentIndex
				tv.currentIndex = int(nmlv.IItem)
				if tv.itemStateChangedEventDelay > 0 {
					tv.delayedCurrentIndexChangedCanceled = false
					if 0 == win.SetTimer(
						tv.hWnd,
						tableViewCurrentIndexChangedTimerId,
						uint32(tv.itemStateChangedEventDelay),
						0) {

						lastError("SetTimer")
					}
				} else {
					tv.currentIndexChangedPublisher.Publish()
				}
			}

			if selectedNow != selectedBefore {
				if !tv.inSetSelectedIndexes && tv.MultiSelection() {
					tv.updateSelectedIndexes()
				}
			}

		case win.LVN_ODSTATECHANGED:
			tv.updateSelectedIndexes()

		case win.LVN_ITEMACTIVATE:
			nmia := (*win.NMITEMACTIVATE)(unsafe.Pointer(lParam))

			if tv.itemStateChangedEventDelay > 0 {
				tv.delayedCurrentIndexChangedCanceled = true
			}

			if int(nmia.IItem) != tv.currentIndex {
				tv.SetCurrentIndex(int(nmia.IItem))
				tv.currentIndexChangedPublisher.Publish()
			}

			tv.itemActivatedPublisher.Publish()
		}

	case win.WM_TIMER:
		switch wParam {
		case tableViewCurrentIndexChangedTimerId:
			if !tv.delayedCurrentIndexChangedCanceled {
				tv.currentIndexChangedPublisher.Publish()
			}

		case tableViewSelectedIndexesChangedTimerId:
			tv.selectedIndexesChangedPublisher.Publish()
		}
	}

	return tv.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
