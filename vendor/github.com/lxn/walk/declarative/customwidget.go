// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type PaintMode int

const (
	PaintNormal   PaintMode = iota // erase background before PaintFunc
	PaintNoErase                   // PaintFunc clears background, single buffered
	PaintBuffered                  // PaintFunc clears background, double buffered
)

type CustomWidget struct {
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

	// CustomWidget

	AssignTo            **walk.CustomWidget
	ClearsBackground    bool
	InvalidatesOnResize bool
	Paint               walk.PaintFunc
	PaintMode           PaintMode
	Style               uint32
}

func (cw CustomWidget) Create(builder *Builder) error {
	w, err := walk.NewCustomWidget(builder.Parent(), uint(cw.Style), cw.Paint)
	if err != nil {
		return err
	}

	if cw.AssignTo != nil {
		*cw.AssignTo = w
	}

	return builder.InitWidget(cw, w, func() error {
		if cw.PaintMode != PaintNormal && cw.ClearsBackground {
			panic("PaintMode and ClearsBackground are incompatible")
		}
		w.SetClearsBackground(cw.ClearsBackground)
		w.SetInvalidatesOnResize(cw.InvalidatesOnResize)
		w.SetPaintMode(walk.PaintMode(cw.PaintMode))

		return nil
	})
}
