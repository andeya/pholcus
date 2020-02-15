// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"strconv"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
)

type CloseReason byte

const (
	CloseReasonUnknown CloseReason = iota
	CloseReasonUser
)

var (
	syncFuncs struct {
		m     sync.Mutex
		funcs []func()
	}

	syncMsgId                 uint32
	taskbarButtonCreatedMsgId uint32
)

func init() {
	syncMsgId = win.RegisterWindowMessage(syscall.StringToUTF16Ptr("WalkSync"))
	taskbarButtonCreatedMsgId = win.RegisterWindowMessage(syscall.StringToUTF16Ptr("TaskbarButtonCreated"))
}

func synchronize(f func()) {
	syncFuncs.m.Lock()
	syncFuncs.funcs = append(syncFuncs.funcs, f)
	syncFuncs.m.Unlock()
}

func runSynchronized() {
	// Clear the list of callbacks first to avoid deadlock
	// if a callback itself calls Synchronize()...
	syncFuncs.m.Lock()
	funcs := syncFuncs.funcs
	syncFuncs.funcs = nil
	syncFuncs.m.Unlock()
	for _, f := range funcs {
		f()
	}
}

type Form interface {
	Container
	AsFormBase() *FormBase
	Run() int
	Starting() *Event
	Closing() *CloseEvent
	Activating() *Event
	Deactivating() *Event
	Activate() error
	Show()
	Hide()
	Title() string
	SetTitle(title string) error
	TitleChanged() *Event
	Icon() Image
	SetIcon(icon Image) error
	IconChanged() *Event
	Owner() Form
	SetOwner(owner Form) error
	ProgressIndicator() *ProgressIndicator

	// RightToLeftLayout returns whether coordinates on the x axis of the
	// Form increase from right to left.
	RightToLeftLayout() bool

	// SetRightToLeftLayout sets whether coordinates on the x axis of the
	// Form increase from right to left.
	SetRightToLeftLayout(rtl bool) error
}

type FormBase struct {
	WindowBase
	clientComposite       *Composite
	owner                 Form
	closingPublisher      CloseEventPublisher
	activatingPublisher   EventPublisher
	deactivatingPublisher EventPublisher
	startingPublisher     EventPublisher
	titleChangedPublisher EventPublisher
	iconChangedPublisher  EventPublisher
	progressIndicator     *ProgressIndicator
	icon                  Image
	prevFocusHWnd         win.HWND
	proposedSize          Size
	isInRestoreState      bool
	started               bool
	didSetFocus           bool
	closeReason           CloseReason
}

func (fb *FormBase) init(form Form) error {
	var err error
	if fb.clientComposite, err = NewComposite(form); err != nil {
		return err
	}
	fb.clientComposite.SetName("clientComposite")
	fb.clientComposite.background = nil

	fb.clientComposite.children.observer = form.AsFormBase()

	fb.MustRegisterProperty("Icon", NewProperty(
		func() interface{} {
			return fb.Icon()
		},
		func(v interface{}) error {
			var icon *Icon

			switch val := v.(type) {
			case *Icon:
				icon = val

			case int:
				var err error
				if icon, err = Resources.Icon(strconv.Itoa(val)); err != nil {
					return err
				}

			case string:
				var err error
				if icon, err = Resources.Icon(val); err != nil {
					return err
				}

			default:
				return ErrInvalidType
			}

			fb.SetIcon(icon)

			return nil
		},
		fb.iconChangedPublisher.Event()))

	fb.MustRegisterProperty("Title", NewProperty(
		func() interface{} {
			return fb.Title()
		},
		func(v interface{}) error {
			return fb.SetTitle(assertStringOr(v, ""))
		},
		fb.titleChangedPublisher.Event()))

	version := win.GetVersion()
	if (version&0xFF) > 6 || ((version&0xFF) == 6 && (version&0xFF00>>8) > 0) {
		win.ChangeWindowMessageFilterEx(fb.hWnd, taskbarButtonCreatedMsgId, win.MSGFLT_ALLOW, nil)
	}

	return nil
}

