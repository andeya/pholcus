// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"unsafe"
)

import (
	"github.com/lxn/win"
)

type Margins struct {
	HNear, VNear, HFar, VFar int
}

type Layout interface {
	Container() Container
	SetContainer(value Container)
	Margins() Margins
	SetMargins(value Margins) error
	Spacing() int
	SetSpacing(value int) error
	LayoutFlags() LayoutFlags
	MinSize() Size
	Update(reset bool) error
}

func shouldLayoutWidget(widget Widget) bool {
	if widget == nil {
		return false
	}

	_, isSpacer := widget.(*Spacer)

	return isSpacer || widget.AsWindowBase().visible || widget.AlwaysConsumeSpace()
}

func DescendantByName(container Container, name string) Widget {
	var widget Widget

	walkDescendants(container.AsContainerBase(), func(w Window) bool {
		if w.Name() == name {
			widget = w.(Widget)
			return false
		}

		return true
	})

	if widget == nil {
		return nil
	}

	return widget
}

type Container interface {
	Window
	AsContainerBase() *ContainerBase
	Children() *WidgetList
	Layout() Layout
	SetLayout(value Layout) error
	DataBinder() *DataBinder
	SetDataBinder(dbm *DataBinder)
}

type ContainerBase struct {
	WidgetBase
	layout     Layout
	children   *WidgetList
	dataBinder *DataBinder
	persistent bool
}

func (cb *ContainerBase) AsWidgetBase() *WidgetBase {
	return &cb.WidgetBase
}

func (cb *ContainerBase) AsContainerBase() *ContainerBase {
	return cb
}

func (cb *ContainerBase) LayoutFlags() LayoutFlags {
	if cb.layout == nil {
		return 0
	}

	return cb.layout.LayoutFlags()
}

func (cb *ContainerBase) MinSizeHint() Size {
	if cb.layout == nil {
		return Size{}
	}

	return cb.layout.MinSize()
}

func (cb *ContainerBase) applyEnabled(enabled bool) {
	cb.WidgetBase.applyEnabled(enabled)

	applyEnabledToDescendants(cb.window.(Widget), enabled)
}

func (cb *ContainerBase) applyFont(font *Font) {
	cb.WidgetBase.applyFont(font)

	applyFontToDescendants(cb.window.(Widget), font)
}

func (cb *ContainerBase) Children() *WidgetList {
	return cb.children
}

func (cb *ContainerBase) Layout() Layout {
	return cb.layout
}

func (cb *ContainerBase) SetLayout(value Layout) error {
	if cb.layout != value {
		if cb.layout != nil {
			cb.layout.SetContainer(nil)
		}

		cb.layout = value

		if value != nil && value.Container() != Container(cb) {
			value.SetContainer(cb)
		}
	}

	return nil
}

func (cb *ContainerBase) DataBinder() *DataBinder {
	return cb.dataBinder
}

func (cb *ContainerBase) SetDataBinder(db *DataBinder) {
	if db == cb.dataBinder {
		return
	}

	if cb.dataBinder != nil {
		cb.dataBinder.SetBoundWidgets(nil)
	}

	cb.dataBinder = db

	if db != nil {
		var boundWidgets []Widget

		walkDescendants(cb.window, func(w Window) bool {
			if w.Handle() == cb.hWnd {
				return true
			}

			if c, ok := w.(Container); ok && c.DataBinder() != nil {
				return false
			}

			for _, prop := range w.AsWindowBase().name2Property {
				if _, ok := prop.Source().(string); ok {
					boundWidgets = append(boundWidgets, w.(Widget))
					break
				}
			}

			return true
		})

		db.SetBoundWidgets(boundWidgets)
	}
}

