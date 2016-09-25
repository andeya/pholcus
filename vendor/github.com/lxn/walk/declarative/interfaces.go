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

type Layout interface {
	Create() (walk.Layout, error)
}

type Widget interface {
	Create(builder *Builder) error
	WidgetInfo() (name string, disabled, hidden bool, font *Font, toolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler)
}

type Container interface {
	ContainerInfo() (DataBinder, Layout, []Widget)
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

type topLevelWindowInfo struct {
	Name             string
	Disabled         bool
	Hidden           bool
	Font             Font
	ToolTipText      string
	MinSize          Size
	MaxSize          Size
	ContextMenuItems []MenuItem
	OnKeyDown        walk.KeyEventHandler
	OnKeyPress       walk.KeyEventHandler
	OnKeyUp          walk.KeyEventHandler
	OnMouseDown      walk.MouseEventHandler
	OnMouseMove      walk.MouseEventHandler
	OnMouseUp        walk.MouseEventHandler
	OnSizeChanged    walk.EventHandler
	DataBinder       DataBinder
	Layout           Layout
	Children         []Widget
}

func (topLevelWindowInfo) Create(builder *Builder) error {
	return nil
}

func (i topLevelWindowInfo) WidgetInfo() (name string, disabled, hidden bool, font *Font, ToolTipText string, minSize, maxSize Size, stretchFactor, row, rowSpan, column, columnSpan int, alwaysConsumeSpace bool, contextMenuItems []MenuItem, OnKeyDown walk.KeyEventHandler, OnKeyPress walk.KeyEventHandler, OnKeyUp walk.KeyEventHandler, OnMouseDown walk.MouseEventHandler, OnMouseMove walk.MouseEventHandler, OnMouseUp walk.MouseEventHandler, OnSizeChanged walk.EventHandler) {
	return i.Name, i.Disabled, i.Hidden, &i.Font, i.ToolTipText, i.MinSize, i.MaxSize, 0, 0, 0, 0, 0, false, i.ContextMenuItems, i.OnKeyDown, i.OnKeyPress, i.OnKeyUp, i.OnMouseDown, i.OnMouseMove, i.OnMouseUp, i.OnSizeChanged
}

func (i topLevelWindowInfo) ContainerInfo() (DataBinder, Layout, []Widget) {
	return i.DataBinder, i.Layout, i.Children
}
