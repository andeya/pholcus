// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type CheckBox struct {
	// Window

	Background         Brush
	ContextMenuItems   []MenuItem
	DoubleBuffering    bool
	Enabled            Property
	Font               Font
	MaxSize            Size
	MinSize            Size
	Name               string
	OnBoundsChanged    walk.EventHandler
	OnKeyDown          walk.KeyEventHandler
	OnKeyPress         walk.KeyEventHandler
	OnKeyUp            walk.KeyEventHandler
	OnMouseDown        walk.MouseEventHandler
	OnMouseMove        walk.MouseEventHandler
	OnMouseUp          walk.MouseEventHandler
	OnSizeChanged      walk.EventHandler
	Persistent         bool
	RightToLeftReading bool
	ToolTipText        Property
	Visible            Property

	// Widget

	Alignment          Alignment2D
	AlwaysConsumeSpace bool
	Column             int
	ColumnSpan         int
	GraphicsEffects    []walk.WidgetGraphicsEffect
	Row                int
	RowSpan            int
	StretchFactor      int

	// Button

	Checked          Property
	OnCheckedChanged walk.EventHandler
	OnClicked        walk.EventHandler
	Text             Property

	// CheckBox

	AssignTo            **walk.CheckBox
	CheckState          Property
	OnCheckStateChanged walk.EventHandler
	TextOnLeftSide      bool
	Tristate            bool
}

func (cb CheckBox) Create(builder *Builder) error {
	w, err := walk.NewCheckBox(builder.Parent())
	if err != nil {
		return err
	}

	if cb.AssignTo != nil {
		*cb.AssignTo = w
	}

	return builder.InitWidget(cb, w, func() error {
		w.SetPersistent(cb.Persistent)

		if err := w.SetTextOnLeftSide(cb.TextOnLeftSide); err != nil {
			return err
		}

		if err := w.SetTristate(cb.Tristate); err != nil {
			return err
		}

		if cb.Tristate && cb.CheckState == nil {
			w.SetCheckState(walk.CheckIndeterminate)
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

		return nil
	})
}
