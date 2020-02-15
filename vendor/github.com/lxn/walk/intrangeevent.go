// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type IntRangeEventHandler func(from, to int)

type IntRangeEvent struct {
	handlers []IntRangeEventHandler
}

func (e *IntRangeEvent) Attach(handler IntRangeEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *IntRangeEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type IntRangeEventPublisher struct {
	event IntRangeEvent
}

func (p *IntRangeEventPublisher) Event() *IntRangeEvent {
	return &p.event
}

func (p *IntRangeEventPublisher) Publish(from, to int) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(from, to)
		}
	}
}
