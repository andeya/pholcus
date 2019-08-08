// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type ToolButton struct {
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

	Image     Property
	OnClicked walk.EventHandler
	Text      Property

	// ToolButton

	AssignTo **walk.ToolButton
}

func (tb ToolButton) Create(builder *Builder) error {
	w, err := walk.NewToolButton(builder.Parent())
	if err != nil {
		return err
	}

	if tb.AssignTo != nil {
		*tb.AssignTo = w
	}

	return builder.InitWidget(tb, w, func() error {
		if tb.OnClicked != nil {
			w.Clicked().Attach(tb.OnClicked)
		}

		return nil
	})
}
