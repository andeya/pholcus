// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
)

var (
	inProgressEventsByForm     = make(map[Form][]*Event)
	scheduledLayoutsByForm     = make(map[Form][]Layout)
	performingScheduledLayouts bool
	formResizeScheduled        bool
)

func scheduleLayout(layout Layout) bool {
	if appSingleton.activeForm == nil {
		inProgressEventsByForm = make(map[Form][]*Event)
		return false
	}

	events := inProgressEventsByForm[appSingleton.activeForm]
	if len(events) == 0 {
		return false
	}

	layouts := scheduledLayoutsByForm[appSingleton.activeForm]

	for _, l := range layouts {
		if l == layout {
			return true
		}
	}

	layouts = append(layouts, layout)

	scheduledLayoutsByForm[appSingleton.activeForm] = layouts

	return true
}

func performScheduledLayouts() {
	layouts := scheduledLayoutsByForm[appSingleton.activeForm]
	delete(scheduledLayoutsByForm, appSingleton.activeForm)
	if len(layouts) == 0 {
		return
	}

	old := performingScheduledLayouts
	performingScheduledLayouts = true
	defer func() {
		performingScheduledLayouts = old
	}()

	if formResizeScheduled {
		formResizeScheduled = false

		bounds := appSingleton.activeForm.BoundsPixels()

		if appSingleton.activeForm.AsFormBase().fixedSize() {
			bounds.Width, bounds.Height = 0, 0
		}

		appSingleton.activeForm.SetBoundsPixels(bounds)
	} else {
		for _, layout := range layouts {
			if widget, ok := layout.Container().(Widget); ok && widget.Form() != appSingleton.activeForm {
				continue
			}

			layout.Update(false)
		}
	}
}

type Margins struct {
	HNear, VNear, HFar, VFar int
}

func (m Margins) isZero() bool {
	return m.HNear == 0 && m.HFar == 0 && m.VNear == 0 && m.VFar == 0
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
	MinSizeForSize(size Size) Size
	Update(reset bool) error
}

type LayoutBase struct {
	layout             Layout
	container          Container
	sizeAndDPI2MinSize map[sizeAndDPI]Size
	margins96dpi       Margins
	margins            Margins
	spacing96dpi       int
	spacing            int
	alignment          Alignment2D
	resetNeeded        bool
	dirty              bool
}

func (l *LayoutBase) asLayoutBase() *LayoutBase {
	return l
}

func (l *LayoutBase) sizeAndDPIToMinSize() map[sizeAndDPI]Size {
	return l.sizeAndDPI2MinSize
}

func (l *LayoutBase) Container() Container {
	return l.container
}

func (l *LayoutBase) SetContainer(value Container) {
	if value == l.container {
		return
	}

	if l.container != nil {
		l.container.SetLayout(nil)
	}

	l.container = value

	if value != nil && value.Layout() != l.layout {
		value.SetLayout(l.layout)
	}

	l.updateMargins()
	l.updateSpacing()

	l.layout.Update(true)
}

func (l *LayoutBase) Margins() Margins {
	return l.margins96dpi
}

func (l *LayoutBase) SetMargins(value Margins) error {
	if value == l.margins96dpi {
		return nil
	}

	if value.HNear < 0 || value.VNear < 0 || value.HFar < 0 || value.VFar < 0 {
		return newError("margins must be positive")
	}

	l.margins96dpi = value

	l.updateMargins()

	l.layout.Update(false)

	return nil
}

func (l *LayoutBase) Spacing() int {
	return l.spacing96dpi
}

func (l *LayoutBase) SetSpacing(value int) error {
	if value == l.spacing96dpi {
		return nil
	}

	if value < 0 {
		return newError("spacing cannot be negative")
	}

	l.spacing96dpi = value

	l.updateSpacing()

	l.layout.Update(false)

	return nil
}

func (l *LayoutBase) updateMargins() {
	if l.container != nil {
		l.margins = l.container.AsWindowBase().MarginsFrom96DPI(l.margins96dpi)
	}
}

