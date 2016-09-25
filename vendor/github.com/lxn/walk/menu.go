// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

type Menu struct {
	hMenu   win.HMENU
	hWnd    win.HWND
	actions *ActionList
}

func newMenuBar(hWnd win.HWND) (*Menu, error) {
	hMenu := win.CreateMenu()
	if hMenu == 0 {
		return nil, lastError("CreateMenu")
	}

	m := &Menu{hMenu: hMenu, hWnd: hWnd}
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

	m := &Menu{hMenu: hMenu}
	m.actions = newActionList(m)

	return m, nil
}

func (m *Menu) Dispose() {
	if m.hMenu != 0 {
		win.DestroyMenu(m.hMenu)
		m.hMenu = 0
	}
}

func (m *Menu) IsDisposed() bool {
	return m.hMenu == 0
}

func (m *Menu) Actions() *ActionList {
	return m.actions
}

func (m *Menu) initMenuItemInfoFromAction(mii *win.MENUITEMINFO, action *Action) {
	mii.CbSize = uint32(unsafe.Sizeof(*mii))
	mii.FMask = win.MIIM_FTYPE | win.MIIM_ID | win.MIIM_STATE | win.MIIM_STRING
	if action.image != nil {
		mii.FMask |= win.MIIM_BITMAP
		mii.HbmpItem = action.image.handle()
	}
	if action.IsSeparator() {
		mii.FType = win.MFT_SEPARATOR
	} else {
		mii.FType = win.MFT_STRING
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

	menu := action.menu
	if menu != nil {
		mii.FMask |= win.MIIM_SUBMENU
		mii.HSubMenu = menu.hMenu
	}
}

func (m *Menu) onActionChanged(action *Action) error {
	if !action.Visible() {
		return nil
	}

	var mii win.MENUITEMINFO

	m.initMenuItemInfoFromAction(&mii, action)

	if !win.SetMenuItemInfo(m.hMenu, uint32(m.actions.indexInObserver(action)), true, &mii) {
		return newError("SetMenuItemInfo failed")
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

	menu := action.menu
	if menu != nil {
		menu.hWnd = m.hWnd
	}

	if m.hWnd != 0 {
		win.DrawMenuBar(m.hWnd)
	}

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

	if m.hWnd != 0 {
		win.DrawMenuBar(m.hWnd)
	}

	return nil
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
