// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type CaseMode uint32

const (
	CaseModeMixed CaseMode = CaseMode(walk.CaseModeMixed)
	CaseModeUpper CaseMode = CaseMode(walk.CaseModeUpper)
	CaseModeLower CaseMode = CaseMode(walk.CaseModeLower)
)

type LineEdit struct {
	AssignTo           **walk.LineEdit
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
	Text               Property
	ReadOnly           Property
	CueBanner          string
	MaxLength          int
	PasswordMode       bool
	OnEditingFinished  walk.EventHandler
	OnTextChanged      walk.EventHandler
	CaseMode           CaseMode
}

func (le LineEdit) Create(builder *Builder) error {
	w, err := walk.NewLineEdit(builder.Parent())
	if err != nil {
		return err
	}

	return builder.InitWidget(le, w, func() error {
		if le.CueBanner != "" {
			if err := w.SetCueBanner(le.CueBanner); err != nil {
				return err
			}
		}
		w.SetMaxLength(le.MaxLength)
		w.SetPasswordMode(le.PasswordMode)

		if err := w.SetCaseMode(walk.CaseMode(le.CaseMode)); err != nil {
			return err
		}

		if le.OnEditingFinished != nil {
			w.EditingFinished().Attach(le.OnEditingFinished)
		}
		if le.OnTextChanged != nil {
			w.TextChanged().Attach(le.OnTextChanged)
		}

		if le.AssignTo != nil {
			*le.AssignTo = w
		}

		return nil
	})
}

func (w LineEdit) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
