// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type LinkLabel struct {
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

	// LinkLabel

	AssignTo        **walk.LinkLabel
	OnLinkActivated walk.LinkLabelLinkEventHandler
	Text            Property
}

func (ll LinkLabel) Create(builder *Builder) error {
	w, err := walk.NewLinkLabel(builder.Parent())
	if err != nil {
		return err
	}

	if ll.AssignTo != nil {
		*ll.AssignTo = w
	}

	return builder.InitWidget(ll, w, func() error {
		if ll.OnLinkActivated != nil {
			w.LinkActivated().Attach(ll.OnLinkActivated)
		}

		return nil
	})
}