func (fb *FormBase) AsContainerBase() *ContainerBase {
	if fb.clientComposite == nil {
		return nil
	}

	return fb.clientComposite.AsContainerBase()
}

func (fb *FormBase) AsFormBase() *FormBase {
	return fb
}

func (fb *FormBase) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (fb *FormBase) SizeHint() Size {
	return fb.dialogBaseUnitsToPixels(Size{252, 218})
}

func (fb *FormBase) Children() *WidgetList {
	if fb.clientComposite == nil {
		return nil
	}

	return fb.clientComposite.Children()
}

func (fb *FormBase) Layout() Layout {
	if fb.clientComposite == nil {
		return nil
	}

	return fb.clientComposite.Layout()
}

func (fb *FormBase) SetLayout(value Layout) error {
	if fb.clientComposite == nil {
		return newError("clientComposite not initialized")
	}

	return fb.clientComposite.SetLayout(value)
}

func (fb *FormBase) SetBoundsPixels(bounds Rectangle) error {
	if layout := fb.Layout(); layout != nil {
		minSize := fb.sizeFromClientSizePixels(layout.MinSizeForSize(bounds.Size()))

		if bounds.Width < minSize.Width {
			bounds.Width = minSize.Width
		}
		if bounds.Height < minSize.Height {
			bounds.Height = minSize.Height
		}
	}

	if err := fb.WindowBase.SetBoundsPixels(bounds); err != nil {
		return err
	}

	walkDescendants(fb, func(wnd Window) bool {
		if container, ok := wnd.(Container); ok {
			if layout := container.Layout(); layout != nil {
				layout.Update(false)
			}
		}

		return true
	})

	return nil
}

func (fb *FormBase) fixedSize() bool {
	return !fb.hasStyleBits(win.WS_THICKFRAME)
}

func (fb *FormBase) DataBinder() *DataBinder {
	return fb.clientComposite.DataBinder()
}

func (fb *FormBase) SetDataBinder(db *DataBinder) {
	fb.clientComposite.SetDataBinder(db)
}

func (fb *FormBase) Suspended() bool {
	if fb.clientComposite == nil {
		return false
	}

	return fb.clientComposite.Suspended()
}

func (fb *FormBase) SetSuspended(suspended bool) {
	fb.clientComposite.SetSuspended(suspended)
}

func (fb *FormBase) MouseDown() *MouseEvent {
	return fb.clientComposite.MouseDown()
}

func (fb *FormBase) MouseMove() *MouseEvent {
	return fb.clientComposite.MouseMove()
}

func (fb *FormBase) MouseUp() *MouseEvent {
	return fb.clientComposite.MouseUp()
}

func (fb *FormBase) onInsertingWidget(index int, widget Widget) error {
	return fb.clientComposite.onInsertingWidget(index, widget)
}

func (fb *FormBase) onInsertedWidget(index int, widget Widget) error {
	err := fb.clientComposite.onInsertedWidget(index, widget)
	if err == nil {
		if layout := fb.Layout(); layout != nil && !fb.Suspended() {
			minClientSize := fb.Layout().MinSize()
			clientSize := fb.clientComposite.SizePixels()

			if clientSize.Width < minClientSize.Width || clientSize.Height < minClientSize.Height {
				fb.SetClientSizePixels(minClientSize)
			}
		}
	}

	return err
}

func (fb *FormBase) onRemovingWidget(index int, widget Widget) error {
	return fb.clientComposite.onRemovingWidget(index, widget)
}

func (fb *FormBase) onRemovedWidget(index int, widget Widget) error {
	return fb.clientComposite.onRemovedWidget(index, widget)
}

func (fb *FormBase) onClearingWidgets() error {
	return fb.clientComposite.onClearingWidgets()
}

