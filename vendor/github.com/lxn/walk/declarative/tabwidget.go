// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type TabWidget struct {
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

	// TabWidget

	AssignTo              **walk.TabWidget
	ContentMargins        Margins
	ContentMarginsZero    bool
	OnCurrentIndexChanged walk.EventHandler
	Pages                 []TabPage
}

func (tw TabWidget) Create(builder *Builder) error {
	w, err := walk.NewTabWidget(builder.Parent())
	if err != nil {
		return err
	}

	if tw.AssignTo != nil {
		*tw.AssignTo = w
	}

	return builder.InitWidget(tw, w, func() error {
		for _, tp := range tw.Pages {
			var wp *walk.TabPage
			if tp.AssignTo == nil {
				tp.AssignTo = &wp
			}

			if tp.Content != nil && len(tp.Children) == 0 {
				tp.Layout = HBox{Margins: tw.ContentMargins, MarginsZero: tw.ContentMarginsZero}
			}

			if err := tp.Create(builder); err != nil {
				return err
			}

			if err := w.Pages().Add(*tp.AssignTo); err != nil {
				return err
			}
		}

		if tw.OnCurrentIndexChanged != nil {
			w.CurrentIndexChanged().Attach(tw.OnCurrentIndexChanged)
		}

		return nil
	})
}
