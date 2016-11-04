// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type CheckBox struct {
	AssignTo            **walk.CheckBox
	Name                string
	Enabled             Property
	Visible             Property
	Font                Font
	ToolTipText         Property
	MinSize             Size
	MaxSize             Size
	StretchFactor       int
	Row                 int
	RowSpan             int
	Column              int
	ColumnSpan          int
	AlwaysConsumeSpace  bool
	ContextMenuItems    []MenuItem
	OnKeyDown           walk.KeyEventHandler
	OnKeyPress          walk.KeyEventHandler
	OnKeyUp             walk.KeyEventHandler
	OnMouseDown         walk.MouseEventHandler
	OnMouseMove         walk.MouseEventHandler
	OnMouseUp           walk.MouseEventHandler
	OnSizeChanged       walk.EventHandler
	Text                Property
	Checked             Property
	CheckState          Property
	Tristate            bool
	OnClicked           walk.EventHandler
	OnCheckedChanged    walk.EventHandler
	OnCheckStateChanged walk.EventHandler
}

func (cb CheckBox) Create(builder *Builder) error {
	w, err := walk.NewCheckBox(builder.Parent())
	if err != nil {
		return err
	}

	return builder.InitWidget(cb, w, func() error {
		if err := w.SetTristate(cb.Tristate); err != nil {
			return err
		}

		if cb.OnClicked != nil {
			w.Clicked().Attach(cb.OnClicked)
		}

		if cb.OnCheckedChanged != nil {
			w.CheckedChanged().Attach(cb.OnCheckedChanged)
		}

		if cb.OnCheckStateChanged != nil {
			w.CheckStateChanged().Attach(cb.OnCheckStateChanged)
		}

		if cb.AssignTo != nil {
			*cb.AssignTo = w
		}

		return nil
	})
}

func (w CheckBox) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
