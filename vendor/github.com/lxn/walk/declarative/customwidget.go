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
	AssignTo            **walk.CustomWidget
	Name                string
	Enabled             Property
	Visible             Property
	Font                Font
	ToolTipText         Property
	MinSize             Size
	MaxSize             Size
	StretchFactor       int
	Row                 int
	RowSpan             int
	Column              int
	ColumnSpan          int
	AlwaysConsumeSpace  bool
	ContextMenuItems    []MenuItem
	OnKeyDown           walk.KeyEventHandler
	OnKeyPress          walk.KeyEventHandler
	OnKeyUp             walk.KeyEventHandler
	OnMouseDown         walk.MouseEventHandler
	OnMouseMove         walk.MouseEventHandler
	OnMouseUp           walk.MouseEventHandler
	OnSizeChanged       walk.EventHandler
	Style               uint32
	Paint               walk.PaintFunc
	ClearsBackground    bool
	InvalidatesOnResize bool
	PaintMode           PaintMode
}

func (cw CustomWidget) Create(builder *Builder) error {
	w, err := walk.NewCustomWidget(builder.Parent(), uint(cw.Style), cw.Paint)
	if err != nil {
		return err
	}

	return builder.InitWidget(cw, w, func() error {
		if cw.PaintMode != PaintNormal && cw.ClearsBackground {
			panic("PaintMode and ClearsBackground are incompatible")
		}
		w.SetClearsBackground(cw.ClearsBackground)
		w.SetInvalidatesOnResize(cw.InvalidatesOnResize)
		w.SetPaintMode(walk.PaintMode(cw.PaintMode))

		if cw.AssignTo != nil {
			*cw.AssignTo = w
		}

		return nil
	})
}

func (w CustomWidget) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