func (l *LayoutBase) updateSpacing() {
	if l.container != nil {
		l.spacing = l.container.AsWindowBase().IntFrom96DPI(l.spacing96dpi)
	}
}

func (l *LayoutBase) Alignment() Alignment2D {
	return l.alignment
}

func (l *LayoutBase) SetAlignment(alignment Alignment2D) error {
	if alignment != l.alignment {
		if alignment < AlignHVDefault || alignment > AlignHFarVFar {
			return newError("invalid Alignment value")
		}

		l.alignment = alignment

		l.layout.Update(false)
	}

	return nil
}

func clearDescendantsLayoutMinSizeCache(root Container) {
	clearLayoutMinSizeCache(root)

	walkDescendants(root, func(w Window) bool {
		if w.Handle() == root.Handle() {
			return true
		}

		if container, ok := w.(Container); ok {
			clearLayoutMinSizeCache(container)
		}

		return true
	})
}

func clearLayoutMinSizeCache(container Container) {
	layout := container.Layout()
	if layout == nil {
		return
	}

	if lb, ok := layout.(interface{ sizeAndDPIToMinSize() map[sizeAndDPI]Size }); ok {
		sizeAndDPI2MinSize := lb.sizeAndDPIToMinSize()

		for k := range sizeAndDPI2MinSize {
			delete(sizeAndDPI2MinSize, k)
		}
	}
}

type sizeAndDPI struct {
	size Size
	dpi  int
}

type HeightForWidther interface {
	HeightForWidth(width int) int
}

type minSizeForSize struct {
	size    Size
	minSize Size
}

type layoutResultItem struct {
	widget Widget
	bounds Rectangle
}

func applyLayoutResults(container Container, items []layoutResultItem) error {
	hdwp := win.BeginDeferWindowPos(int32(len(items)))
	if hdwp == 0 {
		return lastError("BeginDeferWindowPos")
	}

	maybeInvalidate := container.AsContainerBase().hasComplexBackground()

	for _, item := range items {
		widget := item.widget
		x, y, w, h := item.bounds.X, item.bounds.Y, item.bounds.Width, item.bounds.Height

		b := widget.BoundsPixels()

		if w == b.Width && h == b.Height {
			if container, ok := widget.(Container); ok {
				if layout := container.Layout(); layout != nil {
					var dirty bool
					if lb, ok := layout.(interface{ asLayoutBase() *LayoutBase }); ok {
						dirty = lb.asLayoutBase().dirty
					}
					if dirty {
						layout.Update(false)
					} else {
						walkDescendants(container, func(window Window) bool {
							if asTV, ok := window.(interface{ asTableView() *TableView }); ok {
								asTV.asTableView().updateLVSizesWithSpecialCare(true)
							}

							return true
						})
					}
				}
			}

			if x == b.X && y == b.Y {
				continue
			}
		}

		if b.X == x && b.Y == y && b.Width == w {
			if _, ok := widget.(*ComboBox); ok {
				if b.Height+1 == h {
					continue
				}
			} else if b.Height == h {
				continue
			}
		}

		if maybeInvalidate {
			if w == b.Width && h == b.Height && (x != b.X || y != b.Y) {
				widget.Invalidate()
			}
		}

		if hdwp = win.DeferWindowPos(
			hdwp,
			widget.Handle(),
			0,
			int32(x),
			int32(y),
			int32(w),
			int32(h),
			win.SWP_NOACTIVATE|win.SWP_NOOWNERZORDER|win.SWP_NOZORDER); hdwp == 0 {

			return lastError("DeferWindowPos")
		}

		if item.widget.GraphicsEffects().Len() == 0 {
			continue
		}

		item.widget.AsWidgetBase().invalidateBorderInParent()
	}

	if !win.EndDeferWindowPos(hdwp) {
		return lastError("EndDeferWindowPos")
	}

	return nil
}

