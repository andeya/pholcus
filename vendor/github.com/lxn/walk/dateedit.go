// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"strings"
	"syscall"
	"time"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

type DateEdit struct {
	WidgetBase
	dateChangedPublisher EventPublisher
	format               string
}

func newDateEdit(parent Container, style uint32) (*DateEdit, error) {
	de := new(DateEdit)

	if err := InitWidget(
		de,
		parent,
		"SysDateTimePick32",
		win.WS_TABSTOP|win.WS_VISIBLE|win.DTS_SHORTDATEFORMAT|style,
		0); err != nil {
		return nil, err
	}

	de.MustRegisterProperty("Date", NewProperty(
		func() interface{} {
			return de.Date()
		},
		func(v interface{}) error {
			return de.SetDate(v.(time.Time))
		},
		de.dateChangedPublisher.Event()))

	return de, nil
}

func NewDateEdit(parent Container) (*DateEdit, error) {
	return newDateEdit(parent, 0)
}

func NewDateEditWithNoneOption(parent Container) (*DateEdit, error) {
	return newDateEdit(parent, win.DTS_SHOWNONE)
}

func (*DateEdit) LayoutFlags() LayoutFlags {
	return GrowableHorz
}

func (de *DateEdit) MinSizeHint() Size {
	return de.dialogBaseUnitsToPixels(Size{80, 12})
}

func (de *DateEdit) SizeHint() Size {
	return de.MinSizeHint()
}

func (de *DateEdit) systemTimeToTime(st *win.SYSTEMTIME) time.Time {
	if st == nil || !de.hasStyleBits(win.DTS_SHOWNONE) && st.WYear == 1601 && st.WMonth == 1 && st.WDay == 1 {
		return time.Time{}
	}

	var hour, minute, second int
	if de.timeOfDayDisplayed() {
		hour = int(st.WHour)
		minute = int(st.WMinute)
		second = int(st.WSecond)
	}

	return time.Date(int(st.WYear), time.Month(st.WMonth), int(st.WDay), hour, minute, second, 0, time.Local)
}

func (de *DateEdit) timeToSystemTime(t time.Time) *win.SYSTEMTIME {
	if t.Year() < 1601 {
		if de.hasStyleBits(win.DTS_SHOWNONE) {
			return nil
		} else {
			return &win.SYSTEMTIME{
				WYear:  uint16(1601),
				WMonth: uint16(1),
				WDay:   uint16(1),
			}
		}
	}

	st := &win.SYSTEMTIME{
		WYear:  uint16(t.Year()),
		WMonth: uint16(t.Month()),
		WDay:   uint16(t.Day()),
	}

	if de.timeOfDayDisplayed() {
		st.WHour = uint16(t.Hour())
		st.WMinute = uint16(t.Minute())
		st.WSecond = uint16(t.Second())
	}

	return st
}

func (de *DateEdit) systemTime() (*win.SYSTEMTIME, error) {
	var st win.SYSTEMTIME

	switch de.SendMessage(win.DTM_GETSYSTEMTIME, 0, uintptr(unsafe.Pointer(&st))) {
	case win.GDT_VALID:
		return &st, nil

	case win.GDT_NONE:
		return nil, nil
	}

	return nil, newError("SendMessage(DTM_GETSYSTEMTIME)")
}

func (de *DateEdit) setSystemTime(st *win.SYSTEMTIME) error {
	var wParam uintptr

	if st != nil {
		wParam = win.GDT_VALID
	} else {
		// Ensure today's date is displayed.
		de.setSystemTime(de.timeToSystemTime(time.Now()))

		wParam = win.GDT_NONE
	}

	if 0 == de.SendMessage(win.DTM_SETSYSTEMTIME, wParam, uintptr(unsafe.Pointer(st))) {
		return newError("SendMessage(DTM_SETSYSTEMTIME)")
	}

	de.dateChangedPublisher.Publish()

	return nil
}

func (de *DateEdit) timeOfDayDisplayed() bool {
	return strings.ContainsAny(de.format, "Hhms")
}

func (de *DateEdit) Format() string {
	return de.format
}

func (de *DateEdit) SetFormat(format string) error {
	lp := uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(format)))

	if 0 == de.SendMessage(win.DTM_SETFORMAT, 0, lp) {
		return newErr("DTM_SETFORMAT failed")
	}

	de.format = format

	return nil
}

func (de *DateEdit) Range() (min, max time.Time) {
	var st [2]win.SYSTEMTIME

	ret := de.SendMessage(win.DTM_GETRANGE, 0, uintptr(unsafe.Pointer(&st[0])))

	if ret&win.GDTR_MIN > 0 {
		min = de.systemTimeToTime(&st[0])
	}

	if ret&win.GDTR_MAX > 0 {
		max = de.systemTimeToTime(&st[1])
	}

	return
}

func (de *DateEdit) SetRange(min, max time.Time) error {
	if !min.IsZero() && !max.IsZero() {
		if min.Year() > max.Year() ||
			min.Year() == max.Year() && min.Month() > max.Month() ||
			min.Year() == max.Year() && min.Month() == max.Month() && min.Day() > max.Day() {
			return newError("invalid range")
		}
	}

	var st [2]win.SYSTEMTIME
	var wParam uintptr

	if !min.IsZero() {
		wParam |= win.GDTR_MIN
		st[0] = *de.timeToSystemTime(min)
	}

	if !max.IsZero() {
		wParam |= win.GDTR_MAX
		st[1] = *de.timeToSystemTime(max)
	}

	if 0 == de.SendMessage(win.DTM_SETRANGE, wParam, uintptr(unsafe.Pointer(&st[0]))) {
		return newError("SendMessage(DTM_SETRANGE)")
	}

	return nil
}

func (de *DateEdit) Date() time.Time {
	st, err := de.systemTime()
	if err != nil {
		return time.Time{}
	}

	if st == nil {
		return time.Time{}
	}

	return de.systemTimeToTime(st)
}

func (de *DateEdit) SetDate(date time.Time) error {
	return de.setSystemTime(de.timeToSystemTime(date))
}

func (de *DateEdit) DateChanged() *Event {
	return de.dateChangedPublisher.Event()
}

func (de *DateEdit) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_NOTIFY:
		switch uint32(((*win.NMHDR)(unsafe.Pointer(lParam))).Code) {
		case win.DTN_DATETIMECHANGE:
			de.dateChangedPublisher.Publish()
		}
	}

	return de.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
