// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"bytes"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

const numberEditWindowClass = `\o/ Walk_NumberEdit_Class \o/`

func init() {
	MustRegisterWindowClass(numberEditWindowClass)
}

// NumberEdit is a widget that is suited to edit numeric values.
type NumberEdit struct {
	WidgetBase
	edit *numberLineEdit
}

// NewNumberEdit returns a new NumberEdit widget as child of parent.
func NewNumberEdit(parent Container) (*NumberEdit, error) {
	ne := new(NumberEdit)

	if err := InitWidget(
		ne,
		parent,
		numberEditWindowClass,
		win.WS_VISIBLE,
		win.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	var succeeded bool
	defer func() {
		if !succeeded {
			ne.Dispose()
		}
	}()

	var err error
	if ne.edit, err = newNumberLineEdit(ne); err != nil {
		return nil, err
	}

	ne.edit.applyFont(ne.Font())

	if err = ne.SetValue(0); err != nil {
		return nil, err
	}

	ne.MustRegisterProperty("Value", NewProperty(
		func() interface{} {
			return ne.Value()
		},
		func(v interface{}) error {
			return ne.SetValue(v.(float64))
		},
		ne.edit.valueChangedPublisher.Event()))

	succeeded = true

	return ne, nil
}

func (ne *NumberEdit) applyEnabled(enabled bool) {
	ne.WidgetBase.applyEnabled(enabled)

	if ne.edit == nil {
		return
	}

	ne.edit.applyEnabled(enabled)
}

func (ne *NumberEdit) applyFont(font *Font) {
	ne.WidgetBase.applyFont(font)

	if ne.edit == nil {
		return
	}

	ne.edit.applyFont(font)
}

// LayoutFlags returns information that is mainly interesting to Layout
// implementations.
func (*NumberEdit) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | GrowableHorz
}

// MinSizeHint returns information that is mainly interesting to Layout
// implementations.
func (ne *NumberEdit) MinSizeHint() Size {
	return ne.dialogBaseUnitsToPixels(Size{20, 12})
}

// SizeHint returns information that is mainly interesting to Layout
// implementations.
func (ne *NumberEdit) SizeHint() Size {
	return ne.dialogBaseUnitsToPixels(Size{50, 12})
	//	s := ne.dialogBaseUnitsToPixels(Size{50, 12})
	//	return Size{s.Width, maxi(s.Height, 22)}
}

// Decimals returns the number of decimal places in the NumberEdit.
func (ne *NumberEdit) Decimals() int {
	return ne.edit.decimals
}

// SetDecimals sets the number of decimal places in the NumberEdit.
func (ne *NumberEdit) SetDecimals(decimals int) error {
	if decimals < 0 || decimals > 8 {
		return newError("decimals must >= 0 && <= 8")
	}

	ne.edit.decimals = decimals

	return ne.SetValue(ne.edit.value)
}

// Prefix returns the text that appears in the NumberEdit before the number.
func (ne *NumberEdit) Prefix() string {
	return syscall.UTF16ToString(ne.edit.prefix)
}

// SetPrefix sets the text that appears in the NumberEdit before the number.
func (ne *NumberEdit) SetPrefix(prefix string) error {
	p, err := syscall.UTF16FromString(prefix)
	if err != nil {
		return err
	}

	old := ne.edit.prefix
	ne.edit.prefix = p[:len(p)-1]

	if err := ne.edit.setTextFromValue(ne.edit.value); err != nil {
		ne.edit.prefix = old
		return err
	}

	return nil
}

// Suffix returns the text that appears in the NumberEdit after the number.
func (ne *NumberEdit) Suffix() string {
	return syscall.UTF16ToString(ne.edit.suffix)
}

// SetSuffix sets the text that appears in the NumberEdit after the number.
func (ne *NumberEdit) SetSuffix(suffix string) error {
	s, err := syscall.UTF16FromString(suffix)
	if err != nil {
		return err
	}

	old := ne.edit.suffix
	ne.edit.suffix = s[:len(s)-1]

	if err := ne.edit.setTextFromValue(ne.edit.value); err != nil {
		ne.edit.suffix = old
		return err
	}

	return nil
}

// Increment returns the amount by which the NumberEdit increments or decrements
// its value, when the user presses the KeyDown or KeyUp keys, or when the mouse
// wheel is rotated.
func (ne *NumberEdit) Increment() float64 {
	return ne.edit.increment
}

// SetIncrement sets the amount by which the NumberEdit increments or decrements
// its value, when the user presses the KeyDown or KeyUp keys, or when the mouse
// wheel is rotated.
func (ne *NumberEdit) SetIncrement(increment float64) error {
	ne.edit.increment = increment

	return nil
}

