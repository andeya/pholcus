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

type CaseMode uint32

const (
	CaseModeMixed CaseMode = iota
	CaseModeUpper
	CaseModeLower
)

const (
	lineEditMinChars    = 1  // 10 // number of characters needed to make a LineEdit usable
	lineEditGreedyLimit = 29 // 80 // fields with MaxLength larger than this will be greedy (default length is 32767)
)

type LineEdit struct {
	WidgetBase
	editingFinishedPublisher EventPublisher
	readOnlyChangedPublisher EventPublisher
	textChangedPublisher     EventPublisher
	charWidthFont            *Font
	charWidth                int
}

func newLineEdit(parent Window) (*LineEdit, error) {
	le := new(LineEdit)

	if err := InitWindow(
		le,
		parent,
		"EDIT",
		win.WS_CHILD|win.WS_TABSTOP|win.WS_VISIBLE|win.ES_AUTOHSCROLL,
		win.WS_EX_CLIENTEDGE); err != nil {
		return nil, err
	}

	le.MustRegisterProperty("ReadOnly", NewProperty(
		func() interface{} {
			return le.ReadOnly()
		},
		func(v interface{}) error {
			return le.SetReadOnly(v.(bool))
		},
		le.readOnlyChangedPublisher.Event()))

	le.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return le.Text()
		},
		func(v interface{}) error {
			return le.SetText(v.(string))
		},
		le.textChangedPublisher.Event()))

	return le, nil
}

func NewLineEdit(parent Container) (*LineEdit, error) {
	if parent == nil {
		return nil, newError("parent cannot be nil")
	}

	le, err := newLineEdit(parent)
	if err != nil {
		return nil, err
	}

	var succeeded bool
	defer func() {
		if !succeeded {
			le.Dispose()
		}
	}()

	le.parent = parent
	if err = parent.Children().Add(le); err != nil {
		return nil, err
	}

	succeeded = true

	return le, nil
}

func (le *LineEdit) CueBanner() string {
	buf := make([]uint16, 128)
	if win.FALSE == le.SendMessage(win.EM_GETCUEBANNER, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))) {
		newError("EM_GETCUEBANNER failed")
		return ""
	}

	return syscall.UTF16ToString(buf)
}

func (le *LineEdit) SetCueBanner(value string) error {
	if win.FALSE == le.SendMessage(win.EM_SETCUEBANNER, win.FALSE, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(value)))) {
		return newError("EM_SETCUEBANNER failed")
	}

	return nil
}

func (le *LineEdit) MaxLength() int {
	return int(le.SendMessage(win.EM_GETLIMITTEXT, 0, 0))
}

func (le *LineEdit) SetMaxLength(value int) {
	le.SendMessage(win.EM_LIMITTEXT, uintptr(value), 0)
}

func (le *LineEdit) Text() string {
	return windowText(le.hWnd)
}

func (le *LineEdit) SetText(value string) error {
	return setWindowText(le.hWnd, value)
}

