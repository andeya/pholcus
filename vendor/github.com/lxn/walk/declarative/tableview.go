// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type TableView struct {
	AssignTo                   **walk.TableView
	Name                       string
	Enabled                    Property
	Visible                    Property
	Font                       Font
	ToolTipText                Property
	MinSize                    Size
	MaxSize                    Size
	StretchFactor              int
	Row                        int
	RowSpan                    int
	Column                     int
	ColumnSpan                 int
	AlwaysConsumeSpace         bool
	ContextMenuItems           []MenuItem
	OnKeyDown                  walk.KeyEventHandler
	OnKeyPress                 walk.KeyEventHandler
	OnKeyUp                    walk.KeyEventHandler
	OnMouseDown                walk.MouseEventHandler
	OnMouseMove                walk.MouseEventHandler
	OnMouseUp                  walk.MouseEventHandler
	OnSizeChanged              walk.EventHandler
	Columns                    []TableViewColumn
	Model                      interface{}
	AlternatingRowBGColor      walk.Color
	CheckBoxes                 bool
	ItemStateChangedEventDelay int
	LastColumnStretched        bool
	ColumnsOrderable           Property
	ColumnsSizable             Property
	MultiSelection             bool
	NotSortableByHeaderClick   bool
	OnCurrentIndexChanged      walk.EventHandler
	OnSelectedIndexesChanged   walk.EventHandler
	OnItemActivated            walk.EventHandler
}

func (tv TableView) Create(builder *Builder) error {
	var w *walk.TableView
	var err error
	if tv.NotSortableByHeaderClick {
		w, err = walk.NewTableViewWithStyle(builder.Parent(), win.LVS_NOSORTHEADER)
	} else {
		w, err = walk.NewTableView(builder.Parent())
	}
	if err != nil {
		return err
	}

	return builder.InitWidget(tv, w, func() error {
		for i := range tv.Columns {
			if err := tv.Columns[i].Create(w); err != nil {
				return err
			}
		}

		if err := w.SetModel(tv.Model); err != nil {
			return err
		}

		if tv.AlternatingRowBGColor != 0 {
			w.SetAlternatingRowBGColor(tv.AlternatingRowBGColor)
		}
		w.SetCheckBoxes(tv.CheckBoxes)
		w.SetItemStateChangedEventDelay(tv.ItemStateChangedEventDelay)
		if err := w.SetLastColumnStretched(tv.LastColumnStretched); err != nil {
			return err
		}
		if err := w.SetMultiSelection(tv.MultiSelection); err != nil {
			return err
		}

		if tv.OnCurrentIndexChanged != nil {
			w.CurrentIndexChanged().Attach(tv.OnCurrentIndexChanged)
		}
		if tv.OnSelectedIndexesChanged != nil {
			w.SelectedIndexesChanged().Attach(tv.OnSelectedIndexesChanged)
		}
		if tv.OnItemActivated != nil {
			w.ItemActivated().Attach(tv.OnItemActivated)
		}

		if tv.AssignTo != nil {
			*tv.AssignTo = w
		}

		return nil
	})
}

func (w TableView) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
