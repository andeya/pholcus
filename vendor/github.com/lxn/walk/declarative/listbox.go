// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"errors"
)

import (
	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type ListBox struct {
	AssignTo                 **walk.ListBox
	Name                     string
	Enabled                  Property
	Visible                  Property
	Font                     Font
	ToolTipText              Property
	MinSize                  Size
	MaxSize                  Size
	StretchFactor            int
	Row                      int
	RowSpan                  int
	Column                   int
	ColumnSpan               int
	AlwaysConsumeSpace       bool
	ContextMenuItems         []MenuItem
	OnKeyDown                walk.KeyEventHandler
	OnKeyPress               walk.KeyEventHandler
	OnKeyUp                  walk.KeyEventHandler
	OnMouseDown              walk.MouseEventHandler
	OnMouseMove              walk.MouseEventHandler
	OnMouseUp                walk.MouseEventHandler
	OnSizeChanged            walk.EventHandler
	MultiSelection           bool
	Format                   string
	Precision                int
	DataMember               string
	Model                    interface{}
	OnCurrentIndexChanged    walk.EventHandler
	OnSelectedIndexesChanged walk.EventHandler
	OnItemActivated          walk.EventHandler
}

func (lb ListBox) Create(builder *Builder) error {
	var w *walk.ListBox
	var err error
	if _, ok := lb.Model.([]string); ok && lb.DataMember != "" {
		return errors.New("ListBox.Create: DataMember must be empty for []string models.")
	}

	if lb.MultiSelection {
		w, err = walk.NewListBoxWithStyle(builder.Parent(), win.LBS_EXTENDEDSEL)
	} else {
		w, err = walk.NewListBox(builder.Parent())
	}
	if err != nil {
		return err
	}

	return builder.InitWidget(lb, w, func() error {
		w.SetFormat(lb.Format)
		w.SetPrecision(lb.Precision)

		if err := w.SetDataMember(lb.DataMember); err != nil {
			return err
		}

		if err := w.SetModel(lb.Model); err != nil {
			return err
		}

		if lb.OnCurrentIndexChanged != nil {
			w.CurrentIndexChanged().Attach(lb.OnCurrentIndexChanged)
		}
		if lb.OnSelectedIndexesChanged != nil {
			w.SelectedIndexesChanged().Attach(lb.OnSelectedIndexesChanged)
		}
		if lb.OnItemActivated != nil {
			w.ItemActivated().Attach(lb.OnItemActivated)
		}

		if lb.AssignTo != nil {
			*lb.AssignTo = w
		}

		return nil
	})
}

func (w ListBox) WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return w.Name, false, false, &w.Font, "", w.MinSize, w.MaxSize, w.StretchFactor, w.Row, w.RowSpan, w.Column, w.ColumnSpan, w.AlwaysConsumeSpace, w.ContextMenuItems, w.OnKeyDown, w.OnKeyPress, w.OnKeyUp, w.OnMouseDown, w.OnMouseMove, w.OnMouseUp, w.OnSizeChanged
}
