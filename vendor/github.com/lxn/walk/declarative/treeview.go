// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type TreeView struct {
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

	// TreeView

	AssignTo             **walk.TreeView
	ItemHeight           int
	Model                walk.TreeModel
	OnCurrentItemChanged walk.EventHandler
	OnExpandedChanged    walk.TreeItemEventHandler
	OnItemActivated      walk.EventHandler
}

func (tv TreeView) Create(builder *Builder) error {
	w, err := walk.NewTreeView(builder.Parent())
	if err != nil {
		return err
	}

	if tv.AssignTo != nil {
		*tv.AssignTo = w
	}

	return builder.InitWidget(tv, w, func() error {
		if tv.ItemHeight > 0 {
			w.SetItemHeight(tv.ItemHeight)
		}

		if err := w.SetModel(tv.Model); err != nil {
			return err
		}

		if tv.OnCurrentItemChanged != nil {
			w.CurrentItemChanged().Attach(tv.OnCurrentItemChanged)
		}

		if tv.OnExpandedChanged != nil {
			w.ExpandedChanged().Attach(tv.OnExpandedChanged)
		}

		if tv.OnItemActivated != nil {
			w.ItemActivated().Attach(tv.OnItemActivated)
		}

		return nil
	})
}
