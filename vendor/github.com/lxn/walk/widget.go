// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

// LayoutFlags specify how a Widget wants to be treated when used with a Layout.
//
// These flags are interpreted in respect to Widget.SizeHint.
type LayoutFlags byte

const (
	// ShrinkableHorz allows a Widget to be shrunk horizontally.
	ShrinkableHorz LayoutFlags = 1 << iota

	// ShrinkableVert allows a Widget to be shrunk vertically.
	ShrinkableVert

	// GrowableHorz allows a Widget to be enlarged horizontally.
	GrowableHorz

	// GrowableVert allows a Widget to be enlarged vertically.
	GrowableVert

	// GreedyHorz specifies that the widget prefers to take up as much space as
	// possible, horizontally.
	GreedyHorz

	// GreedyVert specifies that the widget prefers to take up as much space as
	// possible, vertically.
	GreedyVert
)

type Widget interface {
	Window

	// Alignment returns the alignment of the Widget.
	Alignment() Alignment2D

	// AlwaysConsumeSpace returns if the Widget should consume space even if it
	// is not visible.
	AlwaysConsumeSpace() bool

	// AsWidgetBase returns a *WidgetBase that implements Widget.
	AsWidgetBase() *WidgetBase

	// Form returns the root ancestor Form of the Widget.
	Form() Form

	// GraphicsEffects returns a list of WidgetGraphicsEffects that are applied to the Widget.
	GraphicsEffects() *WidgetGraphicsEffectList

	// LayoutFlags returns a combination of LayoutFlags that specify how the
	// Widget wants to be treated by Layout implementations.
	LayoutFlags() LayoutFlags

	// MinSizeHint returns the minimum outer Size, including decorations, that
	// makes sense for the respective type of Widget.
	MinSizeHint() Size

	// Parent returns the Container of the Widget.
	Parent() Container

	// SetAlignment sets the alignment of the widget.
	SetAlignment(alignment Alignment2D) error

	// SetAlwaysConsumeSpace sets if the Widget should consume space even if it
	// is not visible.
	SetAlwaysConsumeSpace(b bool) error

	// SetParent sets the parent of the Widget and adds the Widget to the
	// Children list of the Container.
	SetParent(value Container) error

	// SetToolTipText sets the tool tip text of the Widget.
	SetToolTipText(s string) error

	// SizeHint returns the preferred Size for the respective type of Widget.
	SizeHint() Size

	// ToolTipText returns the tool tip text of the Widget.
	ToolTipText() string
}

type WidgetBase struct {
	WindowBase
	form                        Form
	parent                      Container
	toolTipTextProperty         Property
	toolTipTextChangedPublisher EventPublisher
	graphicsEffects             *WidgetGraphicsEffectList
	alignment                   Alignment2D
	alwaysConsumeSpace          bool
}

// InitWidget initializes a Widget.
func InitWidget(widget Widget, parent Window, className string, style, exStyle uint32) error {
	if parent == nil {
		return newError("parent cannot be nil")
	}

	if err := InitWindow(widget, parent, className, style|win.WS_CHILD, exStyle); err != nil {
		return err
	}

	if container, ok := parent.(Container); ok {
		if container.Children() == nil {
			// Required by parents like MainWindow and GroupBox.
			if win.SetParent(widget.Handle(), container.Handle()) == 0 {
				return lastError("SetParent")
			}
		} else {
			if err := container.Children().Add(widget); err != nil {
				return err
			}
		}
	}

	return nil
}

func (wb *WidgetBase) init(widget Widget) error {
	wb.graphicsEffects = newWidgetGraphicsEffectList(wb)

	if err := globalToolTip.AddTool(wb); err != nil {
		return err
	}

	wb.toolTipTextProperty = NewProperty(
		func() interface{} {
			return wb.window.(Widget).ToolTipText()
		},
		func(v interface{}) error {
			wb.window.(Widget).SetToolTipText(assertStringOr(v, ""))
			return nil
		},
		wb.toolTipTextChangedPublisher.Event())

	wb.MustRegisterProperty("ToolTipText", wb.toolTipTextProperty)

	return nil
}

