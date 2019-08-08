// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type HSeparator struct {
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

	// Separator

	AssignTo **walk.Separator
}

func (s HSeparator) Create(builder *Builder) error {
	w, err := walk.NewHSeparator(builder.Parent())
	if err != nil {
		return err
	}

	if s.AssignTo != nil {
		*s.AssignTo = w
	}

	return builder.InitWidget(s, w, func() error {
		return nil
	})
}

type VSeparator struct {
	// Window

	ContextMenuItems []MenuItem
	Enabled          Property
	Font             Font
	MaxSize          Size
	MinSize          Size
	Name             string
	OnBoundsChanged  walk.EventHandler
	OnKeyDown        walk.KeyEventHandler
	OnKeyPress       walk.KeyEventHandler
	OnKeyUp          walk.KeyEventHandler
	OnMouseDown      walk.MouseEventHandler
	OnMouseMove      walk.MouseEventHandler
	OnMouseUp        walk.MouseEventHandler
	OnSizeChanged    walk.EventHandler
	Persistent       bool
	ToolTipText      Property
	Visible          Property

	// Widget

	AlwaysConsumeSpace bool
	Column             int
	ColumnSpan         int
	Row                int
	RowSpan            int
	StretchFactor      int

	// Separator

	AssignTo **walk.Separator
}

func (s VSeparator) Create(builder *Builder) error {
	w, err := walk.NewVSeparator(builder.Parent())
	if err != nil {
		return err
	}

	if s.AssignTo != nil {
		*s.AssignTo = w
	}

	return builder.InitWidget(s, w, func() error {
		return nil
	})
}