func (cb *ContainerBase) forEachPersistableChild(f func(p Persistable) error) error {
	if cb.children == nil {
		return nil
	}

	for _, child := range cb.children.items {
		if persistable, ok := child.(Persistable); ok && persistable.Persistent() {
			if err := f(persistable); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cb *ContainerBase) Persistent() bool {
	return cb.persistent
}

func (cb *ContainerBase) SetPersistent(value bool) {
	cb.persistent = value
}

func (cb *ContainerBase) SaveState() error {
	return cb.forEachPersistableChild(func(p Persistable) error {
		return p.SaveState()
	})
}

func (cb *ContainerBase) RestoreState() error {
	return cb.forEachPersistableChild(func(p Persistable) error {
		return p.RestoreState()
	})
}

func (cb *ContainerBase) SetSuspended(suspend bool) {
	wasSuspended := cb.Suspended()

	cb.WidgetBase.SetSuspended(suspend)

	if !suspend && wasSuspended && cb.layout != nil {
		cb.layout.Update(false)
	}
}

func (cb *ContainerBase) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_COMMAND:
		if lParam == 0 {
			switch win.HIWORD(uint32(wParam)) {
			case 0:
				cmdId := win.LOWORD(uint32(wParam))
				switch cmdId {
				case win.IDOK, win.IDCANCEL:
					form := ancestor(cb)
					if form == nil {
						break
					}

					dlg, ok := form.(dialogish)
					if !ok {
						break
					}

					var button *PushButton
					if cmdId == win.IDOK {
						button = dlg.DefaultButton()
					} else {
						button = dlg.CancelButton()
					}

					if button != nil && button.Visible() && button.Enabled() {
						button.raiseClicked()
					}

					break
				}

				// Menu
				actionId := uint16(win.LOWORD(uint32(wParam)))
				if action, ok := actionsById[actionId]; ok {
					action.raiseTriggered()
					return 0
				}

			case 1:
				// Accelerator
			}
		} else {
			// The window that sent the notification shall handle it itself.
			hWnd := win.HWND(lParam)
			if window := windowFromHandle(hWnd); window != nil {
				window.WndProc(hwnd, msg, wParam, lParam)
				return 0
			}
		}

	case win.WM_NOTIFY:
		nmh := (*win.NMHDR)(unsafe.Pointer(lParam))
		if window := windowFromHandle(nmh.HwndFrom); window != nil {
			// The window that sent the notification shall handle it itself.
			return window.WndProc(hwnd, msg, wParam, lParam)
		}

	case win.WM_HSCROLL, win.WM_VSCROLL:
		if window := windowFromHandle(win.HWND(lParam)); window != nil {
			// The window that sent the notification shall handle it itself.
			return window.WndProc(hwnd, msg, wParam, lParam)
		}

	case win.WM_SIZE, win.WM_SIZING:
		if cb.layout != nil {
			cb.layout.Update(false)
		}
	}

	return cb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (cb *ContainerBase) onInsertingWidget(index int, widget Widget) (err error) {
	return nil
}

func (cb *ContainerBase) onInsertedWidget(index int, widget Widget) (err error) {
	if parent := widget.Parent(); parent == nil || parent.Handle() != cb.hWnd {
		if err = widget.SetParent(cb.window.(Container)); err != nil {
			return
		}
	}

	if cb.layout != nil {
		cb.layout.Update(true)
	}

	widget.(applyFonter).applyFont(cb.Font())

	return
}

func (cb *ContainerBase) onRemovingWidget(index int, widget Widget) (err error) {
	if widget.Parent() == nil {
		return
	}

	if widget.Parent().Handle() == cb.hWnd {
		err = widget.SetParent(nil)
	}

	return
}

func (cb *ContainerBase) onRemovedWidget(index int, widget Widget) (err error) {
	if cb.layout != nil {
		cb.layout.Update(true)
	}

	return
}

func (cb *ContainerBase) onClearingWidgets() (err error) {
	for _, widget := range cb.children.items {
		if widget.Parent().Handle() == cb.hWnd {
			if err = widget.SetParent(nil); err != nil {
				return
			}
		}
	}

	return
}

func (cb *ContainerBase) onClearedWidgets() (err error) {
	if cb.layout != nil {
		cb.layout.Update(true)
	}

	return
}
