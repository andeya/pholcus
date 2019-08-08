// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

const splitterHandleWindowClass = `\o/ Walk_SplitterHandle_Class \o/`

func init() {
	MustRegisterWindowClass(splitterHandleWindowClass)
}

type splitterHandle struct {
	WidgetBase
}

func newSplitterHandle(splitter *Splitter) (*splitterHandle, error) {
	if splitter == nil {
		return nil, newError("splitter cannot be nil")
	}

	sh := new(splitterHandle)
	sh.parent = splitter

	if err := InitWindow(
		sh,
		splitter,
		splitterHandleWindowClass,
		win.WS_CHILD|win.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	sh.SetBackground(NullBrush())

	if err := sh.setAndClearStyleBits(0, win.WS_CLIPSIBLINGS); err != nil {
		return nil, err
	}

	return sh, nil
}

func (sh *splitterHandle) LayoutFlags() LayoutFlags {
	splitter, ok := sh.Parent().(*Splitter)
	if !ok {
		return 0
	}

	if splitter.Orientation() == Horizontal {
		return ShrinkableVert | GrowableVert | GreedyVert
	}

	return ShrinkableHorz | GrowableHorz | GreedyHorz
}

func (sh *splitterHandle) MinSizeHint() Size {
	return sh.SizeHint()
}

func (sh *splitterHandle) SizeHint() Size {
	splitter, ok := sh.Parent().(*Splitter)
	if !ok {
		return Size{}
	}

	handleWidth := splitter.HandleWidth()
	var size Size

	if splitter.Orientation() == Horizontal {
		size.Width = handleWidth
	} else {
		size.Height = handleWidth
	}

	return size
}

func (sh *splitterHandle) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_ERASEBKGND:
		if sh.Background() == nullBrushSingleton {
			return 1
		}

	case win.WM_PAINT:
		if sh.Background() == nullBrushSingleton {
			var ps win.PAINTSTRUCT

			win.BeginPaint(hwnd, &ps)
			defer win.EndPaint(hwnd, &ps)

			return 0
		}
	}

	return sh.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
