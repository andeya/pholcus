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
	AssignTo           **walk.ToolBar
	Name               string
	Enabled            Property
	Visible            Property
	Font               Font
	ToolTipText        Property
	MinSize            Size
	MaxSize            Size
	StretchFactor      int
	Row                int
	RowSpan            int
	Column             int
	ColumnSpan         int
	AlwaysConsumeSpace bool
	ContextMenuItems   []MenuItem
	OnKeyDown          walk.KeyEventHandler
	OnKeyPress         walk.KeyEventHandler
	OnKeyUp            walk.KeyEventHandler
	OnMouseDown        walk.MouseEventHandler
	OnMouseMove        walk.MouseEventHandler
	OnMouseUp          walk.MouseEventHandler
	OnSizeChanged      walk.EventHandler
	Actions            []*walk.Action // Deprecated, use Items instead
	Items              []MenuItem
	MaxTextRows        int
	Orientation        Orientation
	ButtonStyle        ToolBarButtonStyle
}

func (tb ToolBar) Create(builder *Builder) error {
	w, err := walk.NewToolBarWithOrientationAndButtonStyle(builder.Parent(), walk.Orientation(tb.Orientation), walk.ToolBarButtonStyle(tb.ButtonStyle))
	if err != nil {
		return err
	}

	return builder.InitWidget(tb, w, func() error {
		imageList, err := walk.NewImageList(walk.Size{16, 16}, 0)
		if err != nil {
			return err
		}
		w.SetImageList(imageList)

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

		if tb.AssignTo != nil {
			*tb.AssignTo = w
		}

		return nil
	})
}

func (w ToolBar) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
