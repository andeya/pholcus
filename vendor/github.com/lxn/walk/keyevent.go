// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type KeyEventHandler func(key Key)

type KeyEvent struct {
	handlers []KeyEventHandler
}

func (e *KeyEvent) Attach(handler KeyEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *KeyEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type KeyEventPublisher struct {
	event KeyEvent
}

func (p *KeyEventPublisher) Event() *KeyEvent {
	return &p.event
}

func (p *KeyEventPublisher) Publish(key Key) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(key)
		}
	}
}
