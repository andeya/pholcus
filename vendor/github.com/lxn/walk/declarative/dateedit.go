// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"time"
)

import (
	"github.com/lxn/walk"
)

type DateEdit struct {
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

	// DateEdit

	AssignTo      **walk.DateEdit
	Date          Property
	Format        string
	MaxDate       time.Time
	MinDate       time.Time
	NoneOption    bool // Deprecated: use Optional instead
	OnDateChanged walk.EventHandler
	Optional      bool
}

func (de DateEdit) Create(builder *Builder) error {
	var w *walk.DateEdit
	var err error

	if de.Optional || de.NoneOption {
		w, err = walk.NewDateEditWithNoneOption(builder.Parent())
	} else {
		w, err = walk.NewDateEdit(builder.Parent())
	}
	if err != nil {
		return err
	}

	if de.AssignTo != nil {
		*de.AssignTo = w
	}

	return builder.InitWidget(de, w, func() error {
		if err := w.SetFormat(de.Format); err != nil {
			return err
		}

		if err := w.SetRange(de.MinDate, de.MaxDate); err != nil {
			return err
		}

		if de.OnDateChanged != nil {
			w.DateChanged().Attach(de.OnDateChanged)
		}

		return nil
	})
}
