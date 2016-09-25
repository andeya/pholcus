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

type TextEdit struct {
	WidgetBase
	readOnlyChangedPublisher EventPublisher
	textChangedPublisher     EventPublisher
}

func NewTextEdit(parent Container) (*TextEdit, error) {
	te := new(TextEdit)

	if err := InitWidget(
		te,
		parent,
		"EDIT",
		win.WS_TABSTOP|win.WS_VISIBLE|win.WS_VSCROLL|win.ES_MULTILINE|win.ES_WANTRETURN,
		win.WS_EX_CLIENTEDGE); err != nil {
		return nil, err
	}

	te.MustRegisterProperty("ReadOnly", NewProperty(
		func() interface{} {
			return te.ReadOnly()
		},
		func(v interface{}) error {
			return te.SetReadOnly(v.(bool))
		},
		te.readOnlyChangedPublisher.Event()))

	te.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return te.Text()
		},
		func(v interface{}) error {
			return te.SetText(v.(string))
		},
		te.textChangedPublisher.Event()))

	return te, nil
}

func (*TextEdit) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (te *TextEdit) MinSizeHint() Size {
	return te.dialogBaseUnitsToPixels(Size{20, 12})
}

func (te *TextEdit) SizeHint() Size {
	return Size{100, 100}
}

func (te *TextEdit) Text() string {
	return windowText(te.hWnd)
}

func (te *TextEdit) TextLength() int {
	return int(te.SendMessage(win.WM_GETTEXTLENGTH, 0, 0))
}

func (te *TextEdit) SetText(value string) error {
	return setWindowText(te.hWnd, value)
}

func (te *TextEdit) MaxLength() int {
	return int(te.SendMessage(win.EM_GETLIMITTEXT, 0, 0))
}

func (te *TextEdit) SetMaxLength(value int) {
	te.SendMessage(win.EM_SETLIMITTEXT, uintptr(value), 0)
}

func (te *TextEdit) TextSelection() (start, end int) {
	te.SendMessage(win.EM_GETSEL, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
	return
}

func (te *TextEdit) SetTextSelection(start, end int) {
	te.SendMessage(win.EM_SETSEL, uintptr(start), uintptr(end))
}

func (te *TextEdit) ReplaceSelectedText(text string, canUndo bool) {
	te.SendMessage(win.EM_REPLACESEL,
		uintptr(win.BoolToBOOL(canUndo)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))))
}

func (te *TextEdit) AppendText(value string) {
	s, e := te.TextSelection()
	l := te.TextLength()
	te.SetTextSelection(l, l)
	te.ReplaceSelectedText(value, false)
	te.SetTextSelection(s, e)
}

func (te *TextEdit) ReadOnly() bool {
	return te.hasStyleBits(win.ES_READONLY)
}

func (te *TextEdit) SetReadOnly(readOnly bool) error {
	if 0 == te.SendMessage(win.EM_SETREADONLY, uintptr(win.BoolToBOOL(readOnly)), 0) {
		return newError("SendMessage(EM_SETREADONLY)")
	}

	te.readOnlyChangedPublisher.Publish()

	return nil
}

func (te *TextEdit) TextChanged() *Event {
	return te.textChangedPublisher.Event()
}

func (te *TextEdit) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case win.EN_CHANGE:
			te.textChangedPublisher.Publish()
		}

	case win.WM_GETDLGCODE:
		if wParam == win.VK_RETURN {
			return win.DLGC_WANTALLKEYS
		}

		return win.DLGC_HASSETSEL | win.DLGC_WANTARROWS | win.DLGC_WANTCHARS

	case win.WM_KEYDOWN:
		if Key(wParam) == KeyA && ControlDown() {
			te.SetTextSelection(0, -1)
		}
	}

	return te.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
