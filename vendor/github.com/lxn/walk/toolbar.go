// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

type ToolBarButtonStyle int

const (
	ToolBarButtonImageOnly ToolBarButtonStyle = iota
	ToolBarButtonTextOnly
	ToolBarButtonImageBeforeText
	ToolBarButtonImageAboveText
)

type ToolBar struct {
	WidgetBase
	imageList          *ImageList
	actions            *ActionList
	defaultButtonWidth int
	maxTextRows        int
	buttonStyle        ToolBarButtonStyle
}

func NewToolBarWithOrientationAndButtonStyle(parent Container, orientation Orientation, buttonStyle ToolBarButtonStyle) (*ToolBar, error) {
	var style uint32
	if orientation == Vertical {
		style = win.CCS_VERT | win.CCS_NORESIZE
	} else {
		style = win.TBSTYLE_WRAPABLE
	}

	if buttonStyle != ToolBarButtonImageAboveText {
		style |= win.TBSTYLE_LIST
	}

	tb := &ToolBar{buttonStyle: buttonStyle}
	tb.actions = newActionList(tb)

	if orientation == Vertical {
		tb.defaultButtonWidth = 100
	}

	if err := InitWidget(
		tb,
		parent,
		"ToolbarWindow32",
		win.CCS_NODIVIDER|win.TBSTYLE_FLAT|win.TBSTYLE_TOOLTIPS|style,
		0); err != nil {
		return nil, err
	}

	exStyle := tb.SendMessage(win.TB_GETEXTENDEDSTYLE, 0, 0)
	exStyle |= win.TBSTYLE_EX_DRAWDDARROWS | win.TBSTYLE_EX_MIXEDBUTTONS
	tb.SendMessage(win.TB_SETEXTENDEDSTYLE, 0, exStyle)

	return tb, nil
}

func NewToolBar(parent Container) (*ToolBar, error) {
	return NewToolBarWithOrientationAndButtonStyle(parent, Horizontal, ToolBarButtonImageOnly)
}

func NewVerticalToolBar(parent Container) (*ToolBar, error) {
	return NewToolBarWithOrientationAndButtonStyle(parent, Vertical, ToolBarButtonImageAboveText)
}

func (tb *ToolBar) LayoutFlags() LayoutFlags {
	style := win.GetWindowLong(tb.hWnd, win.GWL_STYLE)

	if style&win.CCS_VERT > 0 {
		return ShrinkableVert | GrowableVert | GreedyVert
	}

	// FIXME: Since reimplementation of BoxLayout we must return 0 here,
	// otherwise the ToolBar contained in MainWindow will eat half the space.
	return 0 //ShrinkableHorz | GrowableHorz
}

func (tb *ToolBar) MinSizeHint() Size {
	return tb.SizeHint()
}

func (tb *ToolBar) SizeHint() Size {
	if tb.actions.Len() == 0 {
		return Size{}
	}

	size := uint32(tb.SendMessage(win.TB_GETBUTTONSIZE, 0, 0))

	width := tb.defaultButtonWidth
	if width == 0 {
		width = int(win.LOWORD(size))
	}

	height := int(win.HIWORD(size))

	return Size{width, height}
}

func (tb *ToolBar) ButtonStyle() ToolBarButtonStyle {
	return tb.buttonStyle
}

func (tb *ToolBar) applyDefaultButtonWidth() error {
	if tb.defaultButtonWidth == 0 {
		return nil
	}

	lParam := uintptr(
		win.MAKELONG(uint16(tb.defaultButtonWidth), uint16(tb.defaultButtonWidth)))
	if 0 == tb.SendMessage(win.TB_SETBUTTONWIDTH, 0, lParam) {
		return newError("SendMessage(TB_SETBUTTONWIDTH)")
	}

	size := uint32(tb.SendMessage(win.TB_GETBUTTONSIZE, 0, 0))
	height := win.HIWORD(size)

	lParam = uintptr(win.MAKELONG(uint16(tb.defaultButtonWidth), height))
	if win.FALSE == tb.SendMessage(win.TB_SETBUTTONSIZE, 0, lParam) {
		return newError("SendMessage(TB_SETBUTTONSIZE)")
	}

	return nil
}

