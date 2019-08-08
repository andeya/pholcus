// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
)

const tableViewWindowClass = `\o/ Walk_TableView_Class \o/`

func init() {
	MustRegisterWindowClass(tableViewWindowClass)
}

var (
	white                       = win.COLORREF(RGB(255, 255, 255))
	checkmark                   = string([]byte{0xE2, 0x9C, 0x94})
	tableViewFrozenLVWndProcPtr = syscall.NewCallback(tableViewFrozenLVWndProc)
	tableViewNormalLVWndProcPtr = syscall.NewCallback(tableViewNormalLVWndProc)
	tableViewHdrWndProcPtr      = syscall.NewCallback(tableViewHdrWndProc)
)

const (
	tableViewCurrentIndexChangedTimerId = 1 + iota
	tableViewSelectedIndexesChangedTimerId
)

type TableViewCfg struct {
	Style              uint32
	CustomHeaderHeight int
	CustomRowHeight    int
}

// TableView is a model based widget for record centric, tabular data.
//
// TableView is implemented as a virtual mode list view to support quite large
// amounts of data.
type TableView struct {
	WidgetBase
	hwndFrozenLV                       win.HWND
	hwndFrozenHdr                      win.HWND
	frozenLVOrigWndProcPtr             uintptr
	frozenHdrOrigWndProcPtr            uintptr
	hwndNormalLV                       win.HWND
	hwndNormalHdr                      win.HWND
	normalLVOrigWndProcPtr             uintptr
	normalHdrOrigWndProcPtr            uintptr
	state                              *tableViewState
	columns                            *TableViewColumnList
	model                              TableModel
	providedModel                      interface{}
	itemChecker                        ItemChecker
	imageProvider                      ImageProvider
	styler                             CellStyler
	style                              CellStyle
	itemFont                           *Font
	hIml                               win.HIMAGELIST
	usingSysIml                        bool
	imageUintptr2Index                 map[uintptr]int32
	filePath2IconIndex                 map[string]int32
	rowsResetHandlerHandle             int
	rowChangedHandlerHandle            int
	rowsInsertedHandlerHandle          int
	rowsRemovedHandlerHandle           int
	sortChangedHandlerHandle           int
	selectedIndexes                    []int
	prevIndex                          int
	currentIndex                       int
	itemIndexOfLastMouseButtonDown     int
	hwndItemChanged                    win.HWND
	currentIndexChangedPublisher       EventPublisher
	selectedIndexesChangedPublisher    EventPublisher
	itemActivatedPublisher             EventPublisher
	columnClickedPublisher             IntEventPublisher
	columnsOrderableChangedPublisher   EventPublisher
	columnsSizableChangedPublisher     EventPublisher
	itemCountChangedPublisher          EventPublisher
	publishNextSelClear                bool
	inSetSelectedIndexes               bool
	lastColumnStretched                bool
	persistent                         bool
	itemStateChangedEventDelay         int
	themeNormalBGColor                 Color
	themeNormalTextColor               Color
	themeSelectedBGColor               Color
	themeSelectedTextColor             Color
	themeSelectedNotFocusedBGColor     Color
	itemBGColor                        Color
	itemTextColor                      Color
	alternatingRowBGColor              Color
	hasDarkAltBGColor                  bool
	delayedCurrentIndexChangedCanceled bool
	sortedColumnIndex                  int
	sortOrder                          SortOrder
	formActivatingHandle               int
	customHeaderHeight                 int
	customRowHeight                    int
	scrolling                          bool
	inSetCurrentIndex                  bool
	inMouseEvent                       bool
	hasFrozenColumn                    bool
	inEraseBkgnd                       bool
	focused                            bool
	ignoreNowhere                      bool
	updateLVSizesNeedsSpecialCare      bool
}

// NewTableView creates and returns a *TableView as child of the specified
// Container.
func NewTableView(parent Container) (*TableView, error) {
	return NewTableViewWithStyle(parent, win.LVS_SHOWSELALWAYS)
}

// NewTableViewWithStyle creates and returns a *TableView as child of the specified
// Container and with the provided additional style bits set.
func NewTableViewWithStyle(parent Container, style uint32) (*TableView, error) {
	return NewTableViewWithCfg(parent, &TableViewCfg{Style: style})
}

