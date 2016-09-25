// Copyright 2016 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"unsafe"
)

import (
	"github.com/lxn/win"
)

type SplitButton struct {
	Button
	menu *Menu
}

func NewSplitButton(parent Container) (*SplitButton, error) {
	sb := new(SplitButton)

	menu, err := NewMenu()
	if err != nil {
		return nil, err
	}

	if err := InitWidget(
		sb,
		parent,
		"BUTTON",
		win.WS_TABSTOP|win.WS_VISIBLE|win.BS_SPLITBUTTON,
		0); err != nil {
		return nil, err
	}

	sb.Button.init()

	sb.menu = menu

	return sb, nil
}

func (sb *SplitButton) Dispose() {
	sb.Button.Dispose()

	sb.menu.Dispose()
}

func (*SplitButton) LayoutFlags() LayoutFlags {
	return GrowableHorz
}

func (sb *SplitButton) MinSizeHint() Size {
	var s win.SIZE

	sb.SendMessage(win.BCM_GETIDEALSIZE, 0, uintptr(unsafe.Pointer(&s)))

	return maxSize(Size{int(s.CX), int(s.CY)}, sb.dialogBaseUnitsToPixels(Size{50, 14}))
}

func (sb *SplitButton) SizeHint() Size {
	return sb.MinSizeHint()
}

func (sb *SplitButton) ImageAboveText() bool {
	return sb.hasStyleBits(win.BS_TOP)
}

func (sb *SplitButton) SetImageAboveText(value bool) error {
	if err := sb.ensureStyleBits(win.BS_TOP, value); err != nil {
		return err
	}

	// We need to set the image again, or Windows will fail to calculate the
	// button control size correctly.
	return sb.SetImage(sb.image)
}

func (sb *SplitButton) Menu() *Menu {
	return sb.menu
}

func (sb *SplitButton) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_NOTIFY:
		switch ((*win.NMHDR)(unsafe.Pointer(lParam))).Code {
		case win.BCN_DROPDOWN:
			dd := (*win.NMBCDROPDOWN)(unsafe.Pointer(lParam))

			p := win.POINT{dd.RcButton.Left, dd.RcButton.Bottom}

			win.ClientToScreen(sb.hWnd, &p)

			win.TrackPopupMenuEx(
				sb.menu.hMenu,
				win.TPM_NOANIMATION,
				p.X,
				p.Y,
				sb.hWnd,
				nil)
			return 0
		}
	}

	return sb.Button.WndProc(hwnd, msg, wParam, lParam)
}