func (fb *FormBase) onClearedWidgets() error {
	return fb.clientComposite.onClearedWidgets()
}

func (fb *FormBase) ContextMenu() *Menu {
	return fb.clientComposite.ContextMenu()
}

func (fb *FormBase) SetContextMenu(contextMenu *Menu) {
	fb.clientComposite.SetContextMenu(contextMenu)
}

func (fb *FormBase) applyEnabled(enabled bool) {
	fb.WindowBase.applyEnabled(enabled)

	fb.clientComposite.applyEnabled(enabled)
}

func (fb *FormBase) applyFont(font *Font) {
	fb.WindowBase.applyFont(font)

	fb.clientComposite.applyFont(font)
}

func (fb *FormBase) ApplySysColors() {
	fb.WindowBase.ApplySysColors()
	fb.clientComposite.ApplySysColors()
}

func (fb *FormBase) Background() Brush {
	return fb.clientComposite.Background()
}

func (fb *FormBase) SetBackground(background Brush) {
	fb.clientComposite.SetBackground(background)
}

func (fb *FormBase) Title() string {
	return fb.text()
}

func (fb *FormBase) SetTitle(value string) error {
	return fb.setText(value)
}

func (fb *FormBase) TitleChanged() *Event {
	return fb.titleChangedPublisher.Event()
}

// RightToLeftLayout returns whether coordinates on the x axis of the
// FormBase increase from right to left.
func (fb *FormBase) RightToLeftLayout() bool {
	return fb.hasExtendedStyleBits(win.WS_EX_LAYOUTRTL)
}

// SetRightToLeftLayout sets whether coordinates on the x axis of the
// FormBase increase from right to left.
func (fb *FormBase) SetRightToLeftLayout(rtl bool) error {
	return fb.ensureExtendedStyleBits(win.WS_EX_LAYOUTRTL, rtl)
}

func (fb *FormBase) Run() int {
	if fb.owner != nil {
		win.EnableWindow(fb.owner.Handle(), false)

		invalidateDescendentBorders := func() {
			walkDescendants(fb.owner, func(wnd Window) bool {
				if widget, ok := wnd.(Widget); ok {
					widget.AsWidgetBase().invalidateBorderInParent()
				}

				return true
			})
		}

		invalidateDescendentBorders()
		defer invalidateDescendentBorders()
	}

	if layout := fb.Layout(); layout != nil {
		layout.Update(false)
	}

	fb.clientComposite.focusFirstCandidateDescendant()

	fb.started = true
	fb.startingPublisher.Publish()

	fb.SetBoundsPixels(fb.BoundsPixels())

	return fb.mainLoop()
}

func (fb *FormBase) handleKeyDown(msg *win.MSG) bool {
	ret := false

	key, mods := Key(msg.WParam), ModifiersDown()

	// Tabbing
	if key == KeyTab && (mods&ModControl) != 0 {
		doTabbing := func(tw *TabWidget) {
			index := tw.CurrentIndex()
			if (mods & ModShift) != 0 {
				index--
				if index < 0 {
					index = tw.Pages().Len() - 1
				}
			} else {
				index++
				if index >= tw.Pages().Len() {
					index = 0
				}
			}
			tw.SetCurrentIndex(index)
		}

		hwnd := win.GetFocus()

	LOOP:
		for hwnd != 0 {
			window := windowFromHandle(hwnd)

			switch widget := window.(type) {
			case nil:

			case *TabWidget:
				doTabbing(widget)
				return true

			case Widget:

			default:
				break LOOP
			}

			hwnd = win.GetParent(hwnd)
		}

		walkDescendants(fb.window, func(w Window) bool {
			if tw, ok := w.(*TabWidget); ok {
				doTabbing(tw)
				ret = true
				return false
			}
			return true
		})
		if ret {
			return true
		}
	}

	// Shortcut actions
	hwnd := msg.HWnd
	for hwnd != 0 {
		if window := windowFromHandle(hwnd); window != nil {
			wb := window.AsWindowBase()

			if wb.shortcutActions != nil {
				for _, action := range wb.shortcutActions.actions {
					if action.shortcut.Key == key && action.shortcut.Modifiers == mods && action.Enabled() {
						action.raiseTriggered()
						return true
					}
				}
			}
		}

		hwnd = win.GetParent(hwnd)
	}

	// WebView
	walkDescendants(fb.window, func(w Window) bool {
		if webView, ok := w.(*WebView); ok {
			webViewHWnd := webView.Handle()
			if webViewHWnd == msg.HWnd || win.IsChild(webViewHWnd, msg.HWnd) {
				_ret := webView.translateAccelerator(msg)
				if _ret {
					ret = _ret
				}
			}
		}
		return true
	})
	return ret
}

