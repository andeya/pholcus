// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

type Menu struct {
	hMenu         win.HMENU
	window        Window
	actions       *ActionList
	action2bitmap map[*Action]*Bitmap
	getDPI        func() int
}

func newMenuBar(window Window) (*Menu, error) {
	hMenu := win.CreateMenu()
	if hMenu == 0 {
		return nil, lastError("CreateMenu")
	}

	m := &Menu{
		hMenu:         hMenu,
		window:        window,
		action2bitmap: make(map[*Action]*Bitmap),
	}
	m.actions = newActionList(m)

	return m, nil
}

func NewMenu() (*Menu, error) {
	hMenu := win.CreatePopupMenu()
	if hMenu == 0 {
		return nil, lastError("CreatePopupMenu")
	}

	var mi win.MENUINFO
	mi.CbSize = uint32(unsafe.Sizeof(mi))

	if !win.GetMenuInfo(hMenu, &mi) {
		return nil, lastError("GetMenuInfo")
	}

	mi.FMask |= win.MIM_STYLE
	mi.DwStyle = win.MNS_CHECKORBMP

	if !win.SetMenuInfo(hMenu, &mi) {
		return nil, lastError("SetMenuInfo")
	}

	m := &Menu{
		hMenu:         hMenu,
		action2bitmap: make(map[*Action]*Bitmap),
	}
	m.actions = newActionList(m)

	return m, nil
}

func (m *Menu) Dispose() {
	if m.hMenu != 0 {
		win.DestroyMenu(m.hMenu)
		m.hMenu = 0

		for action, bmp := range m.action2bitmap {
			bmp.Dispose()
			delete(m.action2bitmap, action)
		}
	}
}

func (m *Menu) IsDisposed() bool {
	return m.hMenu == 0
}

func (m *Menu) Actions() *ActionList {
	return m.actions
}

func (m *Menu) updateItemsWithImageForWindow(window Window) {
	if m.window == nil {
		m.window = window
		defer func() {
			m.window = nil
		}()
	}

	for _, action := range m.actions.actions {
		if action.image != nil {
			m.onActionChanged(action)
		}
		if action.menu != nil {
			action.menu.updateItemsWithImageForWindow(window)
		}
	}
}

func (m *Menu) initMenuItemInfoFromAction(mii *win.MENUITEMINFO, action *Action) {
	mii.CbSize = uint32(unsafe.Sizeof(*mii))
	mii.FMask = win.MIIM_FTYPE | win.MIIM_ID | win.MIIM_STATE | win.MIIM_STRING
	if action.image != nil {
		mii.FMask |= win.MIIM_BITMAP
		dpi := 96
		if m.getDPI != nil {
			dpi = m.getDPI()
		} else if m.window != nil {
			dpi = m.window.DPI()
		}
		if bmp, err := iconCache.Bitmap(action.image, dpi); err == nil {
			mii.HbmpItem = bmp.hBmp
		}
	}
	if action.IsSeparator() {
		mii.FType |= win.MFT_SEPARATOR
	} else {
		mii.FType |= win.MFT_STRING
		var text string
		if s := action.shortcut; s.Key != 0 {
			text = fmt.Sprintf("%s\t%s", action.text, s.String())
		} else {
			text = action.text
		}
		mii.DwTypeData = syscall.StringToUTF16Ptr(text)
		mii.Cch = uint32(len([]rune(action.text)))
	}
	mii.WID = uint32(action.id)

	if action.Enabled() {
		mii.FState &^= win.MFS_DISABLED
	} else {
		mii.FState |= win.MFS_DISABLED
	}

	if action.Checkable() {
		mii.FMask |= win.MIIM_CHECKMARKS
	}
	if action.Checked() {
		mii.FState |= win.MFS_CHECKED
	}
	if action.Exclusive() {
		mii.FType |= win.MFT_RADIOCHECK
	}

	menu := action.menu
	if menu != nil {
		mii.FMask |= win.MIIM_SUBMENU
		mii.HSubMenu = menu.hMenu
	}
}

