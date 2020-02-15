// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

const compositeWindowClass = `\o/ Walk_Composite_Class \o/`

func init() {
	MustRegisterWindowClass(compositeWindowClass)
}

type Composite struct {
	ContainerBase
}

func NewCompositeWithStyle(parent Window, style uint32) (*Composite, error) {
	c := new(Composite)
	c.children = newWidgetList(c)
	c.SetPersistent(true)

	if err := InitWidget(
		c,
		parent,
		compositeWindowClass,
		win.WS_CHILD|win.WS_VISIBLE|style,
		win.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	c.SetBackground(NullBrush())

	return c, nil
}

func NewComposite(parent Container) (*Composite, error) {
	return NewCompositeWithStyle(parent, 0)
}

func (c *Composite) onInsertedWidget(index int, widget Widget) (err error) {
	err = c.ContainerBase.onInsertedWidget(index, widget)

	c.ensureAppropriateParentScrollViewCompositeSize()

	return
}

func (c *Composite) onRemovedWidget(index int, widget Widget) (err error) {
	err = c.ContainerBase.onRemovedWidget(index, widget)

	c.ensureAppropriateParentScrollViewCompositeSize()

	return
}

func (c *Composite) onClearedWidgets() error {
	c.ensureAppropriateParentScrollViewCompositeSize()

	return c.ContainerBase.onClearedWidgets()
}

func (c *Composite) ensureAppropriateParentScrollViewCompositeSize() {
	if parent := c.Parent(); parent != nil {
		if sv, ok := parent.(*ScrollView); ok {
			sv.updateCompositeSize()
		}
	}
}
