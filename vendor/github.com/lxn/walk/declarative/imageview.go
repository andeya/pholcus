// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type ImageViewMode int

const (
	ImageViewModeIdeal   = ImageViewMode(walk.ImageViewModeIdeal)
	ImageViewModeCorner  = ImageViewMode(walk.ImageViewModeCorner)
	ImageViewModeCenter  = ImageViewMode(walk.ImageViewModeCenter)
	ImageViewModeShrink  = ImageViewMode(walk.ImageViewModeShrink)
	ImageViewModeZoom    = ImageViewMode(walk.ImageViewModeZoom)
	ImageViewModeStretch = ImageViewMode(walk.ImageViewModeStretch)
)

type ImageView struct {
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

	// ImageView

	AssignTo **walk.ImageView
	Image    Property
	Margin   Property
	Mode     ImageViewMode
}

func (iv ImageView) Create(builder *Builder) error {
	w, err := walk.NewImageView(builder.Parent())
	if err != nil {
		return err
	}

	if iv.AssignTo != nil {
		*iv.AssignTo = w
	}

	return builder.InitWidget(iv, w, func() error {
		w.SetMode(walk.ImageViewMode(iv.Mode))

		return nil
	})
}
