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

const webViewWindowClass = `\o/ Walk_WebView_Class \o/`

func init() {
	MustRegisterWindowClass(webViewWindowClass)
}

type WebView struct {
	WidgetBase
	clientSite          webViewIOleClientSite // IMPORTANT: Must remain first member after WidgetBase
	browserObject       *win.IOleObject
	urlChangedPublisher EventPublisher
}

func NewWebView(parent Container) (*WebView, error) {
	if hr := win.OleInitialize(); hr != win.S_OK && hr != win.S_FALSE {
		return nil, newError(fmt.Sprint("OleInitialize Error: ", hr))
	}

	wv := &WebView{
		clientSite: webViewIOleClientSite{
			IOleClientSite: win.IOleClientSite{
				LpVtbl: webViewIOleClientSiteVtbl,
			},
			inPlaceSite: webViewIOleInPlaceSite{
				IOleInPlaceSite: win.IOleInPlaceSite{
					LpVtbl: webViewIOleInPlaceSiteVtbl,
				},
				inPlaceFrame: webViewIOleInPlaceFrame{
					IOleInPlaceFrame: win.IOleInPlaceFrame{
						LpVtbl: webViewIOleInPlaceFrameVtbl,
					},
				},
			},
			docHostUIHandler: webViewIDocHostUIHandler{
				IDocHostUIHandler: win.IDocHostUIHandler{
					LpVtbl: webViewIDocHostUIHandlerVtbl,
				},
			},
			webBrowserEvents2: webViewDWebBrowserEvents2{
				DWebBrowserEvents2: win.DWebBrowserEvents2{
					LpVtbl: webViewDWebBrowserEvents2Vtbl,
				},
			},
		},
	}

	if err := InitWidget(
		wv,
		parent,
		webViewWindowClass,
		win.WS_CLIPCHILDREN|win.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	wv.clientSite.inPlaceSite.inPlaceFrame.webView = wv

	succeeded := false

	defer func() {
		if !succeeded {
			wv.Dispose()
		}
	}()

	var classFactoryPtr unsafe.Pointer
	if hr := win.CoGetClassObject(&win.CLSID_WebBrowser, win.CLSCTX_INPROC_HANDLER|win.CLSCTX_INPROC_SERVER, nil, &win.IID_IClassFactory, &classFactoryPtr); win.FAILED(hr) {
		return nil, errorFromHRESULT("CoGetClassObject", hr)
	}
	classFactory := (*win.IClassFactory)(classFactoryPtr)
	defer classFactory.Release()

	var browserObjectPtr unsafe.Pointer
	if hr := classFactory.CreateInstance(nil, &win.IID_IOleObject, &browserObjectPtr); win.FAILED(hr) {
		return nil, errorFromHRESULT("IClassFactory.CreateInstance", hr)
	}
	browserObject := (*win.IOleObject)(browserObjectPtr)

	wv.browserObject = browserObject

	if hr := browserObject.SetClientSite((*win.IOleClientSite)(unsafe.Pointer(&wv.clientSite))); win.FAILED(hr) {
		return nil, errorFromHRESULT("IOleObject.SetClientSite", hr)
	}

	if hr := browserObject.SetHostNames(syscall.StringToUTF16Ptr("Walk.WebView"), nil); win.FAILED(hr) {
		return nil, errorFromHRESULT("IOleObject.SetHostNames", hr)
	}

	if hr := win.OleSetContainedObject((*win.IUnknown)(unsafe.Pointer(browserObject)), true); win.FAILED(hr) {
		return nil, errorFromHRESULT("OleSetContainedObject", hr)
	}

	var rect win.RECT
	win.GetClientRect(wv.hWnd, &rect)

	if hr := browserObject.DoVerb(win.OLEIVERB_SHOW, nil, (*win.IOleClientSite)(unsafe.Pointer(&wv.clientSite)), -1, wv.hWnd, &rect); win.FAILED(hr) {
		return nil, errorFromHRESULT("IOleObject.DoVerb", hr)
	}

	var cpcPtr unsafe.Pointer
	if hr := browserObject.QueryInterface(&win.IID_IConnectionPointContainer, &cpcPtr); win.FAILED(hr) {
		return nil, errorFromHRESULT("IOleObject.QueryInterface(IID_IConnectionPointContainer)", hr)
	}
	cpc := (*win.IConnectionPointContainer)(cpcPtr)
	defer cpc.Release()

	var cp *win.IConnectionPoint
	if hr := cpc.FindConnectionPoint(&win.DIID_DWebBrowserEvents2, &cp); win.FAILED(hr) {
		return nil, errorFromHRESULT("IConnectionPointContainer.FindConnectionPoint(DIID_DWebBrowserEvents2)", hr)
	}
	defer cp.Release()

	var cookie uint32
	if hr := cp.Advise(unsafe.Pointer(&wv.clientSite.webBrowserEvents2), &cookie); win.FAILED(hr) {
		return nil, errorFromHRESULT("IConnectionPoint.Advise", hr)
	}

	wv.onResize()

	wv.MustRegisterProperty("URL", NewProperty(
		func() interface{} {
			url, _ := wv.URL()
			return url
		},
		func(v interface{}) error {
			return wv.SetURL(v.(string))
		},
		wv.urlChangedPublisher.Event()))

	succeeded = true

	return wv, nil
}

func (wv *WebView) Dispose() {
	if wv.browserObject != nil {
		wv.browserObject.Close(win.OLECLOSE_NOSAVE)
		wv.browserObject.Release()

		wv.browserObject = nil

		win.OleUninitialize()
	}

	wv.WidgetBase.Dispose()
}

func (*WebView) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (*WebView) SizeHint() Size {
	return Size{100, 100}
}

func (wv *WebView) URL() (url string, err error) {
	err = wv.withWebBrowser2(func(webBrowser2 *win.IWebBrowser2) error {
		var urlBstr *uint16 /*BSTR*/
		if hr := webBrowser2.Get_LocationURL(&urlBstr); win.FAILED(hr) {
			return errorFromHRESULT("IWebBrowser2.Get_LocationURL", hr)
		}
		defer win.SysFreeString(urlBstr)

		url = win.BSTRToString(urlBstr)

		return nil
	})

	return
}

func (wv *WebView) SetURL(url string) error {
	return wv.withWebBrowser2(func(webBrowser2 *win.IWebBrowser2) error {
		urlBstr := win.StringToVariantBSTR(url)
		flags := win.IntToVariantI4(0)
		targetFrameName := win.StringToVariantBSTR("_self")

		if hr := webBrowser2.Navigate2(urlBstr, flags, targetFrameName, nil, nil); win.FAILED(hr) {
			return errorFromHRESULT("IWebBrowser2.Navigate2", hr)
		}

		return nil
	})
}

func (wv *WebView) URLChanged() *Event {
	return wv.urlChangedPublisher.Event()
}

func (wv *WebView) Refresh() error {
	return wv.withWebBrowser2(func(webBrowser2 *win.IWebBrowser2) error {
		if hr := webBrowser2.Refresh(); win.FAILED(hr) {
			return errorFromHRESULT("IWebBrowser2.Refresh", hr)
		}

		return nil
	})
}

func (wv *WebView) withWebBrowser2(f func(webBrowser2 *win.IWebBrowser2) error) error {
	var webBrowser2Ptr unsafe.Pointer
	if hr := wv.browserObject.QueryInterface(&win.IID_IWebBrowser2, &webBrowser2Ptr); win.FAILED(hr) {
		return errorFromHRESULT("IOleObject.QueryInterface", hr)
	}
	webBrowser2 := (*win.IWebBrowser2)(webBrowser2Ptr)
	defer webBrowser2.Release()

	return f(webBrowser2)
}

func (wv *WebView) onResize() {
	// FIXME: handle error?
	wv.withWebBrowser2(func(webBrowser2 *win.IWebBrowser2) error {
		bounds := wv.ClientBounds()

		webBrowser2.Put_Left(0)
		webBrowser2.Put_Top(0)
		webBrowser2.Put_Width(int32(bounds.Width))
		webBrowser2.Put_Height(int32(bounds.Height))

		return nil
	})
}

func (wv *WebView) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_SIZE, win.WM_SIZING:
		if wv.clientSite.inPlaceSite.inPlaceFrame.webView == nil {
			break
		}

		wv.onResize()
	}

	return wv.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
