// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"strconv"
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

const tabWidgetWindowClass = `\o/ Walk_TabWidget_Class \o/`

func init() {
	MustRegisterWindowClass(tabWidgetWindowClass)
}

type TabWidget struct {
	WidgetBase
	hWndTab                      win.HWND
	imageList                    *ImageList
	pages                        *TabPageList
	currentIndex                 int
	currentIndexChangedPublisher EventPublisher
	persistent                   bool
}

func NewTabWidget(parent Container) (*TabWidget, error) {
	tw := &TabWidget{currentIndex: -1}
	tw.pages = newTabPageList(tw)

	if err := InitWidget(
		tw,
		parent,
		tabWidgetWindowClass,
		win.WS_VISIBLE,
		win.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			tw.Dispose()
		}
	}()

	tw.SetPersistent(true)

	tw.hWndTab = win.CreateWindowEx(
		0, syscall.StringToUTF16Ptr("SysTabControl32"), nil,
		win.WS_CHILD|win.WS_CLIPSIBLINGS|win.WS_TABSTOP|win.WS_VISIBLE,
		0, 0, 0, 0, tw.hWnd, 0, 0, nil)
	if tw.hWndTab == 0 {
		return nil, lastError("CreateWindowEx")
	}
	win.SendMessage(tw.hWndTab, win.WM_SETFONT, uintptr(defaultFont.handleForDPI(0)), 1)

	setWindowFont(tw.hWndTab, tw.Font())

	tw.MustRegisterProperty("HasCurrentPage", NewReadOnlyBoolProperty(
		func() bool {
			return tw.CurrentIndex() != -1
		},
		tw.CurrentIndexChanged()))

	succeeded = true

	return tw, nil
}

func (tw *TabWidget) Dispose() {
	tw.WidgetBase.Dispose()

	if tw.imageList != nil {
		tw.imageList.Dispose()
		tw.imageList = nil
	}
}

func (tw *TabWidget) LayoutFlags() LayoutFlags {
	if tw.pages.Len() == 0 {
		return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
	}

	var flags LayoutFlags

	for i := tw.pages.Len() - 1; i >= 0; i-- {
		flags |= tw.pages.At(i).LayoutFlags()
	}

	return flags
}

func (tw *TabWidget) MinSizeHint() Size {
	if tw.pages.Len() == 0 {
		return tw.SizeHint()
	}

	var min Size

	for i := tw.pages.Len() - 1; i >= 0; i-- {
		s := tw.pages.At(i).MinSizeHint()

		min.Width = maxi(min.Width, s.Width)
		min.Height = maxi(min.Height, s.Height)
	}

	b := tw.Bounds()
	pb := tw.pages.At(0).Bounds()

	size := Size{b.Width - pb.Width + min.Width, b.Height - pb.Height + min.Height}

	return size

}

func (tw *TabWidget) SizeHint() Size {
	return Size{100, 100}
}

func (tw *TabWidget) applyEnabled(enabled bool) {
	tw.WidgetBase.applyEnabled(enabled)

	setWindowEnabled(tw.hWndTab, enabled)

	applyEnabledToDescendants(tw, enabled)
}

func (tw *TabWidget) applyFont(font *Font) {
	tw.WidgetBase.applyFont(font)

	setWindowFont(tw.hWndTab, font)

	applyFontToDescendants(tw, font)
}

func (tw *TabWidget) CurrentIndex() int {
	return tw.currentIndex
}

func (tw *TabWidget) SetCurrentIndex(index int) error {
	if index == tw.currentIndex {
		return nil
	}

	if index < 0 || index >= tw.pages.Len() {
		return newError("invalid index")
	}

	ret := int(win.SendMessage(tw.hWndTab, win.TCM_SETCURSEL, uintptr(index), 0))
	if ret == -1 {
		return newError("SendMessage(TCM_SETCURSEL) failed")
	}

	// FIXME: The SendMessage(TCM_SETCURSEL) call above doesn't cause a
	// TCN_SELCHANGE notification, so we use this workaround.
	tw.onSelChange()

	return nil
}

func (tw *TabWidget) CurrentIndexChanged() *Event {
	return tw.currentIndexChangedPublisher.Event()
}

