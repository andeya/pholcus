// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

// see https://msdn.microsoft.com/en-us/library/windows/desktop/bb760416(v=vs.85).aspx
const maxToolTipTextLen = 80 // including NUL terminator

func init() {
	var err error
	if globalToolTip, err = NewToolTip(); err != nil {
		panic(err)
	}
}

type ToolTip struct {
	WindowBase
}

var globalToolTip *ToolTip

func NewToolTip() (*ToolTip, error) {
	tt := new(ToolTip)

	if err := InitWindow(
		tt,
		nil,
		"tooltips_class32",
		win.WS_POPUP|win.TTS_ALWAYSTIP,
		win.WS_EX_TOPMOST); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			tt.Dispose()
		}
	}()

	win.SetWindowPos(tt.hWnd, win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOACTIVATE)

	succeeded = true

	return tt, nil
}

func (*ToolTip) LayoutFlags() LayoutFlags {
	return 0
}

func (tt *ToolTip) SizeHint() Size {
	return Size{0, 0}
}

func (tt *ToolTip) Title() string {
	var gt win.TTGETTITLE

	buf := make([]uint16, 100)

	gt.DwSize = uint32(unsafe.Sizeof(gt))
	gt.Cch = uint32(len(buf))
	gt.PszTitle = &buf[0]

	tt.SendMessage(win.TTM_GETTITLE, 0, uintptr(unsafe.Pointer(&gt)))

	return syscall.UTF16ToString(buf)
}

func (tt *ToolTip) SetTitle(title string) error {
	return tt.setTitle(title, win.TTI_NONE)
}

func (tt *ToolTip) SetInfoTitle(title string) error {
	return tt.setTitle(title, win.TTI_INFO)
}

func (tt *ToolTip) SetWarningTitle(title string) error {
	return tt.setTitle(title, win.TTI_WARNING)
}

func (tt *ToolTip) SetErrorTitle(title string) error {
	return tt.setTitle(title, win.TTI_ERROR)
}

func (tt *ToolTip) setTitle(title string, icon uintptr) error {
	if len(title) > 99 {
		title = title[:99]
	}

	if win.FALSE == tt.SendMessage(win.TTM_SETTITLE, icon, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title)))) {
		return newError("TTM_SETTITLE failed")
	}

	return nil
}

func (tt *ToolTip) AddTool(tool Widget) error {
	hwnd := tool.Handle()

	var ti win.TOOLINFO
	ti.CbSize = uint32(unsafe.Sizeof(ti))
	ti.Hwnd = hwnd
	ti.UFlags = win.TTF_IDISHWND | win.TTF_SUBCLASS
	ti.UId = uintptr(hwnd)

	if win.FALSE == tt.SendMessage(win.TTM_ADDTOOL, 0, uintptr(unsafe.Pointer(&ti))) {
		return newError("TTM_ADDTOOL failed")
	}

	return nil
}

func (tt *ToolTip) RemoveTool(tool Widget) error {
	hwnd := tool.Handle()

	var ti win.TOOLINFO
	ti.CbSize = uint32(unsafe.Sizeof(ti))
	ti.Hwnd = hwnd
	ti.UId = uintptr(hwnd)

	tt.SendMessage(win.TTM_DELTOOL, 0, uintptr(unsafe.Pointer(&ti)))

	return nil
}

func (tt *ToolTip) Text(tool Widget) string {
	ti := tt.toolInfo(tool)
	if ti == nil {
		return ""
	}

	return win.UTF16PtrToString(ti.LpszText)
}

func (tt *ToolTip) SetText(tool Widget, text string) error {
	ti := tt.toolInfo(tool)
	if ti == nil {
		return newError("unknown tool")
	}

	n := 0
	for i, r := range text {
		if r < 0x10000 {
			n++
		} else {
			n += 2 // surrogate pair
		}
		if n >= maxToolTipTextLen {
			text = text[:i]
			break
		}
	}

	ti.LpszText = syscall.StringToUTF16Ptr(text)

	tt.SendMessage(win.TTM_SETTOOLINFO, 0, uintptr(unsafe.Pointer(ti)))

	return nil
}

func (tt *ToolTip) toolInfo(tool Widget) *win.TOOLINFO {
	var ti win.TOOLINFO
	var buf [maxToolTipTextLen]uint16

	hwnd := tool.Handle()

	ti.CbSize = uint32(unsafe.Sizeof(ti))
	ti.Hwnd = hwnd
	ti.UId = uintptr(hwnd)
	ti.LpszText = &buf[0]

	if win.FALSE == tt.SendMessage(win.TTM_GETTOOLINFO, 0, uintptr(unsafe.Pointer(&ti))) {
		return nil
	}

	return &ti
}
