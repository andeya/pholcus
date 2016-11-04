// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type MouseButton int

const (
	LeftButton   MouseButton = win.MK_LBUTTON
	RightButton  MouseButton = win.MK_RBUTTON
	MiddleButton MouseButton = win.MK_MBUTTON
)

type MouseEventHandler func(x, y int, button MouseButton)

type MouseEvent struct {
	handlers []MouseEventHandler
}

func (e *MouseEvent) Attach(handler MouseEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *MouseEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type MouseEventPublisher struct {
	event MouseEvent
}

func (p *MouseEventPublisher) Event() *MouseEvent {
	return &p.event
}

func (p *MouseEventPublisher) Publish(x, y int, button MouseButton) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(x, y, button)
		}
	}
}

func MouseWheelEventDelta(button MouseButton) int {
	return int(int32(button) >> 16)
}

func MouseWheelEventKeyState(button MouseButton) int {
	return int(int32(button) & 0xFFFF)
}
