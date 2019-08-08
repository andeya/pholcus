// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var notifyIcons = make(map[*NotifyIcon]bool)

func notifyIconWndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) (result uintptr) {
	// Retrieve our *NotifyIcon from the message window.
	ptr := win.GetWindowLongPtr(hwnd, win.GWLP_USERDATA)
	ni := (*NotifyIcon)(unsafe.Pointer(ptr))

	switch lParam {
	case win.WM_LBUTTONDOWN:
		ni.publishMouseEvent(&ni.mouseDownPublisher, LeftButton)

	case win.WM_LBUTTONUP:
		ni.publishMouseEvent(&ni.mouseUpPublisher, LeftButton)

	case win.WM_RBUTTONDOWN:
		ni.publishMouseEvent(&ni.mouseDownPublisher, RightButton)

	case win.WM_RBUTTONUP:
		ni.publishMouseEvent(&ni.mouseUpPublisher, RightButton)

		win.SendMessage(hwnd, msg, wParam, win.WM_CONTEXTMENU)

	case win.WM_CONTEXTMENU:
		if ni.contextMenu.Actions().Len() == 0 {
			break
		}

		win.SetForegroundWindow(hwnd)

		var p win.POINT
		if !win.GetCursorPos(&p) {
			lastError("GetCursorPos")
		}

		ni.applyDPI()

		actionId := uint16(win.TrackPopupMenuEx(
			ni.contextMenu.hMenu,
			win.TPM_NOANIMATION|win.TPM_RETURNCMD,
			p.X,
			p.Y,
			hwnd,
			nil))
		if actionId != 0 {
			if action, ok := actionsById[actionId]; ok {
				action.raiseTriggered()
			}
		}

		return 0
	case win.NIN_BALLOONUSERCLICK:
		ni.messageClickedPublisher.Publish()
	}

	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}

// NotifyIcon represents an icon in the taskbar notification area.
type NotifyIcon struct {
	id                      uint32
	hWnd                    win.HWND
	lastDPI                 int
	contextMenu             *Menu
	icon                    Image
	toolTip                 string
	visible                 bool
	mouseDownPublisher      MouseEventPublisher
	mouseUpPublisher        MouseEventPublisher
	messageClickedPublisher EventPublisher
}

// NewNotifyIcon creates and returns a new NotifyIcon.
//
// The NotifyIcon is initially not visible.
func NewNotifyIcon(form Form) (*NotifyIcon, error) {
	fb := form.AsFormBase()
	// Add our notify icon to the status area and make sure it is hidden.
	nid := win.NOTIFYICONDATA{
		HWnd:             fb.hWnd,
		UFlags:           win.NIF_MESSAGE | win.NIF_STATE,
		DwState:          win.NIS_HIDDEN,
		DwStateMask:      win.NIS_HIDDEN,
		UCallbackMessage: notifyIconMessageId,
	}
	nid.CbSize = uint32(unsafe.Sizeof(nid) - unsafe.Sizeof(win.HICON(0)))

	if !win.Shell_NotifyIcon(win.NIM_ADD, &nid) {
		return nil, newError("Shell_NotifyIcon")
	}

	// We want XP-compatible message behavior.
	nid.UVersion = win.NOTIFYICON_VERSION

	if !win.Shell_NotifyIcon(win.NIM_SETVERSION, &nid) {
		return nil, newError("Shell_NotifyIcon")
	}

	// Create and initialize the NotifyIcon already.
	menu, err := NewMenu()
	if err != nil {
		return nil, err
	}
	menu.window = form

	ni := &NotifyIcon{
		id:          nid.UID,
		hWnd:        fb.hWnd,
		contextMenu: menu,
	}

	menu.getDPI = ni.DPI

	// Set our *NotifyIcon as user data for the message window.
	win.SetWindowLongPtr(fb.hWnd, win.GWLP_USERDATA, uintptr(unsafe.Pointer(ni)))

	notifyIcons[ni] = true
	return ni, nil
}

