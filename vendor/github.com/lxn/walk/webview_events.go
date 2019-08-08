// Copyright 2011 The Walk Authors. All rights reserved.
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

type WebViewNavigatingEventData struct {
	pDisp           *win.IDispatch
	url             *win.VARIANT
	flags           *win.VARIANT
	targetFrameName *win.VARIANT
	postData        *win.VARIANT
	headers         *win.VARIANT
	cancel          *win.VARIANT_BOOL
}

func (eventData *WebViewNavigatingEventData) Url() string {
	url := eventData.url
	if url != nil && url.MustBSTR() != nil {
		return win.BSTRToString(url.MustBSTR())
	}
	return ""
}

func (eventData *WebViewNavigatingEventData) Flags() int32 {
	flags := eventData.flags
	if flags != nil {
		return flags.MustLong()
	}
	return 0
}

func (eventData *WebViewNavigatingEventData) PostData() string {
	postData := eventData.postData
	if postData != nil {
		pvar := postData.MustPVariant()
		if pvar != nil && pvar.Vt == win.VT_ARRAY|win.VT_UI1 {
			psa := pvar.MustPSafeArray()
			if psa != nil && psa.CDims == 1 && psa.CbElements == 1 {
				postDataSize := psa.Rgsabound[0].CElements * psa.CbElements
				byteAryPtr := (*[200000000]byte)(unsafe.Pointer(psa.PvData))
				byteArySlice := (*byteAryPtr)[0 : postDataSize-1]
				return string(byteArySlice)
			}
		}
	}
	return ""
}

func (eventData *WebViewNavigatingEventData) Headers() string {
	headers := eventData.headers
	if headers != nil && headers.MustBSTR() != nil {
		return win.BSTRToString(headers.MustBSTR())
	}
	return ""
}

func (eventData *WebViewNavigatingEventData) TargetFrameName() string {
	targetFrameName := eventData.targetFrameName
	if targetFrameName != nil && targetFrameName.MustBSTR() != nil {
		return win.BSTRToString(targetFrameName.MustBSTR())
	}
	return ""
}

func (eventData *WebViewNavigatingEventData) Canceled() bool {
	cancel := eventData.cancel
	if cancel != nil {
		if *cancel != win.VARIANT_FALSE {
			return true
		} else {
			return false
		}
	}
	return false
}

func (eventData *WebViewNavigatingEventData) SetCanceled(value bool) {
	cancel := eventData.cancel
	if cancel != nil {
		if value {
			*cancel = win.VARIANT_TRUE
		} else {
			*cancel = win.VARIANT_FALSE
		}
	}
}

type WebViewNavigatingEventHandler func(eventData *WebViewNavigatingEventData)

type WebViewNavigatingEvent struct {
	handlers []WebViewNavigatingEventHandler
}

func (e *WebViewNavigatingEvent) Attach(handler WebViewNavigatingEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *WebViewNavigatingEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type WebViewNavigatingEventPublisher struct {
	event WebViewNavigatingEvent
}

func (p *WebViewNavigatingEventPublisher) Event() *WebViewNavigatingEvent {
	return &p.event
}

func (p *WebViewNavigatingEventPublisher) Publish(eventData *WebViewNavigatingEventData) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(eventData)
		}
	}
}

type WebViewNavigatedErrorEventData struct {
	pDisp           *win.IDispatch
	url             *win.VARIANT
	targetFrameName *win.VARIANT
	statusCode      *win.VARIANT
	cancel          *win.VARIANT_BOOL
}

func (eventData *WebViewNavigatedErrorEventData) Url() string {
	url := eventData.url
	if url != nil && url.MustBSTR() != nil {
		return win.BSTRToString(url.MustBSTR())
	}
	return ""
}

func (eventData *WebViewNavigatedErrorEventData) TargetFrameName() string {
	targetFrameName := eventData.targetFrameName
	if targetFrameName != nil && targetFrameName.MustBSTR() != nil {
		return win.BSTRToString(targetFrameName.MustBSTR())
	}
	return ""
}

func (eventData *WebViewNavigatedErrorEventData) StatusCode() int32 {
	statusCode := eventData.statusCode
	if statusCode != nil {
		return statusCode.MustLong()
	}
	return 0
}

func (eventData *WebViewNavigatedErrorEventData) Canceled() bool {
	cancel := eventData.cancel
	if cancel != nil {
		if *cancel != win.VARIANT_FALSE {
			return true
		} else {
			return false
		}
	}
	return false
}

