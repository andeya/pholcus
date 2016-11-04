// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type TreeItemEventHandler func(item TreeItem)

type TreeItemEvent struct {
	handlers []TreeItemEventHandler
}

func (e *TreeItemEvent) Attach(handler TreeItemEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *TreeItemEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type TreeItemEventPublisher struct {
	event TreeItemEvent
}

func (p *TreeItemEventPublisher) Event() *TreeItemEvent {
	return &p.event
}

func (p *TreeItemEventPublisher) Publish(item TreeItem) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(item)
		}
	}
}
