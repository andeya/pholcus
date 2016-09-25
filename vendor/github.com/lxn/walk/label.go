// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type Label struct {
	WidgetBase
	textChangedPublisher EventPublisher
}

func NewLabel(parent Container) (*Label, error) {
	l := new(Label)

	if err := InitWidget(
		l,
		parent,
		"STATIC",
		win.WS_VISIBLE|win.SS_CENTERIMAGE,
		0); err != nil {
		return nil, err
	}

	l.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return l.Text()
		},
		func(v interface{}) error {
			return l.SetText(v.(string))
		},
		l.textChangedPublisher.Event()))

	return l, nil
}

func (*Label) LayoutFlags() LayoutFlags {
	return GrowableVert
}

func (l *Label) MinSizeHint() Size {
	return l.calculateTextSize()
}

func (l *Label) SizeHint() Size {
	return l.MinSizeHint()
}

func (l *Label) Text() string {
	return windowText(l.hWnd)
}

func (l *Label) SetText(value string) error {
	if value == l.Text() {
		return nil
	}

	if err := setWindowText(l.hWnd, value); err != nil {
		return err
	}

	return l.updateParentLayout()
}

func (l *Label) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_SETTEXT:
		l.textChangedPublisher.Publish()

	case win.WM_SIZE, win.WM_SIZING:
		l.Invalidate()
	}

	return l.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
