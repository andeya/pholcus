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

var webViewIOleInPlaceFrameVtbl *win.IOleInPlaceFrameVtbl

func init() {
	webViewIOleInPlaceFrameVtbl = &win.IOleInPlaceFrameVtbl{
		syscall.NewCallback(webView_IOleInPlaceFrame_QueryInterface),
		syscall.NewCallback(webView_IOleInPlaceFrame_AddRef),
		syscall.NewCallback(webView_IOleInPlaceFrame_Release),
		syscall.NewCallback(webView_IOleInPlaceFrame_GetWindow),
		syscall.NewCallback(webView_IOleInPlaceFrame_ContextSensitiveHelp),
		syscall.NewCallback(webView_IOleInPlaceFrame_GetBorder),
		syscall.NewCallback(webView_IOleInPlaceFrame_RequestBorderSpace),
		syscall.NewCallback(webView_IOleInPlaceFrame_SetBorderSpace),
		syscall.NewCallback(webView_IOleInPlaceFrame_SetActiveObject),
		syscall.NewCallback(webView_IOleInPlaceFrame_InsertMenus),
		syscall.NewCallback(webView_IOleInPlaceFrame_SetMenu),
		syscall.NewCallback(webView_IOleInPlaceFrame_RemoveMenus),
		syscall.NewCallback(webView_IOleInPlaceFrame_SetStatusText),
		syscall.NewCallback(webView_IOleInPlaceFrame_EnableModeless),
		syscall.NewCallback(webView_IOleInPlaceFrame_TranslateAccelerator),
	}
}

type webViewIOleInPlaceFrame struct {
	win.IOleInPlaceFrame
	webView *WebView
}

func webView_IOleInPlaceFrame_QueryInterface(inPlaceFrame *webViewIOleInPlaceFrame, riid win.REFIID, ppvObj *uintptr) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_AddRef(inPlaceFrame *webViewIOleInPlaceFrame) uintptr {
	return 1
}

func webView_IOleInPlaceFrame_Release(inPlaceFrame *webViewIOleInPlaceFrame) uintptr {
	return 1
}

func webView_IOleInPlaceFrame_GetWindow(inPlaceFrame *webViewIOleInPlaceFrame, lphwnd *win.HWND) uintptr {
	*lphwnd = inPlaceFrame.webView.hWnd

	return win.S_OK
}

func webView_IOleInPlaceFrame_ContextSensitiveHelp(inPlaceFrame *webViewIOleInPlaceFrame, fEnterMode win.BOOL) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_GetBorder(inPlaceFrame *webViewIOleInPlaceFrame, lprectBorder *win.RECT) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_RequestBorderSpace(inPlaceFrame *webViewIOleInPlaceFrame, pborderwidths uintptr) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_SetBorderSpace(inPlaceFrame *webViewIOleInPlaceFrame, pborderwidths uintptr) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_SetActiveObject(inPlaceFrame *webViewIOleInPlaceFrame, pActiveObject uintptr, pszObjName *uint16) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceFrame_InsertMenus(inPlaceFrame *webViewIOleInPlaceFrame, hmenuShared win.HMENU, lpMenuWidths uintptr) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_SetMenu(inPlaceFrame *webViewIOleInPlaceFrame, hmenuShared win.HMENU, holemenu win.HMENU, hwndActiveObject win.HWND) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceFrame_RemoveMenus(inPlaceFrame *webViewIOleInPlaceFrame, hmenuShared win.HMENU) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_SetStatusText(inPlaceFrame *webViewIOleInPlaceFrame, pszStatusText *uint16) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceFrame_EnableModeless(inPlaceFrame *webViewIOleInPlaceFrame, fEnable win.BOOL) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceFrame_TranslateAccelerator(inPlaceFrame *webViewIOleInPlaceFrame, lpmsg *win.MSG, wID uint32) uintptr {
	return win.E_NOTIMPL
}