type applyLayoutResultsItem struct {
	hwnd                           win.HWND
	x                              int32
	y                              int32
	w                              int32
	h                              int32
	oldBounds                      win.RECT
	shouldInvalidateBorderInParent win.BOOL
}

func widgetsToLayout(allWidgets *WidgetList) []Widget {
	filteredWidgets := make([]Widget, 0, allWidgets.Len())

	for i := 0; i < cap(filteredWidgets); i++ {
		widget := allWidgets.At(i)

		if !shouldLayoutWidget(widget) {
			continue
		}

		ps := widget.SizeHint()
		if ps.Width == 0 && ps.Height == 0 && widget.LayoutFlags() == 0 {
			continue
		}

		filteredWidgets = append(filteredWidgets, widget)
	}

	return filteredWidgets
}

func shouldLayoutWidget(widget Widget) bool {
	if widget == nil {
		return false
	}

	_, isSpacer := widget.(*Spacer)

	return isSpacer || widget.AsWindowBase().visible || widget.AlwaysConsumeSpace()
}

func anyVisibleWidgetInHierarchy(root *WidgetBase) bool {
	if root == nil || !root.visible {
		return false
	}

	if container, ok := root.window.(Container); ok && container.Children() != nil {
		for _, child := range container.Children().items {
			if anyVisibleWidgetInHierarchy(child) {
				return true
			}
		}
	} else if _, ok := root.window.(*Spacer); !ok {
		return true
	}

	return false
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
	layout      Layout
	children    *WidgetList
	dataBinder  *DataBinder
	nextChildID int32
	persistent  bool
}

func (cb *ContainerBase) AsWidgetBase() *WidgetBase {
	return &cb.WidgetBase
}

func (cb *ContainerBase) AsContainerBase() *ContainerBase {
	return cb
}