// MinValue returns the minimum value the NumberEdit will accept.
func (ne *NumberEdit) MinValue() float64 {
	return ne.edit.minValue
}

// MinValue returns the maximum value the NumberEdit will accept.
func (ne *NumberEdit) MaxValue() float64 {
	return ne.edit.maxValue
}

// SetRange sets the minimum and maximum values the NumberEdit will accept.
//
// If the current value is out of this range, it will be adjusted.
func (ne *NumberEdit) SetRange(min, max float64) error {
	if min > max {
		return newError("invalid range")
	}

	ne.edit.minValue = min
	ne.edit.maxValue = max
	if min != max {
		if ne.edit.value < min {
			if err := ne.edit.setValue(min, true); err != nil {
				return err
			}
		} else if ne.edit.value > max {
			if err := ne.edit.setValue(max, true); err != nil {
				return err
			}
		}
	}

	return nil
}

// Value returns the value of the NumberEdit.
func (ne *NumberEdit) Value() float64 {
	return ne.edit.value
}

// SetValue sets the value of the NumberEdit.
func (ne *NumberEdit) SetValue(value float64) error {
	if ne.edit.minValue != ne.edit.maxValue &&
		(value < ne.edit.minValue || value > ne.edit.maxValue) {

		return newError("value out of range")
	}

	return ne.edit.setValue(value, true)
}

// ValueChanged returns an Event that can be used to track changes to Value.
func (ne *NumberEdit) ValueChanged() *Event {
	return ne.edit.valueChangedPublisher.Event()
}

// SetFocus sets the keyboard input focus to the NumberEdit.
func (ne *NumberEdit) SetFocus() error {
	if win.SetFocus(ne.edit.hWnd) == 0 {
		return lastError("SetFocus")
	}

	return nil
}

// TextSelection returns the range of the current text selection of the
// NumberEdit.
func (ne *NumberEdit) TextSelection() (start, end int) {
	return ne.edit.TextSelection()
}

// SetTextSelection sets the range of the current text selection of the
// NumberEdit.
func (ne *NumberEdit) SetTextSelection(start, end int) {
	ne.edit.SetTextSelection(start, end)
}

// WndProc is the window procedure of the NumberEdit.
//
// When implementing your own WndProc to add or modify behavior, call the
// WndProc of the embedded NumberEdit for messages you don't handle yourself.
func (ne *NumberEdit) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_SIZE, win.WM_SIZING:
		if ne.edit == nil {
			break
		}

		cb := ne.ClientBounds()
		if err := ne.edit.SetBounds(cb); err != nil {
			break
		}
	}

	return ne.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

type numberLineEdit struct {
	*LineEdit
	buf                   *bytes.Buffer
	prefix                []uint16
	suffix                []uint16
	value                 float64
	minValue              float64
	maxValue              float64
	increment             float64
	decimals              int
	valueChangedPublisher EventPublisher
	inEditMode            bool
}

func newNumberLineEdit(parent Widget) (*numberLineEdit, error) {
	nle := &numberLineEdit{
		buf:       new(bytes.Buffer),
		increment: 1,
	}

	var err error
	if nle.LineEdit, err = newLineEdit(parent); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			nle.Dispose()
		}
	}()

	if err := nle.LineEdit.setAndClearStyleBits(win.ES_RIGHT, win.ES_LEFT|win.ES_CENTER); err != nil {
		return nil, err
	}

	if err := InitWrapperWindow(nle); err != nil {
		return nil, err
	}

	succeeded = true

	return nle, nil
}

func (nle *numberLineEdit) setValue(value float64, setText bool) error {
	if setText {
		if err := nle.setTextFromValue(value); err != nil {
			return err
		}
	}

	if value == nle.value {
		return nil
	}

	nle.value = value

	nle.valueChangedPublisher.Publish()

	return nil
}

func (nle *numberLineEdit) setTextFromValue(value float64) error {
	nle.buf.Reset()

	nle.buf.WriteString(syscall.UTF16ToString(nle.prefix))

	if nle.decimals > 0 {
		nle.buf.WriteString(FormatFloatGrouped(value, nle.decimals))
	} else {
		nle.buf.WriteString(FormatFloat(value, nle.decimals))
	}

	nle.buf.WriteString(syscall.UTF16ToString(nle.suffix))

	return nle.SetText(nle.buf.String())
}

func (nle *numberLineEdit) endEdit() error {
	if err := nle.setTextFromValue(nle.value); err != nil {
		return err
	}

	nle.inEditMode = false

	return nil
}

