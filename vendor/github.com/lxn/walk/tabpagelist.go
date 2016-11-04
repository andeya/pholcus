// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type tabPageListObserver interface {
	onInsertingPage(index int, page *TabPage) error
	onInsertedPage(index int, page *TabPage) error
	onRemovingPage(index int, page *TabPage) error
	onRemovedPage(index int, page *TabPage) error
	onClearingPages(pages []*TabPage) error
	onClearedPages(pages []*TabPage) error
}

type TabPageList struct {
	items    []*TabPage
	observer tabPageListObserver
}

func newTabPageList(observer tabPageListObserver) *TabPageList {
	return &TabPageList{observer: observer}
}

func (l *TabPageList) Add(item *TabPage) error {
	return l.Insert(len(l.items), item)
}

func (l *TabPageList) At(index int) *TabPage {
	return l.items[index]
}

func (l *TabPageList) Clear() error {
	observer := l.observer
	if observer != nil {
		if err := observer.onClearingPages(l.items); err != nil {
			return err
		}
	}

	oldItems := l.items
	l.items = l.items[:0]

	if observer != nil {
		if err := observer.onClearedPages(oldItems); err != nil {
			l.items = oldItems
			return err
		}
	}

	return nil
}

func (l *TabPageList) Index(item *TabPage) int {
	for i, lvi := range l.items {
		if lvi == item {
			return i
		}
	}

	return -1
}

func (l *TabPageList) Contains(item *TabPage) bool {
	return l.Index(item) > -1
}

func (l *TabPageList) indexHandle(handle win.HWND) int {
	for i, page := range l.items {
		if page.Handle() == handle {
			return i
		}
	}

	return -1
}

func (l *TabPageList) containsHandle(handle win.HWND) bool {
	return l.indexHandle(handle) > -1
}

func (l *TabPageList) insertIntoSlice(index int, item *TabPage) {
	l.items = append(l.items, nil)
	copy(l.items[index+1:], l.items[index:])
	l.items[index] = item
}

func (l *TabPageList) Insert(index int, item *TabPage) error {
	observer := l.observer
	if observer != nil {
		if err := observer.onInsertingPage(index, item); err != nil {
			return err
		}
	}

	l.insertIntoSlice(index, item)

	if observer != nil {
		if err := observer.onInsertedPage(index, item); err != nil {
			l.items = append(l.items[:index], l.items[index+1:]...)
			return err
		}
	}

	return nil
}

func (l *TabPageList) Len() int {
	return len(l.items)
}

func (l *TabPageList) Remove(item *TabPage) error {
	index := l.Index(item)
	if index == -1 {
		return nil
	}

	return l.RemoveAt(index)
}

func (l *TabPageList) RemoveAt(index int) error {
	observer := l.observer
	item := l.items[index]
	if observer != nil {
		if err := observer.onRemovingPage(index, item); err != nil {
			return err
		}
	}

	l.items = append(l.items[:index], l.items[index+1:]...)

	if observer != nil {
		if err := observer.onRemovedPage(index, item); err != nil {
			l.insertIntoSlice(index, item)
			return err
		}
	}

	return nil
}
