// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type CheckState int

const (
	CheckUnchecked     CheckState = win.BST_UNCHECKED
	CheckChecked       CheckState = win.BST_CHECKED
	CheckIndeterminate CheckState = win.BST_INDETERMINATE
)

type CheckBox struct {
	Button

	checkStateChangedPublisher EventPublisher
}

func NewCheckBox(parent Container) (*CheckBox, error) {
	cb := new(CheckBox)

	if err := InitWidget(
		cb,
		parent,
		"BUTTON",
		win.WS_TABSTOP|win.WS_VISIBLE|win.BS_AUTOCHECKBOX,
		0); err != nil {
		return nil, err
	}

	cb.Button.init()

	cb.MustRegisterProperty("CheckState", NewProperty(
		func() interface{} {
			return cb.CheckState()
		},
		func(v interface{}) error {
			cb.SetCheckState(v.(CheckState))

			return nil
		},
		cb.CheckStateChanged()))

	return cb, nil
}

func (*CheckBox) LayoutFlags() LayoutFlags {
	return 0
}

func (cb *CheckBox) MinSizeHint() Size {
	defaultSize := cb.dialogBaseUnitsToPixels(Size{50, 10})
	textSize := cb.calculateTextSizeImpl("n" + windowText(cb.hWnd))

	// FIXME: Use GetThemePartSize instead of GetSystemMetrics?
	w := textSize.Width + int(win.GetSystemMetrics(win.SM_CXMENUCHECK))
	h := maxi(defaultSize.Height, textSize.Height)

	return Size{w, h}
}

func (cb *CheckBox) SizeHint() Size {
	return cb.MinSizeHint()
}

func (cb *CheckBox) setChecked(checked bool) {
	cb.Button.setChecked(checked)

	cb.checkStateChangedPublisher.Publish()
}

func (cb *CheckBox) Tristate() bool {
	return cb.hasStyleBits(win.BS_AUTO3STATE)
}

func (cb *CheckBox) SetTristate(tristate bool) error {
	var set, clear uint32
	if tristate {
		set, clear = win.BS_AUTO3STATE, win.BS_AUTOCHECKBOX
	} else {
		set, clear = win.BS_AUTOCHECKBOX, win.BS_AUTO3STATE
	}

	return cb.setAndClearStyleBits(set, clear)
}

func (cb *CheckBox) CheckState() CheckState {
	return CheckState(cb.SendMessage(win.BM_GETCHECK, 0, 0))
}

func (cb *CheckBox) SetCheckState(state CheckState) {
	if state == cb.CheckState() {
		return
	}

	cb.SendMessage(win.BM_SETCHECK, uintptr(state), 0)

	cb.checkedChangedPublisher.Publish()
	cb.checkStateChangedPublisher.Publish()
}

func (cb *CheckBox) CheckStateChanged() *Event {
	return cb.checkStateChangedPublisher.Event()
}

func (cb *CheckBox) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case win.BN_CLICKED:
			cb.checkedChangedPublisher.Publish()
			cb.checkStateChangedPublisher.Publish()
		}
	}

	return cb.Button.WndProc(hwnd, msg, wParam, lParam)
}
