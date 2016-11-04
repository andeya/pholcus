// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type ImageView struct {
	AssignTo           **walk.ImageView
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
	Image              walk.Image
}

func (iv ImageView) Create(builder *Builder) error {
	w, err := walk.NewImageView(builder.Parent())
	if err != nil {
		return err
	}

	return builder.InitWidget(iv, w, func() error {
		if err := w.SetImage(iv.Image); err != nil {
			return err
		}

		if iv.AssignTo != nil {
			*iv.AssignTo = w
		}

		return nil
	})
}

func (w ImageView) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