// DefaultButtonWidth returns the default button width of the ToolBar.
//
// The default value for a horizontal ToolBar is 0, resulting in automatic
// sizing behavior. For a vertical ToolBar, the default is 100 pixels.
func (tb *ToolBar) DefaultButtonWidth() int {
	return tb.defaultButtonWidth
}

// SetDefaultButtonWidth sets the default button width of the ToolBar.
//
// Calling this method affects all buttons in the ToolBar, no matter if they are
// added before or after the call. A width of 0 results in automatic sizing
// behavior. Negative values are not allowed.
func (tb *ToolBar) SetDefaultButtonWidth(width int) error {
	if width == tb.defaultButtonWidth {
		return nil
	}

	if width < 0 {
		return newError("width must be >= 0")
	}

	old := tb.defaultButtonWidth

	tb.defaultButtonWidth = width

	for _, action := range tb.actions.actions {
		if err := tb.onActionChanged(action); err != nil {
			tb.defaultButtonWidth = old

			return err
		}
	}

	return tb.applyDefaultButtonWidth()
}

func (tb *ToolBar) MaxTextRows() int {
	return tb.maxTextRows
}

func (tb *ToolBar) SetMaxTextRows(maxTextRows int) error {
	if 0 == tb.SendMessage(win.TB_SETMAXTEXTROWS, uintptr(maxTextRows), 0) {
		return newError("SendMessage(TB_SETMAXTEXTROWS)")
	}

	tb.maxTextRows = maxTextRows

	return nil
}

func (tb *ToolBar) Actions() *ActionList {
	return tb.actions
}

func (tb *ToolBar) ImageList() *ImageList {
	return tb.imageList
}

func (tb *ToolBar) SetImageList(value *ImageList) {
	var hIml win.HIMAGELIST

	if tb.buttonStyle != ToolBarButtonTextOnly && value != nil {
		hIml = value.hIml
	}

	tb.SendMessage(win.TB_SETIMAGELIST, 0, uintptr(hIml))

	tb.imageList = value
}

func (tb *ToolBar) imageIndex(image *Bitmap) (imageIndex int32, err error) {
	imageIndex = -1
	if image != nil {
		// FIXME: Protect against duplicate insertion
		if imageIndex, err = tb.imageList.AddMasked(image); err != nil {
			return
		}
	}

	return
}

func (tb *ToolBar) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case win.BN_CLICKED:
			actionId := uint16(win.LOWORD(uint32(wParam)))
			if action, ok := actionsById[actionId]; ok {
				action.raiseTriggered()
				return 0
			}
		}

	case win.WM_NOTIFY:
		nmhdr := (*win.NMHDR)(unsafe.Pointer(lParam))

		switch int32(nmhdr.Code) {
		case win.TBN_DROPDOWN:
			nmtb := (*win.NMTOOLBAR)(unsafe.Pointer(lParam))
			actionId := uint16(nmtb.IItem)
			if action := actionsById[actionId]; action != nil {
				var r win.RECT
				if 0 == tb.SendMessage(win.TB_GETRECT, uintptr(actionId), uintptr(unsafe.Pointer(&r))) {
					break
				}

				p := win.POINT{r.Left, r.Bottom}

				if !win.ClientToScreen(tb.hWnd, &p) {
					break
				}

				win.TrackPopupMenuEx(
					action.menu.hMenu,
					win.TPM_NOANIMATION,
					p.X,
					p.Y,
					tb.hWnd,
					nil)

				return win.TBDDRET_DEFAULT
			}
		}
	}

	return tb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (tb *ToolBar) initButtonForAction(action *Action, state, style *byte, image *int32, text *uintptr) (err error) {
	if tb.hasStyleBits(win.CCS_VERT) {
		*state |= win.TBSTATE_WRAP
	} else if tb.defaultButtonWidth == 0 {
		*style |= win.BTNS_AUTOSIZE
	}

	if action.checked {
		*state |= win.TBSTATE_CHECKED
	}

	if action.enabled {
		*state |= win.TBSTATE_ENABLED
	}

	if action.checkable {
		*style |= win.BTNS_CHECK
	}

	if action.exclusive {
		*style |= win.BTNS_GROUP
	}

	if tb.buttonStyle != ToolBarButtonImageOnly {
		*style |= win.BTNS_SHOWTEXT
	}

	if action.menu != nil {
		if len(action.Triggered().handlers) > 0 {
			*style |= win.BTNS_DROPDOWN
		} else {
			*style |= win.BTNS_WHOLEDROPDOWN
		}
	}

	if action.IsSeparator() {
		*style = win.BTNS_SEP
	}

	if tb.buttonStyle != ToolBarButtonTextOnly {
		if *image, err = tb.imageIndex(action.image); err != nil {
			return
		}
	}

	var actionText string
	if s := action.shortcut; tb.buttonStyle == ToolBarButtonImageOnly && s.Key != 0 {
		actionText = fmt.Sprintf("%s (%s)", action.Text(), s.String())
	} else {
		actionText = action.Text()
	}

	*text = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(actionText)))

	return
}

