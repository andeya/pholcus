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

var webViewIOleClientSiteVtbl *win.IOleClientSiteVtbl

func init() {
	webViewIOleClientSiteVtbl = &win.IOleClientSiteVtbl{
		syscall.NewCallback(webView_IOleClientSite_QueryInterface),
		syscall.NewCallback(webView_IOleClientSite_AddRef),
		syscall.NewCallback(webView_IOleClientSite_Release),
		syscall.NewCallback(webView_IOleClientSite_SaveObject),
		syscall.NewCallback(webView_IOleClientSite_GetMoniker),
		syscall.NewCallback(webView_IOleClientSite_GetContainer),
		syscall.NewCallback(webView_IOleClientSite_ShowObject),
		syscall.NewCallback(webView_IOleClientSite_OnShowWindow),
		syscall.NewCallback(webView_IOleClientSite_RequestNewObjectLayout),
	}
}

type webViewIOleClientSite struct {
	win.IOleClientSite
	inPlaceSite       webViewIOleInPlaceSite
	docHostUIHandler  webViewIDocHostUIHandler
	webBrowserEvents2 webViewDWebBrowserEvents2
}

func webView_IOleClientSite_QueryInterface(clientSite *webViewIOleClientSite, riid win.REFIID, ppvObject *unsafe.Pointer) uintptr {
	if win.EqualREFIID(riid, &win.IID_IUnknown) {
		*ppvObject = unsafe.Pointer(clientSite)
	} else if win.EqualREFIID(riid, &win.IID_IOleClientSite) {
		*ppvObject = unsafe.Pointer(clientSite)
	} else if win.EqualREFIID(riid, &win.IID_IOleInPlaceSite) {
		*ppvObject = unsafe.Pointer(&clientSite.inPlaceSite)
	} else if win.EqualREFIID(riid, &win.IID_IDocHostUIHandler) {
		*ppvObject = unsafe.Pointer(&clientSite.docHostUIHandler)
	} else if win.EqualREFIID(riid, &win.DIID_DWebBrowserEvents2) {
		*ppvObject = unsafe.Pointer(&clientSite.webBrowserEvents2)
	} else {
		*ppvObject = nil
		return win.E_NOINTERFACE
	}

	return win.S_OK
}

func webView_IOleClientSite_AddRef(clientSite *webViewIOleClientSite) uintptr {
	return 1
}

func webView_IOleClientSite_Release(clientSite *webViewIOleClientSite) uintptr {
	return 1
}

func webView_IOleClientSite_SaveObject(clientSite *webViewIOleClientSite) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleClientSite_GetMoniker(clientSite *webViewIOleClientSite, dwAssign, dwWhichMoniker uint32, ppmk *unsafe.Pointer) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleClientSite_GetContainer(clientSite *webViewIOleClientSite, ppContainer *unsafe.Pointer) uintptr {
	*ppContainer = nil

	return win.E_NOINTERFACE
}

func webView_IOleClientSite_ShowObject(clientSite *webViewIOleClientSite) uintptr {
	return win.S_OK
}

func webView_IOleClientSite_OnShowWindow(clientSite *webViewIOleClientSite, fShow win.BOOL) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleClientSite_RequestNewObjectLayout(clientSite *webViewIOleClientSite) uintptr {
	return win.E_NOTIMPL
}
