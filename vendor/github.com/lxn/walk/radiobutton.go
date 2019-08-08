// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type RadioButtonGroup struct {
	buttons       []*RadioButton
	checkedButton *RadioButton
}

func (rbg *RadioButtonGroup) Buttons() []*RadioButton {
	buttons := make([]*RadioButton, len(rbg.buttons))
	copy(buttons, rbg.buttons)
	return buttons
}

func (rbg *RadioButtonGroup) CheckedButton() *RadioButton {
	return rbg.checkedButton
}

type radioButtonish interface {
	radioButton() *RadioButton
}

type RadioButton struct {
	Button
	group *RadioButtonGroup
	value interface{}
}

func NewRadioButton(parent Container) (*RadioButton, error) {
	rb := new(RadioButton)

	if count := parent.Children().Len(); count > 0 {
		if prevRB, ok := parent.Children().At(count - 1).(radioButtonish); ok {
			rb.group = prevRB.radioButton().group
		}
	}
	var groupBit uint32
	if rb.group == nil {
		groupBit = win.WS_GROUP
		rb.group = new(RadioButtonGroup)
	}

	if err := InitWidget(
		rb,
		parent,
		"BUTTON",
		groupBit|win.WS_TABSTOP|win.WS_VISIBLE|win.BS_AUTORADIOBUTTON,
		0); err != nil {
		return nil, err
	}

	rb.Button.init()

	rb.SetBackground(nullBrushSingleton)

	rb.GraphicsEffects().Add(InteractionEffect)
	rb.GraphicsEffects().Add(FocusEffect)

	rb.MustRegisterProperty("CheckedValue", NewProperty(
		func() interface{} {
			if rb.Checked() {
				return rb.value
			}

			return nil
		},
		func(v interface{}) error {
			checked := v == rb.value
			if checked {
				rb.group.checkedButton = rb
			}
			rb.SetChecked(checked)

			return nil
		},
		rb.CheckedChanged()))

	rb.group.buttons = append(rb.group.buttons, rb)

	return rb, nil
}

func (rb *RadioButton) radioButton() *RadioButton {
	return rb
}

func (*RadioButton) LayoutFlags() LayoutFlags {
	return 0
}

func (rb *RadioButton) SizeHint() Size {
	return rb.MinSizeHint()
}

func (rb *RadioButton) TextOnLeftSide() bool {
	return rb.hasStyleBits(win.BS_LEFTTEXT)
}

func (rb *RadioButton) SetTextOnLeftSide(textLeft bool) error {
	return rb.ensureStyleBits(win.BS_LEFTTEXT, textLeft)
}

func (rb *RadioButton) Group() *RadioButtonGroup {
	return rb.group
}

func (rb *RadioButton) Value() interface{} {
	return rb.value
}

func (rb *RadioButton) SetValue(value interface{}) {
	rb.value = value
}

func (rb *RadioButton) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case win.BN_CLICKED:
			prevChecked := rb.group.checkedButton
			rb.group.checkedButton = rb

			if prevChecked != rb {
				if prevChecked != nil {
					prevChecked.setChecked(false)
				}

				rb.setChecked(true)
			}
		}
	}

	return rb.Button.WndProc(hwnd, msg, wParam, lParam)
}
