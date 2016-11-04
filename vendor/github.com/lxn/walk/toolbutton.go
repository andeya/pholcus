// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type ToolButton struct {
	Button
}

func NewToolButton(parent Container) (*ToolButton, error) {
	tb := new(ToolButton)

	if err := InitWidget(
		tb,
		parent,
		"BUTTON",
		win.WS_TABSTOP|win.WS_VISIBLE|win.BS_BITMAP|win.BS_PUSHBUTTON,
		0); err != nil {
		return nil, err
	}

	tb.Button.init()

	return tb, nil
}

func (*ToolButton) LayoutFlags() LayoutFlags {
	return 0
}

func (tb *ToolButton) MinSizeHint() Size {
	return tb.SizeHint()
}

func (tb *ToolButton) SizeHint() Size {
	return tb.dialogBaseUnitsToPixels(Size{16, 12})
}

func (tb *ToolButton) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_GETDLGCODE:
		return win.DLGC_BUTTON
	}

	return tb.Button.WndProc(hwnd, msg, wParam, lParam)
}
