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

var webViewDWebBrowserEvents2Vtbl *win.DWebBrowserEvents2Vtbl

func init() {
	webViewDWebBrowserEvents2Vtbl = &win.DWebBrowserEvents2Vtbl{
		syscall.NewCallback(webView_DWebBrowserEvents2_QueryInterface),
		syscall.NewCallback(webView_DWebBrowserEvents2_AddRef),
		syscall.NewCallback(webView_DWebBrowserEvents2_Release),
		syscall.NewCallback(webView_DWebBrowserEvents2_GetTypeInfoCount),
		syscall.NewCallback(webView_DWebBrowserEvents2_GetTypeInfo),
		syscall.NewCallback(webView_DWebBrowserEvents2_GetIDsOfNames),
		syscall.NewCallback(webView_DWebBrowserEvents2_Invoke),
	}
}

type webViewDWebBrowserEvents2 struct {
	win.DWebBrowserEvents2
}

func webView_DWebBrowserEvents2_QueryInterface(wbe2 *webViewDWebBrowserEvents2, riid win.REFIID, ppvObject *unsafe.Pointer) uintptr {
	// Just reuse the QueryInterface implementation we have for IOleClientSite.
	// We need to adjust object, which initially points at our
	// webViewDWebBrowserEvents2, so it refers to the containing
	// webViewIOleClientSite for the call.
	var clientSite win.IOleClientSite
	var webViewInPlaceSite webViewIOleInPlaceSite
	var docHostUIHandler webViewIDocHostUIHandler

	ptr := uintptr(unsafe.Pointer(wbe2)) -
		uintptr(unsafe.Sizeof(clientSite)) -
		uintptr(unsafe.Sizeof(webViewInPlaceSite)) -
		uintptr(unsafe.Sizeof(docHostUIHandler))

	return webView_IOleClientSite_QueryInterface((*webViewIOleClientSite)(unsafe.Pointer(ptr)), riid, ppvObject)
}

func webView_DWebBrowserEvents2_AddRef(args *uintptr) uintptr {
	return 1
}

func webView_DWebBrowserEvents2_Release(args *uintptr) uintptr {
	return 1
}

func webView_DWebBrowserEvents2_GetTypeInfoCount(args *uintptr) uintptr {
	/*	p := (*struct {
			wbe2    *webViewDWebBrowserEvents2
			pctinfo *uint
		})(unsafe.Pointer(args))

		*p.pctinfo = 0

		return S_OK*/

	return win.E_NOTIMPL
}

func webView_DWebBrowserEvents2_GetTypeInfo(args *uintptr) uintptr {
	/*	p := (*struct {
				wbe2         *webViewDWebBrowserEvents2
			})(unsafe.Pointer(args))

		    unsigned int  iTInfo,
		    LCID  lcid,
		    ITypeInfo FAR* FAR*  ppTInfo*/

	return win.E_NOTIMPL
}

func webView_DWebBrowserEvents2_GetIDsOfNames(args *uintptr) uintptr {
	/*	p := (*struct {
		wbe2      *webViewDWebBrowserEvents2
		riid      REFIID
		rgszNames **uint16
		cNames    uint32
		lcid      LCID
		rgDispId  *DISPID
	})(unsafe.Pointer(args))*/

	return win.E_NOTIMPL
}

func webView_DWebBrowserEvents2_Invoke(
	wbe2 *webViewDWebBrowserEvents2,
	dispIdMember win.DISPID,
	riid win.REFIID,
	lcid uint32, // LCID
	wFlags uint16,
	pDispParams *win.DISPPARAMS,
	pVarResult *win.VARIANT,
	pExcepInfo unsafe.Pointer, // *EXCEPINFO
	puArgErr *uint32) uintptr {

	var wb WidgetBase
	var wvcs webViewIOleClientSite

	wv := (*WebView)(unsafe.Pointer(uintptr(unsafe.Pointer(wbe2)) +
		uintptr(unsafe.Sizeof(*wbe2)) -
		uintptr(unsafe.Sizeof(wvcs)) -
		uintptr(unsafe.Sizeof(wb))))

	switch dispIdMember {
	case win.DISPID_NAVIGATECOMPLETE2:
		wv.urlChangedPublisher.Publish()
	}

	return win.DISP_E_MEMBERNOTFOUND
}