func (nle *numberLineEdit) processChar(text []uint16, start, end int, key Key, char uint16) {
	hadSelection := start != end

	if !nle.inEditMode {
		var groupSepsBeforeStart int
		if nle.decimals > 0 {
			groupSepsBeforeStart = uint16CountUint16(text[:start], groupSepUint16)
		}

		if hadSelection {
			text = append(text[:start], text[end:]...)
		}

		if nle.decimals > 0 {
			text = uint16RemoveUint16(text, groupSepUint16)
			start -= groupSepsBeforeStart
		}

		nle.inEditMode = true
	} else {
		if hadSelection {
			text = append(text[:start], text[end:]...)
		}
	}

	end = start

	switch key {
	case KeyBack:
		if !hadSelection && start > 0 {
			start -= 1
			text = append(text[:start], text[start+1:]...)
		}

	case KeyDelete:
		if !hadSelection && start < len(text) {
			text = append(text[:start], text[start+1:]...)
		}

	default:
		t := make([]uint16, len(text[:start]), len(text)+1)
		copy(t, text[:start])
		t = append(t, char)
		text = append(t, text[start:]...)
		start += 1
	}

	nle.buf.Reset()

	str := syscall.UTF16ToString(text)

	nle.buf.WriteString(syscall.UTF16ToString(nle.prefix))
	nle.buf.WriteString(str)
	nle.buf.WriteString(syscall.UTF16ToString(nle.suffix))

	nle.SetText(nle.buf.String())

	start += len(nle.prefix)
	nle.SetTextSelection(start, start)

	nle.tryUpdateValue(false)
}

func (nle *numberLineEdit) tryUpdateValue(setText bool) bool {
	t := nle.textUTF16()
	t = t[len(nle.prefix) : len(t)-len(nle.suffix)]

	text := strings.Replace(syscall.UTF16ToString(t), decimalSepS, ".", 1)

	switch text {
	case "", ".":
		text = "0"
	}

	if value, err := strconv.ParseFloat(text, 64); err == nil {
		if nle.minValue == nle.maxValue || value >= nle.minValue && value <= nle.maxValue {
			return nle.setValue(value, setText) == nil
		}
	}

	return false
}

func (nle *numberLineEdit) selectNumber() {
	nle.SetTextSelection(len(nle.prefix), len(nle.textUTF16())-len(nle.suffix))
}

func (nle *numberLineEdit) textUTF16() []uint16 {
	textLength := nle.SendMessage(win.WM_GETTEXTLENGTH, 0, 0)
	buf := make([]uint16, textLength+1)
	nle.SendMessage(win.WM_GETTEXT, uintptr(textLength+1), uintptr(unsafe.Pointer(&buf[0])))

	return buf[:len(buf)-1]
}

func (nle *numberLineEdit) incrementValue(delta float64) {
	value := nle.value + delta

	if nle.minValue != nle.maxValue {
		if value < nle.minValue {
			value = nle.minValue
		} else if value > nle.maxValue {
			value = nle.maxValue
		}
	}

	nle.setValue(value, true)
	nle.selectNumber()
}

