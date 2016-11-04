// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type HSpacer struct {
	Name          string
	MinSize       Size
	MaxSize       Size
	StretchFactor int
	Row           int
	RowSpan       int
	Column        int
	ColumnSpan    int
	Size          int
}

func (hs HSpacer) Create(builder *Builder) (err error) {
	var w *walk.Spacer
	if hs.Size > 0 {
		if w, err = walk.NewHSpacerFixed(builder.Parent(), hs.Size); err != nil {
			return
		}
	} else {
		if w, err = walk.NewHSpacer(builder.Parent()); err != nil {
			return
		}
	}

	return builder.InitWidget(hs, w, nil)
}

func (hs HSpacer) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return hs.Name, false, false, nil, "", hs.MinSize, hs.MaxSize, hs.StretchFactor, hs.Row, hs.RowSpan, hs.Column, hs.ColumnSpan, false, nil, nil, nil, nil, nil, nil, nil, nil
}

type VSpacer struct {
	Name          string
	MinSize       Size
	MaxSize       Size
	StretchFactor int
	Row           int
	RowSpan       int
	Column        int
	ColumnSpan    int
	Size          int
}

func (vs VSpacer) Create(builder *Builder) (err error) {
	var w *walk.Spacer
	if vs.Size > 0 {
		if w, err = walk.NewVSpacerFixed(builder.Parent(), vs.Size); err != nil {
			return
		}
	} else {
		if w, err = walk.NewVSpacer(builder.Parent()); err != nil {
			return
		}
	}

	return builder.InitWidget(vs, w, nil)
}

func (vs VSpacer) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return vs.Name, false, false, nil, "", vs.MinSize, vs.MaxSize, vs.StretchFactor, vs.Row, vs.RowSpan, vs.Column, vs.ColumnSpan, false, nil, nil, nil, nil, nil, nil, nil, nil
}
