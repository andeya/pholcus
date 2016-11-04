// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
	"syscall"
)

type DropFilesEventHandler func([]string)

type DropFilesEvent struct {
	hWnd     win.HWND
	handlers []DropFilesEventHandler
}

func (e *DropFilesEvent) Attach(handler DropFilesEventHandler) int {
	if len(e.handlers) == 0 {
		win.DragAcceptFiles(e.hWnd, true)
	}
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *DropFilesEvent) Detach(handle int) {
	e.handlers[handle] = nil
	for _, h := range e.handlers {
		if h != nil {
			return
		}
	}
	win.DragAcceptFiles(e.hWnd, false)
}

type DropFilesEventPublisher struct {
	event DropFilesEvent
}

func (p *DropFilesEventPublisher) Event(hWnd win.HWND) *DropFilesEvent {
	p.event.hWnd = hWnd
	return &p.event
}

func (p *DropFilesEventPublisher) Publish(hDrop win.HDROP) {
	var files []string

	n := win.DragQueryFile(hDrop, 0xFFFFFFFF, nil, 0)
	for i := 0; i < int(n); i++ {
		bufSize := uint(512)
		buf := make([]uint16, bufSize)
		if win.DragQueryFile(hDrop, uint(i), &buf[0], bufSize) > 0 {
			files = append(files, syscall.UTF16ToString(buf))
		}
	}
	win.DragFinish(hDrop)

	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(files)
		}
	}
}
