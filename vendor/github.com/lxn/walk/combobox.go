// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"math/big"
	"syscall"
	"time"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

type ComboBox struct {
	WidgetBase
	bindingValueProvider         BindingValueProvider
	model                        ListModel
	providedModel                interface{}
	bindingMember                string
	displayMember                string
	format                       string
	precision                    int
	itemsResetHandlerHandle      int
	itemChangedHandlerHandle     int
	maxItemTextWidth             int
	prevCurIndex                 int
	selChangeIndex               int
	maxLength                    int
	currentIndexChangedPublisher EventPublisher
	textChangedPublisher         EventPublisher
	editOrigWndProcPtr           uintptr
}

var comboBoxEditWndProcPtr = syscall.NewCallback(comboBoxEditWndProc)

func comboBoxEditWndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	cb := (*ComboBox)(unsafe.Pointer(win.GetWindowLongPtr(hwnd, win.GWLP_USERDATA)))

	switch msg {
	case win.WM_GETDLGCODE:
		if form := ancestor(cb); form != nil {
			if dlg, ok := form.(dialogish); ok {
				if dlg.DefaultButton() != nil {
					// If the ComboBox lives in a Dialog that has a
					// DefaultButton, we won't swallow the return key.
					break
				}
			}
		}

		if wParam == win.VK_RETURN {
			return win.DLGC_WANTALLKEYS
		}

	case win.WM_KEYDOWN:
		if wParam != win.VK_RETURN || 0 == cb.SendMessage(win.CB_GETDROPPEDSTATE, 0, 0) {
			cb.handleKeyDown(wParam, lParam)
		}

	case win.WM_KEYUP:
		if wParam != win.VK_RETURN || 0 == cb.SendMessage(win.CB_GETDROPPEDSTATE, 0, 0) {
			cb.handleKeyUp(wParam, lParam)
		}
	}

	return win.CallWindowProc(cb.editOrigWndProcPtr, hwnd, msg, wParam, lParam)
}

func NewComboBox(parent Container) (*ComboBox, error) {
	cb, err := newComboBoxWithStyle(parent, win.CBS_DROPDOWN)
	if err != nil {
		return nil, err
	}

	editHwnd := win.GetWindow(cb.hWnd, win.GW_CHILD)

	win.SetWindowLongPtr(editHwnd, win.GWLP_USERDATA, uintptr(unsafe.Pointer(cb)))
	cb.editOrigWndProcPtr = win.SetWindowLongPtr(editHwnd, win.GWLP_WNDPROC, comboBoxEditWndProcPtr)

	return cb, nil
}

func NewDropDownBox(parent Container) (*ComboBox, error) {
	return newComboBoxWithStyle(parent, win.CBS_DROPDOWNLIST)
}

func newComboBoxWithStyle(parent Container, style uint32) (*ComboBox, error) {
	cb := &ComboBox{prevCurIndex: -1, selChangeIndex: -1, precision: 2}

	if err := InitWidget(
		cb,
		parent,
		"COMBOBOX",
		win.WS_TABSTOP|win.WS_VISIBLE|win.WS_VSCROLL|style,
		0); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			cb.Dispose()
		}
	}()

	cb.MustRegisterProperty("CurrentIndex", NewProperty(
		func() interface{} {
			return cb.CurrentIndex()
		},
		func(v interface{}) error {
			return cb.SetCurrentIndex(v.(int))
		},
		cb.CurrentIndexChanged()))

	cb.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return cb.Text()
		},
		func(v interface{}) error {
			return cb.SetText(v.(string))
		},
		cb.TextChanged()))

	cb.MustRegisterProperty("HasCurrentItem", NewReadOnlyBoolProperty(
		func() bool {
			return cb.CurrentIndex() != -1
		},
		cb.CurrentIndexChanged()))

	cb.MustRegisterProperty("TextNotEmpty", NewReadOnlyBoolProperty(
		func() bool {
			return cb.Text() != ""
		},
		cb.CurrentIndexChanged()))

	cb.MustRegisterProperty("Value", NewProperty(
		func() interface{} {
			if cb.Editable() {
				return cb.Text()
			}

			index := cb.CurrentIndex()

			if cb.bindingValueProvider == nil || index == -1 {
				return nil
			}

			return cb.bindingValueProvider.BindingValue(index)
		},
		func(v interface{}) error {
			if cb.Editable() {
				return cb.SetText(v.(string))
			}

			if cb.bindingValueProvider == nil {
				if cb.model == nil {
					return nil
				} else {
					return newError("Data binding is only supported using a model that implements BindingValueProvider.")
				}
			}

			index := -1

			count := cb.model.ItemCount()
			for i := 0; i < count; i++ {
				if cb.bindingValueProvider.BindingValue(i) == v {
					index = i
					break
				}
			}

			return cb.SetCurrentIndex(index)
		},
		cb.CurrentIndexChanged()))

	succeeded = true

	return cb, nil
}