func (wb *WidgetBase) Dispose() {
	if wb.hWnd == 0 {
		return
	}

	globalToolTip.RemoveTool(wb)

	wb.WindowBase.Dispose()
}

// AsWidgetBase just returns the receiver.
func (wb *WidgetBase) AsWidgetBase() *WidgetBase {
	return wb
}

// Bounds returns the outer bounding box Rectangle of the WidgetBase, including
// decorations.
//
// The coordinates are relative to the parent of the Widget.
func (wb *WidgetBase) Bounds() Rectangle {
	return wb.RectangleTo96DPI(wb.BoundsPixels())
}

// BoundsPixels returns the outer bounding box Rectangle of the WidgetBase, including
// decorations.
//
// The coordinates are relative to the parent of the Widget.
func (wb *WidgetBase) BoundsPixels() Rectangle {
	b := wb.WindowBase.BoundsPixels()

	if wb.parent != nil {
		p := win.POINT{int32(b.X), int32(b.Y)}
		if !win.ScreenToClient(wb.parent.Handle(), &p) {
			newError("ScreenToClient failed")
			return Rectangle{}
		}
		b.X = int(p.X)
		b.Y = int(p.Y)
	}

	return b
}

// BringToTop moves the WidgetBase to the top of the keyboard focus order.
func (wb *WidgetBase) BringToTop() error {
	if wb.parent != nil {
		if err := wb.parent.BringToTop(); err != nil {
			return err
		}
	}

	return wb.WindowBase.BringToTop()
}

// Enabled returns if the WidgetBase is enabled for user interaction.
func (wb *WidgetBase) Enabled() bool {
	if wb.parent != nil {
		return wb.enabled && wb.parent.Enabled()
	}

	return wb.enabled
}

// Font returns the Font of the WidgetBase.
//
// By default this is a MS Shell Dlg 2, 8 point font.
func (wb *WidgetBase) Font() *Font {
	if wb.font != nil {
		return wb.font
	} else if wb.parent != nil {
		return wb.parent.Font()
	}

	return defaultFont
}

func (wb *WidgetBase) applyFont(font *Font) {
	wb.WindowBase.applyFont(font)

	wb.updateParentLayout()
}

// Form returns the root ancestor Form of the Widget.
func (wb *WidgetBase) Form() Form {
	if wb.form == nil {
		wb.form = ancestor(wb)
	}

	return wb.form
}

// Alignment return the alignment ot the *WidgetBase.
func (wb *WidgetBase) Alignment() Alignment2D {
	return wb.alignment
}

// SetAlignment sets the alignment of the *WidgetBase.
func (wb *WidgetBase) SetAlignment(alignment Alignment2D) error {
	if alignment != wb.alignment {
		if alignment < AlignHVDefault || alignment > AlignHFarVFar {
			return newError("invalid Alignment value")
		}

		wb.alignment = alignment

		wb.updateParentLayout()
	}

	return nil
}

// LayoutFlags returns a combination of LayoutFlags that specify how the
// WidgetBase wants to be treated by Layout implementations.
func (wb *WidgetBase) LayoutFlags() LayoutFlags {
	return 0
}

// SetMinMaxSize sets the minimum and maximum outer Size of the *WidgetBase,
// including decorations.
//
// Use walk.Size{} to make the respective limit be ignored.
func (wb *WidgetBase) SetMinMaxSize(min, max Size) (err error) {
	err = wb.WindowBase.SetMinMaxSize(min, max)

	wb.updateParentLayout()

	return
}

// AlwaysConsumeSpace returns if the Widget should consume space even if it is
// not visible.
func (wb *WidgetBase) AlwaysConsumeSpace() bool {
	return wb.alwaysConsumeSpace
}

// SetAlwaysConsumeSpace sets if the Widget should consume space even if it is
// not visible.
func (wb *WidgetBase) SetAlwaysConsumeSpace(b bool) error {
	wb.alwaysConsumeSpace = b

	return wb.updateParentLayout()
}

// MinSizeHint returns the minimum outer Size, including decorations, that
// makes sense for the respective type of Widget.
func (wb *WidgetBase) MinSizeHint() Size {
	return Size{10, 10}
}

