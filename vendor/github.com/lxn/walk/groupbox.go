// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
)

import (
	"github.com/lxn/win"
)

const groupBoxWindowClass = `\o/ Walk_GroupBox_Class \o/`

func init() {
	MustRegisterWindowClass(groupBoxWindowClass)
}

type GroupBox struct {
	WidgetBase
	hWndGroupBox          win.HWND
	checkBox              *CheckBox
	composite             *Composite
	titleChangedPublisher EventPublisher
}

func NewGroupBox(parent Container) (*GroupBox, error) {
	gb := new(GroupBox)

	if err := InitWidget(
		gb,
		parent,
		groupBoxWindowClass,
		win.WS_VISIBLE,
		win.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			gb.Dispose()
		}
	}()

	gb.hWndGroupBox = win.CreateWindowEx(
		0, syscall.StringToUTF16Ptr("BUTTON"), nil,
		win.WS_CHILD|win.WS_VISIBLE|win.BS_GROUPBOX,
		0, 0, 80, 24, gb.hWnd, 0, 0, nil)
	if gb.hWndGroupBox == 0 {
		return nil, lastError("CreateWindowEx(BUTTON)")
	}

	setWindowFont(gb.hWndGroupBox, gb.Font())

	var err error

	gb.checkBox, err = NewCheckBox(gb)
	if err != nil {
		return nil, err
	}

	gb.SetCheckable(false)
	gb.checkBox.SetChecked(true)

	gb.checkBox.CheckedChanged().Attach(func() {
		gb.applyEnabledFromCheckBox(gb.checkBox.Checked())
	})

	setWindowVisible(gb.checkBox.hWnd, false)

	gb.composite, err = NewComposite(gb)
	if err != nil {
		return nil, err
	}

	win.SetWindowPos(gb.checkBox.hWnd, win.HWND_TOP, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE)

	gb.MustRegisterProperty("Title", NewProperty(
		func() interface{} {
			return gb.Title()
		},
		func(v interface{}) error {
			return gb.SetTitle(v.(string))
		},
		gb.titleChangedPublisher.Event()))

	gb.MustRegisterProperty("Checked", NewBoolProperty(
		func() bool {
			return gb.Checked()
		},
		func(v bool) error {
			gb.SetChecked(v)
			return nil
		},
		gb.CheckedChanged()))

	succeeded = true

	return gb, nil
}

func (gb *GroupBox) AsContainerBase() *ContainerBase {
	return gb.composite.AsContainerBase()
}

func (gb *GroupBox) LayoutFlags() LayoutFlags {
	if gb.composite == nil {
		return 0
	}

	return gb.composite.LayoutFlags()
}

func (gb *GroupBox) MinSizeHint() Size {
	if gb.composite == nil {
		return Size{100, 100}
	}

	cmsh := gb.composite.MinSizeHint()

	if gb.Checkable() {
		s := gb.checkBox.SizeHint()

		cmsh.Width = maxi(cmsh.Width, s.Width)
		cmsh.Height += s.Height
	}

	return Size{cmsh.Width + 2, cmsh.Height + 14}
}

func (gb *GroupBox) SizeHint() Size {
	return gb.MinSizeHint()
}

func (gb *GroupBox) ClientBounds() Rectangle {
	cb := windowClientBounds(gb.hWndGroupBox)

	if gb.Layout() == nil {
		return cb
	}

	if gb.Checkable() {
		s := gb.checkBox.SizeHint()

		cb.Y += s.Height
		cb.Height -= s.Height
	}

	// FIXME: Use appropriate margins
	return Rectangle{cb.X + 1, cb.Y + 14, cb.Width - 2, cb.Height - 14}
}

func (gb *GroupBox) applyEnabled(enabled bool) {
	gb.WidgetBase.applyEnabled(enabled)

	if gb.hWndGroupBox != 0 {
		setWindowEnabled(gb.hWndGroupBox, enabled)
	}

	if gb.checkBox != nil {
		gb.checkBox.applyEnabled(enabled)
	}

	if gb.composite != nil {
		gb.composite.applyEnabled(enabled)
	}
}

