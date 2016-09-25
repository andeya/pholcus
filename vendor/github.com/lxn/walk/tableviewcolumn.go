// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

// TableViewColumn represents a column in a TableView.
type TableViewColumn struct {
	tv            *TableView
	name          string
	dataMember    string
	alignment     Alignment1D
	format        string
	precision     int
	title         string
	titleOverride string
	visible       bool
	width         int
}

// NewTableViewColumn returns a new TableViewColumn.
func NewTableViewColumn() *TableViewColumn {
	return &TableViewColumn{
		format:  "%v",
		visible: true,
		width:   50,
	}
}

// Alignment returns the alignment of the TableViewColumn.
func (tvc *TableViewColumn) Alignment() Alignment1D {
	return tvc.alignment
}

// SetAlignment sets the alignment of the TableViewColumn.
func (tvc *TableViewColumn) SetAlignment(alignment Alignment1D) (err error) {
	if alignment == tvc.alignment {
		return nil
	}

	old := tvc.alignment
	defer func() {
		if err != nil {
			tvc.alignment = old
		}
	}()

	tvc.alignment = alignment

	return tvc.update()
}

// DataMember returns the data member this TableViewColumn is bound against.
func (tvc *TableViewColumn) DataMember() string {
	return tvc.dataMember
}

// DataMemberEffective returns the effective data member this TableViewColumn is
// bound against.
func (tvc *TableViewColumn) DataMemberEffective() string {
	if tvc.dataMember != "" {
		return tvc.dataMember
	}

	return tvc.name
}

// SetDataMember sets the data member this TableViewColumn is bound against.
func (tvc *TableViewColumn) SetDataMember(dataMember string) {
	tvc.dataMember = dataMember
}

// Format returns the format string for converting a value into a string.
func (tvc *TableViewColumn) Format() string {
	return tvc.format
}

// SetFormat sets the format string for converting a value into a string.
func (tvc *TableViewColumn) SetFormat(format string) (err error) {
	if format == tvc.format {
		return nil
	}

	old := tvc.format
	defer func() {
		if err != nil {
			tvc.format = old
		}
	}()

	tvc.format = format

	if tvc.tv == nil {
		return nil
	}

	return tvc.tv.Invalidate()
}

// Name returns the name of this TableViewColumn.
func (tvc *TableViewColumn) Name() string {
	return tvc.name
}

// SetName sets the name of this TableViewColumn.
func (tvc *TableViewColumn) SetName(name string) {
	tvc.name = name
}

// Precision returns the number of decimal places for formatting float32,
// float64 or big.Rat values.
func (tvc *TableViewColumn) Precision() int {
	return tvc.precision
}

// SetPrecision sets the number of decimal places for formatting float32,
// float64 or big.Rat values.
func (tvc *TableViewColumn) SetPrecision(precision int) (err error) {
	if precision == tvc.precision {
		return nil
	}

	old := tvc.precision
	defer func() {
		if err != nil {
			tvc.precision = old
		}
	}()

	tvc.precision = precision

	if tvc.tv == nil {
		return nil
	}

	return tvc.tv.Invalidate()
}

// Title returns the (default) text to display in the column header.
func (tvc *TableViewColumn) Title() string {
	return tvc.title
}

// SetTitle sets the (default) text to display in the column header.
func (tvc *TableViewColumn) SetTitle(title string) (err error) {
	if title == tvc.title {
		return nil
	}

	old := tvc.title
	defer func() {
		if err != nil {
			tvc.title = old
		}
	}()

	tvc.title = title

	return tvc.update()
}

// TitleOverride returns the (overridden by user) text to display in the column
// header.
func (tvc *TableViewColumn) TitleOverride() string {
	return tvc.titleOverride
}

// SetTitleOverride sets the (overridden by user) text to display in the column
// header.
func (tvc *TableViewColumn) SetTitleOverride(titleOverride string) (err error) {
	if titleOverride == tvc.titleOverride {
		return nil
	}

	old := tvc.titleOverride
	defer func() {
		if err != nil {
			tvc.titleOverride = old
		}
	}()

	tvc.titleOverride = titleOverride

	return tvc.update()
}