func (le *LineEdit) TextSelection() (start, end int) {
	le.SendMessage(win.EM_GETSEL, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
	return
}

func (le *LineEdit) SetTextSelection(start, end int) {
	le.SendMessage(win.EM_SETSEL, uintptr(start), uintptr(end))
}

func (le *LineEdit) CaseMode() CaseMode {
	style := uint32(win.GetWindowLong(le.hWnd, win.GWL_STYLE))

	if style&win.ES_UPPERCASE != 0 {
		return CaseModeUpper
	} else if style&win.ES_LOWERCASE != 0 {
		return CaseModeLower
	} else {
		return CaseModeMixed
	}
}

func (le *LineEdit) SetCaseMode(mode CaseMode) error {
	var set, clear uint32

	switch mode {
	case CaseModeMixed:
		clear = win.ES_UPPERCASE | win.ES_LOWERCASE

	case CaseModeUpper:
		set = win.ES_UPPERCASE
		clear = win.ES_LOWERCASE

	case CaseModeLower:
		set = win.ES_LOWERCASE
		clear = win.ES_UPPERCASE

	default:
		panic("invalid CaseMode")
	}

	return le.setAndClearStyleBits(set, clear)
}

func (le *LineEdit) PasswordMode() bool {
	return le.SendMessage(win.EM_GETPASSWORDCHAR, 0, 0) != 0
}

func (le *LineEdit) SetPasswordMode(value bool) {
	var c uintptr
	if value {
		c = uintptr('*')
	}

	le.SendMessage(win.EM_SETPASSWORDCHAR, c, 0)
}

func (le *LineEdit) ReadOnly() bool {
	return le.hasStyleBits(win.ES_READONLY)
}

func (le *LineEdit) SetReadOnly(readOnly bool) error {
	if 0 == le.SendMessage(win.EM_SETREADONLY, uintptr(win.BoolToBOOL(readOnly)), 0) {
		return newError("SendMessage(EM_SETREADONLY)")
	}

	le.readOnlyChangedPublisher.Publish()

	return nil
}

func (le *LineEdit) LayoutFlags() (lf LayoutFlags) {
	lf = ShrinkableHorz | GrowableHorz
	if le.MaxLength() > lineEditGreedyLimit {
		lf |= GreedyHorz
	}
	return
}

func (le *LineEdit) MinSizeHint() Size {
	return le.sizeHintForLimit(lineEditMinChars)
}

func (le *LineEdit) SizeHint() (size Size) {
	return le.sizeHintForLimit(lineEditGreedyLimit)
}

func (le *LineEdit) sizeHintForLimit(limit int) (size Size) {
	size = le.dialogBaseUnitsToPixels(Size{50, 12})
	le.initCharWidth()
	n := le.MaxLength()
	if n > limit {
		n = limit
	}
	size.Width = le.charWidth * (n + 1)
	return
}

func (le *LineEdit) initCharWidth() {

	font := le.Font()
	if font == le.charWidthFont {
		return
	}
	le.charWidthFont = font
	le.charWidth = 8

	hdc := win.GetDC(le.hWnd)
	if hdc == 0 {
		newError("GetDC failed")
		return
	}
	defer win.ReleaseDC(le.hWnd, hdc)

	defer win.SelectObject(hdc, win.SelectObject(hdc, win.HGDIOBJ(font.handleForDPI(0))))

	buf := []uint16{'M'}

	var s win.SIZE
	if !win.GetTextExtentPoint32(hdc, &buf[0], int32(len(buf)), &s) {
		newError("GetTextExtentPoint32 failed")
		return
	}
	le.charWidth = int(s.CX)
}

func (le *LineEdit) EditingFinished() *Event {
	return le.editingFinishedPublisher.Event()
}

func (le *LineEdit) TextChanged() *Event {
	return le.textChangedPublisher.Event()
}

func (le *LineEdit) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case win.EN_CHANGE:
			le.textChangedPublisher.Publish()
		}

	case win.WM_GETDLGCODE:
		if form := ancestor(le); form != nil {
			if dlg, ok := form.(dialogish); ok {
				if dlg.DefaultButton() != nil {
					// If the LineEdit lives in a Dialog that has a DefaultButton,
					// we won't swallow the return key.
					break
				}
			}
		}

		if wParam == win.VK_RETURN {
			return win.DLGC_WANTALLKEYS
		}

	case win.WM_KEYDOWN:
		switch Key(wParam) {
		case KeyA:
			if ControlDown() {
				le.SetTextSelection(0, -1)
			}

		case KeyReturn:
			le.editingFinishedPublisher.Publish()
		}

	case win.WM_KILLFOCUS:
		// FIXME: This may be dangerous, see remarks section:
		// http://msdn.microsoft.com/en-us/library/ms646282(v=vs.85).aspx
		le.editingFinishedPublisher.Publish()
	}

	return le.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