func (ni *NotifyIcon) DPI() int {
	fakeWb := WindowBase{hWnd: win.FindWindow(syscall.StringToUTF16Ptr("Shell_TrayWnd"), syscall.StringToUTF16Ptr(""))}
	return fakeWb.DPI()
}

func (ni *NotifyIcon) applyDPI() {
	dpi := ni.DPI()
	if dpi == ni.lastDPI {
		return
	}
	ni.lastDPI = dpi
	for _, action := range ni.contextMenu.actions.actions {
		if action.image != nil {
			ni.contextMenu.onActionChanged(action)
		}
	}
	icon := ni.icon
	ni.icon = nil
	if icon != nil {
		ni.SetIcon(icon)
	}
}

func (ni *NotifyIcon) notifyIconData() *win.NOTIFYICONDATA {
	nid := &win.NOTIFYICONDATA{
		UID:  ni.id,
		HWnd: ni.hWnd,
	}
	nid.CbSize = uint32(unsafe.Sizeof(*nid) - unsafe.Sizeof(win.HICON(0)))

	return nid
}

// Dispose releases the operating system resources associated with the
// NotifyIcon.
//
// The associated Icon is not disposed of.
func (ni *NotifyIcon) Dispose() error {
	if ni.hWnd == 0 {
		return nil
	}
	delete(notifyIcons, ni)

	nid := ni.notifyIconData()

	if !win.Shell_NotifyIcon(win.NIM_DELETE, nid) {
		return newError("Shell_NotifyIcon")
	}

	if !win.DestroyWindow(ni.hWnd) {
		return lastError("DestroyWindow")
	}
	ni.hWnd = 0

	return nil
}

func (ni *NotifyIcon) showMessage(title, info string, iconType uint32, icon Image) error {
	nid := ni.notifyIconData()
	nid.UFlags = win.NIF_INFO
	nid.DwInfoFlags = iconType
	var oldIcon Image
	if iconType == win.NIIF_USER && icon != nil {
		oldIcon = ni.icon
		if err := ni.setNIDIcon(nid, icon); err != nil {
			return err
		}
		nid.UFlags |= win.NIF_ICON
	}
	if title16, err := syscall.UTF16FromString(title); err == nil {
		copy(nid.SzInfoTitle[:], title16)
	}
	if info16, err := syscall.UTF16FromString(info); err == nil {
		copy(nid.SzInfo[:], info16)
	}
	if !win.Shell_NotifyIcon(win.NIM_MODIFY, nid) {
		return newError("Shell_NotifyIcon")
	}
	if oldIcon != nil {
		ni.icon = nil
		ni.SetIcon(oldIcon)
	}

	return nil
}

// ShowMessage displays a neutral message balloon above the NotifyIcon.
//
// The NotifyIcon must be visible before calling this method.
func (ni *NotifyIcon) ShowMessage(title, info string) error {
	return ni.showMessage(title, info, win.NIIF_NONE, nil)
}

// ShowInfo displays an info message balloon above the NotifyIcon.
//
// The NotifyIcon must be visible before calling this method.
func (ni *NotifyIcon) ShowInfo(title, info string) error {
	return ni.showMessage(title, info, win.NIIF_INFO, nil)
}

// ShowWarning displays a warning message balloon above the NotifyIcon.
//
// The NotifyIcon must be visible before calling this method.
func (ni *NotifyIcon) ShowWarning(title, info string) error {
	return ni.showMessage(title, info, win.NIIF_WARNING, nil)
}

// ShowError displays an error message balloon above the NotifyIcon.
//
// The NotifyIcon must be visible before calling this method.
func (ni *NotifyIcon) ShowError(title, info string) error {
	return ni.showMessage(title, info, win.NIIF_ERROR, nil)
}

