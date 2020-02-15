// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type RadioButtonGroupBox struct {
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

	// Container

	Children   []Widget
	DataBinder DataBinder
	Layout     Layout

	// GroupBox

	AssignTo  **walk.GroupBox
	Checkable bool
	Checked   Property
	Title     string

	// RadioButtonGroupBox

	Buttons    []RadioButton
	DataMember string
	Optional   bool
}

func (rbgb RadioButtonGroupBox) Create(builder *Builder) error {
	w, err := walk.NewGroupBox(builder.Parent())
	if err != nil {
		return err
	}

	if rbgb.AssignTo != nil {
		*rbgb.AssignTo = w
	}

	w.SetSuspended(true)
	builder.Defer(func() error {
		w.SetSuspended(false)
		return nil
	})

	return builder.InitWidget(rbgb, w, func() error {
		if err := w.SetTitle(rbgb.Title); err != nil {
			return err
		}

		w.SetCheckable(rbgb.Checkable)

		if err := (RadioButtonGroup{
			DataMember: rbgb.DataMember,
			Optional:   rbgb.Optional,
			Buttons:    rbgb.Buttons,
		}).Create(builder); err != nil {
			return err
		}

		return nil
	})
}
