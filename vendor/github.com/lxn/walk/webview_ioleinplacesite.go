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

var webViewIOleInPlaceSiteVtbl *win.IOleInPlaceSiteVtbl

func init() {
	webViewIOleInPlaceSiteVtbl = &win.IOleInPlaceSiteVtbl{
		syscall.NewCallback(webView_IOleInPlaceSite_QueryInterface),
		syscall.NewCallback(webView_IOleInPlaceSite_AddRef),
		syscall.NewCallback(webView_IOleInPlaceSite_Release),
		syscall.NewCallback(webView_IOleInPlaceSite_GetWindow),
		syscall.NewCallback(webView_IOleInPlaceSite_ContextSensitiveHelp),
		syscall.NewCallback(webView_IOleInPlaceSite_CanInPlaceActivate),
		syscall.NewCallback(webView_IOleInPlaceSite_OnInPlaceActivate),
		syscall.NewCallback(webView_IOleInPlaceSite_OnUIActivate),
		syscall.NewCallback(webView_IOleInPlaceSite_GetWindowContext),
		syscall.NewCallback(webView_IOleInPlaceSite_Scroll),
		syscall.NewCallback(webView_IOleInPlaceSite_OnUIDeactivate),
		syscall.NewCallback(webView_IOleInPlaceSite_OnInPlaceDeactivate),
		syscall.NewCallback(webView_IOleInPlaceSite_DiscardUndoState),
		syscall.NewCallback(webView_IOleInPlaceSite_DeactivateAndUndo),
		syscall.NewCallback(webView_IOleInPlaceSite_OnPosRectChange),
	}
}

type webViewIOleInPlaceSite struct {
	win.IOleInPlaceSite
	inPlaceFrame webViewIOleInPlaceFrame
}

func webView_IOleInPlaceSite_QueryInterface(inPlaceSite *webViewIOleInPlaceSite, riid win.REFIID, ppvObject *unsafe.Pointer) uintptr {
	// Just reuse the QueryInterface implementation we have for IOleClientSite.
	// We need to adjust object from the webViewIDocHostUIHandler to the
	// containing webViewIOleInPlaceSite.
	var clientSite win.IOleClientSite

	ptr := uintptr(unsafe.Pointer(inPlaceSite)) - uintptr(unsafe.Sizeof(clientSite))

	return webView_IOleClientSite_QueryInterface((*webViewIOleClientSite)(unsafe.Pointer(ptr)), riid, ppvObject)
}

func webView_IOleInPlaceSite_AddRef(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return 1
}

func webView_IOleInPlaceSite_Release(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return 1
}

func webView_IOleInPlaceSite_GetWindow(inPlaceSite *webViewIOleInPlaceSite, lphwnd *win.HWND) uintptr {
	*lphwnd = inPlaceSite.inPlaceFrame.webView.hWnd

	return win.S_OK
}

func webView_IOleInPlaceSite_ContextSensitiveHelp(inPlaceSite *webViewIOleInPlaceSite, fEnterMode win.BOOL) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceSite_CanInPlaceActivate(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_OnInPlaceActivate(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_OnUIActivate(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_GetWindowContext(inPlaceSite *webViewIOleInPlaceSite, lplpFrame **webViewIOleInPlaceFrame, lplpDoc *uintptr, lprcPosRect, lprcClipRect *win.RECT, lpFrameInfo *win.OLEINPLACEFRAMEINFO) uintptr {
	*lplpFrame = &inPlaceSite.inPlaceFrame
	*lplpDoc = 0

	lpFrameInfo.FMDIApp = win.FALSE
	lpFrameInfo.HwndFrame = inPlaceSite.inPlaceFrame.webView.hWnd
	lpFrameInfo.Haccel = 0
	lpFrameInfo.CAccelEntries = 0

	return win.S_OK
}

func webView_IOleInPlaceSite_Scroll(inPlaceSite *webViewIOleInPlaceSite, scrollExtentX, scrollExtentY int32) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceSite_OnUIDeactivate(inPlaceSite *webViewIOleInPlaceSite, fUndoable win.BOOL) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_OnInPlaceDeactivate(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_DiscardUndoState(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceSite_DeactivateAndUndo(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceSite_OnPosRectChange(inPlaceSite *webViewIOleInPlaceSite, lprcPosRect *win.RECT) uintptr {
	browserObject := inPlaceSite.inPlaceFrame.webView.browserObject
	var inPlaceObjectPtr unsafe.Pointer
	if hr := browserObject.QueryInterface(&win.IID_IOleInPlaceObject, &inPlaceObjectPtr); win.FAILED(hr) {
		return uintptr(hr)
	}
	inPlaceObject := (*win.IOleInPlaceObject)(inPlaceObjectPtr)
	defer inPlaceObject.Release()

	return uintptr(inPlaceObject.SetObjectRects(lprcPosRect, lprcPosRect))
}