func (gb *GroupBox) applyEnabledFromCheckBox(enabled bool) {
	if gb.hWndGroupBox != 0 {
		setWindowEnabled(gb.hWndGroupBox, enabled)
	}

	if gb.composite != nil {
		gb.composite.applyEnabled(enabled)
	}
}

func (gb *GroupBox) applyFont(font *Font) {
	gb.WidgetBase.applyFont(font)

	if gb.checkBox != nil {
		gb.checkBox.applyFont(font)
	}

	if gb.hWndGroupBox != 0 {
		setWindowFont(gb.hWndGroupBox, font)
	}

	if gb.composite != nil {
		gb.composite.applyFont(font)
	}
}

func (gb *GroupBox) SetSuspended(suspend bool) {
	gb.composite.SetSuspended(suspend)
	gb.WidgetBase.SetSuspended(suspend)
	gb.Invalidate()
}

func (gb *GroupBox) DataBinder() *DataBinder {
	return gb.composite.dataBinder
}

func (gb *GroupBox) SetDataBinder(dataBinder *DataBinder) {
	gb.composite.SetDataBinder(dataBinder)
}

func (gb *GroupBox) Title() string {
	if gb.Checkable() {
		return gb.checkBox.Text()
	}

	return windowText(gb.hWndGroupBox)
}

func (gb *GroupBox) SetTitle(title string) error {
	if gb.Checkable() {
		if err := setWindowText(gb.hWndGroupBox, ""); err != nil {
			return err
		}

		return gb.checkBox.SetText(title)
	}

	return setWindowText(gb.hWndGroupBox, title)
}

func (gb *GroupBox) Checkable() bool {
	return gb.checkBox.visible
}

func (gb *GroupBox) SetCheckable(checkable bool) {
	title := gb.Title()

	gb.checkBox.SetVisible(checkable)

	gb.SetTitle(title)

	gb.updateParentLayout()
}

func (gb *GroupBox) Checked() bool {
	return gb.checkBox.Checked()
}

func (gb *GroupBox) SetChecked(checked bool) {
	gb.checkBox.SetChecked(checked)
}

func (gb *GroupBox) CheckedChanged() *Event {
	return gb.checkBox.CheckedChanged()
}

func (gb *GroupBox) Children() *WidgetList {
	if gb.composite == nil {
		// Without this we would get into trouble in NewComposite.
		return nil
	}

	return gb.composite.Children()
}

func (gb *GroupBox) Layout() Layout {
	if gb.composite == nil {
		// Without this we would get into trouble through the call to
		// SetCheckable in NewGroupBox.
		return nil
	}

	return gb.composite.Layout()
}

func (gb *GroupBox) SetLayout(value Layout) error {
	return gb.composite.SetLayout(value)
}

func (gb *GroupBox) MouseDown() *MouseEvent {
	return gb.composite.MouseDown()
}

func (gb *GroupBox) MouseMove() *MouseEvent {
	return gb.composite.MouseMove()
}

func (gb *GroupBox) MouseUp() *MouseEvent {
	return gb.composite.MouseUp()
}

func (gb *GroupBox) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	if gb.composite != nil {
		switch msg {
		case win.WM_COMMAND, win.WM_NOTIFY:
			gb.composite.WndProc(hwnd, msg, wParam, lParam)

		case win.WM_SETTEXT:
			gb.titleChangedPublisher.Publish()

		case win.WM_PAINT:
			win.UpdateWindow(gb.checkBox.hWnd)

		case win.WM_SIZE, win.WM_SIZING:
			wbcb := gb.WidgetBase.ClientBounds()
			if !win.MoveWindow(
				gb.hWndGroupBox,
				int32(wbcb.X),
				int32(wbcb.Y),
				int32(wbcb.Width),
				int32(wbcb.Height),
				true) {

				lastError("MoveWindow")
				break
			}

			if gb.Checkable() {
				s := gb.checkBox.SizeHint()
				gb.checkBox.SetBounds(Rectangle{9, 14, s.Width, s.Height})
			}

			gbcb := gb.ClientBounds()
			gb.composite.SetBounds(gbcb)
		}
	}

	return gb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