func (cb *ComboBox) LayoutFlags() LayoutFlags {
	if cb.Editable() {
		return GrowableHorz | GreedyHorz
	}

	return GrowableHorz
}

func (cb *ComboBox) MinSizeHint() Size {
	defaultSize := cb.dialogBaseUnitsToPixels(Size{30, 12})

	if cb.model != nil && cb.maxItemTextWidth <= 0 {
		cb.maxItemTextWidth = cb.calculateMaxItemTextWidth()
	}

	// FIXME: Use GetThemePartSize instead of guessing
	w := maxi(defaultSize.Width, cb.maxItemTextWidth+30)
	h := defaultSize.Height + 1

	return Size{w, h}
}

func (cb *ComboBox) SizeHint() Size {
	return cb.MinSizeHint()
}

func (cb *ComboBox) applyFont(font *Font) {
	cb.WidgetBase.applyFont(font)

	if cb.model != nil {
		cb.maxItemTextWidth = cb.calculateMaxItemTextWidth()
		cb.updateParentLayout()
	}
}

func (cb *ComboBox) Editable() bool {
	return !cb.hasStyleBits(win.CBS_DROPDOWNLIST)
}

func (cb *ComboBox) itemString(index int) string {
	switch val := cb.model.Value(index).(type) {
	case string:
		return val

	case time.Time:
		return val.Format(cb.format)

	case *big.Rat:
		return val.FloatString(cb.precision)

	default:
		return fmt.Sprintf(cb.format, val)
	}

	panic("unreachable")
}

func (cb *ComboBox) insertItemAt(index int) error {
	str := cb.itemString(index)
	lp := uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(str)))

	if win.CB_ERR == cb.SendMessage(win.CB_INSERTSTRING, uintptr(index), lp) {
		return newError("SendMessage(CB_INSERTSTRING)")
	}

	return nil
}

func (cb *ComboBox) resetItems() error {
	cb.SetSuspended(true)
	defer cb.SetSuspended(false)

	cb.selChangeIndex = -1

	if win.FALSE == cb.SendMessage(win.CB_RESETCONTENT, 0, 0) {
		return newError("SendMessage(CB_RESETCONTENT)")
	}

	cb.maxItemTextWidth = 0

	cb.SetCurrentIndex(-1)

	if cb.model == nil {
		return nil
	}

	count := cb.model.ItemCount()

	for i := 0; i < count; i++ {
		if err := cb.insertItemAt(i); err != nil {
			return err
		}
	}

	cb.updateParentLayout()

	return nil
}

func (cb *ComboBox) attachModel() {
	itemsResetHandler := func() {
		cb.resetItems()
	}
	cb.itemsResetHandlerHandle = cb.model.ItemsReset().Attach(itemsResetHandler)

	itemChangedHandler := func(index int) {
		if win.CB_ERR == cb.SendMessage(win.CB_DELETESTRING, uintptr(index), 0) {
			newError("SendMessage(CB_DELETESTRING)")
		}

		cb.insertItemAt(index)

		cb.SetCurrentIndex(cb.prevCurIndex)
	}
	cb.itemChangedHandlerHandle = cb.model.ItemChanged().Attach(itemChangedHandler)
}

