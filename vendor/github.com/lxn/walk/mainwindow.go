// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"unsafe"
)

import (
	"github.com/lxn/win"
)

const mainWindowWindowClass = `\o/ Walk_MainWindow_Class \o/`

func init() {
	MustRegisterWindowClass(mainWindowWindowClass)
}

type MainWindow struct {
	FormBase
	windowPlacement *win.WINDOWPLACEMENT
	menu            *Menu
	toolBar         *ToolBar
	statusBar       *StatusBar
}

func NewMainWindow() (*MainWindow, error) {
	mw := new(MainWindow)

	if err := InitWindow(
		mw,
		nil,
		mainWindowWindowClass,
		win.WS_OVERLAPPEDWINDOW,
		win.WS_EX_CONTROLPARENT); err != nil {

		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			mw.Dispose()
		}
	}()

	mw.SetPersistent(true)

	var err error

	if mw.menu, err = newMenuBar(mw.hWnd); err != nil {
		return nil, err
	}
	if !win.SetMenu(mw.hWnd, mw.menu.hMenu) {
		return nil, lastError("SetMenu")
	}

	tb, err := NewToolBar(mw)
	if err != nil {
		return nil, err
	}
	mw.SetToolBar(tb)

	if mw.statusBar, err = NewStatusBar(mw); err != nil {
		return nil, err
	}
	mw.statusBar.parent = nil
	mw.Children().Remove(mw.statusBar)
	mw.statusBar.parent = mw
	win.SetParent(mw.statusBar.hWnd, mw.hWnd)
	mw.statusBar.visibleChangedPublisher.event.Attach(func() {
		mw.SetBounds(mw.Bounds())
	})

	// This forces display of focus rectangles, as soon as the user starts to type.
	mw.SendMessage(win.WM_CHANGEUISTATE, win.UIS_INITIALIZE, 0)

	succeeded = true

	return mw, nil
}

func (mw *MainWindow) Menu() *Menu {
	return mw.menu
}

func (mw *MainWindow) ToolBar() *ToolBar {
	return mw.toolBar
}

func (mw *MainWindow) SetToolBar(tb *ToolBar) {
	if mw.toolBar != nil {
		win.SetParent(mw.toolBar.hWnd, 0)
	}

	if tb != nil {
		parent := tb.parent
		tb.parent = nil
		parent.Children().Remove(tb)
		tb.parent = mw
		win.SetParent(tb.hWnd, mw.hWnd)
	}

	mw.toolBar = tb
}

func (mw *MainWindow) StatusBar() *StatusBar {
	return mw.statusBar
}

func (mw *MainWindow) ClientBounds() Rectangle {
	bounds := mw.FormBase.ClientBounds()

	if mw.toolBar != nil && mw.toolBar.Actions().Len() > 0 {
		tlbBounds := mw.toolBar.Bounds()

		bounds.Y += tlbBounds.Height
		bounds.Height -= tlbBounds.Height
	}

	if mw.statusBar.Visible() {
		bounds.Height -= mw.statusBar.Height()
	}

	return bounds
}

func (mw *MainWindow) SetVisible(visible bool) {
	if visible {
		win.DrawMenuBar(mw.hWnd)

		if mw.clientComposite.layout != nil {
			mw.clientComposite.layout.Update(false)
		}
	}

	mw.FormBase.SetVisible(visible)
}

func (mw *MainWindow) applyFont(font *Font) {
	mw.FormBase.applyFont(font)

	if mw.toolBar != nil {
		mw.toolBar.applyFont(font)
	}

	if mw.statusBar != nil {
		mw.statusBar.applyFont(font)
	}
}

func (mw *MainWindow) Fullscreen() bool {
	return win.GetWindowLong(mw.hWnd, win.GWL_STYLE)&win.WS_OVERLAPPEDWINDOW == 0
}

func (mw *MainWindow) SetFullscreen(fullscreen bool) error {
	if fullscreen == mw.Fullscreen() {
		return nil
	}

	if fullscreen {
		var mi win.MONITORINFO
		mi.CbSize = uint32(unsafe.Sizeof(mi))

		if mw.windowPlacement == nil {
			mw.windowPlacement = new(win.WINDOWPLACEMENT)
		}

		if !win.GetWindowPlacement(mw.hWnd, mw.windowPlacement) {
			return lastError("GetWindowPlacement")
		}
		if !win.GetMonitorInfo(win.MonitorFromWindow(
			mw.hWnd, win.MONITOR_DEFAULTTOPRIMARY), &mi) {

			return newError("GetMonitorInfo")
		}

		if err := mw.ensureStyleBits(win.WS_OVERLAPPEDWINDOW, false); err != nil {
			return err
		}

		if r := mi.RcMonitor; !win.SetWindowPos(
			mw.hWnd, win.HWND_TOP,
			r.Left, r.Top, r.Right-r.Left, r.Bottom-r.Top,
			win.SWP_FRAMECHANGED|win.SWP_NOOWNERZORDER) {

			return lastError("SetWindowPos")
		}
	} else {
		if err := mw.ensureStyleBits(win.WS_OVERLAPPEDWINDOW, true); err != nil {
			return err
		}

		if !win.SetWindowPlacement(mw.hWnd, mw.windowPlacement) {
			return lastError("SetWindowPlacement")
		}

		if !win.SetWindowPos(mw.hWnd, 0, 0, 0, 0, 0, win.SWP_FRAMECHANGED|win.SWP_NOMOVE|
			win.SWP_NOOWNERZORDER|win.SWP_NOSIZE|win.SWP_NOZORDER) {

			return lastError("SetWindowPos")
		}
	}

	return nil
}

func (mw *MainWindow) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_SIZE, win.WM_SIZING:
		cb := mw.ClientBounds()

		if mw.toolBar != nil {
			mw.toolBar.SetBounds(Rectangle{0, 0, cb.Width, mw.toolBar.Height()})
		}

		mw.statusBar.SetBounds(Rectangle{0, cb.Y + cb.Height, cb.Width, mw.statusBar.Height()})
	}

	return mw.FormBase.WndProc(hwnd, msg, wParam, lParam)
}
