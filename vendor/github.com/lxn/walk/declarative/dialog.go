// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type Dialog struct {
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
	RightToLeftLayout  bool
	RightToLeftReading bool
	ToolTipText        Property
	Visible            Property

	// Container

	DataBinder DataBinder
	Layout     Layout
	Children   []Widget

	// Form

	Expressions func() map[string]walk.Expression
	Functions   map[string]func(args ...interface{}) (interface{}, error)
	Icon        Property
	Title       Property
	Size        Size

	// Dialog

	AssignTo      **walk.Dialog
	CancelButton  **walk.PushButton
	DefaultButton **walk.PushButton
	FixedSize     bool
}

func (d Dialog) Create(owner walk.Form) error {
	var w *walk.Dialog
	var err error

	if d.FixedSize {
		w, err = walk.NewDialogWithFixedSize(owner)
	} else {
		w, err = walk.NewDialog(owner)
	}
	if err != nil {
		return err
	}

	if d.AssignTo != nil {
		*d.AssignTo = w
	}

	fi := formInfo{
		// Window
		Background:         d.Background,
		ContextMenuItems:   d.ContextMenuItems,
		DoubleBuffering:    d.DoubleBuffering,
		Enabled:            d.Enabled,
		Font:               d.Font,
		MaxSize:            d.MaxSize,
		MinSize:            d.MinSize,
		Name:               d.Name,
		OnBoundsChanged:    d.OnBoundsChanged,
		OnKeyDown:          d.OnKeyDown,
		OnKeyPress:         d.OnKeyPress,
		OnKeyUp:            d.OnKeyUp,
		OnMouseDown:        d.OnMouseDown,
		OnMouseMove:        d.OnMouseMove,
		OnMouseUp:          d.OnMouseUp,
		OnSizeChanged:      d.OnSizeChanged,
		RightToLeftReading: d.RightToLeftReading,
		ToolTipText:        "",
		Visible:            d.Visible,

		// Container
		Children:   d.Children,
		DataBinder: d.DataBinder,
		Layout:     d.Layout,

		// Form
		Icon:  d.Icon,
		Title: d.Title,
	}

	var db *walk.DataBinder
	if d.DataBinder.AssignTo == nil {
		d.DataBinder.AssignTo = &db
	}

	builder := NewBuilder(nil)

	w.SetSuspended(true)
	builder.Defer(func() error {
		w.SetSuspended(false)
		w.SetBoundsPixels(w.BoundsPixels())
		return nil
	})

	if err := w.SetRightToLeftLayout(d.RightToLeftLayout); err != nil {
		return err
	}

	return builder.InitWidget(fi, w, func() error {
		if err := w.SetSizePixels(d.Size.toW()); err != nil {
			return err
		}

		if d.DefaultButton != nil {
			if err := w.SetDefaultButton(*d.DefaultButton); err != nil {
				return err
			}

			if db := *d.DataBinder.AssignTo; db != nil {
				if db.DataSource() != nil {
					(*d.DefaultButton).SetEnabled(db.CanSubmit())
				}

				db.CanSubmitChanged().Attach(func() {
					(*d.DefaultButton).SetEnabled(db.CanSubmit())
				})
			}
		}
		if d.CancelButton != nil {
			if err := w.SetCancelButton(*d.CancelButton); err != nil {
				return err
			}
		}

		if d.Expressions != nil {
			for name, expr := range d.Expressions() {
				builder.expressions[name] = expr
			}
		}
		if d.Functions != nil {
			for name, fn := range d.Functions {
				builder.functions[name] = fn
			}
		}

		return nil
	})
}

func (d Dialog) Run(owner walk.Form) (int, error) {
	var w *walk.Dialog

	if d.AssignTo == nil {
		d.AssignTo = &w
	}

	if err := d.Create(owner); err != nil {
		return 0, err
	}

	return (*d.AssignTo).Run(), nil
}
