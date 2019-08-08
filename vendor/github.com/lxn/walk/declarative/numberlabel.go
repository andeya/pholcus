// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type NumberLabel struct {
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

	// static

	TextColor walk.Color

	// NumberLabel

	AssignTo      **walk.NumberLabel
	Decimals      Property
	Suffix        Property
	TextAlignment Alignment1D
	Value         Property
}

func (nl NumberLabel) Create(builder *Builder) error {
	w, err := walk.NewNumberLabel(builder.Parent())
	if err != nil {
		return err
	}

	if nl.AssignTo != nil {
		*nl.AssignTo = w
	}

	return builder.InitWidget(nl, w, func() error {
		if err := w.SetTextAlignment(walk.Alignment1D(nl.TextAlignment)); err != nil {
			return err
		}

		w.SetTextColor(nl.TextColor)

		return nil
	})
}
