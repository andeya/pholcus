// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"time"
)

import (
	"github.com/lxn/walk"
)

type DateEdit struct {
	AssignTo           **walk.DateEdit
	Name               string
	Enabled            Property
	Visible            Property
	Font               Font
	ToolTipText        Property
	MinSize            Size
	MaxSize            Size
	StretchFactor      int
	Row                int
	RowSpan            int
	Column             int
	ColumnSpan         int
	AlwaysConsumeSpace bool
	ContextMenuItems   []MenuItem
	OnKeyDown          walk.KeyEventHandler
	OnKeyPress         walk.KeyEventHandler
	OnKeyUp            walk.KeyEventHandler
	OnMouseDown        walk.MouseEventHandler
	OnMouseMove        walk.MouseEventHandler
	OnMouseUp          walk.MouseEventHandler
	OnSizeChanged      walk.EventHandler
	Format             string
	NoneOption         bool
	MinDate            time.Time
	MaxDate            time.Time
	Date               Property
	OnDateChanged      walk.EventHandler
}

func (de DateEdit) Create(builder *Builder) error {
	var w *walk.DateEdit
	var err error

	if de.NoneOption {
		w, err = walk.NewDateEditWithNoneOption(builder.Parent())
	} else {
		w, err = walk.NewDateEdit(builder.Parent())
	}
	if err != nil {
		return err
	}

	return builder.InitWidget(de, w, func() error {
		if err := w.SetFormat(de.Format); err != nil {
			return err
		}

		if err := w.SetRange(de.MinDate, de.MaxDate); err != nil {
			return err
		}

		if de.OnDateChanged != nil {
			w.DateChanged().Attach(de.OnDateChanged)
		}

		if de.AssignTo != nil {
			*de.AssignTo = w
		}

		return nil
	})
}

func (w DateEdit) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
