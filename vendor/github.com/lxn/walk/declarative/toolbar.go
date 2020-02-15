// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type ToolBarButtonStyle int

const (
	ToolBarButtonImageOnly ToolBarButtonStyle = iota
	ToolBarButtonTextOnly
	ToolBarButtonImageBeforeText
	ToolBarButtonImageAboveText
)

type ToolBar struct {
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

	// ToolBar

	Actions     []*walk.Action // Deprecated, use Items instead
	AssignTo    **walk.ToolBar
	ButtonStyle ToolBarButtonStyle
	Items       []MenuItem
	MaxTextRows int
	Orientation Orientation
}

func (tb ToolBar) Create(builder *Builder) error {
	w, err := walk.NewToolBarWithOrientationAndButtonStyle(builder.Parent(), walk.Orientation(tb.Orientation), walk.ToolBarButtonStyle(tb.ButtonStyle))
	if err != nil {
		return err
	}

	if tb.AssignTo != nil {
		*tb.AssignTo = w
	}

	return builder.InitWidget(tb, w, func() error {
		mtr := tb.MaxTextRows
		if mtr < 1 {
			mtr = 1
		}
		if err := w.SetMaxTextRows(mtr); err != nil {
			return err
		}

		if len(tb.Items) > 0 {
			builder.deferBuildActions(w.Actions(), tb.Items)
		} else {
			if err := addToActionList(w.Actions(), tb.Actions); err != nil {
				return err
			}
		}

		return nil
	})
}
