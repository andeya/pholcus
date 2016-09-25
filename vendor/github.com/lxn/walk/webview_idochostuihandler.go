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

var webViewIDocHostUIHandlerVtbl *win.IDocHostUIHandlerVtbl

func init() {
	webViewIDocHostUIHandlerVtbl = &win.IDocHostUIHandlerVtbl{
		syscall.NewCallback(webView_IDocHostUIHandler_QueryInterface),
		syscall.NewCallback(webView_IDocHostUIHandler_AddRef),
		syscall.NewCallback(webView_IDocHostUIHandler_Release),
		syscall.NewCallback(webView_IDocHostUIHandler_ShowContextMenu),
		syscall.NewCallback(webView_IDocHostUIHandler_GetHostInfo),
		syscall.NewCallback(webView_IDocHostUIHandler_ShowUI),
		syscall.NewCallback(webView_IDocHostUIHandler_HideUI),
		syscall.NewCallback(webView_IDocHostUIHandler_UpdateUI),
		syscall.NewCallback(webView_IDocHostUIHandler_EnableModeless),
		syscall.NewCallback(webView_IDocHostUIHandler_OnDocWindowActivate),
		syscall.NewCallback(webView_IDocHostUIHandler_OnFrameWindowActivate),
		syscall.NewCallback(webView_IDocHostUIHandler_ResizeBorder),
		syscall.NewCallback(webView_IDocHostUIHandler_TranslateAccelerator),
		syscall.NewCallback(webView_IDocHostUIHandler_GetOptionKeyPath),
		syscall.NewCallback(webView_IDocHostUIHandler_GetDropTarget),
		syscall.NewCallback(webView_IDocHostUIHandler_GetExternal),
		syscall.NewCallback(webView_IDocHostUIHandler_TranslateUrl),
		syscall.NewCallback(webView_IDocHostUIHandler_FilterDataObject),
	}
}

type webViewIDocHostUIHandler struct {
	win.IDocHostUIHandler
}

func webView_IDocHostUIHandler_QueryInterface(docHostUIHandler *webViewIDocHostUIHandler, riid win.REFIID, ppvObject *unsafe.Pointer) uintptr {
	// Just reuse the QueryInterface implementation we have for IOleClientSite.
	// We need to adjust object, which initially points at our
	// webViewIDocHostUIHandler, so it refers to the containing
	// webViewIOleClientSite for the call.
	var clientSite win.IOleClientSite
	var webViewInPlaceSite webViewIOleInPlaceSite

	ptr := uintptr(unsafe.Pointer(docHostUIHandler)) - uintptr(unsafe.Sizeof(clientSite)) -
		uintptr(unsafe.Sizeof(webViewInPlaceSite))

	return webView_IOleClientSite_QueryInterface((*webViewIOleClientSite)(unsafe.Pointer(ptr)), riid, ppvObject)
}

func webView_IDocHostUIHandler_AddRef(docHostUIHandler *webViewIDocHostUIHandler) uintptr {
	return 1
}

func webView_IDocHostUIHandler_Release(docHostUIHandler *webViewIDocHostUIHandler) uintptr {
	return 1
}

func webView_IDocHostUIHandler_ShowContextMenu(docHostUIHandler *webViewIDocHostUIHandler, dwID uint32, ppt *win.POINT, pcmdtReserved *win.IUnknown, pdispReserved uintptr) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_GetHostInfo(docHostUIHandler *webViewIDocHostUIHandler, pInfo *win.DOCHOSTUIINFO) uintptr {
	pInfo.CbSize = uint32(unsafe.Sizeof(*pInfo))
	pInfo.DwFlags = win.DOCHOSTUIFLAG_NO3DBORDER
	pInfo.DwDoubleClick = win.DOCHOSTUIDBLCLK_DEFAULT

	return win.S_OK
}

func webView_IDocHostUIHandler_ShowUI(docHostUIHandler *webViewIDocHostUIHandler, dwID uint32, pActiveObject uintptr, pCommandTarget uintptr, pFrame *win.IOleInPlaceFrame, pDoc uintptr) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_HideUI(docHostUIHandler *webViewIDocHostUIHandler) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_UpdateUI(docHostUIHandler *webViewIDocHostUIHandler) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_EnableModeless(docHostUIHandler *webViewIDocHostUIHandler, fEnable win.BOOL) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_OnDocWindowActivate(docHostUIHandler *webViewIDocHostUIHandler, fActivate win.BOOL) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_OnFrameWindowActivate(docHostUIHandler *webViewIDocHostUIHandler, fActivate win.BOOL) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_ResizeBorder(docHostUIHandler *webViewIDocHostUIHandler, prcBorder *win.RECT, pUIWindow uintptr, fRameWindow win.BOOL) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_TranslateAccelerator(docHostUIHandler *webViewIDocHostUIHandler, lpMsg *win.MSG, pguidCmdGroup *syscall.GUID, nCmdID uint) uintptr {
	return win.S_FALSE
}

func webView_IDocHostUIHandler_GetOptionKeyPath(docHostUIHandler *webViewIDocHostUIHandler, pchKey *uint16, dw uint) uintptr {
	return win.S_FALSE
}

func webView_IDocHostUIHandler_GetDropTarget(docHostUIHandler *webViewIDocHostUIHandler, pDropTarget uintptr, ppDropTarget *uintptr) uintptr {
	return win.S_FALSE
}

func webView_IDocHostUIHandler_GetExternal(docHostUIHandler *webViewIDocHostUIHandler, ppDispatch *uintptr) uintptr {
	*ppDispatch = 0

	return win.S_FALSE
}

func webView_IDocHostUIHandler_TranslateUrl(docHostUIHandler *webViewIDocHostUIHandler, dwTranslate uint32, pchURLIn *uint16, ppchURLOut **uint16) uintptr {
	*ppchURLOut = nil

	return win.S_FALSE
}

func webView_IDocHostUIHandler_FilterDataObject(docHostUIHandler *webViewIDocHostUIHandler, pDO uintptr, ppDORet *uintptr) uintptr {
	*ppDORet = 0

	return win.S_FALSE
}
