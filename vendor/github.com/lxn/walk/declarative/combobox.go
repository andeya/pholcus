// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"errors"
)

import (
	"github.com/lxn/walk"
)

type ComboBox struct {
	AssignTo              **walk.ComboBox
	Name                  string
	Enabled               Property
	Visible               Property
	Font                  Font
	ToolTipText           Property
	MinSize               Size
	MaxSize               Size
	StretchFactor         int
	Row                   int
	RowSpan               int
	Column                int
	ColumnSpan            int
	AlwaysConsumeSpace    bool
	ContextMenuItems      []MenuItem
	OnKeyDown             walk.KeyEventHandler
	OnKeyUp               walk.KeyEventHandler
	OnMouseDown           walk.MouseEventHandler
	OnKeyPress            walk.KeyEventHandler
	OnMouseMove           walk.MouseEventHandler
	OnMouseUp             walk.MouseEventHandler
	OnSizeChanged         walk.EventHandler
	Editable              bool
	Format                string
	Precision             int
	MaxLength             int
	BindingMember         string
	DisplayMember         string
	Model                 interface{}
	Value                 Property
	CurrentIndex          Property
	OnCurrentIndexChanged walk.EventHandler
}

func (cb ComboBox) Create(builder *Builder) error {
	if _, ok := cb.Model.([]string); ok &&
		(cb.BindingMember != "" || cb.DisplayMember != "") {

		return errors.New("ComboBox.Create: BindingMember and DisplayMember must be empty for []string models.")
	}

	var w *walk.ComboBox
	var err error
	if cb.Editable {
		w, err = walk.NewComboBox(builder.Parent())
	} else {
		w, err = walk.NewDropDownBox(builder.Parent())
	}
	if err != nil {
		return err
	}

	return builder.InitWidget(cb, w, func() error {
		w.SetFormat(cb.Format)
		w.SetPrecision(cb.Precision)
		w.SetMaxLength(cb.MaxLength)

		if err := w.SetBindingMember(cb.BindingMember); err != nil {
			return err
		}
		if err := w.SetDisplayMember(cb.DisplayMember); err != nil {
			return err
		}

		if err := w.SetModel(cb.Model); err != nil {
			return err
		}

		if cb.OnCurrentIndexChanged != nil {
			w.CurrentIndexChanged().Attach(cb.OnCurrentIndexChanged)
		}

		if cb.AssignTo != nil {
			*cb.AssignTo = w
		}

		return nil
	})
}

func (w ComboBox) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
