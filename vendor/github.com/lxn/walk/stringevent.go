// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type StringEventHandler func(s string)

type StringEvent struct {
	handlers []StringEventHandler
}

func (e *StringEvent) Attach(handler StringEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *StringEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type StringEventPublisher struct {
	event StringEvent
}

func (p *StringEventPublisher) Event() *StringEvent {
	return &p.event
}

func (p *StringEventPublisher) Publish(s string) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(s)
		}
	}
}
