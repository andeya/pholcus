// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

func tr(source string, context ...string) string {
	if translation := walk.TranslationFunc(); translation != nil {
		return translation(source, context...)
	}

	return source
}

type Property interface{}

type bindData struct {
	expression string
	validator  Validator
}

func Bind(expression string, validators ...Validator) Property {
	bd := bindData{expression: expression}
	switch len(validators) {
	case 0:
		// nop

	case 1:
		bd.validator = validators[0]

	default:
		bd.validator = dMultiValidator{validators}
	}

	return bd
}

type Brush interface {
	Create() (walk.Brush, error)
}

type Layout interface {
	Create() (walk.Layout, error)
}

type Widget interface {
	Create(builder *Builder) error
}

type MenuItem interface {
	createAction(builder *Builder, menu *walk.Menu) (*walk.Action, error)
}

type Validator interface {
	Create() (walk.Validator, error)
}

type ErrorPresenter interface {
	Create() (walk.ErrorPresenter, error)
}

type ErrorPresenterRef struct {
	ErrorPresenter *walk.ErrorPresenter
}

func (epr ErrorPresenterRef) Create() (walk.ErrorPresenter, error) {
	if epr.ErrorPresenter != nil {
		return *epr.ErrorPresenter, nil
	}

	return nil, nil
}

type ToolTipErrorPresenter struct {
}

func (ToolTipErrorPresenter) Create() (walk.ErrorPresenter, error) {
	return walk.NewToolTipErrorPresenter()
}

type formInfo struct {
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
	RightToLeftReading bool
	ToolTipText        string
	Visible            Property

	// Container

	Children   []Widget
	DataBinder DataBinder
	Layout     Layout

	// Form

	Icon  Property
	Title Property
}

func (formInfo) Create(builder *Builder) error {
	return nil
}