// ShowCustom displays a custom icon message balloon above the NotifyIcon.
// If icon is nil, the main notification icon is used instead of a custom one.
//
// The NotifyIcon must be visible before calling this method.
func (ni *NotifyIcon) ShowCustom(title, info string, icon Image) error {
	return ni.showMessage(title, info, win.NIIF_USER, icon)
}

// ContextMenu returns the context menu of the NotifyIcon.
func (ni *NotifyIcon) ContextMenu() *Menu {
	return ni.contextMenu
}

// Icon returns the Icon of the NotifyIcon.
func (ni *NotifyIcon) Icon() Image {
	return ni.icon
}

// SetIcon sets the Icon of the NotifyIcon.
func (ni *NotifyIcon) SetIcon(icon Image) error {
	if icon == ni.icon {
		return nil
	}

	nid := ni.notifyIconData()
	nid.UFlags = win.NIF_ICON
	if icon == nil {
		nid.HIcon = 0
	} else {
		if err := ni.setNIDIcon(nid, icon); err != nil {
			return err
		}
	}

	if !win.Shell_NotifyIcon(win.NIM_MODIFY, nid) {
		return newError("Shell_NotifyIcon")
	}

	ni.icon = icon

	return nil
}

func (ni *NotifyIcon) setNIDIcon(nid *win.NOTIFYICONDATA, icon Image) error {
	dpi := ni.DPI()
	ic, err := iconCache.Icon(icon, dpi)
	if err != nil {
		return err
	}
	nid.HIcon = ic.handleForDPI(dpi)

	return nil
}

// ToolTip returns the tool tip text of the NotifyIcon.
func (ni *NotifyIcon) ToolTip() string {
	return ni.toolTip
}

// SetToolTip sets the tool tip text of the NotifyIcon.
func (ni *NotifyIcon) SetToolTip(toolTip string) error {
	if toolTip == ni.toolTip {
		return nil
	}

	nid := ni.notifyIconData()
	nid.UFlags = win.NIF_TIP
	copy(nid.SzTip[:], syscall.StringToUTF16(toolTip))

	if !win.Shell_NotifyIcon(win.NIM_MODIFY, nid) {
		return newError("Shell_NotifyIcon")
	}

	ni.toolTip = toolTip

	return nil
}

// Visible returns if the NotifyIcon is visible.
func (ni *NotifyIcon) Visible() bool {
	return ni.visible
}

// SetVisible sets if the NotifyIcon is visible.
func (ni *NotifyIcon) SetVisible(visible bool) error {
	if visible == ni.visible {
		return nil
	}

	nid := ni.notifyIconData()
	nid.UFlags = win.NIF_STATE
	nid.DwStateMask = win.NIS_HIDDEN
	if !visible {
		nid.DwState = win.NIS_HIDDEN
	}

	if !win.Shell_NotifyIcon(win.NIM_MODIFY, nid) {
		return newError("Shell_NotifyIcon")
	}

	ni.visible = visible

	return nil
}

func (ni *NotifyIcon) publishMouseEvent(publisher *MouseEventPublisher, button MouseButton) {
	var p win.POINT
	if !win.GetCursorPos(&p) {
		lastError("GetCursorPos")
	}

	publisher.Publish(int(p.X), int(p.Y), button)
}

// MouseDown returns the event that is published when a mouse button is pressed
// while the cursor is over the NotifyIcon.
func (ni *NotifyIcon) MouseDown() *MouseEvent {
	return ni.mouseDownPublisher.Event()
}

// MouseDown returns the event that is published when a mouse button is released
// while the cursor is over the NotifyIcon.
func (ni *NotifyIcon) MouseUp() *MouseEvent {
	return ni.mouseUpPublisher.Event()
}

// MessageClicked occurs when the user clicks a message shown with ShowMessage or
// one of its iconed variants.
func (ni *NotifyIcon) MessageClicked() *Event {
	return ni.messageClickedPublisher.Event()
}
