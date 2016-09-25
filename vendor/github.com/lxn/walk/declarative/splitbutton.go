// Copyright 2016 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type SplitButton struct {
	AssignTo           **walk.SplitButton
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
	Text               Property
	Image              interface{}
	ImageAboveText     bool
	MenuItems          []MenuItem
	OnClicked          walk.EventHandler
}

func (sb SplitButton) Create(builder *Builder) error {
	w, err := walk.NewSplitButton(builder.Parent())
	if err != nil {
		return err
	}

	builder.deferBuildMenuActions(w.Menu(), sb.MenuItems)

	return builder.InitWidget(sb, w, func() error {
		img := sb.Image
		if s, ok := img.(string); ok {
			var err error
			if img, err = imageFromFile(s); err != nil {
				return err
			}
		}
		if img != nil {
			if err := w.SetImage(img.(walk.Image)); err != nil {
				return err
			}
		}

		if err := w.SetImageAboveText(sb.ImageAboveText); err != nil {
			return err
		}

		if sb.OnClicked != nil {
			w.Clicked().Attach(sb.OnClicked)
		}

		if sb.AssignTo != nil {
			*sb.AssignTo = w
		}

		return nil
	})
}

func (w SplitButton) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