func (cb *ComboBox) detachModel() {
	cb.model.ItemsReset().Detach(cb.itemsResetHandlerHandle)
	cb.model.ItemChanged().Detach(cb.itemChangedHandlerHandle)
}

// Model returns the model of the ComboBox.
func (cb *ComboBox) Model() interface{} {
	return cb.providedModel
}

// SetModel sets the model of the ComboBox.
//
// It is required that mdl either implements walk.ListModel or
// walk.ReflectListModel or be a slice of pointers to struct or a []string.
func (cb *ComboBox) SetModel(mdl interface{}) error {
	model, ok := mdl.(ListModel)
	if !ok && mdl != nil {
		var err error
		if model, err = newReflectListModel(mdl); err != nil {
			return err
		}

		if _, ok := mdl.([]string); !ok {
			if badms, ok := model.(bindingAndDisplayMemberSetter); ok {
				var bindingMember string
				if cb.Editable() {
					bindingMember = cb.displayMember
				} else {
					bindingMember = cb.bindingMember
				}
				badms.setBindingMember(bindingMember)
				badms.setDisplayMember(cb.displayMember)
			}
		}
	}
	cb.providedModel = mdl

	if cb.model != nil {
		cb.detachModel()
	}

	cb.model = model
	cb.bindingValueProvider, _ = model.(BindingValueProvider)

	if model != nil {
		cb.attachModel()
	}

	if err := cb.resetItems(); err != nil {
		return err
	}

	if !cb.Editable() && model != nil && model.ItemCount() == 1 {
		cb.SetCurrentIndex(0)
	}

	return cb.Invalidate()
}

// BindingMember returns the member from the model of the ComboBox that is bound
// to a field of the data source managed by an associated DataBinder.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
func (cb *ComboBox) BindingMember() string {
	return cb.bindingMember
}

// SetBindingMember sets the member from the model of the ComboBox that is bound
// to a field of the data source managed by an associated DataBinder.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
//
// For a model consisting of items of type S, data source field of type T and
// bindingMember "Foo", this can be one of the following:
//
//	A field		Foo T
//	A method	func (s S) Foo() T
//	A method	func (s S) Foo() (T, error)
//
// If bindingMember is not a simple member name like "Foo", but a path to a
// member like "A.B.Foo", members "A" and "B" both must be one of the options
// mentioned above, but with T having type pointer to struct.
func (cb *ComboBox) SetBindingMember(bindingMember string) error {
	if bindingMember != "" {
		if _, ok := cb.providedModel.([]string); ok {
			return newError("invalid for []string model")
		}
	}

	cb.bindingMember = bindingMember

	if badms, ok := cb.model.(bindingAndDisplayMemberSetter); ok {
		badms.setBindingMember(bindingMember)
	}

	return nil
}

// DisplayMember returns the member from the model of the ComboBox that is
// displayed in the ComboBox.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
func (cb *ComboBox) DisplayMember() string {
	return cb.displayMember
}

// SetDisplayMember sets the member from the model of the ComboBox that is
// displayed in the ComboBox.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
//
// For a model consisting of items of type S, the type of the specified member T
// and displayMember "Foo", this can be one of the following:
//
//	A field		Foo T
//	A method	func (s S) Foo() T
//	A method	func (s S) Foo() (T, error)
//
// If displayMember is not a simple member name like "Foo", but a path to a
// member like "A.B.Foo", members "A" and "B" both must be one of the options
// mentioned above, but with T having type pointer to struct.
func (cb *ComboBox) SetDisplayMember(displayMember string) error {
	if displayMember != "" {
		if _, ok := cb.providedModel.([]string); ok {
			return newError("invalid for []string model")
		}
	}

	cb.displayMember = displayMember

	if badms, ok := cb.model.(bindingAndDisplayMemberSetter); ok {
		badms.setDisplayMember(displayMember)
	}

	return nil
}

func (cb *ComboBox) Format() string {
	return cb.format
}