func (fb *FormBase) Starting() *Event {
	return fb.startingPublisher.Event()
}

func (fb *FormBase) Activating() *Event {
	return fb.activatingPublisher.Event()
}

func (fb *FormBase) Deactivating() *Event {
	return fb.deactivatingPublisher.Event()
}

func (fb *FormBase) Activate() error {
	if hwndPrevActive := win.SetActiveWindow(fb.hWnd); hwndPrevActive == 0 {
		return lastError("SetActiveWindow")
	}

	return nil
}

func (fb *FormBase) Owner() Form {
	return fb.owner
}

func (fb *FormBase) SetOwner(value Form) error {
	fb.owner = value

	var ownerHWnd win.HWND
	if value != nil {
		ownerHWnd = value.Handle()
	}

	win.SetLastError(0)
	if 0 == win.SetWindowLong(
		fb.hWnd,
		win.GWL_HWNDPARENT,
		int32(ownerHWnd)) && win.GetLastError() != 0 {

		return lastError("SetWindowLong")
	}

	return nil
}

func (fb *FormBase) Icon() Image {
	return fb.icon
}

func (fb *FormBase) SetIcon(icon Image) error {
	var hIconSmall, hIconBig uintptr

	if icon != nil {
		smallIcon, err := iconCache.Icon(icon, fb.DPI())
		if err != nil {
			return err
		}
		hIconSmall = uintptr(smallIcon.handleForDPI(fb.DPI()))

		bigDPI := int(48.0 / float64(icon.Size().Width) * 96.0)
		bigIcon, err := iconCache.Icon(icon, bigDPI)
		if err != nil {
			return err
		}
		hIconBig = uintptr(bigIcon.handleForDPI(bigDPI))
	}

	fb.SendMessage(win.WM_SETICON, 0, hIconSmall)
	fb.SendMessage(win.WM_SETICON, 1, hIconBig)

	fb.icon = icon

	fb.iconChangedPublisher.Publish()

	return nil
}

func (fb *FormBase) IconChanged() *Event {
	return fb.iconChangedPublisher.Event()
}

func (fb *FormBase) Hide() {
	fb.window.SetVisible(false)
}

func (fb *FormBase) Show() {
	fb.proposedSize = fb.minSize

	if p, ok := fb.window.(Persistable); ok && p.Persistent() && appSingleton.settings != nil {
		p.RestoreState()
	}

	fb.window.SetVisible(true)
}

func (fb *FormBase) close() error {
	if p, ok := fb.window.(Persistable); ok && p.Persistent() && appSingleton.settings != nil {
		p.SaveState()
	}

	fb.window.Dispose()

	return nil
}

func (fb *FormBase) Close() error {
	fb.SendMessage(win.WM_CLOSE, 0, 0)

	return nil
}

func (fb *FormBase) Persistent() bool {
	return fb.clientComposite.persistent
}

func (fb *FormBase) SetPersistent(value bool) {
	fb.clientComposite.persistent = value
}