// NewTableViewWithCfg creates and returns a *TableView as child of the specified
// Container and with the provided additional configuration.
func NewTableViewWithCfg(parent Container, cfg *TableViewCfg) (*TableView, error) {
	tv := &TableView{
		imageUintptr2Index:   make(map[uintptr]int32),
		filePath2IconIndex:   make(map[string]int32),
		formActivatingHandle: -1,
		customHeaderHeight:   cfg.CustomHeaderHeight,
		customRowHeight:      cfg.CustomRowHeight,
	}

	tv.columns = newTableViewColumnList(tv)

	if err := InitWidget(
		tv,
		parent,
		tableViewWindowClass,
		win.WS_BORDER|win.WS_VISIBLE,
		win.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			tv.Dispose()
		}
	}()

	var rowHeightStyle uint32
	if cfg.CustomRowHeight > 0 {
		rowHeightStyle = win.LVS_OWNERDRAWFIXED
	}

	if tv.hwndFrozenLV = win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr("SysListView32"),
		nil,
		win.WS_CHILD|win.WS_CLIPSIBLINGS|win.WS_TABSTOP|win.WS_VISIBLE|win.LVS_OWNERDATA|win.LVS_REPORT|cfg.Style|rowHeightStyle,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		tv.hWnd,
		0,
		0,
		nil,
	); tv.hwndFrozenLV == 0 {
		return nil, newError("creating frozen lv failed")
	}

	tv.frozenLVOrigWndProcPtr = win.SetWindowLongPtr(tv.hwndFrozenLV, win.GWLP_WNDPROC, tableViewFrozenLVWndProcPtr)
	if tv.frozenLVOrigWndProcPtr == 0 {
		return nil, lastError("SetWindowLongPtr")
	}

	tv.hwndFrozenHdr = win.HWND(win.SendMessage(tv.hwndFrozenLV, win.LVM_GETHEADER, 0, 0))
	tv.frozenHdrOrigWndProcPtr = win.SetWindowLongPtr(tv.hwndFrozenHdr, win.GWLP_WNDPROC, tableViewHdrWndProcPtr)
	if tv.frozenHdrOrigWndProcPtr == 0 {
		return nil, lastError("SetWindowLongPtr")
	}

	if tv.hwndNormalLV = win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr("SysListView32"),
		nil,
		win.WS_CHILD|win.WS_CLIPSIBLINGS|win.WS_TABSTOP|win.WS_VISIBLE|win.LVS_OWNERDATA|win.LVS_REPORT|cfg.Style|rowHeightStyle,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		tv.hWnd,
		0,
		0,
		nil,
	); tv.hwndNormalLV == 0 {
		return nil, newError("creating normal lv failed")
	}

	tv.normalLVOrigWndProcPtr = win.SetWindowLongPtr(tv.hwndNormalLV, win.GWLP_WNDPROC, tableViewNormalLVWndProcPtr)
	if tv.normalLVOrigWndProcPtr == 0 {
		return nil, lastError("SetWindowLongPtr")
	}

	tv.hwndNormalHdr = win.HWND(win.SendMessage(tv.hwndNormalLV, win.LVM_GETHEADER, 0, 0))
	tv.normalHdrOrigWndProcPtr = win.SetWindowLongPtr(tv.hwndNormalHdr, win.GWLP_WNDPROC, tableViewHdrWndProcPtr)
	if tv.normalHdrOrigWndProcPtr == 0 {
		return nil, lastError("SetWindowLongPtr")
	}

	tv.SetPersistent(true)

	exStyle := win.SendMessage(tv.hwndFrozenLV, win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	exStyle |= win.LVS_EX_DOUBLEBUFFER | win.LVS_EX_FULLROWSELECT | win.LVS_EX_HEADERDRAGDROP | win.LVS_EX_LABELTIP | win.LVS_EX_SUBITEMIMAGES
	win.SendMessage(tv.hwndFrozenLV, win.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
	win.SendMessage(tv.hwndNormalLV, win.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)

	if hr := win.SetWindowTheme(tv.hwndFrozenLV, syscall.StringToUTF16Ptr("Explorer"), nil); win.FAILED(hr) {
		return nil, errorFromHRESULT("SetWindowTheme", hr)
	}
	if hr := win.SetWindowTheme(tv.hwndNormalLV, syscall.StringToUTF16Ptr("Explorer"), nil); win.FAILED(hr) {
		return nil, errorFromHRESULT("SetWindowTheme", hr)
	}

	win.SendMessage(tv.hwndFrozenLV, win.WM_CHANGEUISTATE, uintptr(win.MAKELONG(win.UIS_SET, win.UISF_HIDEFOCUS)), 0)
	win.SendMessage(tv.hwndNormalLV, win.WM_CHANGEUISTATE, uintptr(win.MAKELONG(win.UIS_SET, win.UISF_HIDEFOCUS)), 0)

	tv.applyFont(parent.Font())

	tv.style.dpi = tv.DPI()
	tv.ApplySysColors()

	tv.currentIndex = -1

	tv.GraphicsEffects().Add(InteractionEffect)
	tv.GraphicsEffects().Add(FocusEffect)

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

	tv.MustRegisterProperty("CurrentIndex", NewProperty(
		func() interface{} {
			return tv.CurrentIndex()
		},
		func(v interface{}) error {
			return tv.SetCurrentIndex(assertIntOr(v, -1))
		},
		tv.CurrentIndexChanged()))

	tv.MustRegisterProperty("CurrentItem", NewReadOnlyProperty(
		func() interface{} {
			if i := tv.CurrentIndex(); i > -1 {
				if rm, ok := tv.providedModel.(reflectModel); ok {
					return reflect.ValueOf(rm.Items()).Index(i).Interface()
				}
			}

			return nil
		},
		tv.CurrentIndexChanged()))

	tv.MustRegisterProperty("HasCurrentItem", NewReadOnlyBoolProperty(
		func() bool {
			return tv.CurrentIndex() != -1
		},
		tv.CurrentIndexChanged()))

	tv.MustRegisterProperty("ItemCount", NewReadOnlyProperty(
		func() interface{} {
			if tv.model == nil {
				return 0
			}
			return tv.model.RowCount()
		},
		tv.itemCountChangedPublisher.Event()))

	tv.MustRegisterProperty("SelectedCount", NewReadOnlyProperty(
		func() interface{} {
			return len(tv.selectedIndexes)
		},
		tv.SelectedIndexesChanged()))

	succeeded = true

	return tv, nil
}

func (tv *TableView) asTableView() *TableView {
	return tv
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

	if tv.hwndFrozenLV != 0 {
		win.DestroyWindow(tv.hwndFrozenLV)
		tv.hwndFrozenLV = 0
	}

	if tv.hwndNormalLV != 0 {
		win.DestroyWindow(tv.hwndNormalLV)
		tv.hwndNormalLV = 0
	}

	if tv.formActivatingHandle > -1 {
		if form := tv.Form(); form != nil {
			form.Activating().Detach(tv.formActivatingHandle)
		}
		tv.formActivatingHandle = -1
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

func (tv *TableView) applyEnabled(enabled bool) {
	tv.WidgetBase.applyEnabled(enabled)

	win.EnableWindow(tv.hwndFrozenLV, enabled)
	win.EnableWindow(tv.hwndNormalLV, enabled)
}

func (tv *TableView) applyFont(font *Font) {
	tv.WidgetBase.applyFont(font)

	hFont := uintptr(font.handleForDPI(tv.DPI()))

	win.SendMessage(tv.hwndFrozenLV, win.WM_SETFONT, hFont, 0)
	win.SendMessage(tv.hwndNormalLV, win.WM_SETFONT, hFont, 0)
}

func (tv *TableView) ApplyDPI(dpi int) {
	tv.style.dpi = dpi
	if tv.style.canvas != nil {
		tv.style.canvas.dpix = dpi
		tv.style.canvas.dpiy = dpi
	}

	tv.WidgetBase.ApplyDPI(dpi)

	for _, column := range tv.columns.items {
		column.update()
	}

	tv.disposeImageListAndCaches()
}

func (tv *TableView) ApplySysColors() {
	tv.WidgetBase.ApplySysColors()

	// As some combinations of property and state may be invalid for any theme,
	// we set some defaults here.
	tv.themeNormalBGColor = Color(win.GetSysColor(win.COLOR_WINDOW))
	tv.themeNormalTextColor = Color(win.GetSysColor(win.COLOR_WINDOWTEXT))
	tv.themeSelectedBGColor = tv.themeNormalBGColor
	tv.themeSelectedTextColor = tv.themeNormalTextColor
	tv.themeSelectedNotFocusedBGColor = tv.themeNormalBGColor

	if hTheme := win.OpenThemeData(tv.hwndNormalLV, syscall.StringToUTF16Ptr("Listview")); hTheme != 0 {
		defer win.CloseThemeData(hTheme)

		type item struct {
			stateID    int32
			propertyID int32
			color      *Color
		}

		items := [...]item{
			{win.LISS_NORMAL, win.TMT_FILLCOLOR, &tv.themeNormalBGColor},
			{win.LISS_NORMAL, win.TMT_TEXTCOLOR, &tv.themeNormalTextColor},
			{win.LISS_SELECTED, win.TMT_FILLCOLOR, &tv.themeSelectedBGColor},
			{win.LISS_SELECTED, win.TMT_TEXTCOLOR, &tv.themeSelectedTextColor},
			{win.LISS_SELECTEDNOTFOCUS, win.TMT_FILLCOLOR, &tv.themeSelectedNotFocusedBGColor},
		}
		for _, item := range items {
			var c win.COLORREF
			if result := win.GetThemeColor(hTheme, win.LVP_LISTITEM, item.stateID, item.propertyID, &c); !win.FAILED(result) {
				(*item.color) = Color(c)
			}
		}
	} else {
		// The others already have been retrieved above.
		tv.themeSelectedBGColor = Color(win.GetSysColor(win.COLOR_HIGHLIGHT))
		tv.themeSelectedTextColor = Color(win.GetSysColor(win.COLOR_HIGHLIGHTTEXT))
		tv.themeSelectedNotFocusedBGColor = Color(win.GetSysColor(win.COLOR_BTNFACE))
	}

	win.SendMessage(tv.hwndNormalLV, win.LVM_SETBKCOLOR, 0, uintptr(tv.themeNormalBGColor))
	win.SendMessage(tv.hwndFrozenLV, win.LVM_SETBKCOLOR, 0, uintptr(tv.themeNormalBGColor))
}

// ColumnsOrderable returns if the user can reorder columns by dragging and
// dropping column headers.
func (tv *TableView) ColumnsOrderable() bool {
	exStyle := win.SendMessage(tv.hwndNormalLV, win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	return exStyle&win.LVS_EX_HEADERDRAGDROP > 0
}

// SetColumnsOrderable sets if the user can reorder columns by dragging and
// dropping column headers.
func (tv *TableView) SetColumnsOrderable(enabled bool) {
	var hwnd win.HWND
	if tv.hasFrozenColumn {
		hwnd = tv.hwndFrozenLV
	} else {
		hwnd = tv.hwndNormalLV
	}

	exStyle := win.SendMessage(hwnd, win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	if enabled {
		exStyle |= win.LVS_EX_HEADERDRAGDROP
	} else {
		exStyle &^= win.LVS_EX_HEADERDRAGDROP
	}
	win.SendMessage(tv.hwndFrozenLV, win.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
	win.SendMessage(tv.hwndNormalLV, win.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)

	tv.columnsOrderableChangedPublisher.Publish()
}

// ColumnsSizable returns if the user can change column widths by dragging
// dividers in the header.
func (tv *TableView) ColumnsSizable() bool {
	style := win.GetWindowLong(tv.hwndNormalHdr, win.GWL_STYLE)

	return style&win.HDS_NOSIZING == 0
}

// SetColumnsSizable sets if the user can change column widths by dragging
// dividers in the header.
func (tv *TableView) SetColumnsSizable(b bool) error {
	updateStyle := func(headerHWnd win.HWND) error {
		style := win.GetWindowLong(headerHWnd, win.GWL_STYLE)

		if b {
			style &^= win.HDS_NOSIZING
		} else {
			style |= win.HDS_NOSIZING
		}

		if 0 == win.SetWindowLong(headerHWnd, win.GWL_STYLE, style) {
			return lastError("SetWindowLong(GWL_STYLE)")
		}

		return nil
	}

	if err := updateStyle(tv.hwndFrozenHdr); err != nil {
		return err
	}
	if err := updateStyle(tv.hwndNormalHdr); err != nil {
		return err
	}

	tv.columnsSizableChangedPublisher.Publish()

	return nil
}

// SortableByHeaderClick returns if the user can change sorting by clicking the header.
func (tv *TableView) SortableByHeaderClick() bool {
	return !hasWindowLongBits(tv.hwndFrozenLV, win.GWL_STYLE, win.LVS_NOSORTHEADER) ||
		!hasWindowLongBits(tv.hwndNormalLV, win.GWL_STYLE, win.LVS_NOSORTHEADER)
}

// HeaderHidden returns whether the column header is hidden.
func (tv *TableView) HeaderHidden() bool {
	style := win.GetWindowLong(tv.hwndNormalLV, win.GWL_STYLE)

	return style&win.LVS_NOCOLUMNHEADER != 0
}

// SetHeaderHidden sets whether the column header is hidden.
func (tv *TableView) SetHeaderHidden(hidden bool) error {
	updateStyle := func(hwnd win.HWND) error {
		style := win.GetWindowLong(hwnd, win.GWL_STYLE)

		if hidden {
			style |= win.LVS_NOCOLUMNHEADER
		} else {
			style &^= win.LVS_NOCOLUMNHEADER
		}

		if 0 == win.SetWindowLong(hwnd, win.GWL_STYLE, style) {
			return lastError("SetWindowLong(GWL_STYLE)")
		}

		return nil
	}

	if err := updateStyle(tv.hwndFrozenLV); err != nil {
		return err
	}

	return updateStyle(tv.hwndNormalLV)
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

	frozenCount := tv.visibleFrozenColumnCount()
	normalCount := len(visibleCols) - frozenCount

	if frozenCount > 0 {
		if win.FALSE == win.SendMessage(tv.hwndFrozenLV, win.LVM_GETCOLUMNORDERARRAY, uintptr(frozenCount), uintptr(unsafe.Pointer(&indices[0]))) {
			newError("LVM_GETCOLUMNORDERARRAY")
			return nil
		}
	}
	if normalCount > 0 {
		if win.FALSE == win.SendMessage(tv.hwndNormalLV, win.LVM_GETCOLUMNORDERARRAY, uintptr(normalCount), uintptr(unsafe.Pointer(&indices[frozenCount]))) {
			newError("LVM_GETCOLUMNORDERARRAY")
			return nil
		}
	}

	orderedCols := make([]*TableViewColumn, len(visibleCols))

	for i, j := range indices {
		if i >= frozenCount {
			j += int32(frozenCount)
		}
		orderedCols[i] = visibleCols[j]
	}

	return orderedCols
}

// RowsPerPage returns the number of fully visible rows.
func (tv *TableView) RowsPerPage() int {
	return int(win.SendMessage(tv.hwndNormalLV, win.LVM_GETCOUNTPERPAGE, 0, 0))
}

func (tv *TableView) Invalidate() error {
	win.InvalidateRect(tv.hwndFrozenLV, nil, true)
	win.InvalidateRect(tv.hwndNormalLV, nil, true)

	return tv.WidgetBase.Invalidate()
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
		if win.FALSE == win.SendMessage(tv.hwndFrozenLV, win.LVM_UPDATE, uintptr(index), 0) {
			return newError("LVM_UPDATE")
		}
		if win.FALSE == win.SendMessage(tv.hwndNormalLV, win.LVM_UPDATE, uintptr(index), 0) {
			return newError("LVM_UPDATE")
		}
	}

	return nil
}

func (tv *TableView) attachModel() {
	tv.rowsResetHandlerHandle = tv.model.RowsReset().Attach(func() {
		tv.setItemCount()

		tv.SetCurrentIndex(-1)

		tv.itemCountChangedPublisher.Publish()
	})

	tv.rowChangedHandlerHandle = tv.model.RowChanged().Attach(func(row int) {
		tv.UpdateItem(row)
	})

	tv.rowsInsertedHandlerHandle = tv.model.RowsInserted().Attach(func(from, to int) {
		i := tv.currentIndex

		tv.setItemCount()

		if from <= i {
			i += 1 + to - from

			tv.SetCurrentIndex(i)
		}

		tv.itemCountChangedPublisher.Publish()
	})

	tv.rowsRemovedHandlerHandle = tv.model.RowsRemoved().Attach(func(from, to int) {
		i := tv.currentIndex

		tv.setItemCount()

		index := i

		if from <= i && i <= to {
			index = -1
		} else if from < i {
			index -= 1 + to - from
		}

		if index != i {
			tv.SetCurrentIndex(index)
		}

		tv.itemCountChangedPublisher.Publish()
	})

	if sorter, ok := tv.model.(Sorter); ok {
		tv.sortChangedHandlerHandle = sorter.SortChanged().Attach(func() {
			col := sorter.SortedColumn()
			tv.setSortIcon(col, sorter.SortOrder())
			tv.Invalidate()
		})
	}
}

func (tv *TableView) detachModel() {
	tv.model.RowsReset().Detach(tv.rowsResetHandlerHandle)
	tv.model.RowChanged().Detach(tv.rowChangedHandlerHandle)
	tv.model.RowsInserted().Detach(tv.rowsInsertedHandlerHandle)
	tv.model.RowsRemoved().Detach(tv.rowsRemovedHandlerHandle)
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

	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	if tv.model != nil {
		tv.detachModel()

		tv.disposeImageListAndCaches()
	}

	oldProvidedModelStyler, _ := tv.providedModel.(CellStyler)
	if styler, ok := mdl.(CellStyler); ok || tv.styler == oldProvidedModelStyler {
		tv.styler = styler
	}

	tv.providedModel = mdl
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

		if lfs, ok := model.(lessFuncsSetter); ok {
			lessFuncs := make([]func(i, j int) bool, tv.columns.Len())
			for i, c := range tv.columns.items {
				lessFuncs[i] = c.lessFunc
			}
			lfs.setLessFuncs(lessFuncs)
		}

		if sorter, ok := tv.model.(Sorter); ok {
			if tv.sortedColumnIndex >= tv.visibleColumnCount() {
				tv.sortedColumnIndex = maxi(-1, mini(0, tv.visibleColumnCount()-1))
				tv.sortOrder = SortAscending
			}

			sorter.Sort(tv.sortedColumnIndex, tv.sortOrder)
		}
	}

	tv.SetCurrentIndex(-1)

	tv.setItemCount()

	tv.itemCountChangedPublisher.Publish()

	return nil
}

// TableModel returns the TableModel of the TableView.
func (tv *TableView) TableModel() TableModel {
	return tv.model
}

// ItemChecker returns the ItemChecker of the TableView.
func (tv *TableView) ItemChecker() ItemChecker {
	return tv.itemChecker
}

// SetItemChecker sets the ItemChecker of the TableView.
func (tv *TableView) SetItemChecker(itemChecker ItemChecker) {
	tv.itemChecker = itemChecker
}

// CellStyler returns the CellStyler of the TableView.
func (tv *TableView) CellStyler() CellStyler {
	return tv.styler
}

// SetCellStyler sets the CellStyler of the TableView.
func (tv *TableView) SetCellStyler(styler CellStyler) {
	tv.styler = styler
}

func (tv *TableView) setItemCount() error {
	var count int

	if tv.model != nil {
		count = tv.model.RowCount()
	}

	if 0 == win.SendMessage(tv.hwndFrozenLV, win.LVM_SETITEMCOUNT, uintptr(count), win.LVSICF_NOSCROLL) {
		return newError("SendMessage(LVM_SETITEMCOUNT)")
	}
	if 0 == win.SendMessage(tv.hwndNormalLV, win.LVM_SETITEMCOUNT, uintptr(count), win.LVSICF_NOSCROLL) {
		return newError("SendMessage(LVM_SETITEMCOUNT)")
	}

	return nil
}

// CheckBoxes returns if the *TableView has check boxes.
func (tv *TableView) CheckBoxes() bool {
	var hwnd win.HWND
	if tv.hasFrozenColumn {
		hwnd = tv.hwndFrozenLV
	} else {
		hwnd = tv.hwndNormalLV
	}

	return win.SendMessage(hwnd, win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)&win.LVS_EX_CHECKBOXES > 0
}

// SetCheckBoxes sets if the *TableView has check boxes.
func (tv *TableView) SetCheckBoxes(checkBoxes bool) {
	var hwnd, hwndOther win.HWND
	if tv.hasFrozenColumn {
		hwnd, hwndOther = tv.hwndFrozenLV, tv.hwndNormalLV
	} else {
		hwnd, hwndOther = tv.hwndNormalLV, tv.hwndFrozenLV
	}

	exStyle := win.SendMessage(hwnd, win.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	oldStyle := exStyle
	if checkBoxes {
		exStyle |= win.LVS_EX_CHECKBOXES
	} else {
		exStyle &^= win.LVS_EX_CHECKBOXES
	}
	if exStyle != oldStyle {
		win.SendMessage(hwnd, win.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
	}

	win.SendMessage(hwndOther, win.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle&^win.LVS_EX_CHECKBOXES)

	mask := win.SendMessage(hwnd, win.LVM_GETCALLBACKMASK, 0, 0)

	if checkBoxes {
		mask |= win.LVIS_STATEIMAGEMASK
	} else {
		mask &^= win.LVIS_STATEIMAGEMASK
	}

	if win.FALSE == win.SendMessage(hwnd, win.LVM_SETCALLBACKMASK, mask, 0) {
		newError("SendMessage(LVM_SETCALLBACKMASK)")
	}
}

func (tv *TableView) fromLVColIdx(frozen bool, index int32) int {
	var idx int32

	for i, tvc := range tv.columns.items {
		if frozen == tvc.frozen && tvc.visible {
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

func (tv *TableView) visibleFrozenColumnCount() int {
	var count int

	for _, tvc := range tv.columns.items {
		if tvc.frozen && tvc.visible {
			count++
		}
	}

	return count
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

// func (tv *TableView) setSelectedColumnIndex(index int) {
// 	tv.SendMessage(win.LVM_SETSELECTEDCOLUMN, uintptr(tv.toLVColIdx(index)), 0)
// }

func (tv *TableView) setSortIcon(index int, order SortOrder) error {
	idx := int(tv.toLVColIdx(index))

	frozenCount := tv.visibleFrozenColumnCount()

	for i, col := range tv.visibleColumns() {
		item := win.HDITEM{
			Mask: win.HDI_FORMAT,
		}

		var headerHwnd win.HWND
		var offset int
		if col.frozen {
			headerHwnd = tv.hwndFrozenHdr
		} else {
			headerHwnd = tv.hwndNormalHdr
			offset = -frozenCount
		}

		iPtr := uintptr(offset + i)
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
func (tv *TableView) SetCurrentIndex(index int) error {
	if tv.inSetCurrentIndex {
		return nil
	}
	tv.inSetCurrentIndex = true
	defer func() {
		tv.inSetCurrentIndex = false
	}()

	var lvi win.LVITEM

	lvi.StateMask = win.LVIS_FOCUSED | win.LVIS_SELECTED

	if tv.MultiSelection() {
		if win.FALSE == win.SendMessage(tv.hwndFrozenLV, win.LVM_SETITEMSTATE, ^uintptr(0), uintptr(unsafe.Pointer(&lvi))) {
			return newError("SendMessage(LVM_SETITEMSTATE)")
		}
		if win.FALSE == win.SendMessage(tv.hwndNormalLV, win.LVM_SETITEMSTATE, ^uintptr(0), uintptr(unsafe.Pointer(&lvi))) {
			return newError("SendMessage(LVM_SETITEMSTATE)")
		}
	}

	if index > -1 {
		lvi.State = win.LVIS_FOCUSED | win.LVIS_SELECTED
	}

	if win.FALSE == win.SendMessage(tv.hwndFrozenLV, win.LVM_SETITEMSTATE, uintptr(index), uintptr(unsafe.Pointer(&lvi))) {
		return newError("SendMessage(LVM_SETITEMSTATE)")
	}
	if win.FALSE == win.SendMessage(tv.hwndNormalLV, win.LVM_SETITEMSTATE, uintptr(index), uintptr(unsafe.Pointer(&lvi))) {
		return newError("SendMessage(LVM_SETITEMSTATE)")
	}

	if index != -1 {
		if win.FALSE == win.SendMessage(tv.hwndFrozenLV, win.LVM_ENSUREVISIBLE, uintptr(index), uintptr(0)) {
			return newError("SendMessage(LVM_ENSUREVISIBLE)")
		}
		// Windows bug? Sometimes a second LVM_ENSUREVISIBLE is required.
		if win.FALSE == win.SendMessage(tv.hwndFrozenLV, win.LVM_ENSUREVISIBLE, uintptr(index), uintptr(0)) {
			return newError("SendMessage(LVM_ENSUREVISIBLE)")
		}
		if win.FALSE == win.SendMessage(tv.hwndNormalLV, win.LVM_ENSUREVISIBLE, uintptr(index), uintptr(0)) {
			return newError("SendMessage(LVM_ENSUREVISIBLE)")
		}
		// Windows bug? Sometimes a second LVM_ENSUREVISIBLE is required.
		if win.FALSE == win.SendMessage(tv.hwndNormalLV, win.LVM_ENSUREVISIBLE, uintptr(index), uintptr(0)) {
			return newError("SendMessage(LVM_ENSUREVISIBLE)")
		}
	}

	tv.currentIndex = index

	if index == -1 || tv.itemStateChangedEventDelay == 0 {
		tv.currentIndexChangedPublisher.Publish()
	}

	if tv.MultiSelection() {
		tv.updateSelectedIndexes()
	}

	return nil
}

// CurrentIndexChanged is the event that is published after CurrentIndex has
// changed.
func (tv *TableView) CurrentIndexChanged() *Event {
	return tv.currentIndexChangedPublisher.Event()
}

// ItemVisible returns whether the item at position index is visible.
func (tv *TableView) ItemVisible(index int) bool {
	return 0 != win.SendMessage(tv.hwndNormalLV, win.LVM_ISITEMVISIBLE, uintptr(index), 0)
}

// EnsureItemVisible ensures the item at position index is visible, scrolling if necessary.
func (tv *TableView) EnsureItemVisible(index int) {
	win.SendMessage(tv.hwndNormalLV, win.LVM_ENSUREVISIBLE, uintptr(index), 0)
}

// SelectionHiddenWithoutFocus returns whether selection indicators are visible when the TableView does not have the keyboard input focus.
func (tv *TableView) SelectionHiddenWithoutFocus() bool {
	style := uint(win.GetWindowLong(tv.hwndNormalLV, win.GWL_STYLE))
	if style == 0 {
		lastError("GetWindowLong")
		return false
	}

	return style&win.LVS_SHOWSELALWAYS == 0
}

// SetSelectionHiddenWithoutFocus sets whether selection indicators are visible when the TableView does not have the keyboard input focus.
func (tv *TableView) SetSelectionHiddenWithoutFocus(hidden bool) error {
	if err := ensureWindowLongBits(tv.hwndFrozenLV, win.GWL_STYLE, win.LVS_SHOWSELALWAYS, !hidden); err != nil {
		return err
	}

	return ensureWindowLongBits(tv.hwndNormalLV, win.GWL_STYLE, win.LVS_SHOWSELALWAYS, !hidden)
}

// MultiSelection returns whether multiple items can be selected at once.
//
// By default only a single item can be selected at once.
func (tv *TableView) MultiSelection() bool {
	style := uint(win.GetWindowLong(tv.hwndNormalLV, win.GWL_STYLE))
	if style == 0 {
		lastError("GetWindowLong")
		return false
	}

	return style&win.LVS_SINGLESEL == 0
}

// SetMultiSelection sets whether multiple items can be selected at once.
func (tv *TableView) SetMultiSelection(multiSel bool) error {
	if err := ensureWindowLongBits(tv.hwndFrozenLV, win.GWL_STYLE, win.LVS_SINGLESEL, !multiSel); err != nil {
		return err
	}

	return ensureWindowLongBits(tv.hwndNormalLV, win.GWL_STYLE, win.LVS_SINGLESEL, !multiSel)
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

	if win.FALSE == win.SendMessage(tv.hwndFrozenLV, win.LVM_SETITEMSTATE, ^uintptr(0), lp) {
		return newError("SendMessage(LVM_SETITEMSTATE)")
	}
	if win.FALSE == win.SendMessage(tv.hwndNormalLV, win.LVM_SETITEMSTATE, ^uintptr(0), lp) {
		return newError("SendMessage(LVM_SETITEMSTATE)")
	}

	selectAll := false
	lvi.State = win.LVIS_FOCUSED | win.LVIS_SELECTED
	for _, i := range indexes {
		val := uintptr(i)
		if i == -1 {
			selectAll = true
			val = ^uintptr(0)
		}
		if win.FALSE == win.SendMessage(tv.hwndFrozenLV, win.LVM_SETITEMSTATE, val, lp) && i != -1 {
			return newError("SendMessage(LVM_SETITEMSTATE)")
		}
		if win.FALSE == win.SendMessage(tv.hwndNormalLV, win.LVM_SETITEMSTATE, val, lp) && i != -1 {
			return newError("SendMessage(LVM_SETITEMSTATE)")
		}
	}

	if !selectAll {
		idxs := make([]int, len(indexes))

		for i, j := range indexes {
			idxs[i] = j
		}

		tv.selectedIndexes = idxs
	} else {
		count := int(win.SendMessage(tv.hwndNormalLV, win.LVM_GETSELECTEDCOUNT, 0, 0))
		idxs := make([]int, count)
		for i := range idxs {
			idxs[i] = i
		}
		tv.selectedIndexes = idxs
	}

	return nil
}

func (tv *TableView) updateSelectedIndexes() {
	count := int(win.SendMessage(tv.hwndNormalLV, win.LVM_GETSELECTEDCOUNT, 0, 0))
	indexes := make([]int, count)

	j := -1
	for i := 0; i < count; i++ {
		j = int(win.SendMessage(tv.hwndNormalLV, win.LVM_GETNEXTITEM, uintptr(j), win.LVNI_SELECTED))
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

func (tv *TableView) copySelectedIndexes(hwndTo, hwndFrom win.HWND) error {
	count := int(win.SendMessage(hwndFrom, win.LVM_GETSELECTEDCOUNT, 0, 0))

	lvi := &win.LVITEM{StateMask: win.LVIS_FOCUSED | win.LVIS_SELECTED}
	lp := uintptr(unsafe.Pointer(lvi))

	if win.FALSE == win.SendMessage(hwndTo, win.LVM_SETITEMSTATE, ^uintptr(0), lp) {
		return newError("SendMessage(LVM_SETITEMSTATE)")
	}

	lvi.StateMask = win.LVIS_SELECTED
	lvi.State = win.LVIS_SELECTED

	j := -1
	for i := 0; i < count; i++ {
		j = int(win.SendMessage(hwndFrom, win.LVM_GETNEXTITEM, uintptr(j), win.LVNI_SELECTED))

		if win.FALSE == win.SendMessage(hwndTo, win.LVM_SETITEMSTATE, uintptr(j), lp) {
			return newError("SendMessage(LVM_SETITEMSTATE)")
		}
	}

	return nil
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

	if 0 == win.SendMessage(tv.hwndNormalLV, win.LVM_SETCOLUMNWIDTH, uintptr(colCount-1), win.LVSCW_AUTOSIZE_USEHEADER) {
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

// IgnoreNowhere returns if the *TableView should ignore left mouse clicks in the
// empty space. It forbids the user from unselecting the current index, or when
// multi selection is enabled, disables click drag selection.
func (tv *TableView) IgnoreNowhere() bool {
	return tv.ignoreNowhere
}

// IgnoreNowhere sets if the *TableView should ignore left mouse clicks in the
// empty space. It forbids the user from unselecting the current index, or when
// multi selection is enabled, disables click drag selection.
func (tv *TableView) SetIgnoreNowhere(value bool) {
	tv.ignoreNowhere = value
}

type tableViewState struct {
	SortColumnName     string
	SortOrder          SortOrder
	ColumnDisplayOrder []string
	Columns            []*tableViewColumnState
}

type tableViewColumnState struct {
	Name         string
	Title        string
	Width        int
	Visible      bool
	Frozen       bool
	LastSeenDate string
}

// SaveState writes the UI state of the *TableView to the settings.
func (tv *TableView) SaveState() error {
	if tv.columns.Len() == 0 {
		return nil
	}

	if tv.state == nil {
		tv.state = new(tableViewState)
	}

	tvs := tv.state

	tvs.SortColumnName = tv.columns.items[tv.sortedColumnIndex].name
	tvs.SortOrder = tv.sortOrder

	// tvs.Columns = make([]tableViewColumnState, tv.columns.Len())

	for _, tvc := range tv.columns.items {
		var tvcs *tableViewColumnState
		for _, cur := range tvs.Columns {
			if cur.Name == tvc.name {
				tvcs = cur
				break
			}
		}

		// tvcs := &tvs.Columns[i]

		if tvcs == nil {
			tvs.Columns = append(tvs.Columns, new(tableViewColumnState))
			tvcs = tvs.Columns[len(tvs.Columns)-1]
		}

		tvcs.Name = tvc.name
		tvcs.Title = tvc.titleOverride
		tvcs.Width = tvc.Width()
		tvcs.Visible = tvc.Visible()
		tvcs.Frozen = tvc.Frozen()
		tvcs.LastSeenDate = time.Now().Format("2006-01-02")
	}

	visibleCols := tv.visibleColumns()
	frozenCount := tv.visibleFrozenColumnCount()
	normalCount := len(visibleCols) - frozenCount
	indices := make([]int32, len(visibleCols))
	var lp uintptr
	if frozenCount > 0 {
		lp = uintptr(unsafe.Pointer(&indices[0]))

		if 0 == win.SendMessage(tv.hwndFrozenLV, win.LVM_GETCOLUMNORDERARRAY, uintptr(frozenCount), lp) {
			return newError("LVM_GETCOLUMNORDERARRAY")
		}
	}
	if normalCount > 0 {
		lp = uintptr(unsafe.Pointer(&indices[frozenCount]))

		if 0 == win.SendMessage(tv.hwndNormalLV, win.LVM_GETCOLUMNORDERARRAY, uintptr(normalCount), lp) {
			return newError("LVM_GETCOLUMNORDERARRAY")
		}
	}

	tvs.ColumnDisplayOrder = make([]string, len(visibleCols))
	for i, j := range indices {
		if i >= frozenCount {
			j += int32(frozenCount)
		}
		tvs.ColumnDisplayOrder[i] = visibleCols[j].name
	}

	state, err := json.Marshal(tvs)
	if err != nil {
		return err
	}

	return tv.WriteState(string(state))
}

// RestoreState restores the UI state of the *TableView from the settings.
func (tv *TableView) RestoreState() error {
	state, err := tv.ReadState()
	if err != nil {
		return err
	}
	if state == "" {
		return nil
	}

	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	if tv.state == nil {
		tv.state = new(tableViewState)
	}

	tvs := tv.state

	if err := json.Unmarshal(([]byte)(state), tvs); err != nil {
		return err
	}

	name2tvc := make(map[string]*TableViewColumn)

	for _, tvc := range tv.columns.items {
		name2tvc[tvc.name] = tvc
	}

	name2tvcs := make(map[string]*tableViewColumnState)

	tvcsRetained := make([]*tableViewColumnState, 0, len(tvs.Columns))
	for _, tvcs := range tvs.Columns {
		if tvcs.LastSeenDate != "" {
			if lastSeen, err := time.Parse("2006-02-01", tvcs.LastSeenDate); err != nil {
				tvcs.LastSeenDate = ""
			} else if name2tvc[tvcs.Name] == nil && lastSeen.Add(time.Hour*24*90).Before(time.Now()) {
				continue
			}
		}
		tvcsRetained = append(tvcsRetained, tvcs)

		name2tvcs[tvcs.Name] = tvcsRetained[len(tvcsRetained)-1]

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
			if err := tvc.SetVisible(tvc.visible && (visible || tvcs.Visible)); err != nil {
				return err
			}
			if err := tvc.SetFrozen(tvcs.Frozen); err != nil {
				return err
			}
		}
	}
	tvs.Columns = tvcsRetained

	visibleCount := tv.visibleColumnCount()
	frozenCount := tv.visibleFrozenColumnCount()
	normalCount := visibleCount - frozenCount

	indices := make([]int32, visibleCount)

	knownNames := make(map[string]struct{})

	displayOrder := make([]string, 0, visibleCount)
	for _, name := range tvs.ColumnDisplayOrder {
		knownNames[name] = struct{}{}
		if tvc, ok := name2tvc[name]; ok && tvc.visible {
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
			if tvc.name == name && j < visibleCount {
				idx := i
				if j >= frozenCount {
					idx -= frozenCount
				}
				indices[j] = int32(idx)
				break
			}
		}
	}

	var lp uintptr
	if frozenCount > 0 {
		lp = uintptr(unsafe.Pointer(&indices[0]))

		if 0 == win.SendMessage(tv.hwndFrozenLV, win.LVM_SETCOLUMNORDERARRAY, uintptr(frozenCount), lp) {
			return newError("LVM_SETCOLUMNORDERARRAY")
		}
	}
	if normalCount > 0 {
		lp = uintptr(unsafe.Pointer(&indices[frozenCount]))

		if 0 == win.SendMessage(tv.hwndNormalLV, win.LVM_SETCOLUMNORDERARRAY, uintptr(normalCount), lp) {
			return newError("LVM_SETCOLUMNORDERARRAY")
		}
	}

	for i, c := range tvs.Columns {
		if c.Name == tvs.SortColumnName && i < visibleCount {
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
					break
				}
			}
		}

		sorter.Sort(tv.sortedColumnIndex, tvs.SortOrder)
	}

	return nil
}

func (tv *TableView) toggleItemChecked(index int) error {
	checked := tv.itemChecker.Checked(index)

	if err := tv.itemChecker.SetChecked(index, !checked); err != nil {
		return wrapError(err)
	}

	if win.FALSE == win.SendMessage(tv.hwndFrozenLV, win.LVM_UPDATE, uintptr(index), 0) {
		return newError("SendMessage(LVM_UPDATE)")
	}
	if win.FALSE == win.SendMessage(tv.hwndNormalLV, win.LVM_UPDATE, uintptr(index), 0) {
		return newError("SendMessage(LVM_UPDATE)")
	}

	return nil
}

func (tv *TableView) applyImageListForImage(image interface{}) {
	tv.hIml, tv.usingSysIml, _ = imageListForImage(image, tv.DPI())

	tv.applyImageList()

	tv.imageUintptr2Index = make(map[uintptr]int32)
	tv.filePath2IconIndex = make(map[string]int32)
}

func (tv *TableView) applyImageList() {
	win.SendMessage(tv.hwndFrozenLV, win.LVM_SETIMAGELIST, win.LVSIL_SMALL, uintptr(tv.hIml))
	win.SendMessage(tv.hwndNormalLV, win.LVM_SETIMAGELIST, win.LVSIL_SMALL, uintptr(tv.hIml))
}

func (tv *TableView) disposeImageListAndCaches() {
	if tv.hIml != 0 && !tv.usingSysIml {
		win.SendMessage(tv.hwndFrozenLV, win.LVM_SETIMAGELIST, win.LVSIL_SMALL, 0)
		win.SendMessage(tv.hwndNormalLV, win.LVM_SETIMAGELIST, win.LVSIL_SMALL, 0)

		win.ImageList_Destroy(tv.hIml)
	}
	tv.hIml = 0

	tv.imageUintptr2Index = nil
	tv.filePath2IconIndex = nil
}

func (tv *TableView) Focused() bool {
	focused := win.GetFocus()

	return focused == tv.hwndFrozenLV || focused == tv.hwndNormalLV
}

func (tv *TableView) maybePublishFocusChanged(hwnd win.HWND, msg uint32, wp uintptr) {
	focused := msg == win.WM_SETFOCUS

	if focused != tv.focused && wp != uintptr(tv.hwndFrozenLV) && wp != uintptr(tv.hwndNormalLV) {
		tv.focused = focused
		tv.focusedChangedPublisher.Publish()
	}
}

func tableViewFrozenLVWndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	tv := (*TableView)(unsafe.Pointer(windowFromHandle(win.GetParent(hwnd)).AsWindowBase()))

	ensureWindowLongBits(hwnd, win.GWL_STYLE, win.WS_HSCROLL|win.WS_VSCROLL, false)

	switch msg {
	case win.WM_SETFOCUS:
		win.SetFocus(tv.hwndNormalLV)
		tv.maybePublishFocusChanged(hwnd, msg, wp)

	case win.WM_KILLFOCUS:
		tv.maybePublishFocusChanged(hwnd, msg, wp)

	case win.WM_MOUSEWHEEL:
		tableViewNormalLVWndProc(tv.hwndNormalLV, msg, wp, lp)
	}

	return tv.lvWndProc(tv.frozenLVOrigWndProcPtr, hwnd, msg, wp, lp)
}

func tableViewNormalLVWndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	tv := (*TableView)(unsafe.Pointer(windowFromHandle(win.GetParent(hwnd)).AsWindowBase()))

	switch msg {
	case win.WM_LBUTTONDOWN, win.WM_RBUTTONDOWN:
		win.SetFocus(tv.hwndFrozenLV)

	case win.WM_SETFOCUS:
		tv.WndProc(tv.hWnd, msg, wp, lp)
		tv.maybePublishFocusChanged(hwnd, msg, wp)

	case win.WM_KILLFOCUS:
		win.SendMessage(tv.hwndFrozenLV, msg, wp, lp)
		tv.WndProc(tv.hWnd, msg, wp, lp)
		tv.maybePublishFocusChanged(hwnd, msg, wp)
	}

	return tv.lvWndProc(tv.normalLVOrigWndProcPtr, hwnd, msg, wp, lp)
}

func (tv *TableView) lvWndProc(origWndProcPtr uintptr, hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	var hwndOther win.HWND
	if hwnd == tv.hwndFrozenLV {
		hwndOther = tv.hwndNormalLV
	} else {
		hwndOther = tv.hwndFrozenLV
	}

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
		if wp == win.VK_RETURN {
			return win.DLGC_WANTALLKEYS
		}

	case win.WM_LBUTTONDOWN, win.WM_RBUTTONDOWN, win.WM_LBUTTONDBLCLK, win.WM_RBUTTONDBLCLK:
		var hti win.LVHITTESTINFO
		hti.Pt = win.POINT{win.GET_X_LPARAM(lp), win.GET_Y_LPARAM(lp)}
		win.SendMessage(hwnd, win.LVM_HITTEST, 0, uintptr(unsafe.Pointer(&hti)))

		tv.itemIndexOfLastMouseButtonDown = int(hti.IItem)

		if hti.Flags == win.LVHT_NOWHERE {
			if tv.MultiSelection() {
				tv.publishNextSelClear = true
			} else {
				if tv.CheckBoxes() {
					if tv.currentIndex > -1 {
						tv.SetCurrentIndex(-1)
					}
				} else {
					// We keep the current item, if in single item selection mode without check boxes.
					win.SetFocus(tv.hwndFrozenLV)
					return 0
				}
			}

			if tv.IgnoreNowhere() {
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

	case win.WM_LBUTTONUP, win.WM_RBUTTONUP:
		tv.itemIndexOfLastMouseButtonDown = -1

	case win.WM_MOUSEMOVE, win.WM_MOUSELEAVE:
		if tv.inMouseEvent {
			break
		}
		tv.inMouseEvent = true
		defer func() {
			tv.inMouseEvent = false
		}()

		if msg == win.WM_MOUSEMOVE {
			y := int(win.GET_Y_LPARAM(lp))
			lp = uintptr(win.MAKELONG(0, uint16(y)))
		}

		win.SendMessage(hwndOther, msg, wp, lp)

	case win.WM_KEYDOWN:
		if wp == win.VK_SPACE &&
			tv.currentIndex > -1 &&
			tv.itemChecker != nil &&
			tv.CheckBoxes() {

			tv.toggleItemChecked(tv.currentIndex)
		}

		tv.handleKeyDown(wp, lp)

	case win.WM_KEYUP:
		tv.handleKeyUp(wp, lp)

	case win.WM_NOTIFY:
		nmh := ((*win.NMHDR)(unsafe.Pointer(lp)))
		switch nmh.HwndFrom {
		case tv.hwndFrozenHdr, tv.hwndNormalHdr:
			if nmh.Code == win.NM_CUSTOMDRAW {
				return tableViewHdrWndProc(nmh.HwndFrom, msg, wp, lp)
			}
		}

		switch nmh.Code {
		case win.LVN_GETDISPINFO:
			di := (*win.NMLVDISPINFO)(unsafe.Pointer(lp))

			row := int(di.Item.IItem)
			col := tv.fromLVColIdx(hwnd == tv.hwndFrozenLV, di.Item.ISubItem)
			if col == -1 {
				break
			}

			if di.Item.Mask&win.LVIF_TEXT > 0 {
				value := tv.model.Value(row, col)
				var text string
				if format := tv.columns.items[col].formatFunc; format != nil {
					text = format(value)
				} else {
					switch val := value.(type) {
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

					case bool:
						if val {
							text = checkmark
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
				}

				utf16 := syscall.StringToUTF16(text)
				buf := (*[264]uint16)(unsafe.Pointer(di.Item.PszText))
				max := mini(len(utf16), int(di.Item.CchTextMax))
				copy((*buf)[:], utf16[:max])
				(*buf)[max-1] = 0
			}

			if (tv.imageProvider != nil || tv.styler != nil) && di.Item.Mask&win.LVIF_IMAGE > 0 {
				var image interface{}
				if di.Item.ISubItem == 0 {
					if ip := tv.imageProvider; ip != nil && image == nil {
						image = ip.Image(row)
					}
				}
				if styler := tv.styler; styler != nil && image == nil {
					tv.style.row = row
					tv.style.col = col
					tv.style.bounds = Rectangle{}
					tv.style.Image = nil

					styler.StyleCell(&tv.style)

					image = tv.style.Image
				}

				if image != nil {
					if tv.hIml == 0 {
						tv.applyImageListForImage(image)
					}

					di.Item.IImage = imageIndexMaybeAdd(
						image,
						tv.hIml,
						tv.usingSysIml,
						tv.imageUintptr2Index,
						tv.filePath2IconIndex,
						tv.DPI())
				}
			}

			if di.Item.ISubItem == 0 && di.Item.StateMask&win.LVIS_STATEIMAGEMASK > 0 &&
				tv.itemChecker != nil {
				checked := tv.itemChecker.Checked(row)

				if checked {
					di.Item.State = 0x2000
				} else {
					di.Item.State = 0x1000
				}
			}

		case win.NM_CUSTOMDRAW:
			nmlvcd := (*win.NMLVCUSTOMDRAW)(unsafe.Pointer(lp))

			if nmlvcd.IIconPhase == 0 {
				row := int(nmlvcd.Nmcd.DwItemSpec)
				col := tv.fromLVColIdx(hwnd == tv.hwndFrozenLV, nmlvcd.ISubItem)
				if col == -1 {
					break
				}

				applyCellStyle := func() int {
					if tv.styler != nil {
						tv.style.row = row
						tv.style.col = col
						tv.style.bounds = tv.RectangleTo96DPI(rectangleFromRECT(nmlvcd.Nmcd.Rc))
						tv.style.hdc = nmlvcd.Nmcd.Hdc
						tv.style.BackgroundColor = tv.itemBGColor
						tv.style.TextColor = tv.itemTextColor
						tv.style.Font = nil
						tv.style.Image = nil

						tv.styler.StyleCell(&tv.style)

						defer func() {
							tv.style.bounds = Rectangle{}
							if tv.style.canvas != nil {
								tv.style.canvas.Dispose()
								tv.style.canvas = nil
							}
							tv.style.hdc = 0
						}()

						if tv.style.canvas != nil {
							return win.CDRF_SKIPDEFAULT
						}

						nmlvcd.ClrTextBk = win.COLORREF(tv.style.BackgroundColor)
						nmlvcd.ClrText = win.COLORREF(tv.style.TextColor)

						if font := tv.style.Font; font != nil {
							win.SelectObject(nmlvcd.Nmcd.Hdc, win.HGDIOBJ(font.handleForDPI(tv.DPI())))
						}
					}

					return 0
				}

				switch nmlvcd.Nmcd.DwDrawStage {
				case win.CDDS_PREPAINT:
					return win.CDRF_NOTIFYITEMDRAW

				case win.CDDS_ITEMPREPAINT:
					var selected bool
					if itemState := win.SendMessage(hwnd, win.LVM_GETITEMSTATE, nmlvcd.Nmcd.DwItemSpec, win.LVIS_SELECTED); itemState&win.LVIS_SELECTED != 0 {
						selected = true

						tv.itemBGColor = tv.themeSelectedBGColor
						tv.itemTextColor = tv.themeSelectedTextColor
					} else {
						tv.itemBGColor = tv.themeNormalBGColor
						tv.itemTextColor = tv.themeNormalTextColor
					}

					if !selected && tv.alternatingRowBGColor != 0 && row%2 == 1 {
						tv.itemBGColor = tv.alternatingRowBGColor
					}

					tv.style.BackgroundColor = tv.itemBGColor
					tv.style.TextColor = tv.itemTextColor

					if tv.styler != nil {
						tv.style.row = row
						tv.style.col = -1
						tv.style.bounds = tv.RectangleTo96DPI(rectangleFromRECT(nmlvcd.Nmcd.Rc))
						tv.style.hdc = 0
						tv.style.Font = nil
						tv.style.Image = nil

						tv.styler.StyleCell(&tv.style)

						tv.itemFont = tv.style.Font
					}

					if selected {
						tv.style.BackgroundColor = tv.itemBGColor
						tv.style.TextColor = tv.itemTextColor
					} else {
						tv.itemBGColor = tv.style.BackgroundColor
						tv.itemTextColor = tv.style.TextColor
					}

					if tv.style.BackgroundColor != tv.themeNormalBGColor {
						var color Color
						if selected && !tv.Focused() {
							color = tv.themeSelectedNotFocusedBGColor
						} else {
							color = tv.style.BackgroundColor
						}

						if brush, _ := NewSolidColorBrush(color); brush != nil {
							defer brush.Dispose()

							canvas, _ := newCanvasFromHDC(nmlvcd.Nmcd.Hdc)
							canvas.fillRectanglePixels(brush, rectangleFromRECT(nmlvcd.Nmcd.Rc))
						}
					}

					nmlvcd.ClrText = win.COLORREF(tv.style.TextColor)
					nmlvcd.ClrTextBk = win.COLORREF(tv.style.BackgroundColor)

					return win.CDRF_NOTIFYSUBITEMDRAW

				case win.CDDS_ITEMPREPAINT | win.CDDS_SUBITEM:
					if tv.itemFont != nil {
						win.SelectObject(nmlvcd.Nmcd.Hdc, win.HGDIOBJ(tv.itemFont.handleForDPI(0)))
					}

					if applyCellStyle() == win.CDRF_SKIPDEFAULT && win.IsAppThemed() {
						return win.CDRF_SKIPDEFAULT
					}

					return win.CDRF_NEWFONT | win.CDRF_SKIPPOSTPAINT | win.CDRF_NOTIFYPOSTPAINT

				case win.CDDS_ITEMPOSTPAINT | win.CDDS_SUBITEM:
					if applyCellStyle() == win.CDRF_SKIPDEFAULT {
						return win.CDRF_SKIPDEFAULT
					}

					return win.CDRF_NEWFONT | win.CDRF_SKIPPOSTPAINT
				}

				return win.CDRF_SKIPPOSTPAINT
			}

			return win.CDRF_SKIPPOSTPAINT

		case win.LVN_BEGINSCROLL:
			if tv.scrolling {
				break
			}
			tv.scrolling = true
			defer func() {
				tv.scrolling = false
			}()

			var rc win.RECT
			win.SendMessage(hwnd, win.LVM_GETITEMRECT, 0, uintptr(unsafe.Pointer(&rc)))

			nmlvs := (*win.NMLVSCROLL)(unsafe.Pointer(lp))
			win.SendMessage(hwndOther, win.LVM_SCROLL, 0, uintptr(nmlvs.Dy*(rc.Bottom-rc.Top)))

		case win.LVN_COLUMNCLICK:
			nmlv := (*win.NMLISTVIEW)(unsafe.Pointer(lp))

			col := tv.fromLVColIdx(hwnd == tv.hwndFrozenLV, nmlv.ISubItem)

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
			nmlv := (*win.NMLISTVIEW)(unsafe.Pointer(lp))

			if tv.hwndItemChanged != 0 && tv.hwndItemChanged != hwnd {
				break
			}
			tv.hwndItemChanged = hwnd
			defer func() {
				tv.hwndItemChanged = 0
			}()

			tv.copySelectedIndexes(hwndOther, hwnd)

			if nmlv.IItem == -1 && !tv.publishNextSelClear {
				break
			}
			tv.publishNextSelClear = false

			selectedNow := nmlv.UNewState&win.LVIS_SELECTED > 0
			selectedBefore := nmlv.UOldState&win.LVIS_SELECTED > 0
			if tv.itemIndexOfLastMouseButtonDown != -1 && selectedNow && !selectedBefore && ModifiersDown()&(ModControl|ModShift) == 0 {
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

					tv.SetCurrentIndex(int(nmlv.IItem))
				} else {
					tv.SetCurrentIndex(int(nmlv.IItem))
				}
			}

			if selectedNow != selectedBefore {
				if !tv.inSetSelectedIndexes && tv.MultiSelection() {
					tv.updateSelectedIndexes()
				}
			}

		case win.LVN_ODSTATECHANGED:
			if tv.hwndItemChanged != 0 && tv.hwndItemChanged != hwnd {
				break
			}
			tv.hwndItemChanged = hwnd
			defer func() {
				tv.hwndItemChanged = 0
			}()

			tv.copySelectedIndexes(hwndOther, hwnd)

			tv.updateSelectedIndexes()

		case win.LVN_ITEMACTIVATE:
			nmia := (*win.NMITEMACTIVATE)(unsafe.Pointer(lp))

			if tv.itemStateChangedEventDelay > 0 {
				tv.delayedCurrentIndexChangedCanceled = true
			}

			if int(nmia.IItem) != tv.currentIndex {
				tv.SetCurrentIndex(int(nmia.IItem))
				tv.currentIndexChangedPublisher.Publish()
			}

			tv.itemActivatedPublisher.Publish()

		case win.HDN_ITEMCHANGING:
			tv.updateLVSizes()
		}

	case win.WM_UPDATEUISTATE:
		switch win.LOWORD(uint32(wp)) {
		case win.UIS_SET:
			wp |= win.UISF_HIDEFOCUS << 16

		case win.UIS_CLEAR, win.UIS_INITIALIZE:
			wp &^= ^uintptr(win.UISF_HIDEFOCUS << 16)
		}
	}

	lpFixed := lp
	fixXInLP := func() {
		// fmt.Printf("hwnd == tv.hwndNormalLV: %t, tv.hasFrozenColumn: %t\n", hwnd == tv.hwndNormalLV, tv.hasFrozenColumn)
		if hwnd == tv.hwndNormalLV && tv.hasFrozenColumn {
			var rc win.RECT
			if win.GetWindowRect(tv.hwndFrozenLV, &rc) {
				x := int(win.GET_X_LPARAM(lp)) + int(rc.Right-rc.Left)
				y := int(win.GET_Y_LPARAM(lp))

				lpFixed = uintptr(win.MAKELONG(uint16(x), uint16(y)))
			}
		}
	}

	switch msg {
	case win.WM_LBUTTONDOWN, win.WM_MBUTTONDOWN, win.WM_RBUTTONDOWN:
		fixXInLP()
		tv.publishMouseEvent(&tv.mouseDownPublisher, msg, wp, lpFixed)

	case win.WM_LBUTTONUP, win.WM_MBUTTONUP, win.WM_RBUTTONUP:
		fixXInLP()
		tv.publishMouseEvent(&tv.mouseUpPublisher, msg, wp, lpFixed)

	case win.WM_MOUSEMOVE:
		fixXInLP()
		tv.publishMouseEvent(&tv.mouseMovePublisher, msg, wp, lpFixed)

	case win.WM_MOUSEWHEEL:
		fixXInLP()
		tv.publishMouseWheelEvent(&tv.mouseWheelPublisher, wp, lpFixed)
	}

	return win.CallWindowProc(origWndProcPtr, hwnd, msg, wp, lp)
}

func tableViewHdrWndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	tv := (*TableView)(unsafe.Pointer(windowFromHandle(win.GetParent(win.GetParent(hwnd))).AsWindowBase()))

	var origWndProcPtr uintptr
	if hwnd == tv.hwndFrozenHdr {
		origWndProcPtr = tv.frozenHdrOrigWndProcPtr
	} else {
		origWndProcPtr = tv.normalHdrOrigWndProcPtr
	}

	switch msg {
	case win.WM_NOTIFY:
		switch ((*win.NMHDR)(unsafe.Pointer(lp))).Code {
		case win.NM_CUSTOMDRAW:
			if tv.customHeaderHeight == 0 {
				break
			}

			nmcd := (*win.NMCUSTOMDRAW)(unsafe.Pointer(lp))

			switch nmcd.DwDrawStage {
			case win.CDDS_PREPAINT:
				return win.CDRF_NOTIFYITEMDRAW

			case win.CDDS_ITEMPREPAINT:
				return win.CDRF_NOTIFYPOSTPAINT

			case win.CDDS_ITEMPOSTPAINT:
				col := tv.fromLVColIdx(hwnd == tv.hwndFrozenHdr, int32(nmcd.DwItemSpec))
				if tv.styler != nil && col > -1 {
					tv.style.row = -1
					tv.style.col = col
					tv.style.bounds = tv.RectangleTo96DPI(rectangleFromRECT(nmcd.Rc))
					tv.style.hdc = nmcd.Hdc
					tv.style.TextColor = tv.themeNormalTextColor
					tv.style.Font = nil

					tv.styler.StyleCell(&tv.style)

					defer func() {
						tv.style.bounds = Rectangle{}
						if tv.style.canvas != nil {
							tv.style.canvas.Dispose()
							tv.style.canvas = nil
						}
						tv.style.hdc = 0
					}()
				}

				return win.CDRF_DODEFAULT
			}

			return win.CDRF_DODEFAULT
		}

	case win.HDM_LAYOUT:
		if tv.customHeaderHeight == 0 {
			break
		}

		result := win.CallWindowProc(origWndProcPtr, hwnd, msg, wp, lp)

		hdl := (*win.HDLAYOUT)(unsafe.Pointer(lp))
		hdl.Prc.Top = int32(tv.customHeaderHeight)
		hdl.Pwpos.Cy = int32(tv.customHeaderHeight)

		return result
	}

	return win.CallWindowProc(origWndProcPtr, hwnd, msg, wp, lp)
}

func (tv *TableView) WndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case win.WM_NOTIFY:
		nmh := (*win.NMHDR)(unsafe.Pointer(lp))
		switch nmh.HwndFrom {
		case tv.hwndFrozenLV:
			return tableViewFrozenLVWndProc(nmh.HwndFrom, msg, wp, lp)

		case tv.hwndNormalLV:
			return tableViewNormalLVWndProc(nmh.HwndFrom, msg, wp, lp)
		}

	case win.WM_SIZE:
		if tv.formActivatingHandle == -1 {
			if form := tv.Form(); form != nil {
				tv.formActivatingHandle = form.Activating().Attach(func() {
					if tv.hwndNormalLV == win.GetFocus() {
						win.SetFocus(tv.hwndFrozenLV)
					}
				})
			}
		}

		tv.updateLVSizes()

	case win.WM_TIMER:
		if !win.KillTimer(tv.hWnd, wp) {
			lastError("KillTimer")
		}

		switch wp {
		case tableViewCurrentIndexChangedTimerId:
			if !tv.delayedCurrentIndexChangedCanceled {
				tv.currentIndexChangedPublisher.Publish()
			}

		case tableViewSelectedIndexesChangedTimerId:
			tv.selectedIndexesChangedPublisher.Publish()
		}

	case win.WM_MEASUREITEM:
		mis := (*win.MEASUREITEMSTRUCT)(unsafe.Pointer(lp))
		mis.ItemHeight = uint32(tv.customRowHeight)

		ensureWindowLongBits(tv.hwndFrozenLV, win.GWL_STYLE, win.LVS_OWNERDRAWFIXED, false)
		ensureWindowLongBits(tv.hwndNormalLV, win.GWL_STYLE, win.LVS_OWNERDRAWFIXED, false)

	case win.WM_SETFOCUS:
		win.SetFocus(tv.hwndNormalLV)

	case win.WM_DESTROY:
		// As we subclass all windows of system classes, we prevented the
		// clean-up code in the WM_NCDESTROY handlers of some windows from
		// being called. To fix this, we restore the original window
		// procedures here.
		if tv.frozenHdrOrigWndProcPtr != 0 {
			win.SetWindowLongPtr(tv.hwndFrozenHdr, win.GWLP_WNDPROC, tv.frozenHdrOrigWndProcPtr)
		}
		if tv.frozenLVOrigWndProcPtr != 0 {
			win.SetWindowLongPtr(tv.hwndFrozenLV, win.GWLP_WNDPROC, tv.frozenLVOrigWndProcPtr)
		}
		if tv.normalHdrOrigWndProcPtr != 0 {
			win.SetWindowLongPtr(tv.hwndNormalHdr, win.GWLP_WNDPROC, tv.normalHdrOrigWndProcPtr)
		}
		if tv.normalLVOrigWndProcPtr != 0 {
			win.SetWindowLongPtr(tv.hwndNormalLV, win.GWLP_WNDPROC, tv.normalLVOrigWndProcPtr)
		}
	}

	return tv.WidgetBase.WndProc(hwnd, msg, wp, lp)
}

func (tv *TableView) updateLVSizes() {
	tv.updateLVSizesWithSpecialCare(false)
}

func (tv *TableView) updateLVSizesWithSpecialCare(needSpecialCare bool) {
	var width int
	for i := tv.columns.Len() - 1; i >= 0; i-- {
		if col := tv.columns.At(i); col.frozen {
			width += col.Width()
		}
	}

	width = tv.IntFrom96DPI(width)

	cb := tv.ClientBoundsPixels()

	win.MoveWindow(tv.hwndNormalLV, int32(width), 0, int32(cb.Width-width), int32(cb.Height), true)

	var sbh int
	if hasWindowLongBits(tv.hwndNormalLV, win.GWL_STYLE, win.WS_HSCROLL) {
		sbh = int(win.GetSystemMetrics(win.SM_CYHSCROLL))
	}

	win.MoveWindow(tv.hwndFrozenLV, 0, 0, int32(width), int32(cb.Height-sbh), true)

	if needSpecialCare {
		tv.updateLVSizesNeedsSpecialCare = true
	}

	if tv.updateLVSizesNeedsSpecialCare {
		win.ShowWindow(tv.hwndNormalLV, win.SW_HIDE)
		win.ShowWindow(tv.hwndNormalLV, win.SW_SHOW)
	}

	if !needSpecialCare {
		tv.updateLVSizesNeedsSpecialCare = false
	}
}
