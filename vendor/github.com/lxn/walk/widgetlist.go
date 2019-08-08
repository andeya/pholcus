// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type widgetListObserver interface {
	onInsertingWidget(index int, widget Widget) error
	onInsertedWidget(index int, widget Widget) error
	onRemovingWidget(index int, widget Widget) error
	onRemovedWidget(index int, widget Widget) error
	onClearingWidgets() error
	onClearedWidgets() error
}

type WidgetList struct {
	items           []*WidgetBase
	observer        widgetListObserver
	widgetInRemoval *WidgetBase
}

func newWidgetList(observer widgetListObserver) *WidgetList {
	return &WidgetList{observer: observer}
}

func (l *WidgetList) Add(item Widget) error {
	return l.Insert(len(l.items), item)
}

func (l *WidgetList) At(index int) Widget {
	return l.items[index].window.(Widget)
}

func (l *WidgetList) Clear() error {
	observer := l.observer
	if observer != nil {
		if err := observer.onClearingWidgets(); err != nil {
			return err
		}
	}

	oldItems := l.items
	l.items = l.items[:0]

	if observer != nil {
		if err := observer.onClearedWidgets(); err != nil {
			l.items = oldItems
			return err
		}
	}

	for _, item := range oldItems {
		item.form = nil
	}

	return nil
}

func (l *WidgetList) Index(item Widget) int {
	wb := item.AsWidgetBase()

	for i, widget := range l.items {
		if widget == wb {
			return i
		}
	}

	return -1
}

func (l *WidgetList) Contains(item Widget) bool {
	return l.Index(item) > -1
}

func (l *WidgetList) indexHandle(handle win.HWND) int {
	for i, widget := range l.items {
		if widget.Handle() == handle {
			return i
		}
	}

	return -1
}

func (l *WidgetList) containsHandle(handle win.HWND) bool {
	return l.indexHandle(handle) > -1
}

func (l *WidgetList) insertIntoSlice(index int, item Widget) {
	l.items = append(l.items, nil)
	copy(l.items[index+1:], l.items[index:])
	l.items[index] = item.AsWidgetBase()
}

func (l *WidgetList) Insert(index int, item Widget) error {
	if l.Contains(item) {
		return newError("cannot insert same widget multiple times")
	}

	observer := l.observer
	if observer != nil {
		if err := observer.onInsertingWidget(index, item); err != nil {
			return err
		}
	}

	l.insertIntoSlice(index, item)

	if observer != nil {
		if err := observer.onInsertedWidget(index, item); err != nil {
			l.items = append(l.items[:index], l.items[index+1:]...)
			return err
		}
	}

	return nil
}

func (l *WidgetList) Len() int {
	return len(l.items)
}

func (l *WidgetList) Remove(item Widget) error {
	index := l.Index(item)
	if index == -1 {
		return nil
	}

	return l.RemoveAt(index)
}

func (l *WidgetList) RemoveAt(index int) error {
	item := l.items[index]

	if item == l.widgetInRemoval {
		return nil
	}

	observer := l.observer
	widget := item.window.(Widget)

	if observer != nil {
		l.widgetInRemoval = item
		defer func() {
			l.widgetInRemoval = nil
		}()

		if err := observer.onRemovingWidget(index, widget); err != nil {
			return err
		}
	}

	l.items = append(l.items[:index], l.items[index+1:]...)

	if observer != nil {
		if err := observer.onRemovedWidget(index, widget); err != nil {
			l.insertIntoSlice(index, widget)
			return err
		}
	}

	item.form = nil

	return nil
}
