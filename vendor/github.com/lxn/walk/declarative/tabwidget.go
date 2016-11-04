// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type TabWidget struct {
	AssignTo              **walk.TabWidget
	Name                  string
	Enabled               Property
	Visible               Property
	Font                  Font
	ToolTipText           Property
	MinSize               Size
	MaxSize               Size
	StretchFactor         int
	Row                   int
	RowSpan               int
	Column                int
	ColumnSpan            int
	AlwaysConsumeSpace    bool
	ContextMenuItems      []MenuItem
	OnKeyDown             walk.KeyEventHandler
	OnKeyPress            walk.KeyEventHandler
	OnKeyUp               walk.KeyEventHandler
	OnMouseDown           walk.MouseEventHandler
	OnMouseMove           walk.MouseEventHandler
	OnMouseUp             walk.MouseEventHandler
	OnSizeChanged         walk.EventHandler
	ContentMargins        Margins
	ContentMarginsZero    bool
	Pages                 []TabPage
	OnCurrentIndexChanged walk.EventHandler
}

func (tw TabWidget) Create(builder *Builder) error {
	w, err := walk.NewTabWidget(builder.Parent())
	if err != nil {
		return err
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

		if tw.AssignTo != nil {
			*tw.AssignTo = w
		}

		return nil
	})
}

func (w TabWidget) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
