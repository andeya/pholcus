// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type CancelEventHandler func(canceled *bool)

type CancelEvent struct {
	handlers []CancelEventHandler
}

func (e *CancelEvent) Attach(handler CancelEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *CancelEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type CancelEventPublisher struct {
	event CancelEvent
}

func (p *CancelEventPublisher) Event() *CancelEvent {
	return &p.event
}

func (p *CancelEventPublisher) Publish(canceled *bool) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(canceled)
		}
	}
}
