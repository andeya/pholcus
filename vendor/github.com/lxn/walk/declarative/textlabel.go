// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type Alignment2D uint

const (
	AlignHVDefault      = Alignment2D(walk.AlignHVDefault)
	AlignHNearVNear     = Alignment2D(walk.AlignHNearVNear)
	AlignHCenterVNear   = Alignment2D(walk.AlignHCenterVNear)
	AlignHFarVNear      = Alignment2D(walk.AlignHFarVNear)
	AlignHNearVCenter   = Alignment2D(walk.AlignHNearVCenter)
	AlignHCenterVCenter = Alignment2D(walk.AlignHCenterVCenter)
	AlignHFarVCenter    = Alignment2D(walk.AlignHFarVCenter)
	AlignHNearVFar      = Alignment2D(walk.AlignHNearVFar)
	AlignHCenterVFar    = Alignment2D(walk.AlignHCenterVFar)
	AlignHFarVFar       = Alignment2D(walk.AlignHFarVFar)
)

type TextLabel struct {
	// Window

	Background         Brush
	ContextMenuItems   []MenuItem
	DoubleBuffering    bool
	Enabled            Property
	Font               Font
	MaxSize            Size
	MinSize            Size // Set MinSize.Width to a value > 0 to enable dynamic line wrapping.
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

	// Text

	AssignTo      **walk.TextLabel
	TextAlignment Alignment2D
	Text          Property
}

func (tl TextLabel) Create(builder *Builder) error {
	w, err := walk.NewTextLabel(builder.Parent())
	if err != nil {
		return err
	}

	if tl.AssignTo != nil {
		*tl.AssignTo = w
	}

	return builder.InitWidget(tl, w, func() error {
		w.SetTextColor(tl.TextColor)

		if err := w.SetTextAlignment(walk.Alignment2D(tl.TextAlignment)); err != nil {
			return err
		}

		return nil
	})
}