func (cb *ContainerBase) NextChildID() int32 {
	cb.nextChildID++
	return cb.nextChildID
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

func (cb *ContainerBase) HeightForWidth(width int) int {
	if cb.layout == nil {
		return 0
	}

	return cb.layout.MinSizeForSize(Size{Width: width}).Height
}

func (cb *ContainerBase) applyEnabled(enabled bool) {
	cb.WidgetBase.applyEnabled(enabled)

	applyEnabledToDescendants(cb.window.(Widget), enabled)
}

func (cb *ContainerBase) applyFont(font *Font) {
	cb.WidgetBase.applyFont(font)

	applyFontToDescendants(cb.window.(Widget), font)
}

func (cb *ContainerBase) ApplySysColors() {
	cb.WidgetBase.ApplySysColors()

	applySysColorsToDescendants(cb.window.(Widget))
}

func (cb *ContainerBase) ApplyDPI(dpi int) {
	cb.WidgetBase.ApplyDPI(dpi)

	applyDPIToDescendants(cb.window.(Widget), dpi)

	if cb.layout != nil {
		if ums, ok := cb.layout.(interface {
			updateMargins()
			updateSpacing()
		}); ok {
			ums.updateMargins()
			ums.updateSpacing()
		}

		cb.layout.Update(false)
	}
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

	for _, wb := range cb.children.items {
		if persistable, ok := wb.window.(Persistable); ok && persistable.Persistent() {
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
		clearDescendantsLayoutMinSizeCache(cb)

		cb.layout.Update(false)
	}
}

func (cb *ContainerBase) doPaint() error {
	var ps win.PAINTSTRUCT

	hdc := win.BeginPaint(cb.hWnd, &ps)
	defer win.EndPaint(cb.hWnd, &ps)

	canvas, err := newCanvasFromHDC(hdc)
	if err != nil {
		return err
	}
	defer canvas.Dispose()

	for _, wb := range cb.children.items {
		widget := wb.window.(Widget)

		for _, effect := range widget.GraphicsEffects().items {
			switch effect {
			case InteractionEffect:
				type ReadOnlyer interface {
					ReadOnly() bool
				}
				if ro, ok := widget.(ReadOnlyer); ok {
					if ro.ReadOnly() {
						continue
					}
				}

				if hwnd := widget.Handle(); !win.IsWindowEnabled(hwnd) || !win.IsWindowVisible(hwnd) {
					continue
				}

			case FocusEffect:
				continue
			}

			b := widget.BoundsPixels().toRECT()
			win.ExcludeClipRect(hdc, b.Left, b.Top, b.Right, b.Bottom)

			if err := effect.Draw(widget, canvas); err != nil {
				return err
			}
		}
	}

	if FocusEffect != nil {
		hwndFocused := win.GetFocus()
		var widget Widget
		if wnd := windowFromHandle(hwndFocused); wnd != nil {
			widget, _ = wnd.(Widget)
		}
		for hwndFocused != 0 && (widget == nil || widget.Parent() == nil) {
			hwndFocused = win.GetParent(hwndFocused)
			if wnd := windowFromHandle(hwndFocused); wnd != nil {
				widget, _ = wnd.(Widget)
			}
		}

		if widget != nil && widget.Parent() != nil && widget.Parent().Handle() == cb.hWnd {
			for _, effect := range widget.GraphicsEffects().items {
				if effect == FocusEffect {
					b := widget.BoundsPixels().toRECT()
					win.ExcludeClipRect(hdc, b.Left, b.Top, b.Right, b.Bottom)

					if err := FocusEffect.Draw(widget, canvas); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (cb *ContainerBase) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_CTLCOLOREDIT, win.WM_CTLCOLORSTATIC:
		if hBrush := cb.handleWMCTLCOLOR(wParam, lParam); hBrush != 0 {
			return hBrush
		}

	case win.WM_PAINT:
		if err := cb.doPaint(); err != nil {
			panic(err)
		}

		return 0

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

	case win.WM_MEASUREITEM:
		mis := (*win.MEASUREITEMSTRUCT)(unsafe.Pointer(lParam))
		if window := windowFromHandle(win.GetDlgItem(hwnd, int32(mis.CtlID))); window != nil {
			// The window that sent the notification shall handle it itself.
			return window.WndProc(hwnd, msg, wParam, lParam)
		}

	case win.WM_DRAWITEM:
		dis := (*win.DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
		if window := windowFromHandle(dis.HwndItem); window != nil {
			// The window that sent the notification shall handle it itself.
			return window.WndProc(hwnd, msg, wParam, lParam)
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

		if cb.background == nullBrushSingleton {
			cb.Invalidate()
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

func (cb *ContainerBase) focusFirstCandidateDescendant() {
	window := firstFocusableDescendant(cb)
	if window == nil {
		return
	}

	if err := window.SetFocus(); err != nil {
		return
	}

	if textSel, ok := window.(textSelectable); ok {
		time.AfterFunc(time.Millisecond, func() {
			window.Synchronize(func() {
				if window.Focused() {
					textSel.SetTextSelection(0, -1)
				}
			})
		})
	}
}

func firstFocusableDescendantCallback(hwnd win.HWND, lParam uintptr) uintptr {
	if !win.IsWindowVisible(hwnd) || !win.IsWindowEnabled(hwnd) {
		return 1
	}

	if win.GetWindowLong(hwnd, win.GWL_STYLE)&win.WS_TABSTOP > 0 {
		hwndPtr := (*win.HWND)(unsafe.Pointer(lParam))
		*hwndPtr = hwnd
		return 0
	}

	return 1
}

var firstFocusableDescendantCallbackPtr = syscall.NewCallback(firstFocusableDescendantCallback)

func firstFocusableDescendant(container Container) Window {
	var hwnd win.HWND

	win.EnumChildWindows(container.Handle(), firstFocusableDescendantCallbackPtr, uintptr(unsafe.Pointer(&hwnd)))

	window := windowFromHandle(hwnd)

	for hwnd != 0 && window == nil {
		hwnd = win.GetParent(hwnd)
		window = windowFromHandle(hwnd)
	}

	return window
}

type textSelectable interface {
	SetTextSelection(start, end int)
}