func (m *Menu) handleDefaultState(action *Action) {
	if action.Default() {
		// Unset other default actions before we set this one. Otherwise insertion fails.
		win.SetMenuDefaultItem(m.hMenu, ^uint32(0), false)
		for _, otherAction := range m.actions.actions {
			if otherAction != action {
				otherAction.SetDefault(false)
			}
		}
	}
}

func (m *Menu) onActionChanged(action *Action) error {
	m.handleDefaultState(action)

	if !action.Visible() {
		return nil
	}

	var mii win.MENUITEMINFO

	m.initMenuItemInfoFromAction(&mii, action)

	if !win.SetMenuItemInfo(m.hMenu, uint32(m.actions.indexInObserver(action)), true, &mii) {
		return newError("SetMenuItemInfo failed")
	}

	if action.Default() {
		win.SetMenuDefaultItem(m.hMenu, uint32(m.actions.indexInObserver(action)), true)
	}

	if action.Exclusive() && action.Checked() {
		var first, last int

		index := m.actions.Index(action)

		for i := index; i >= 0; i-- {
			first = i
			if !m.actions.At(i).Exclusive() {
				break
			}
		}

		for i := index; i < m.actions.Len(); i++ {
			last = i
			if !m.actions.At(i).Exclusive() {
				break
			}
		}

		if !win.CheckMenuRadioItem(m.hMenu, uint32(first), uint32(last), uint32(index), win.MF_BYPOSITION) {
			return newError("CheckMenuRadioItem failed")
		}
	}

	return nil
}

func (m *Menu) onActionVisibleChanged(action *Action) error {
	if !action.IsSeparator() {
		defer m.actions.updateSeparatorVisibility()
	}

	if action.Visible() {
		return m.insertAction(action, true)
	}

	return m.removeAction(action, true)
}

func (m *Menu) insertAction(action *Action, visibleChanged bool) (err error) {
	m.handleDefaultState(action)

	if !visibleChanged {
		action.addChangedHandler(m)
		defer func() {
			if err != nil {
				action.removeChangedHandler(m)
			}
		}()
	}

	if !action.Visible() {
		return
	}

	index := m.actions.indexInObserver(action)

	var mii win.MENUITEMINFO

	m.initMenuItemInfoFromAction(&mii, action)

	if !win.InsertMenuItem(m.hMenu, uint32(index), true, &mii) {
		return newError("InsertMenuItem failed")
	}

	if action.Default() {
		win.SetMenuDefaultItem(m.hMenu, uint32(m.actions.indexInObserver(action)), true)
	}

	menu := action.menu
	if menu != nil {
		menu.window = m.window
	}

	m.ensureMenuBarRedrawn()

	return
}

func (m *Menu) removeAction(action *Action, visibleChanged bool) error {
	index := m.actions.indexInObserver(action)

	if !win.RemoveMenu(m.hMenu, uint32(index), win.MF_BYPOSITION) {
		return lastError("RemoveMenu")
	}

	if !visibleChanged {
		action.removeChangedHandler(m)
	}

	m.ensureMenuBarRedrawn()

	return nil
}

func (m *Menu) ensureMenuBarRedrawn() {
	if m.window != nil {
		if mw, ok := m.window.(*MainWindow); ok && mw != nil {
			win.DrawMenuBar(mw.Handle())
		}
	}
}

func (m *Menu) onInsertedAction(action *Action) error {
	return m.insertAction(action, false)
}

func (m *Menu) onRemovingAction(action *Action) error {
	return m.removeAction(action, false)
}

func (m *Menu) onClearingActions() error {
	for i := m.actions.Len() - 1; i >= 0; i-- {
		if action := m.actions.At(i); action.Visible() {
			if err := m.onRemovingAction(action); err != nil {
				return err
			}
		}
	}

	return nil
}