func (tb *ToolBar) onActionChanged(action *Action) error {
	tbbi := win.TBBUTTONINFO{
		DwMask: win.TBIF_IMAGE | win.TBIF_STATE | win.TBIF_STYLE | win.TBIF_TEXT,
	}

	tbbi.CbSize = uint32(unsafe.Sizeof(tbbi))

	if err := tb.initButtonForAction(
		action,
		&tbbi.FsState,
		&tbbi.FsStyle,
		&tbbi.IImage,
		&tbbi.PszText); err != nil {

		return err
	}

	if 0 == tb.SendMessage(
		win.TB_SETBUTTONINFO,
		uintptr(action.id),
		uintptr(unsafe.Pointer(&tbbi))) {

		return newError("SendMessage(TB_SETBUTTONINFO) failed")
	}

	return nil
}

func (tb *ToolBar) onActionVisibleChanged(action *Action) error {
	if !action.IsSeparator() {
		defer tb.actions.updateSeparatorVisibility()
	}

	if action.Visible() {
		return tb.insertAction(action, true)
	}

	return tb.removeAction(action, true)
}

func (tb *ToolBar) insertAction(action *Action, visibleChanged bool) (err error) {
	if !visibleChanged {
		action.addChangedHandler(tb)
		defer func() {
			if err != nil {
				action.removeChangedHandler(tb)
			}
		}()
	}

	if !action.Visible() {
		return
	}

	index := tb.actions.indexInObserver(action)

	tbb := win.TBBUTTON{
		IdCommand: int32(action.id),
	}

	if err = tb.initButtonForAction(
		action,
		&tbb.FsState,
		&tbb.FsStyle,
		&tbb.IBitmap,
		&tbb.IString); err != nil {

		return
	}

	tb.SetVisible(true)

	tb.SendMessage(win.TB_BUTTONSTRUCTSIZE, uintptr(unsafe.Sizeof(tbb)), 0)

	if win.FALSE == tb.SendMessage(win.TB_INSERTBUTTON, uintptr(index), uintptr(unsafe.Pointer(&tbb))) {
		return newError("SendMessage(TB_ADDBUTTONS)")
	}

	if err = tb.applyDefaultButtonWidth(); err != nil {
		return
	}

	tb.SendMessage(win.TB_AUTOSIZE, 0, 0)

	return
}

func (tb *ToolBar) removeAction(action *Action, visibleChanged bool) error {
	index := tb.actions.indexInObserver(action)

	if !visibleChanged {
		action.removeChangedHandler(tb)
	}

	if 0 == tb.SendMessage(win.TB_DELETEBUTTON, uintptr(index), 0) {
		return newError("SendMessage(TB_DELETEBUTTON) failed")
	}

	return nil
}

func (tb *ToolBar) onInsertedAction(action *Action) error {
	return tb.insertAction(action, false)
}

func (tb *ToolBar) onRemovingAction(action *Action) error {
	return tb.removeAction(action, false)
}

func (tb *ToolBar) onClearingActions() error {
	for i := tb.actions.Len() - 1; i >= 0; i-- {
		if action := tb.actions.At(i); action.Visible() {
			if err := tb.onRemovingAction(action); err != nil {
				return err
			}
		}
	}

	return nil
}
