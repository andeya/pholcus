// Copyright 2014 The Walk Authors. All rights reserved.
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

const scrollViewWindowClass = `\o/ Walk_ScrollView_Class \o/`

func init() {
	MustRegisterWindowClass(scrollViewWindowClass)
}

type ScrollView struct {
	WidgetBase
	composite *Composite
}

func NewScrollView(parent Container) (*ScrollView, error) {
	sv := new(ScrollView)

	if err := InitWidget(
		sv,
		parent,
		scrollViewWindowClass,
		win.WS_CHILD|win.WS_HSCROLL|win.WS_VISIBLE|win.WS_VSCROLL,
		win.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			sv.Dispose()
		}
	}()

	var err error
	if sv.composite, err = NewComposite(sv); err != nil {
		return nil, err
	}

	sv.composite.SizeChanged().Attach(func() {
		sv.updateScrollBars()
	})

	succeeded = true

	return sv, nil
}

func (sv *ScrollView) AsContainerBase() *ContainerBase {
	return sv.composite.AsContainerBase()
}

func (sv *ScrollView) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (sv *ScrollView) SizeHint() Size {
	return sv.MinSizeHint()
}

func (sv *ScrollView) SetSuspended(suspend bool) {
	sv.composite.SetSuspended(suspend)
	sv.WidgetBase.SetSuspended(suspend)
	sv.Invalidate()
}

func (sv *ScrollView) DataBinder() *DataBinder {
	return sv.composite.dataBinder
}

func (sv *ScrollView) SetDataBinder(dataBinder *DataBinder) {
	sv.composite.SetDataBinder(dataBinder)
}

func (sv *ScrollView) Children() *WidgetList {
	if sv.composite == nil {
		// Without this we would get into trouble in NewComposite.
		return nil
	}

	return sv.composite.Children()
}

func (sv *ScrollView) Layout() Layout {
	return sv.composite.Layout()
}

func (sv *ScrollView) SetLayout(value Layout) error {
	return sv.composite.SetLayout(value)
}

func (sv *ScrollView) Name() string {
	return sv.composite.Name()
}

func (sv *ScrollView) SetName(name string) {
	sv.composite.SetName(name)
}

func (sv *ScrollView) Persistent() bool {
	return sv.composite.Persistent()
}

func (sv *ScrollView) SetPersistent(value bool) {
	sv.composite.SetPersistent(value)
}

func (sv *ScrollView) SaveState() error {
	return sv.composite.SaveState()
}

func (sv *ScrollView) RestoreState() error {
	return sv.composite.RestoreState()
}

func (sv *ScrollView) MouseDown() *MouseEvent {
	return sv.composite.MouseDown()
}

func (sv *ScrollView) MouseMove() *MouseEvent {
	return sv.composite.MouseMove()
}

func (sv *ScrollView) MouseUp() *MouseEvent {
	return sv.composite.MouseUp()
}

func (sv *ScrollView) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	if sv.composite != nil {
		switch msg {
		case win.WM_HSCROLL:
			sv.composite.SetX(sv.scroll(win.SB_HORZ, win.LOWORD(uint32(wParam))))

		case win.WM_VSCROLL:
			sv.composite.SetY(sv.scroll(win.SB_VERT, win.LOWORD(uint32(wParam))))

		case win.WM_MOUSEWHEEL:
			var cmd uint16
			if delta := int16(win.HIWORD(uint32(wParam))); delta < 0 {
				cmd = win.SB_LINEDOWN
			} else {
				cmd = win.SB_LINEUP
			}

			sv.composite.SetY(sv.scroll(win.SB_VERT, cmd))

			return 0

		case win.WM_COMMAND, win.WM_NOTIFY:
			sv.composite.WndProc(hwnd, msg, wParam, lParam)

		case win.WM_SIZE, win.WM_SIZING:
			s := maxSize(sv.composite.layout.MinSize(), sv.ClientBounds().Size())
			sv.composite.SetSize(s)
			sv.updateScrollBars()
		}
	}

	return sv.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (sv *ScrollView) updateScrollBars() {
	s := sv.composite.Size()
	clb := sv.ClientBounds()

	var si win.SCROLLINFO
	si.CbSize = uint32(unsafe.Sizeof(si))
	si.FMask = win.SIF_PAGE | win.SIF_RANGE

	si.NMax = int32(s.Width - 1)
	si.NPage = uint32(clb.Width)
	win.SetScrollInfo(sv.hWnd, win.SB_HORZ, &si, false)
	sv.composite.SetX(sv.scroll(win.SB_HORZ, win.SB_THUMBPOSITION))

	si.NMax = int32(s.Height - 1)
	si.NPage = uint32(clb.Height)
	win.SetScrollInfo(sv.hWnd, win.SB_VERT, &si, false)
	sv.composite.SetY(sv.scroll(win.SB_VERT, win.SB_THUMBPOSITION))
}

func (sv *ScrollView) scroll(sb int32, cmd uint16) int {
	var pos int32
	var si win.SCROLLINFO
	si.CbSize = uint32(unsafe.Sizeof(si))
	si.FMask = win.SIF_PAGE | win.SIF_POS | win.SIF_RANGE | win.SIF_TRACKPOS

	win.GetScrollInfo(sv.hWnd, sb, &si)

	pos = si.NPos

	switch cmd {
	case win.SB_LINELEFT: // == win.SB_LINEUP
		pos -= 20

	case win.SB_LINERIGHT: // == win.SB_LINEDOWN
		pos += 20

	case win.SB_PAGELEFT: // == win.SB_PAGEUP
		pos -= int32(si.NPage)

	case win.SB_PAGERIGHT: // == win.SB_PAGEDOWN
		pos += int32(si.NPage)

	case win.SB_THUMBTRACK:
		pos = si.NTrackPos
	}

	if pos < 0 {
		pos = 0
	}
	if pos > si.NMax+1-int32(si.NPage) {
		pos = si.NMax + 1 - int32(si.NPage)
	}

	si.FMask = win.SIF_POS
	si.NPos = pos
	win.SetScrollInfo(sv.hWnd, sb, &si, true)

	return -int(pos)
}