// TitleEffective returns the effective text to display in the column header.
func (tvc *TableViewColumn) TitleEffective() string {
	if tvc.titleOverride != "" {
		return tvc.titleOverride
	}

	if tvc.title != "" {
		return tvc.title
	}

	return tvc.DataMemberEffective()
}

// Visible returns if the column is visible.
func (tvc *TableViewColumn) Visible() bool {
	return tvc.visible
}

// SetVisible sets if the column is visible.
func (tvc *TableViewColumn) SetVisible(visible bool) (err error) {
	if visible == tvc.visible {
		return nil
	}

	old := tvc.visible
	defer func() {
		if err != nil {
			tvc.visible = old
		}
	}()

	tvc.visible = visible

	if tvc.tv == nil {
		return nil
	}

	if visible {
		return tvc.create()
	}

	return tvc.destroy()
}

// Width returns the width of the column in pixels.
func (tvc *TableViewColumn) Width() int {
	if tvc.tv == nil || !tvc.visible {
		return tvc.width
	}

	return int(tvc.tv.SendMessage(win.LVM_GETCOLUMNWIDTH, uintptr(tvc.indexInListView()), 0))
}

// SetWidth sets the width of the column in pixels.
func (tvc *TableViewColumn) SetWidth(width int) (err error) {
	if width == tvc.width {
		return nil
	}

	old := tvc.width
	defer func() {
		if err != nil {
			tvc.width = old
		}
	}()

	tvc.width = width

	return tvc.update()
}

func (tvc *TableViewColumn) indexInListView() int32 {
	if tvc.tv == nil {
		return -1
	}

	var idx int32

	for _, c := range tvc.tv.columns.items {
		if c == tvc {
			break
		}

		if c.visible {
			idx++
		}
	}

	return idx
}

func (tvc *TableViewColumn) create() error {
	var lvc win.LVCOLUMN

	index := tvc.indexInListView()

	lvc.Mask = win.LVCF_FMT | win.LVCF_WIDTH | win.LVCF_TEXT | win.LVCF_SUBITEM
	lvc.ISubItem = index
	lvc.PszText = syscall.StringToUTF16Ptr(tvc.TitleEffective())
	if tvc.width > 0 {
		lvc.Cx = int32(tvc.width)
	} else {
		lvc.Cx = 100
	}

	switch tvc.alignment {
	case AlignCenter:
		lvc.Fmt = 2

	case AlignFar:
		lvc.Fmt = 1
	}

	j := tvc.tv.SendMessage(win.LVM_INSERTCOLUMN, uintptr(index), uintptr(unsafe.Pointer(&lvc)))
	if int(j) == -1 {
		return newError("TableView.SetModel: Failed to insert column.")
	}

	return nil
}

func (tvc *TableViewColumn) destroy() error {
	width := tvc.Width()

	if win.FALSE == tvc.tv.SendMessage(win.LVM_DELETECOLUMN, uintptr(tvc.indexInListView()), 0) {
		return newError("LVM_DELETECOLUMN")
	}

	tvc.width = width

	return nil
}

func (tvc *TableViewColumn) update() error {
	if tvc.tv == nil || !tvc.visible {
		return nil
	}

	lvc := tvc.getLVCOLUMN()

	if win.FALSE == tvc.tv.SendMessage(win.LVM_SETCOLUMN, uintptr(tvc.indexInListView()), uintptr(unsafe.Pointer(lvc))) {
		return newError("LVM_SETCOLUMN")
	}

	return nil
}

func (tvc *TableViewColumn) getLVCOLUMN() *win.LVCOLUMN {
	var lvc win.LVCOLUMN

	lvc.Mask = win.LVCF_FMT | win.LVCF_WIDTH | win.LVCF_TEXT | win.LVCF_SUBITEM
	lvc.ISubItem = int32(tvc.indexInListView())
	lvc.PszText = syscall.StringToUTF16Ptr(tvc.TitleEffective())
	lvc.Cx = int32(tvc.width)

	switch tvc.alignment {
	case AlignCenter:
		lvc.Fmt = 2

	case AlignFar:
		lvc.Fmt = 1
	}

	return &lvc
}