func (tw *TabWidget) Pages() *TabPageList {
	return tw.pages
}

func (tw *TabWidget) Persistent() bool {
	return tw.persistent
}

func (tw *TabWidget) SetPersistent(value bool) {
	tw.persistent = value
}

func (tw *TabWidget) SaveState() error {
	tw.putState(strconv.Itoa(tw.CurrentIndex()))

	for _, page := range tw.pages.items {
		if err := page.SaveState(); err != nil {
			return err
		}
	}

	return nil
}

func (tw *TabWidget) RestoreState() error {
	state, err := tw.getState()
	if err != nil {
		return err
	}
	if state == "" {
		return nil
	}

	index, err := strconv.Atoi(state)
	if err != nil {
		return err
	}
	if index >= 0 && index < tw.pages.Len() {
		if err := tw.SetCurrentIndex(index); err != nil {
			return err
		}
	}

	for _, page := range tw.pages.items {
		if err := page.RestoreState(); err != nil {
			return err
		}
	}

	return nil
}

func (tw *TabWidget) resizePages() {
	var r win.RECT
	if !win.GetWindowRect(tw.hWndTab, &r) {
		lastError("GetWindowRect")
		return
	}

	p := win.POINT{
		r.Left,
		r.Top,
	}
	if !win.ScreenToClient(tw.hWnd, &p) {
		newError("ScreenToClient failed")
		return
	}

	r = win.RECT{
		p.X,
		p.Y,
		r.Right - r.Left + p.X,
		r.Bottom - r.Top + p.Y,
	}
	win.SendMessage(tw.hWndTab, win.TCM_ADJUSTRECT, 0, uintptr(unsafe.Pointer(&r)))

	for _, page := range tw.pages.items {
		if err := page.SetBounds(
			Rectangle{
				int(r.Left - 2),
				int(r.Top),
				int(r.Right - r.Left + 2),
				int(r.Bottom - r.Top),
			}); err != nil {

			return
		}
	}
}

func (tw *TabWidget) onResize(lParam uintptr) {
	r := win.RECT{0, 0, win.GET_X_LPARAM(lParam), win.GET_Y_LPARAM(lParam)}
	if !win.MoveWindow(tw.hWndTab, r.Left, r.Top, r.Right-r.Left, r.Bottom-r.Top, true) {
		lastError("MoveWindow")
		return
	}

	tw.resizePages()
}

func (tw *TabWidget) onSelChange() {
	pageCount := tw.pages.Len()

	if tw.currentIndex > -1 && tw.currentIndex < pageCount {
		page := tw.pages.At(tw.currentIndex)
		page.SetVisible(false)
	}

	tw.currentIndex = int(win.SendMessage(tw.hWndTab, win.TCM_GETCURSEL, 0, 0))

	if tw.currentIndex > -1 && tw.currentIndex < pageCount {
		page := tw.pages.At(tw.currentIndex)
		page.SetVisible(true)
		page.Invalidate()
	}

	tw.currentIndexChangedPublisher.Publish()
}

func (tw *TabWidget) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	if tw.hWndTab != 0 {
		switch msg {
		case win.WM_SIZE, win.WM_SIZING:
			tw.onResize(lParam)

		case win.WM_NOTIFY:
			nmhdr := (*win.NMHDR)(unsafe.Pointer(lParam))

			switch int32(nmhdr.Code) {
			case win.TCN_SELCHANGE:
				tw.onSelChange()
			}
		}
	}

	return tw.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (tw *TabWidget) onPageChanged(page *TabPage) (err error) {
	index := tw.pages.Index(page)
	item := tw.tcitemFromPage(page)

	if 0 == win.SendMessage(tw.hWndTab, win.TCM_SETITEM, uintptr(index), uintptr(unsafe.Pointer(item))) {
		return newError("SendMessage(TCM_SETITEM) failed")
	}

	return nil
}

func (tw *TabWidget) onInsertingPage(index int, page *TabPage) (err error) {
	return nil
}

