// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type Composite struct {
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

	// Composite

	AssignTo    **walk.Composite
	Border      bool
	Expressions func() map[string]walk.Expression
	Functions   map[string]func(args ...interface{}) (interface{}, error)
}

func (c Composite) Create(builder *Builder) error {
	var style uint32
	if c.Border {
		style |= win.WS_BORDER
	}
	w, err := walk.NewCompositeWithStyle(builder.Parent(), style)
	if err != nil {
		return err
	}

	if c.AssignTo != nil {
		*c.AssignTo = w
	}

	w.SetSuspended(true)
	builder.Defer(func() error {
		w.SetSuspended(false)
		return nil
	})

	return builder.InitWidget(c, w, func() error {
		if c.Expressions != nil {
			for name, expr := range c.Expressions() {
				builder.expressions[name] = expr
			}
		}
		if c.Functions != nil {
			for name, fn := range c.Functions {
				builder.functions[name] = fn
			}
		}

		return nil
	})
}
