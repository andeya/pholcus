// Copyright 2014 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"unsafe"

	"github.com/lxn/win"
)

const scrollViewWindowClass = `\o/ Walk_ScrollView_Class \o/`

func init() {
	MustRegisterWindowClass(scrollViewWindowClass)
}

type ScrollView struct {
	WidgetBase
	composite  *Composite
	horizontal bool
	vertical   bool
}

func NewScrollView(parent Container) (*ScrollView, error) {
	sv := &ScrollView{horizontal: true, vertical: true}

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

	sv.SetBackground(NullBrush())

	succeeded = true

	return sv, nil
}

func (sv *ScrollView) AsContainerBase() *ContainerBase {
	if sv.composite == nil {
		return nil
	}

	return sv.composite.AsContainerBase()
}

func (sv *ScrollView) LayoutFlags() LayoutFlags {
	var flags LayoutFlags

	h, v := sv.Scrollbars()

	if h {
		flags |= ShrinkableHorz | GrowableHorz | GreedyHorz
	}

	if v {
		flags |= ShrinkableVert | GrowableVert | GreedyVert
	}

	return flags
}

func (sv *ScrollView) SizeHint() Size {
	return sv.sizeHint(true)
}

func (sv *ScrollView) MinSizeHint() Size {
	return sv.sizeHint(false)
}

func (sv *ScrollView) sizeHint(ideal bool) Size {
	if sv.composite == nil {
		return Size{}
	}

	s := sv.composite.MinSizeHint()
	cb := sv.ClientBoundsPixels()

	h, v := sv.Scrollbars()

	if h {
		if !v {
			if s.Width > cb.Width {
				s.Height += int(win.GetSystemMetrics(win.SM_CYHSCROLL))
			}
		}

		if !ideal {
			s.Width = 100
		}
	}

	if v {
		if !h {
			if s.Height > cb.Height {
				s.Width += int(win.GetSystemMetrics(win.SM_CXVSCROLL))
			}
		}

		if !ideal {
			s.Height = 100
		}
	}

	return s
}

func (sv *ScrollView) ApplyDPI(dpi int) {
	sv.WidgetBase.ApplyDPI(dpi)
	sv.composite.ApplyDPI(dpi)
}

func (sv *ScrollView) Scrollbars() (horizontal, vertical bool) {
	horizontal = sv.horizontal
	vertical = sv.vertical

	return
}

func (sv *ScrollView) SetScrollbars(horizontal, vertical bool) {
	sv.horizontal = horizontal
	sv.vertical = vertical

	sv.ensureStyleBits(win.WS_HSCROLL, horizontal)
	sv.ensureStyleBits(win.WS_VSCROLL, vertical)
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
	if sv.composite == nil {
		return nil
	}

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
		avoidBGArtifacts := func() {
			if sv.hasComplexBackground() {
				sv.composite.Invalidate()
			}
		}

		switch msg {
		case win.WM_HSCROLL:
			sv.composite.SetX(sv.IntTo96DPI(sv.scroll(win.SB_HORZ, win.LOWORD(uint32(wParam)))))
			if wParam == win.SB_ENDSCROLL {
				avoidBGArtifacts()
			}

		case win.WM_VSCROLL:
			sv.composite.SetY(sv.IntTo96DPI(sv.scroll(win.SB_VERT, win.LOWORD(uint32(wParam)))))
			if wParam == win.SB_ENDSCROLL {
				avoidBGArtifacts()
			}

		case win.WM_MOUSEWHEEL:
			if win.GetWindowLong(sv.hWnd, win.GWL_STYLE)&win.WS_VSCROLL == 0 {
				break
			}

			var cmd uint16
			if delta := int16(win.HIWORD(uint32(wParam))); delta < 0 {
				cmd = win.SB_LINEDOWN
			} else {
				cmd = win.SB_LINEUP
			}

			sv.composite.SetY(sv.IntTo96DPI(sv.scroll(win.SB_VERT, cmd)))
			avoidBGArtifacts()

			return 0

		case win.WM_COMMAND, win.WM_NOTIFY:
			sv.composite.WndProc(hwnd, msg, wParam, lParam)

		case win.WM_SIZE, win.WM_SIZING:
			sv.updateCompositeSize()
		}
	}

	return sv.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (sv *ScrollView) updateCompositeSize() {
	var minSize Size
	if fl, ok := sv.composite.layout.(*FlowLayout); ok {
		minSize = fl.MinSizeForSize(sv.ClientBoundsPixels().Size())
	} else {
		minSize = sv.composite.layout.MinSize()
	}
	s := maxSize(minSize, sv.ClientBoundsPixels().Size())
	sv.composite.SetSizePixels(s)
	sv.updateScrollBars()
}

func (sv *ScrollView) updateScrollBars() {
	s := sv.composite.SizePixels()
	clb := sv.ClientBoundsPixels()

	var si win.SCROLLINFO
	si.CbSize = uint32(unsafe.Sizeof(si))
	si.FMask = win.SIF_PAGE | win.SIF_RANGE

	h, v := sv.Scrollbars()

	sbFlags := win.GetWindowLong(sv.hWnd, win.GWL_STYLE) & (win.WS_HSCROLL | win.WS_VSCROLL)

	if h {
		si.NMax = int32(s.Width - 1)
		si.NPage = uint32(clb.Width)
		win.SetScrollInfo(sv.hWnd, win.SB_HORZ, &si, false)
		sv.composite.SetX(sv.IntTo96DPI(sv.scroll(win.SB_HORZ, win.SB_THUMBPOSITION)))
	}

	if v {
		if h {
			clb = sv.ClientBoundsPixels()
		}

		si.NMax = int32(s.Height - 1)
		si.NPage = uint32(clb.Height)
		win.SetScrollInfo(sv.hWnd, win.SB_VERT, &si, false)
		sv.composite.SetY(sv.IntTo96DPI(sv.scroll(win.SB_VERT, win.SB_THUMBPOSITION)))
	}

	if sbFlags != win.GetWindowLong(sv.hWnd, win.GWL_STYLE)&(win.WS_HSCROLL|win.WS_VSCROLL) {
		sv.updateParentLayout()
	}
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

func ifContainerIsScrollViewDoCoolSpecialLayoutStuff(layout Layout) bool {
	if widget, ok := layout.Container().(Widget); ok {
		if parent := widget.Parent(); parent != nil {
			if sv, ok := parent.(*ScrollView); ok {
				min := layout.MinSize()
				flags := layout.LayoutFlags()

				s := widget.BoundsPixels().Size()

				hsb, vsb := sv.Scrollbars()

				var changeCompositeSize bool
				if min.Width > s.Width || min.Width < s.Width && (!hsb || (flags&GreedyHorz == 0)) {
					s.Width = min.Width
					changeCompositeSize = true
				}
				if min.Height > s.Height || min.Height < s.Height && (!vsb || (flags&GreedyVert == 0)) {
					s.Height = min.Height
					changeCompositeSize = true
				}

				if changeCompositeSize {
					widget.SetBoundsPixels(Rectangle{X: 0, Y: 0, Width: s.Width, Height: s.Height})
					sv.updateScrollBars()
					return false
				}

				parent = sv.Parent()

				for parent != nil {
					if parentLayout := parent.Layout(); parentLayout != nil {
						flags := parentLayout.LayoutFlags()

						if !hsb && flags&GreedyHorz != 0 || !vsb && flags&GreedyVert != 0 {
							parentLayout.Update(false)
							return true
						}
					}

					if widget, ok := parent.(Widget); ok {
						parent = widget.Parent()
					} else {
						break
					}
				}
			}
		}
	}

	return false
}
