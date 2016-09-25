// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type ErrorEventHandler func(err error)

type ErrorEvent struct {
	handlers []ErrorEventHandler
}

func (e *ErrorEvent) Attach(handler ErrorEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *ErrorEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type ErrorEventPublisher struct {
	event ErrorEvent
}

func (p *ErrorEventPublisher) Event() *ErrorEvent {
	return &p.event
}

func (p *ErrorEventPublisher) Publish(err error) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(err)
		}
	}
}