func (nle *numberLineEdit) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_CHAR:
		if AltDown() {
			return 0
		}

		if ControlDown() {
			if wParam == 1 {
				// Ctrl+A
				return 0
			}
			break
		}

		char := uint16(wParam)

		text := nle.textUTF16()
		text = text[len(nle.prefix) : len(text)-len(nle.suffix)]
		start, end := nle.TextSelection()
		start -= len(nle.prefix)
		end -= len(nle.prefix)

		if Key(wParam) == KeyBack {
			nle.processChar(text, start, end, KeyBack, 0)
			return 0
		}

		switch char {
		case uint16('0'), uint16('1'), uint16('2'), uint16('3'), uint16('4'), uint16('5'), uint16('6'), uint16('7'), uint16('8'), uint16('9'):
			if start == end && nle.decimals > 0 {
				if i := uint16IndexUint16(text, decimalSepUint16); i > -1 && i < len(text)-nle.decimals && start > i {
					return 0
				}
			}

			nle.processChar(text, start, end, 0, char)
			return 0

		case uint16('-'):
			if nle.minValue != nle.maxValue && nle.minValue >= 0 {
				return 0
			}

			if start > 0 || uint16ContainsUint16(text, uint16('-')) && end == 0 {
				return 0
			}

			nle.processChar(text, start, end, 0, char)
			return 0

		case decimalSepUint16:
			if nle.decimals == 0 {
				return 0
			}

			if start == 0 && end == 0 && len(text) > 0 && text[0] == '-' {
				return 0
			}

			if end < len(text)-nle.decimals {
				return 0
			}

			if i := uint16IndexUint16(text, decimalSepUint16); i > -1 && i <= start || i > end {
				return 0
			}

			nle.processChar(text, start, end, 0, char)
			return 0

		default:
			return 0
		}

	case win.WM_KEYDOWN:
		switch Key(wParam) {
		case KeyA:
			if ControlDown() {
				nle.selectNumber()
				return 0
			}

		case KeyDelete:
			text := nle.textUTF16()
			text = text[len(nle.prefix) : len(text)-len(nle.suffix)]
			start, end := nle.TextSelection()
			start -= len(nle.prefix)
			end -= len(nle.prefix)

			nle.processChar(text, start, end, KeyDelete, 0)
			return 0

		case KeyDown:
			nle.incrementValue(-nle.increment)
			return 0

		case KeyEnd:
			start, end := nle.TextSelection()
			end = len(nle.textUTF16()) - len(nle.suffix)
			if !ShiftDown() {
				start = end
			}
			nle.SetTextSelection(start, end)
			return 0

		case KeyHome:
			start, end := nle.TextSelection()
			start = len(nle.prefix)
			if !ShiftDown() {
				end = start
			}
			nle.SetTextSelection(start, end)
			return 0

		case KeyLeft:
			var pos win.POINT
			win.GetCaretPos(&pos)

			lParam := uintptr(win.MAKELONG(uint16(pos.X), uint16(pos.Y)))
			i := int(win.LOWORD(uint32(nle.SendMessage(win.EM_CHARFROMPOS, 0, lParam))))

			if min := len(nle.prefix); i <= min {
				if !ShiftDown() {
					nle.SetTextSelection(min, min)
				}
				return 0
			}

		case KeyReturn:
			if nle.inEditMode {
				nle.endEdit()
				nle.selectNumber()
				return 0
			}

		case KeyRight:
			var pos win.POINT
			win.GetCaretPos(&pos)

			lParam := uintptr(win.MAKELONG(uint16(pos.X), uint16(pos.Y)))
			i := int(win.LOWORD(uint32(nle.SendMessage(win.EM_CHARFROMPOS, 0, lParam))))

			if max := len(nle.textUTF16()) - len(nle.suffix); i >= max {
				if !ShiftDown() {
					nle.SetTextSelection(max, max)
				}
				return 0
			}

		case KeyUp:
			nle.incrementValue(nle.increment)
			return 0
		}

	case win.WM_GETDLGCODE:
		if !nle.inEditMode {
			if form := ancestor(nle); form != nil {
				if dlg, ok := form.(dialogish); ok {
					if dlg.DefaultButton() != nil {
						// If the NumberEdit lives in a Dialog that has a
						// DefaultButton, we won't swallow the return key.
						break
					}
				}
			}
		}

		if wParam == win.VK_RETURN {
			return win.DLGC_WANTALLKEYS
		}

	case win.WM_KILLFOCUS:
		nle.endEdit()

	case win.WM_LBUTTONDOWN:
		i := int(win.LOWORD(uint32(nle.SendMessage(win.EM_CHARFROMPOS, 0, lParam))))

		if min := len(nle.prefix); i < min {
			nle.SetFocus()
			nle.SetTextSelection(min, min)
			return 0
		}
		if max := len(nle.textUTF16()) - len(nle.suffix); i > max {
			nle.SetFocus()
			nle.SetTextSelection(max, max)
			return 0
		}

	case win.WM_LBUTTONDBLCLK:
		nle.selectNumber()
		return 0

	case win.WM_MOUSEMOVE:
		i := int(win.LOWORD(uint32(nle.SendMessage(win.EM_CHARFROMPOS, 0, lParam))))

		if min := len(nle.prefix); i < min {
			return 0
		}
		if max := len(nle.textUTF16()) - len(nle.suffix); i > max {
			return 0
		}

	case win.WM_MOUSEWHEEL:
		delta := float64(int16(win.HIWORD(uint32(wParam))))
		nle.incrementValue(delta / 120 * nle.increment)
		return 0

	case win.WM_PASTE:
		ret := nle.LineEdit.WndProc(hwnd, msg, wParam, lParam)
		if !nle.tryUpdateValue(true) {
			nle.setTextFromValue(nle.value)
		}
		nle.selectNumber()
		return ret

	case win.WM_SETFOCUS:
		nle.selectNumber()

	case win.EM_SETSEL:
		start := int(wParam)
		end := int(lParam)
		adjusted := false
		if min := len(nle.prefix); start < min {
			start = min
			adjusted = true
		}
		if max := len(nle.textUTF16()) - len(nle.suffix); end < 0 || end > max {
			end = max
			adjusted = true
		}

		if adjusted {
			nle.SetTextSelection(start, end)
			return 0
		}
	}

	return nle.LineEdit.WndProc(hwnd, msg, wParam, lParam)
}

func (ne *NumberEdit) SetToolTipText(s string) error {
	return ne.edit.SetToolTipText(s)
}
