// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"

	"github.com/lxn/win"
)

const staticWindowClass = `\o/ Walk_Static_Class \o/`

var staticWndProcPtr = syscall.NewCallback(staticWndProc)

func init() {
	MustRegisterWindowClass(staticWindowClass)
}

type static struct {
	WidgetBase
	hwndStatic           win.HWND
	origStaticWndProcPtr uintptr
	textAlignment        Alignment2D
	textColor            Color
}

func (s *static) init(widget Widget, parent Container) error {
	if err := InitWidget(
		widget,
		parent,
		staticWindowClass,
		win.WS_VISIBLE,
		win.WS_EX_CONTROLPARENT); err != nil {
		return err
	}

	if s.hwndStatic = win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr("static"),
		nil,
		win.WS_CHILD|win.WS_CLIPSIBLINGS|win.WS_VISIBLE|win.SS_LEFT,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		s.hWnd,
		0,
		0,
		nil,
	); s.hwndStatic == 0 {
		return newError("creating static failed")
	}

	s.origStaticWndProcPtr = win.SetWindowLongPtr(s.hwndStatic, win.GWLP_WNDPROC, staticWndProcPtr)
	if s.origStaticWndProcPtr == 0 {
		return lastError("SetWindowLongPtr")
	}

	s.applyFont(s.Font())

	s.SetBackground(nullBrushSingleton)

	s.SetAlignment(AlignHNearVCenter)

	return nil
}

func (s *static) Dispose() {
	if s.hwndStatic != 0 {
		win.DestroyWindow(s.hwndStatic)
		s.hwndStatic = 0
	}

	s.WidgetBase.Dispose()
}

func (s *static) LayoutFlags() LayoutFlags {
	if s.textAlignment1D() == AlignNear {
		return 0
	}

	return GrowableHorz
}

func (s *static) MinSizeHint() Size {
	return s.calculateTextSizeForWidth(0)
}

func (s *static) SizeHint() Size {
	return s.MinSizeHint()
}

func (s *static) applyEnabled(enabled bool) {
	s.WidgetBase.applyEnabled(enabled)

	setWindowEnabled(s.hwndStatic, enabled)
}

func (s *static) applyFont(font *Font) {
	s.WidgetBase.applyFont(font)

	SetWindowFont(s.hwndStatic, font)
}

func (s *static) textAlignment1D() Alignment1D {
	switch s.textAlignment {
	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		return AlignCenter

	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		return AlignFar

	default:
		return AlignNear
	}
}

func (s *static) setTextAlignment1D(alignment Alignment1D) error {
	var align Alignment2D

	switch alignment {
	case AlignCenter:
		align = AlignHCenterVCenter

	case AlignFar:
		align = AlignHFarVCenter

	default:
		align = AlignHNearVCenter
	}

	return s.setTextAlignment(align)
}

func (s *static) setTextAlignment(alignment Alignment2D) error {
	if alignment == s.textAlignment {
		return nil
	}

	var styleBit uint32

	switch alignment {
	case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
		styleBit |= win.SS_LEFT

	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		styleBit |= win.SS_CENTER

	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		styleBit |= win.SS_RIGHT
	}

	if err := setAndClearWindowLongBits(s.hwndStatic, win.GWL_STYLE, styleBit, win.SS_LEFT|win.SS_CENTER|win.SS_RIGHT); err != nil {
		return err
	}

	s.textAlignment = alignment

	s.Invalidate()

	return nil
}

func (s *static) setText(text string) (changed bool, err error) {
	if text == s.text() {
		return false, nil
	}

	if err := s.WidgetBase.setText(text); err != nil {
		return false, err
	}

	if err := setWindowText(s.hwndStatic, text); err != nil {
		return false, err
	}

	size := s.BoundsPixels().Size()

	if err := s.updateParentLayout(); err != nil {
		return false, err
	}

	if s.BoundsPixels().Size() == size && size != (Size{}) {
		s.updateStaticBounds()
	}

	return true, nil
}

func (s *static) TextColor() Color {
	return s.textColor
}

func (s *static) SetTextColor(c Color) {
	s.textColor = c

	s.Invalidate()
}

func (s *static) updateStaticBounds() {
	var format DrawTextFormat

	switch s.textAlignment {
	case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
		format |= TextLeft

	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		format |= TextCenter

	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		format |= TextRight
	}

	switch s.textAlignment {
	case AlignHNearVNear, AlignHCenterVNear, AlignHFarVNear:
		format |= TextTop

	case AlignHNearVCenter, AlignHCenterVCenter, AlignHFarVCenter:
		format |= TextVCenter

	case AlignHNearVFar, AlignHCenterVFar, AlignHFarVFar:
		format |= TextBottom
	}

	cb := s.ClientBoundsPixels()

	if format&TextVCenter != 0 || format&TextBottom != 0 {
		var size Size
		if _, ok := s.window.(HeightForWidther); ok {
			size = s.calculateTextSizeForWidth(cb.Width)
		} else {
			size = s.calculateTextSize()
		}

		if format&TextVCenter != 0 {
			cb.Y += (cb.Height - size.Height) / 2
		} else {
			cb.Y += cb.Height - size.Height
		}

		cb.Height = size.Height
	}

	win.MoveWindow(s.hwndStatic, int32(cb.X), int32(cb.Y), int32(cb.Width), int32(cb.Height), true)

	s.Invalidate()
}

func (s *static) WndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case win.WM_CTLCOLORSTATIC:
		if hBrush := s.handleWMCTLCOLOR(wp, uintptr(s.hWnd)); hBrush != 0 {
			return hBrush
		}

	case win.WM_SIZE:
		s.updateStaticBounds()
	}

	return s.WidgetBase.WndProc(hwnd, msg, wp, lp)
}

func staticWndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	as, ok := windowFromHandle(win.GetParent(hwnd)).(interface{ asStatic() *static })
	if !ok {
		return 0
	}

	s := as.asStatic()

	switch msg {
	case win.WM_NCHITTEST:
		return win.HTCLIENT
	}

	return win.CallWindowProc(s.origStaticWndProcPtr, hwnd, msg, wp, lp)
}
