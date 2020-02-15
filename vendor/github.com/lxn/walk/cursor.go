// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"image"
)

import (
	"github.com/lxn/win"
)

type Cursor interface {
	Dispose()
	handle() win.HCURSOR
}

type stockCursor struct {
	hCursor win.HCURSOR
}

func (sc stockCursor) Dispose() {
	// nop
}

func (sc stockCursor) handle() win.HCURSOR {
	return sc.hCursor
}

func CursorArrow() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))}
}

func CursorIBeam() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_IBEAM))}
}

func CursorWait() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_WAIT))}
}

func CursorCross() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_CROSS))}
}

func CursorUpArrow() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_UPARROW))}
}

func CursorSizeNWSE() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_SIZENWSE))}
}

func CursorSizeNESW() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_SIZENESW))}
}

func CursorSizeWE() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_SIZEWE))}
}

func CursorSizeNS() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_SIZENS))}
}

func CursorSizeAll() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_SIZEALL))}
}

func CursorNo() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_NO))}
}

func CursorHand() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_HAND))}
}

func CursorAppStarting() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_APPSTARTING))}
}

func CursorHelp() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_HELP))}
}

func CursorIcon() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ICON))}
}

func CursorSize() Cursor {
	return stockCursor{win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_SIZE))}
}

type customCursor struct {
	hCursor win.HCURSOR
}

func NewCursorFromImage(im image.Image, hotspot image.Point) (Cursor, error) {
	i, err := createAlphaCursorOrIconFromImage(im, hotspot, false)
	if err != nil {
		return nil, err
	}
	return customCursor{win.HCURSOR(i)}, nil
}

func (cc customCursor) Dispose() {
	win.DestroyIcon(win.HICON(cc.hCursor))
}

func (cc customCursor) handle() win.HCURSOR {
	return cc.hCursor
}