// Parent returns the Container of the WidgetBase.
func (wb *WidgetBase) Parent() Container {
	return wb.parent
}

// SetParent sets the parent of the WidgetBase and adds the WidgetBase to the
// Children list of the Container.
func (wb *WidgetBase) SetParent(parent Container) (err error) {
	if parent == wb.parent {
		return nil
	}

	style := uint32(win.GetWindowLong(wb.hWnd, win.GWL_STYLE))
	if style == 0 {
		return lastError("GetWindowLong")
	}

	if parent == nil {
		wb.SetVisible(false)

		style &^= win.WS_CHILD
		style |= win.WS_POPUP

		if win.SetParent(wb.hWnd, 0) == 0 {
			return lastError("SetParent")
		}
		win.SetLastError(0)
		if win.SetWindowLong(wb.hWnd, win.GWL_STYLE, int32(style)) == 0 {
			return lastError("SetWindowLong")
		}
	} else {
		style |= win.WS_CHILD
		style &^= win.WS_POPUP

		win.SetLastError(0)
		if win.SetWindowLong(wb.hWnd, win.GWL_STYLE, int32(style)) == 0 {
			return lastError("SetWindowLong")
		}
		if win.SetParent(wb.hWnd, parent.Handle()) == 0 {
			return lastError("SetParent")
		}

		if cb := parent.AsContainerBase(); cb != nil {
			if win.SetWindowLong(wb.hWnd, win.GWL_ID, cb.NextChildID()) == 0 {
				return lastError("SetWindowLong")
			}
		}
	}

	b := wb.BoundsPixels()

	if !win.SetWindowPos(
		wb.hWnd,
		win.HWND_BOTTOM,
		int32(b.X),
		int32(b.Y),
		int32(b.Width),
		int32(b.Height),
		win.SWP_FRAMECHANGED) {

		return lastError("SetWindowPos")
	}

	oldParent := wb.parent

	wb.parent = parent

	var oldChildren, newChildren *WidgetList
	if oldParent != nil {
		oldChildren = oldParent.Children()
	}
	if parent != nil {
		newChildren = parent.Children()
	}

	if newChildren == oldChildren {
		return nil
	}

	widget := wb.window.(Widget)

	if oldChildren != nil {
		oldChildren.Remove(widget)
	}

	if newChildren != nil && !newChildren.containsHandle(wb.hWnd) {
		newChildren.Add(widget)
	}

	return nil
}

func (wb *WidgetBase) ForEachAncestor(f func(window Window) bool) {
	hwnd := win.GetParent(wb.hWnd)

	for hwnd != 0 {
		if window := windowFromHandle(hwnd); window != nil {
			if !f(window) {
				return
			}
		}

		hwnd = win.GetParent(hwnd)
	}
}

// SizeHint returns a default Size that should be "overidden" by a concrete
// Widget type.
func (wb *WidgetBase) SizeHint() Size {
	return wb.window.(Widget).MinSizeHint()
}

// ToolTipText returns the tool tip text of the WidgetBase.
func (wb *WidgetBase) ToolTipText() string {
	return globalToolTip.Text(wb.window.(Widget))
}

// SetToolTipText sets the tool tip text of the WidgetBase.
func (wb *WidgetBase) SetToolTipText(s string) error {
	if err := globalToolTip.SetText(wb.window.(Widget), s); err != nil {
		return err
	}

	wb.toolTipTextChangedPublisher.Publish()

	return nil
}

// GraphicsEffects returns a list of WidgetGraphicsEffects that are applied to the WidgetBase.
func (wb *WidgetBase) GraphicsEffects() *WidgetGraphicsEffectList {
	return wb.graphicsEffects
}

func (wb *WidgetBase) onInsertedGraphicsEffect(index int, effect WidgetGraphicsEffect) error {
	wb.invalidateBorderInParent()

	return nil
}

func (wb *WidgetBase) onRemovedGraphicsEffect(index int, effect WidgetGraphicsEffect) error {
	wb.invalidateBorderInParent()

	return nil
}

func (wb *WidgetBase) onClearedGraphicsEffects() error {
	wb.invalidateBorderInParent()

	return nil
}