func (eventData *WebViewNavigatedErrorEventData) SetCanceled(value bool) {
	cancel := eventData.cancel
	if cancel != nil {
		if value {
			*cancel = win.VARIANT_TRUE
		} else {
			*cancel = win.VARIANT_FALSE
		}
	}
}

type WebViewNavigatedErrorEventHandler func(eventData *WebViewNavigatedErrorEventData)

type WebViewNavigatedErrorEvent struct {
	handlers []WebViewNavigatedErrorEventHandler
}

func (e *WebViewNavigatedErrorEvent) Attach(handler WebViewNavigatedErrorEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *WebViewNavigatedErrorEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type WebViewNavigatedErrorEventPublisher struct {
	event WebViewNavigatedErrorEvent
}

func (p *WebViewNavigatedErrorEventPublisher) Event() *WebViewNavigatedErrorEvent {
	return &p.event
}

func (p *WebViewNavigatedErrorEventPublisher) Publish(eventData *WebViewNavigatedErrorEventData) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(eventData)
		}
	}
}

type WebViewNewWindowEventData struct {
	ppDisp         **win.IDispatch
	cancel         *win.VARIANT_BOOL
	dwFlags        uint32
	bstrUrlContext *uint16
	bstrUrl        *uint16
}

func (eventData *WebViewNewWindowEventData) Canceled() bool {
	cancel := eventData.cancel
	if cancel != nil {
		if *cancel != win.VARIANT_FALSE {
			return true
		} else {
			return false
		}
	}
	return false
}

func (eventData *WebViewNewWindowEventData) SetCanceled(value bool) {
	cancel := eventData.cancel
	if cancel != nil {
		if value {
			*cancel = win.VARIANT_TRUE
		} else {
			*cancel = win.VARIANT_FALSE
		}
	}
}

func (eventData *WebViewNewWindowEventData) Flags() uint32 {
	return eventData.dwFlags
}

func (eventData *WebViewNewWindowEventData) UrlContext() string {
	bstrUrlContext := eventData.bstrUrlContext
	if bstrUrlContext != nil {
		return win.BSTRToString(bstrUrlContext)
	}
	return ""
}

func (eventData *WebViewNewWindowEventData) Url() string {
	bstrUrl := eventData.bstrUrl
	if bstrUrl != nil {
		return win.BSTRToString(bstrUrl)
	}
	return ""
}

type WebViewNewWindowEventHandler func(eventData *WebViewNewWindowEventData)

type WebViewNewWindowEvent struct {
	handlers []WebViewNewWindowEventHandler
}

func (e *WebViewNewWindowEvent) Attach(handler WebViewNewWindowEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *WebViewNewWindowEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type WebViewNewWindowEventPublisher struct {
	event WebViewNewWindowEvent
}

func (p *WebViewNewWindowEventPublisher) Event() *WebViewNewWindowEvent {
	return &p.event
}

func (p *WebViewNewWindowEventPublisher) Publish(eventData *WebViewNewWindowEventData) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(eventData)
		}
	}
}

type WebViewWindowClosingEventData struct {
	bIsChildWindow win.VARIANT_BOOL
	cancel         *win.VARIANT_BOOL
}

func (eventData *WebViewWindowClosingEventData) IsChildWindow() bool {
	bIsChildWindow := eventData.bIsChildWindow
	if bIsChildWindow != win.VARIANT_FALSE {
		return true
	} else {
		return false
	}
	return false
}

func (eventData *WebViewWindowClosingEventData) Canceled() bool {
	cancel := eventData.cancel
	if cancel != nil {
		if *cancel != win.VARIANT_FALSE {
			return true
		} else {
			return false
		}
	}
	return false
}

func (eventData *WebViewWindowClosingEventData) SetCanceled(value bool) {
	cancel := eventData.cancel
	if cancel != nil {
		if value {
			*cancel = win.VARIANT_TRUE
		} else {
			*cancel = win.VARIANT_FALSE
		}
	}
}

type WebViewWindowClosingEventHandler func(eventData *WebViewWindowClosingEventData)

type WebViewWindowClosingEvent struct {
	handlers []WebViewWindowClosingEventHandler
}

func (e *WebViewWindowClosingEvent) Attach(handler WebViewWindowClosingEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *WebViewWindowClosingEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type WebViewWindowClosingEventPublisher struct {
	event WebViewWindowClosingEvent
}

func (p *WebViewWindowClosingEventPublisher) Event() *WebViewWindowClosingEvent {
	return &p.event
}

func (p *WebViewWindowClosingEventPublisher) Publish(eventData *WebViewWindowClosingEventData) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(eventData)
		}
	}
}