func (tw *TabWidget) onInsertedPage(index int, page *TabPage) (err error) {
	item := tw.tcitemFromPage(page)

	if idx := int(win.SendMessage(tw.hWndTab, win.TCM_INSERTITEM, uintptr(index), uintptr(unsafe.Pointer(item)))); idx == -1 {
		return newError("SendMessage(TCM_INSERTITEM) failed")
	}

	page.SetVisible(false)

	style := uint32(win.GetWindowLong(page.hWnd, win.GWL_STYLE))
	if style == 0 {
		return lastError("GetWindowLong")
	}

	style |= win.WS_CHILD
	style &^= win.WS_POPUP

	win.SetLastError(0)
	if win.SetWindowLong(page.hWnd, win.GWL_STYLE, int32(style)) == 0 {
		return lastError("SetWindowLong")
	}

	if win.SetParent(page.hWnd, tw.hWnd) == 0 {
		return lastError("SetParent")
	}

	if tw.pages.Len() == 1 {
		page.SetVisible(true)
		tw.SetCurrentIndex(0)
	}

	tw.resizePages()

	page.tabWidget = tw

	page.applyFont(tw.Font())

	return
}

func (tw *TabWidget) removePage(page *TabPage) (err error) {
	page.SetVisible(false)

	style := uint32(win.GetWindowLong(page.hWnd, win.GWL_STYLE))
	if style == 0 {
		return lastError("GetWindowLong")
	}

	style &^= win.WS_CHILD
	style |= win.WS_POPUP

	win.SetLastError(0)
	if win.SetWindowLong(page.hWnd, win.GWL_STYLE, int32(style)) == 0 {
		return lastError("SetWindowLong")
	}

	page.tabWidget = nil

	return page.SetParent(nil)
}

func (tw *TabWidget) onRemovingPage(index int, page *TabPage) (err error) {
	return nil
}

func (tw *TabWidget) onRemovedPage(index int, page *TabPage) (err error) {
	err = tw.removePage(page)
	if err != nil {
		return
	}

	win.SendMessage(tw.hWndTab, win.TCM_DELETEITEM, uintptr(index), 0)

	if tw.pages.Len() > 0 {
		tw.currentIndex = 0
		win.SendMessage(tw.hWndTab, win.TCM_SETCURSEL, uintptr(tw.currentIndex), 0)
	} else {
		tw.currentIndex = -1
	}
	tw.onSelChange()

	return

	// FIXME: Either make use of this unreachable code or remove it.
	if index == tw.currentIndex {
		// removal of current visible tabpage...
		tw.currentIndex = -1

		// select new tabpage if any :
		if tw.pages.Len() > 0 {
			// are we removing the rightmost page ?
			if index == tw.pages.Len()-1 {
				// If so, select the page on the left
				index -= 1
			}
		}
	}

	tw.SetCurrentIndex(index)
	//tw.Invalidate()

	return
}

func (tw *TabWidget) onClearingPages(pages []*TabPage) (err error) {
	return nil
}

func (tw *TabWidget) onClearedPages(pages []*TabPage) (err error) {
	win.SendMessage(tw.hWndTab, win.TCM_DELETEALLITEMS, 0, 0)
	for _, page := range pages {
		tw.removePage(page)
	}
	tw.currentIndex = -1
	return nil
}

func (tw *TabWidget) tcitemFromPage(page *TabPage) *win.TCITEM {
	imageIndex, _ := tw.imageIndex(page.image)
	text := syscall.StringToUTF16(page.title)

	item := &win.TCITEM{
		Mask:       win.TCIF_IMAGE | win.TCIF_TEXT,
		IImage:     imageIndex,
		PszText:    &text[0],
		CchTextMax: int32(len(text)),
	}

	return item
}

func (tw *TabWidget) imageIndex(image *Bitmap) (index int32, err error) {
	index = -1
	if image != nil {
		if tw.imageList == nil {
			if tw.imageList, err = NewImageList(Size{16, 16}, 0); err != nil {
				return
			}

			win.SendMessage(tw.hWndTab, win.TCM_SETIMAGELIST, 0, uintptr(tw.imageList.hIml))
		}

		// FIXME: Protect against duplicate insertion
		if index, err = tw.imageList.AddMasked(image); err != nil {
			return
		}
	}

	return
}