func (wb *WidgetBase) invalidateBorderInParent() {
	if wb.parent != nil && wb.parent.Layout() != nil {
		b := wb.BoundsPixels().toRECT()
		s := int32(wb.parent.Layout().Spacing())

		hwnd := wb.parent.Handle()

		rc := win.RECT{Left: b.Left - s, Top: b.Top - s, Right: b.Left, Bottom: b.Bottom + s}
		win.InvalidateRect(hwnd, &rc, true)

		rc = win.RECT{Left: b.Right, Top: b.Top - s, Right: b.Right + s, Bottom: b.Bottom + s}
		win.InvalidateRect(hwnd, &rc, true)

		rc = win.RECT{Left: b.Left, Top: b.Top - s, Right: b.Right, Bottom: b.Top}
		win.InvalidateRect(hwnd, &rc, true)

		rc = win.RECT{Left: b.Left, Top: b.Bottom, Right: b.Right, Bottom: b.Bottom + s}
		win.InvalidateRect(hwnd, &rc, true)
	}
}

func (wb *WidgetBase) hasComplexBackground() bool {
	if bg := wb.window.Background(); bg != nil && !bg.simple() {
		return false
	}

	var complex bool
	wb.ForEachAncestor(func(window Window) bool {
		if bg := window.Background(); bg != nil && !bg.simple() {
			complex = true
			return false
		}

		return true
	})

	return complex
}

func (wb *WidgetBase) updateParentLayout() error {
	return wb.updateParentLayoutWithReset(false)
}

func (wb *WidgetBase) updateParentLayoutWithReset(reset bool) error {
	parent := wb.window.(Widget).Parent()

	if parent == nil || parent.Layout() == nil {
		return nil
	}

	if lb, ok := parent.Layout().(interface{ asLayoutBase() *LayoutBase }); ok {
		lb.asLayoutBase().dirty = true
	}

	return wb.updateParentLayoutWithResetRecursive(reset)
}

func (wb *WidgetBase) updateParentLayoutWithResetRecursive(reset bool) error {
	if form := wb.Form(); form == nil || form.Suspended() {
		return nil
	}

	parent := wb.window.(Widget).Parent()

	if parent == nil || parent.Layout() == nil {
		return nil
	}

	layout := parent.Layout()

	updateLayoutAndMaybeInvalidateBorder := func() {
		layout.Update(reset)

		if FocusEffect != nil {
			if focusedWnd := windowFromHandle(win.GetFocus()); focusedWnd != nil && win.GetParent(focusedWnd.Handle()) == parent.Handle() {
				focusedWnd.(Widget).AsWidgetBase().invalidateBorderInParent()
			}
		}
	}

	if !formResizeScheduled || len(inProgressEventsByForm[appSingleton.activeForm]) == 0 {
		switch wnd := parent.(type) {
		case *ScrollView:
			ifContainerIsScrollViewDoCoolSpecialLayoutStuff(layout)
			wnd.updateCompositeSize()
			updateLayoutAndMaybeInvalidateBorder()
			return nil

		case Widget:
			return wnd.AsWidgetBase().updateParentLayoutWithResetRecursive(reset)

		case Form:
			if len(inProgressEventsByForm[appSingleton.activeForm]) > 0 {
				formResizeScheduled = true
			} else {
				bounds := wnd.BoundsPixels()

				if wnd.AsFormBase().fixedSize() {
					bounds.Width, bounds.Height = 0, 0
				}

				wnd.SetBoundsPixels(bounds)

				return nil
			}
		}
	}

	updateLayoutAndMaybeInvalidateBorder()

	return nil
}

func ancestor(w Widget) Form {
	if w == nil {
		return nil
	}

	hWndRoot := win.GetAncestor(w.Handle(), win.GA_ROOT)

	rw, _ := windowFromHandle(hWndRoot).(Form)
	return rw
}

func minSizeEffective(w Widget) Size {
	s := maxSize(w.MinSizePixels(), w.MinSizeHint())

	max := w.MaxSizePixels()
	if max.Width > 0 && s.Width > max.Width {
		s.Width = max.Width
	}
	if max.Height > 0 && s.Height > max.Height {
		s.Height = max.Height
	}

	return s
}