func (cb *ComboBox) SetFormat(value string) {
	cb.format = value
}

func (cb *ComboBox) Precision() int {
	return cb.precision
}

func (cb *ComboBox) SetPrecision(value int) {
	cb.precision = value
}

func (cb *ComboBox) MaxLength() int {
	return cb.maxLength
}

func (cb *ComboBox) SetMaxLength(value int) {
	cb.SendMessage(win.CB_LIMITTEXT, uintptr(value), 0)

	cb.maxLength = value
}

func (cb *ComboBox) calculateMaxItemTextWidth() int {
	hdc := win.GetDC(cb.hWnd)
	if hdc == 0 {
		newError("GetDC failed")
		return -1
	}
	defer win.ReleaseDC(cb.hWnd, hdc)

	hFontOld := win.SelectObject(hdc, win.HGDIOBJ(cb.Font().handleForDPI(0)))
	defer win.SelectObject(hdc, hFontOld)

	var maxWidth int

	count := cb.model.ItemCount()
	for i := 0; i < count; i++ {
		var s win.SIZE
		str := syscall.StringToUTF16(cb.itemString(i))

		if !win.GetTextExtentPoint32(hdc, &str[0], int32(len(str)-1), &s) {
			newError("GetTextExtentPoint32 failed")
			return -1
		}

		maxWidth = maxi(maxWidth, int(s.CX))
	}

	return maxWidth
}

func (cb *ComboBox) CurrentIndex() int {
	return int(int32(cb.SendMessage(win.CB_GETCURSEL, 0, 0)))
}

func (cb *ComboBox) SetCurrentIndex(value int) error {
	index := int(int32(cb.SendMessage(win.CB_SETCURSEL, uintptr(value), 0)))

	if index != value {
		return newError("invalid index")
	}

	if value != cb.prevCurIndex {
		cb.prevCurIndex = value
		cb.currentIndexChangedPublisher.Publish()
	}

	return nil
}

func (cb *ComboBox) CurrentIndexChanged() *Event {
	return cb.currentIndexChangedPublisher.Event()
}

func (cb *ComboBox) Text() string {
	return windowText(cb.hWnd)
}

func (cb *ComboBox) SetText(value string) error {
	return setWindowText(cb.hWnd, value)
}

func (cb *ComboBox) TextSelection() (start, end int) {
	cb.SendMessage(win.CB_GETEDITSEL, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
	return
}

func (cb *ComboBox) SetTextSelection(start, end int) {
	cb.SendMessage(win.CB_SETEDITSEL, 0, uintptr(win.MAKELONG(uint16(start), uint16(end))))
}

func (cb *ComboBox) TextChanged() *Event {
	return cb.textChangedPublisher.Event()
}

func (cb *ComboBox) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_COMMAND:
		code := win.HIWORD(uint32(wParam))
		selIndex := cb.CurrentIndex()

		switch code {
		case win.CBN_EDITCHANGE:
			cb.selChangeIndex = -1
			cb.textChangedPublisher.Publish()

		case win.CBN_SELCHANGE:
			cb.selChangeIndex = selIndex

		case win.CBN_SELENDCANCEL:
			if cb.selChangeIndex != -1 {
				if cb.selChangeIndex < cb.model.ItemCount() {
					cb.SetCurrentIndex(cb.selChangeIndex)
				}

				cb.selChangeIndex = -1
			}

		case win.CBN_SELENDOK:
			if editable := cb.Editable(); editable || selIndex != cb.prevCurIndex {
				if editable && selIndex > -1 {
					cb.Property("Value").Set(cb.model.Value(selIndex))
				}
				cb.currentIndexChangedPublisher.Publish()
				cb.prevCurIndex = selIndex
				return 0
			}

			cb.selChangeIndex = -1
		}

	case win.WM_MOUSEWHEEL:
		if !cb.Enabled() {
			return 0
		}

	case win.WM_SIZE:
		if cb.Editable() {
			result := cb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)

			cb.SetTextSelection(0, 0)

			return result
		}
	}

	return cb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
