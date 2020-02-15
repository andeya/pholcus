// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

type LinkLabel struct {
	WidgetBase
	textChangedPublisher   EventPublisher
	linkActivatedPublisher LinkLabelLinkEventPublisher
}

func NewLinkLabel(parent Container) (*LinkLabel, error) {
	ll := new(LinkLabel)

	if err := InitWidget(
		ll,
		parent,
		"SysLink",
		win.WS_TABSTOP|win.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	ll.SetBackground(nullBrushSingleton)

	ll.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return ll.Text()
		},
		func(v interface{}) error {
			return ll.SetText(assertStringOr(v, ""))
		},
		ll.textChangedPublisher.Event()))

	return ll, nil
}

func (ll *LinkLabel) MinSizeHint() Size {
	var s win.SIZE

	ll.SendMessage(win.LM_GETIDEALSIZE, uintptr(ll.maxSize.Width), uintptr(unsafe.Pointer(&s)))

	return Size{int(s.CX), int(s.CY)}
}

func (ll *LinkLabel) SizeHint() Size {
	return ll.MinSizeHint()
}

func (ll *LinkLabel) Text() string {
	return ll.text()
}

func (ll *LinkLabel) SetText(value string) error {
	if value == ll.Text() {
		return nil
	}

	if err := ll.setText(value); err != nil {
		return err
	}

	return ll.updateParentLayout()
}

func (ll *LinkLabel) LinkActivated() *LinkLabelLinkEvent {
	return ll.linkActivatedPublisher.Event()
}

func (ll *LinkLabel) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_NOTIFY:
		nml := (*win.NMLINK)(unsafe.Pointer(lParam))

		switch nml.Hdr.Code {
		case win.NM_CLICK, win.NM_RETURN:
			link := &LinkLabelLink{
				ll:    ll,
				index: int(nml.Item.ILink),
				id:    syscall.UTF16ToString(nml.Item.SzID[:]),
				url:   syscall.UTF16ToString(nml.Item.SzUrl[:]),
			}

			ll.linkActivatedPublisher.Publish(link)
		}

	case win.WM_KILLFOCUS:
		ll.ensureStyleBits(win.WS_TABSTOP, true)

	case win.WM_SETTEXT:
		ll.textChangedPublisher.Publish()

	case win.WM_SIZE, win.WM_SIZING:
		ll.Invalidate()
	}

	return ll.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

type LinkLabelLinkEventHandler func(link *LinkLabelLink)

type LinkLabelLinkEvent struct {
	handlers []LinkLabelLinkEventHandler
}

func (e *LinkLabelLinkEvent) Attach(handler LinkLabelLinkEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *LinkLabelLinkEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type LinkLabelLinkEventPublisher struct {
	event LinkLabelLinkEvent
}

func (p *LinkLabelLinkEventPublisher) Event() *LinkLabelLinkEvent {
	return &p.event
}

func (p *LinkLabelLinkEventPublisher) Publish(link *LinkLabelLink) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(link)
		}
	}
}

type LinkLabelLink struct {
	ll    *LinkLabel
	index int
	id    string
	url   string
}

func (lll *LinkLabelLink) Index() int {
	return lll.index
}

func (lll *LinkLabelLink) Id() string {
	return lll.id
}

func (lll *LinkLabelLink) URL() string {
	return lll.url
}

func (lll *LinkLabelLink) Enabled() (bool, error) {
	return lll.hasState(win.LIS_ENABLED)
}

func (lll *LinkLabelLink) SetEnabled(enabled bool) error {
	return lll.setState(win.LIS_ENABLED, enabled)
}

func (lll *LinkLabelLink) Focused() (bool, error) {
	return lll.hasState(win.LIS_FOCUSED)
}

func (lll *LinkLabelLink) SetFocused(focused bool) error {
	return lll.setState(win.LIS_FOCUSED, focused)
}

func (lll *LinkLabelLink) Visited() (bool, error) {
	return lll.hasState(win.LIS_VISITED)
}

func (lll *LinkLabelLink) SetVisited(visited bool) error {
	return lll.setState(win.LIS_VISITED, visited)
}

func (lll *LinkLabelLink) hasState(state uint32) (bool, error) {
	li := win.LITEM{
		ILink:     int32(lll.index),
		Mask:      win.LIF_ITEMINDEX | win.LIF_STATE,
		StateMask: state,
	}

	if win.TRUE != lll.ll.SendMessage(win.LM_GETITEM, 0, uintptr(unsafe.Pointer(&li))) {
		return false, newError("LM_GETITEM")
	}

	return li.State&state == state, nil
}

func (lll *LinkLabelLink) setState(state uint32, set bool) error {
	li := win.LITEM{
		Mask:      win.LIF_STATE,
		StateMask: state,
	}

	if set {
		li.State = state
	}

	li.Mask |= win.LIF_ITEMINDEX
	li.ILink = int32(lll.index)

	if win.TRUE != lll.ll.SendMessage(win.LM_SETITEM, 0, uintptr(unsafe.Pointer(&li))) {
		return newError("LM_SETITEM")
	}

	return nil
}