func (fb *FormBase) SaveState() error {
	if err := fb.clientComposite.SaveState(); err != nil {
		return err
	}

	var wp win.WINDOWPLACEMENT

	wp.Length = uint32(unsafe.Sizeof(wp))

	if !win.GetWindowPlacement(fb.hWnd, &wp) {
		return lastError("GetWindowPlacement")
	}

	state := fmt.Sprint(
		wp.Flags, wp.ShowCmd,
		wp.PtMinPosition.X, wp.PtMinPosition.Y,
		wp.PtMaxPosition.X, wp.PtMaxPosition.Y,
		wp.RcNormalPosition.Left, wp.RcNormalPosition.Top,
		wp.RcNormalPosition.Right, wp.RcNormalPosition.Bottom)

	if err := fb.WriteState(state); err != nil {
		return err
	}

	return nil
}

func (fb *FormBase) RestoreState() error {
	if fb.isInRestoreState {
		return nil
	}
	fb.isInRestoreState = true
	defer func() {
		fb.isInRestoreState = false
	}()

	state, err := fb.ReadState()
	if err != nil {
		return err
	}
	if state == "" {
		return nil
	}

	var wp win.WINDOWPLACEMENT

	if _, err := fmt.Sscan(state,
		&wp.Flags, &wp.ShowCmd,
		&wp.PtMinPosition.X, &wp.PtMinPosition.Y,
		&wp.PtMaxPosition.X, &wp.PtMaxPosition.Y,
		&wp.RcNormalPosition.Left, &wp.RcNormalPosition.Top,
		&wp.RcNormalPosition.Right, &wp.RcNormalPosition.Bottom); err != nil {
		return err
	}

	wp.Length = uint32(unsafe.Sizeof(wp))

	if layout := fb.Layout(); layout != nil && fb.fixedSize() {
		minSize := fb.sizeFromClientSizePixels(layout.MinSize())

		wp.RcNormalPosition.Right = wp.RcNormalPosition.Left + int32(minSize.Width) - 1
		wp.RcNormalPosition.Bottom = wp.RcNormalPosition.Top + int32(minSize.Height) - 1
	}

	if !win.SetWindowPlacement(fb.hWnd, &wp) {
		return lastError("SetWindowPlacement")
	}

	return fb.clientComposite.RestoreState()
}

func (fb *FormBase) Closing() *CloseEvent {
	return fb.closingPublisher.Event()
}

func (fb *FormBase) ProgressIndicator() *ProgressIndicator {
	return fb.progressIndicator
}

