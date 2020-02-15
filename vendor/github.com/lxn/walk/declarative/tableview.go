// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type TableView struct {
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

	// TableView

	AlternatingRowBGColor       walk.Color
	AssignTo                    **walk.TableView
	CellStyler                  walk.CellStyler
	CheckBoxes                  bool
	Columns                     []TableViewColumn
	ColumnsOrderable            Property
	ColumnsSizable              Property
	CustomHeaderHeight          int
	CustomRowHeight             int
	ItemStateChangedEventDelay  int
	HeaderHidden                bool
	LastColumnStretched         bool
	Model                       interface{}
	MultiSelection              bool
	NotSortableByHeaderClick    bool
	OnCurrentIndexChanged       walk.EventHandler
	OnItemActivated             walk.EventHandler
	OnSelectedIndexesChanged    walk.EventHandler
	SelectionHiddenWithoutFocus bool
	StyleCell                   func(style *walk.CellStyle)
}

type tvStyler struct {
	dflt              walk.CellStyler
	colStyleCellFuncs []func(style *walk.CellStyle)
}

func (tvs *tvStyler) StyleCell(style *walk.CellStyle) {
	if tvs.dflt != nil {
		tvs.dflt.StyleCell(style)
	}

	if styleCell := tvs.colStyleCellFuncs[style.Col()]; styleCell != nil {
		styleCell(style)
	}
}

type styleCellFunc func(style *walk.CellStyle)

func (scf styleCellFunc) StyleCell(style *walk.CellStyle) {
	scf(style)
}

func (tv TableView) Create(builder *Builder) error {
	var w *walk.TableView
	var err error
	if tv.NotSortableByHeaderClick {
		w, err = walk.NewTableViewWithStyle(builder.Parent(), win.LVS_NOSORTHEADER)
	} else {
		w, err = walk.NewTableViewWithCfg(builder.Parent(), &walk.TableViewCfg{CustomHeaderHeight: tv.CustomHeaderHeight, CustomRowHeight: tv.CustomRowHeight})
	}
	if err != nil {
		return err
	}

	if tv.AssignTo != nil {
		*tv.AssignTo = w
	}

	return builder.InitWidget(tv, w, func() error {
		for i := range tv.Columns {
			if err := tv.Columns[i].Create(w); err != nil {
				return err
			}
		}

		if err := w.SetModel(tv.Model); err != nil {
			return err
		}

		defaultStyler, _ := tv.Model.(walk.CellStyler)

		if tv.CellStyler != nil {
			defaultStyler = tv.CellStyler
		}

		if tv.StyleCell != nil {
			defaultStyler = styleCellFunc(tv.StyleCell)
		}

		var hasColStyleFunc bool
		for _, c := range tv.Columns {
			if c.StyleCell != nil {
				hasColStyleFunc = true
				break
			}
		}

		if defaultStyler != nil || hasColStyleFunc {
			var styler walk.CellStyler

			if hasColStyleFunc {
				tvs := &tvStyler{
					dflt:              defaultStyler,
					colStyleCellFuncs: make([]func(style *walk.CellStyle), len(tv.Columns)),
				}

				styler = tvs

				for i, c := range tv.Columns {
					tvs.colStyleCellFuncs[i] = c.StyleCell
				}
			} else {
				styler = defaultStyler
			}

			w.SetCellStyler(styler)
		}

		if tv.AlternatingRowBGColor != 0 {
			w.SetAlternatingRowBGColor(tv.AlternatingRowBGColor)
		}
		w.SetCheckBoxes(tv.CheckBoxes)
		w.SetItemStateChangedEventDelay(tv.ItemStateChangedEventDelay)
		if err := w.SetLastColumnStretched(tv.LastColumnStretched); err != nil {
			return err
		}
		if err := w.SetMultiSelection(tv.MultiSelection); err != nil {
			return err
		}
		if err := w.SetSelectionHiddenWithoutFocus(tv.SelectionHiddenWithoutFocus); err != nil {
			return err
		}
		if err := w.SetHeaderHidden(tv.HeaderHidden); err != nil {
			return err
		}

		if tv.OnCurrentIndexChanged != nil {
			w.CurrentIndexChanged().Attach(tv.OnCurrentIndexChanged)
		}
		if tv.OnSelectedIndexesChanged != nil {
			w.SelectedIndexesChanged().Attach(tv.OnSelectedIndexesChanged)
		}
		if tv.OnItemActivated != nil {
			w.ItemActivated().Attach(tv.OnItemActivated)
		}

		return nil
	})
}
