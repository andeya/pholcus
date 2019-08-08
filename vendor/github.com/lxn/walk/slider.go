// Copyright 2016 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"strconv"

	"github.com/lxn/win"
)

type Slider struct {
	WidgetBase
	valueChangedPublisher EventPublisher
	layoutFlags           LayoutFlags
	tracking              bool
	persistent            bool
}

type SliderCfg struct {
	Orientation    Orientation
	ToolTipsHidden bool
}

func NewSlider(parent Container) (*Slider, error) {
	return NewSliderWithOrientation(parent, Horizontal)
}

func NewSliderWithOrientation(parent Container, orientation Orientation) (*Slider, error) {
	return NewSliderWithCfg(parent, &SliderCfg{Orientation: orientation})
}

func NewSliderWithCfg(parent Container, cfg *SliderCfg) (*Slider, error) {
	sl := new(Slider)

	var style uint32 = win.WS_TABSTOP | win.WS_VISIBLE
	if cfg.Orientation == Vertical {
		style |= win.TBS_VERT
		sl.layoutFlags = ShrinkableVert | GrowableVert
	} else {
		sl.layoutFlags = ShrinkableHorz | GrowableHorz
	}
	if !cfg.ToolTipsHidden {
		style |= win.TBS_TOOLTIPS
	}

	if err := InitWidget(
		sl,
		parent,
		"msctls_trackbar32",
		style,
		0); err != nil {
		return nil, err
	}

	sl.SetBackground(nullBrushSingleton)

	sl.GraphicsEffects().Add(InteractionEffect)
	sl.GraphicsEffects().Add(FocusEffect)

	sl.MustRegisterProperty("Value", NewProperty(
		func() interface{} {
			return sl.Value()
		},
		func(v interface{}) error {
			sl.SetValue(assertIntOr(v, 0))
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

func (sl *Slider) Persistent() bool {
	return sl.persistent
}

func (sl *Slider) SetPersistent(value bool) {
	sl.persistent = value
}

func (sl *Slider) SaveState() error {
	return sl.WriteState(strconv.Itoa(sl.Value()))
}

func (sl *Slider) RestoreState() error {
	s, err := sl.ReadState()
	if err != nil {
		return err
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	sl.SetValue(value)

	return nil
}

func (sl *Slider) LineSize() int {
	return int(sl.SendMessage(win.TBM_GETLINESIZE, 0, 0))
}

func (sl *Slider) SetLineSize(lineSize int) {
	sl.SendMessage(win.TBM_SETLINESIZE, 0, uintptr(lineSize))
}

func (sl *Slider) PageSize() int {
	return int(sl.SendMessage(win.TBM_GETPAGESIZE, 0, 0))
}

func (sl *Slider) SetPageSize(pageSize int) {
	sl.SendMessage(win.TBM_SETPAGESIZE, 0, uintptr(pageSize))
}

func (sl *Slider) Tracking() bool {
	return sl.tracking
}

func (sl *Slider) SetTracking(tracking bool) {
	sl.tracking = tracking
}

func (sl *Slider) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_HSCROLL, win.WM_VSCROLL:
		switch win.LOWORD(uint32(wParam)) {
		case win.TB_THUMBPOSITION, win.TB_ENDTRACK:
			sl.valueChangedPublisher.Publish()

		case win.TB_THUMBTRACK:
			if sl.tracking {
				sl.valueChangedPublisher.Publish()
			}
		}
		return 0
	}
	return sl.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
