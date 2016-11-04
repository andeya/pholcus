// Copyright 2016 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type Slider struct {
	WidgetBase
	valueChangedPublisher EventPublisher
	layoutFlags           LayoutFlags
}

func NewSlider(parent Container) (*Slider, error) {
	return NewSliderWithOrientation(parent, Horizontal)
}

func NewSliderWithOrientation(parent Container, orientation Orientation) (*Slider, error) {
	sl := new(Slider)

	var style uint32 = win.WS_TABSTOP | win.WS_VISIBLE | win.TBS_TOOLTIPS
	if orientation == Vertical {
		style |= win.TBS_VERT
		sl.layoutFlags = ShrinkableVert | GrowableVert | GreedyVert
	} else {
		sl.layoutFlags = ShrinkableHorz | GrowableHorz | GreedyHorz
	}

	if err := InitWidget(
		sl,
		parent,
		"msctls_trackbar32",
		style,
		0); err != nil {
		return nil, err
	}

	sl.MustRegisterProperty("Value", NewProperty(
		func() interface{} {
			return sl.Value()
		},
		func(v interface{}) error {
			sl.SetValue(v.(int))
			return nil
		},
		sl.valueChangedPublisher.Event()))

	return sl, nil
}

func (sl *Slider) LayoutFlags() LayoutFlags {
	return sl.layoutFlags
}

func (sl *Slider) SizeHint() Size {
	return sl.MinSizeHint()
}

func (sl *Slider) MinSizeHint() Size {
	return sl.dialogBaseUnitsToPixels(Size{20, 20})
}

func (sl *Slider) MinValue() int {
	return int(sl.SendMessage(win.TBM_GETRANGEMIN, 0, 0))
}

func (sl *Slider) MaxValue() int {
	return int(sl.SendMessage(win.TBM_GETRANGEMAX, 0, 0))
}

func (sl *Slider) SetRange(min, max int) {
	sl.SendMessage(win.TBM_SETRANGEMIN, 0, uintptr(min))
	sl.SendMessage(win.TBM_SETRANGEMAX, 1, uintptr(max))
}

func (sl *Slider) Value() int {
	return int(sl.SendMessage(win.TBM_GETPOS, 0, 0))
}

func (sl *Slider) SetValue(value int) {
	sl.SendMessage(win.TBM_SETPOS, 1, uintptr(value))
	sl.valueChangedPublisher.Publish()
}

// ValueChanged returns an Event that can be used to track changes to Value.
func (sl *Slider) ValueChanged() *Event {
	return sl.valueChangedPublisher.Event()
}

func (sl *Slider) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_HSCROLL, win.WM_VSCROLL:
		switch win.LOWORD(uint32(wParam)) {
		case win.TB_THUMBPOSITION, win.TB_ENDTRACK:
			sl.valueChangedPublisher.Publish()
		}
		return 0
	}
	return sl.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
