// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type IntEventHandler func(n int)

type IntEvent struct {
	handlers []IntEventHandler
}

func (e *IntEvent) Attach(handler IntEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *IntEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type IntEventPublisher struct {
	event IntEvent
}

func (p *IntEventPublisher) Event() *IntEvent {
	return &p.event
}

func (p *IntEventPublisher) Publish(n int) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(n)
		}
	}
}