func (fb *FormBase) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_ACTIVATE:
		switch win.LOWORD(uint32(wParam)) {
		case win.WA_ACTIVE, win.WA_CLICKACTIVE:
			if fb.prevFocusHWnd != 0 {
				win.SetFocus(fb.prevFocusHWnd)
			}

			appSingleton.activeForm = fb.window.(Form)

			fb.activatingPublisher.Publish()

		case win.WA_INACTIVE:
			fb.prevFocusHWnd = win.GetFocus()

			appSingleton.activeForm = nil

			fb.deactivatingPublisher.Publish()
		}

		return 0

	case win.WM_CLOSE:
		fb.closeReason = CloseReasonUnknown
		var canceled bool
		fb.closingPublisher.Publish(&canceled, fb.closeReason)
		if !canceled {
			if fb.owner != nil {
				win.EnableWindow(fb.owner.Handle(), true)
				if !win.SetWindowPos(fb.owner.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_SHOWWINDOW) {
					lastError("SetWindowPos")
				}
			}

			fb.close()
		}
		return 0

	case win.WM_COMMAND:
		return fb.clientComposite.WndProc(hwnd, msg, wParam, lParam)

	case win.WM_GETMINMAXINFO:
		if fb.Suspended() || fb.proposedSize == (Size{}) {
			break
		}

		mmi := (*win.MINMAXINFO)(unsafe.Pointer(lParam))

		var min Size
		if layout := fb.clientComposite.layout; layout != nil {
			size := fb.clientSizeFromSizePixels(fb.proposedSize)
			min = fb.sizeFromClientSizePixels(layout.MinSizeForSize(size))

			if fb.proposedSize.Width < min.Width {
				min = fb.sizeFromClientSizePixels(layout.MinSizeForSize(min))
			}
		}

		mmi.PtMinTrackSize = win.POINT{
			int32(maxi(min.Width, fb.minSize.Width)),
			int32(maxi(min.Height, fb.minSize.Height)),
		}
		return 0

	case win.WM_NOTIFY:
		return fb.clientComposite.WndProc(hwnd, msg, wParam, lParam)

	case win.WM_SETTEXT:
		fb.titleChangedPublisher.Publish()

	case win.WM_SIZING:
		rc := (*win.RECT)(unsafe.Pointer(lParam))

		fb.proposedSize = rectangleFromRECT(*rc).Size()

		fb.clientComposite.SetBoundsPixels(fb.window.ClientBoundsPixels())

	case win.WM_SIZE:
		fb.clientComposite.SetBoundsPixels(fb.window.ClientBoundsPixels())

	case win.WM_SHOWWINDOW:
		if wParam == win.FALSE {
			fb.didSetFocus = false
		}

	case win.WM_PAINT:
		if !fb.didSetFocus && fb.Visible() {
			fb.didSetFocus = true
			fb.clientComposite.focusFirstCandidateDescendant()
		}

	case win.WM_SYSCOLORCHANGE:
		fb.ApplySysColors()

	case win.WM_DPICHANGED:
		wasSuspended := fb.Suspended()
		fb.SetSuspended(true)
		defer fb.SetSuspended(wasSuspended)

		dpi := int(win.HIWORD(uint32(wParam)))

		seenInApplyFontToDescendantsDuringDPIChange = make(map[*WindowBase]bool)
		seenInApplyDPIToDescendantsDuringDPIChange = make(map[*WindowBase]bool)
		defer func() {
			seenInApplyFontToDescendantsDuringDPIChange = nil
			seenInApplyDPIToDescendantsDuringDPIChange = nil
		}()

		fb.clientComposite.ApplyDPI(dpi)
		fb.ApplyDPI(dpi)
		if fb.progressIndicator != nil {
			fb.progressIndicator.SetOverlayIcon(fb.progressIndicator.overlayIcon, fb.progressIndicator.overlayIconDescription)
		}
		applyDPIToDescendants(fb.window, dpi)

		rc := (*win.RECT)(unsafe.Pointer(lParam))
		fb.window.SetBoundsPixels(rectangleFromRECT(*rc))

		fb.SetIcon(fb.icon)

		time.AfterFunc(time.Second, func() {
			if fb.hWnd == 0 {
				return
			}
			fb.Synchronize(func() {
				for ni := range notifyIcons {
					// We do this on all NotifyIcons, not just ones attached to this form or descendents, because
					// the notify icon might be on a different screen, and since it can't get notifications itself
					// we hope that one of the forms did for it. We also have to delay it by a second, because the
					// tray usually gets resized sometime after us. This is a nasty hack!
					ni.applyDPI()
				}
			})
		})

	case win.WM_SYSCOMMAND:
		if wParam == win.SC_CLOSE {
			fb.closeReason = CloseReasonUser
		}

	case taskbarButtonCreatedMsgId:
		version := win.GetVersion()
		major := version & 0xFF
		minor := version & 0xFF00 >> 8
		// Check that the OS is Win 7 or later (Win 7 is v6.1).
		if fb.progressIndicator == nil && (major > 6 || (major == 6 && minor > 0)) {
			fb.progressIndicator, _ = newTaskbarList3(fb.hWnd)
		}
	}

	return fb.WindowBase.WndProc(hwnd, msg, wParam, lParam)
}
