// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type ProgressBar struct {
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

	// ProgressBar

	AssignTo    **walk.ProgressBar
	MarqueeMode bool
	MaxValue    int
	MinValue    int
	Value       int
}

func (pb ProgressBar) Create(builder *Builder) error {
	w, err := walk.NewProgressBar(builder.Parent())
	if err != nil {
		return err
	}

	if pb.AssignTo != nil {
		*pb.AssignTo = w
	}

	return builder.InitWidget(pb, w, func() error {
		if pb.MaxValue > pb.MinValue {
			w.SetRange(pb.MinValue, pb.MaxValue)
		}
		w.SetValue(pb.Value)

		if err := w.SetMarqueeMode(pb.MarqueeMode); err != nil {
			return err
		}

		return nil
	})
}
