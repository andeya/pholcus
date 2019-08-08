// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"bytes"
	"fmt"
	"image"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

// App-specific message ids for internal use in Walk.
// TODO: Document reserved range somewhere (when we have an idea how many we need).
const (
	notifyIconMessageId = win.WM_APP + iota
)

// Window is an interface that provides operations common to all windows.
type Window interface {
	// AddDisposable adds a Disposable resource that should be disposed of
	// together with this Window.
	AddDisposable(d Disposable)

	// AsWindowBase returns a *WindowBase, a pointer to an instance of the
	// struct that implements most operations common to all windows.
	AsWindowBase() *WindowBase

	// Background returns the background Brush of the Window.
	//
	// By default this is nil.
	Background() Brush

	// Bounds returns the outer bounding box Rectangle of the Window, including
	// decorations.
	//
	// For a Form, like *MainWindow or *Dialog, the Rectangle is in screen
	// coordinates, for a child Window the coordinates are relative to its
	// parent.
	Bounds() Rectangle

	// BoundsPixels returns the outer bounding box Rectangle of the Window, including
	// decorations.
	//
	// For a Form, like *MainWindow or *Dialog, the Rectangle is in screen
	// coordinates, for a child Window the coordinates are relative to its
	// parent.
	BoundsPixels() Rectangle

	// BoundsChanged returns an *Event that you can attach to for handling bounds
	// changed events for the Window.
	BoundsChanged() *Event

	// BringToTop moves the Window to the top of the keyboard focus order.
	BringToTop() error

	// ClientBounds returns the inner bounding box Rectangle of the Window,
	// excluding decorations.
	ClientBounds() Rectangle

	// ClientBoundsPixels returns the inner bounding box Rectangle of the Window,
	// excluding decorations.
	ClientBoundsPixels() Rectangle

	// ContextMenu returns the context menu of the Window.
	//
	// By default this is nil.
	ContextMenu() *Menu

	// CreateCanvas creates and returns a *Canvas that can be used to draw
	// inside the ClientBoundsPixels of the Window.
	//
	// Remember to call the Dispose method on the canvas to release resources,
	// when you no longer need it.
	CreateCanvas() (*Canvas, error)

	// Cursor returns the Cursor of the Window.
	//
	// By default this is nil.
	Cursor() Cursor

	// Dispose releases the operating system resources, associated with the
	// Window.
	//
	// If a user closes a *MainWindow or *Dialog, it is automatically released.
	// Also, if a Container is disposed of, all its descendants will be released
	// as well.
	Dispose()

	// Disposing returns an Event that is published when the Window is disposed
	// of.
	Disposing() *Event

	// DoubleBuffering returns whether double buffering of the
	// drawing is enabled, which may help reduce flicker.
	DoubleBuffering() bool

	// DPI returns the current DPI value of the Window.
	DPI() int

	// Enabled returns if the Window is enabled for user interaction.
	Enabled() bool

	// Focused returns whether the Window has the keyboard input focus.
	Focused() bool

	// FocusedChanged returns an Event that you can attach to for handling focus
	// changed events for the Window.
	FocusedChanged() *Event

	// Font returns the *Font of the Window.
	//
	// By default this is a MS Shell Dlg 2, 8 point font.
	Font() *Font

	// Handle returns the window handle of the Window.
	Handle() win.HWND

	// Height returns the outer height of the Window, including decorations.
	Height() int

	// HeightPixels returns the outer height of the Window, including decorations.
	HeightPixels() int

	// Invalidate schedules a full repaint of the Window.
	Invalidate() error

	// IsDisposed returns if the Window has been disposed of.
	IsDisposed() bool

	// KeyDown returns a *KeyEvent that you can attach to for handling key down
	// events for the Window.
	KeyDown() *KeyEvent

	// KeyPress returns a *KeyEvent that you can attach to for handling key
	// press events for the Window.
	KeyPress() *KeyEvent

	// KeyUp returns a *KeyEvent that you can attach to for handling key up
	// events for the Window.
	KeyUp() *KeyEvent

	// MaxSize returns the maximum allowed outer Size for the Window, including
	// decorations.
	//
	// For child windows, this is only relevant when the parent of the Window
	// has a Layout. RootWidgets, like *MainWindow and *Dialog, also honor this.
	MaxSize() Size

	// MaxSizePixels returns the maximum allowed outer Size for the Window, including
	// decorations.
	//
	// For child windows, this is only relevant when the parent of the Window
	// has a Layout. RootWidgets, like *MainWindow and *Dialog, also honor this.
	MaxSizePixels() Size

	// MinSize returns the minimum allowed outer Size for the Window, including
	// decorations.
	//
	// For child windows, this is only relevant when the parent of the Window
	// has a Layout. RootWidgets, like *MainWindow and *Dialog, also honor this.
	MinSize() Size

	// MinSizePixels returns the minimum allowed outer Size for the Window, including
	// decorations.
	//
	// For child windows, this is only relevant when the parent of the Window
	// has a Layout. RootWidgets, like *MainWindow and *Dialog, also honor this.
	MinSizePixels() Size

	// MouseDown returns a *MouseEvent that you can attach to for handling
	// mouse down events for the Window.
	MouseDown() *MouseEvent

	// MouseMove returns a *MouseEvent that you can attach to for handling
	// mouse move events for the Window.
	MouseMove() *MouseEvent

	// MouseUp returns a *MouseEvent that you can attach to for handling
	// mouse up events for the Window.
	MouseUp() *MouseEvent

	// Name returns the name of the Window.
	Name() string

	// RightToLeftReading returns whether the reading order of the Window
	// is from right to left.
	RightToLeftReading() bool

	// Screenshot returns an image of the window.
	Screenshot() (*image.RGBA, error)

	// SendMessage sends a message to the window and returns the result.
	SendMessage(msg uint32, wParam, lParam uintptr) uintptr

	// SetBackground sets the background Brush of the Window.
	SetBackground(value Brush)

	// SetBounds sets the outer bounding box Rectangle of the Window, including
	// decorations.
	//
	// For a Form, like *MainWindow or *Dialog, the Rectangle is in screen
	// coordinates, for a child Window the coordinates are relative to its
	// parent.
	SetBounds(value Rectangle) error

	// SetBoundsPixels sets the outer bounding box Rectangle of the Window, including
	// decorations.
	//
	// For a Form, like *MainWindow or *Dialog, the Rectangle is in screen
	// coordinates, for a child Window the coordinates are relative to its
	// parent.
	SetBoundsPixels(value Rectangle) error

	// SetClientSize sets the Size of the inner bounding box of the Window,
	// excluding decorations.
	SetClientSize(value Size) error

	// SetClientSizePixels sets the Size of the inner bounding box of the Window,
	// excluding decorations.
	SetClientSizePixels(value Size) error

	// SetContextMenu sets the context menu of the Window.
	SetContextMenu(value *Menu)

	// SetCursor sets the Cursor of the Window.
	SetCursor(value Cursor)

	// SetDoubleBuffering enables or disables double buffering of the
	// drawing, which may help reduce flicker.
	SetDoubleBuffering(value bool) error

	// SetEnabled sets if the Window is enabled for user interaction.
	SetEnabled(value bool)

	// SetFocus sets the keyboard input focus to the Window.
	SetFocus() error

	// SetFont sets the *Font of the Window.
	SetFont(value *Font)

	// SetHeight sets the outer height of the Window, including decorations.
	SetHeight(value int) error

	// SetHeightPixels sets the outer height of the Window, including decorations.
	SetHeightPixels(value int) error

	// SetMinMaxSize sets the minimum and maximum outer Size of the Window,
	// including decorations.
	//
	// Use walk.Size{} to make the respective limit be ignored.
	SetMinMaxSize(min, max Size) error

	// SetMinMaxSizePixels sets the minimum and maximum outer Size of the Window,
	// including decorations.
	//
	// Use walk.Size{} to make the respective limit be ignored.
	SetMinMaxSizePixels(min, max Size) error

	// SetName sets the name of the Window.
	//
	// This is important if you want to make use of the built-in UI persistence.
	// Some windows support automatic state persistence. See Settings for
	// details.
	SetName(name string)

	// SetRightToLeftReading sets whether the reading order of the Window
	// is from right to left.
	SetRightToLeftReading(rtl bool) error

	// SetSize sets the outer Size of the Window, including decorations.
	SetSize(value Size) error

	// SetSizePixels sets the outer Size of the Window, including decorations.
	SetSizePixels(value Size) error

	// SetSuspended sets if the Window is suspended for layout and repainting
	// purposes.
	//
	// You should call SetSuspended(true), before doing a batch of modifications
	// that would cause multiple layout or drawing updates. Remember to call
	// SetSuspended(false) afterwards, which will update the Window accordingly.
	SetSuspended(suspend bool)

	// SetVisible sets if the Window is visible.
	SetVisible(value bool)

	// SetWidth sets the outer width of the Window, including decorations.
	SetWidth(value int) error

	// SetWidthPixels sets the outer width of the Window, including decorations.
	SetWidthPixels(value int) error

	// SetX sets the x coordinate of the Window, relative to the screen for
	// RootWidgets like *MainWindow or *Dialog and relative to the parent for
	// child Windows.
	SetX(value int) error

	// SetXPixels sets the x coordinate of the Window, relative to the screen for
	// RootWidgets like *MainWindow or *Dialog and relative to the parent for
	// child Windows.
	SetXPixels(value int) error

	// SetY sets the y coordinate of the Window, relative to the screen for
	// RootWidgets like *MainWindow or *Dialog and relative to the parent for
	// child Windows.
	SetY(value int) error

	// SetYPixels sets the y coordinate of the Window, relative to the screen for
	// RootWidgets like *MainWindow or *Dialog and relative to the parent for
	// child Windows.
	SetYPixels(value int) error

	// Size returns the outer Size of the Window, including decorations.
	Size() Size

	// SizePixels returns the outer Size of the Window, including decorations.
	SizePixels() Size

	// SizeChanged returns an *Event that you can attach to for handling size
	// changed events for the Window.
	SizeChanged() *Event

	// Suspended returns if the Window is suspended for layout and repainting
	// purposes.
	Suspended() bool

	// Synchronize enqueues func f to be called some time later by the main
	// goroutine from inside a message loop.
	Synchronize(f func())

	// Visible returns if the Window is visible.
	Visible() bool

	// VisibleChanged returns an Event that you can attach to for handling
	// visible changed events for the Window.
	VisibleChanged() *Event

	// Width returns the outer width of the Window, including decorations.
	Width() int

	// WidthPixels returns the outer width of the Window, including decorations.
	WidthPixels() int

	// WndProc is the window procedure of the window.
	//
	// When implementing your own WndProc to add or modify behavior, call the
	// WndProc of the embedded window for messages you don't handle yourself.
	WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr

	// X returns the x coordinate of the Window, relative to the screen for
	// RootWidgets like *MainWindow or *Dialog and relative to the parent for
	// child Windows.
	X() int

	// XPixels returns the x coordinate of the Window, relative to the screen for
	// RootWidgets like *MainWindow or *Dialog and relative to the parent for
	// child Windows.
	XPixels() int

	// Y returns the y coordinate of the Window, relative to the screen for
	// RootWidgets like *MainWindow or *Dialog and relative to the parent for
	// child Windows.
	Y() int

	// YPixels returns the y coordinate of the Window, relative to the screen for
	// RootWidgets like *MainWindow or *Dialog and relative to the parent for
	// child Windows.
	YPixels() int
}

type calcTextSizeInfo struct {
	width int
	font  fontInfo
	text  string
	size  Size
	dpi   int
}

// WindowBase implements many operations common to all Windows.
type WindowBase struct {
	nopActionListObserver
	window                  Window
	hWnd                    win.HWND
	origWndProcPtr          uintptr
	name                    string
	font                    *Font
	hFont                   win.HFONT
	contextMenu             *Menu
	shortcutActions         *ActionList
	disposables             []Disposable
	disposingPublisher      EventPublisher
	dropFilesPublisher      DropFilesEventPublisher
	keyDownPublisher        KeyEventPublisher
	keyPressPublisher       KeyEventPublisher
	keyUpPublisher          KeyEventPublisher
	mouseDownPublisher      MouseEventPublisher
	mouseUpPublisher        MouseEventPublisher
	mouseMovePublisher      MouseEventPublisher
	mouseWheelPublisher     MouseEventPublisher
	boundsChangedPublisher  EventPublisher
	sizeChangedPublisher    EventPublisher
	maxSize                 Size
	minSize                 Size
	background              Brush
	cursor                  Cursor
	name2Property           map[string]Property
	enabledProperty         Property
	enabledChangedPublisher EventPublisher
	visibleProperty         Property
	visibleChangedPublisher EventPublisher
	focusedProperty         Property
	focusedChangedPublisher EventPublisher
	calcTextSizeInfoPrev    *calcTextSizeInfo
	suspended               bool
	visible                 bool
	enabled                 bool
}

var (
	registeredWindowClasses = make(map[string]bool)
	defaultWndProcPtr       = syscall.NewCallback(defaultWndProc)
	hwnd2WindowBase         = make(map[win.HWND]*WindowBase)
)

// MustRegisterWindowClass registers the specified window class.
//
// MustRegisterWindowClass must be called once for every window type that is not
// based on any system provided control, before calling InitChildWidget or
// InitWidget. Calling MustRegisterWindowClass twice with the same className
// results in a panic.
func MustRegisterWindowClass(className string) {
	MustRegisterWindowClassWithWndProcPtr(className, defaultWndProcPtr)
}

func MustRegisterWindowClassWithStyle(className string, style uint32) {
	MustRegisterWindowClassWithWndProcPtrAndStyle(className, defaultWndProcPtr, style)
}

func MustRegisterWindowClassWithWndProcPtr(className string, wndProcPtr uintptr) {
	MustRegisterWindowClassWithWndProcPtrAndStyle(className, wndProcPtr, 0)
}

func MustRegisterWindowClassWithWndProcPtrAndStyle(className string, wndProcPtr uintptr, style uint32) {
	if registeredWindowClasses[className] {
		panic("window class already registered")
	}

	hInst := win.GetModuleHandle(nil)
	if hInst == 0 {
		panic("GetModuleHandle")
	}

	hIcon := win.LoadIcon(hInst, win.MAKEINTRESOURCE(7)) // rsrc uses 7 for app icon
	if hIcon == 0 {
		hIcon = win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_APPLICATION))
	}
	if hIcon == 0 {
		panic("LoadIcon")
	}

	hCursor := win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))
	if hCursor == 0 {
		panic("LoadCursor")
	}

	var wc win.WNDCLASSEX
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = wndProcPtr
	wc.HInstance = hInst
	wc.HIcon = hIcon
	wc.HCursor = hCursor
	wc.HbrBackground = win.COLOR_BTNFACE + 1
	wc.LpszClassName = syscall.StringToUTF16Ptr(className)
	wc.Style = style

	if atom := win.RegisterClassEx(&wc); atom == 0 {
		panic("RegisterClassEx")
	}

	registeredWindowClasses[className] = true
}

var lockedOSThread bool

// InitWindow initializes a window.
//
// Widgets should be initialized using InitWidget instead.
func InitWindow(window, parent Window, className string, style, exStyle uint32) error {
	if !lockedOSThread {
		runtime.LockOSThread()
	}

	wb := window.AsWindowBase()
	wb.window = window
	wb.enabled = true
	wb.visible = true

	wb.name2Property = make(map[string]Property)

	var hwndParent win.HWND
	var hMenu win.HMENU
	if parent != nil {
		hwndParent = parent.Handle()

		if widget, ok := window.(Widget); ok {
			if container, ok := parent.(Container); ok {
				if cb := container.AsContainerBase(); cb != nil {
					hMenu = win.HMENU(cb.NextChildID())
				}
				widget.AsWidgetBase().parent = container
			}
		}
	}

	var windowName *uint16
	if len(wb.name) != 0 {
		windowName = syscall.StringToUTF16Ptr(wb.name)
	}

	if hwnd := window.Handle(); hwnd == 0 {
		wb.hWnd = win.CreateWindowEx(
			exStyle,
			syscall.StringToUTF16Ptr(className),
			windowName,
			style|win.WS_CLIPSIBLINGS,
			win.CW_USEDEFAULT,
			win.CW_USEDEFAULT,
			win.CW_USEDEFAULT,
			win.CW_USEDEFAULT,
			hwndParent,
			hMenu,
			0,
			nil)
		if wb.hWnd == 0 {
			return lastError("CreateWindowEx")
		}
	} else {
		wb.hWnd = hwnd
	}

	succeeded := false
	defer func() {
		if !succeeded {
			wb.Dispose()
		}
	}()

	hwnd2WindowBase[wb.hWnd] = wb

	if !registeredWindowClasses[className] {
		// We subclass all windows of system classes.
		wb.origWndProcPtr = win.SetWindowLongPtr(wb.hWnd, win.GWLP_WNDPROC, defaultWndProcPtr)
		if wb.origWndProcPtr == 0 {
			return lastError("SetWindowLongPtr")
		}
	}

	SetWindowFont(wb.hWnd, defaultFont)

	if form, ok := window.(Form); ok {
		if fb := form.AsFormBase(); fb != nil {
			if err := fb.init(form); err != nil {
				return err
			}
		}
	}

	if widget, ok := window.(Widget); ok {
		if wb := widget.AsWidgetBase(); wb != nil {
			if err := wb.init(widget); err != nil {
				return err
			}
		}
	}

	wb.enabledProperty = NewBoolProperty(
		func() bool {
			return wb.window.Enabled()
		},
		func(b bool) error {
			wb.window.SetEnabled(b)
			return nil
		},
		wb.enabledChangedPublisher.Event())

	wb.visibleProperty = NewBoolProperty(
		func() bool {
			return window.AsWindowBase().visible
		},
		func(b bool) error {
			wb.window.SetVisible(b)
			return nil
		},
		wb.visibleChangedPublisher.Event())

	wb.focusedProperty = NewReadOnlyBoolProperty(
		func() bool {
			return wb.window.Focused()
		},
		wb.focusedChangedPublisher.Event())

	wb.MustRegisterProperty("Enabled", wb.enabledProperty)
	wb.MustRegisterProperty("Visible", wb.visibleProperty)
	wb.MustRegisterProperty("Focused", wb.focusedProperty)

	succeeded = true

	return nil
}

// InitWrapperWindow initializes a window that wraps (embeds) another window.
//
// Calling this method is necessary, if you want to be able to override the
// WndProc method of the embedded window. The embedded window should only be
// used as inseparable part of the wrapper window to avoid undefined behavior.
func InitWrapperWindow(window Window) error {
	wb := window.AsWindowBase()

	wb.window = window

	if container, ok := window.(Container); ok {
		children := container.Children()

		if wlo, ok := window.(widgetListObserver); ok {
			children.observer = wlo
		}

		for _, child := range children.items {
			child.parent = container
		}
	}

	return nil
}

func (wb *WindowBase) MustRegisterProperty(name string, property Property) {
	if property == nil {
		panic("property must not be nil")
	}
	if wb.name2Property[name] != nil {
		panic("property already registered")
	}

	wb.name2Property[name] = property
}

func (wb *WindowBase) Property(name string) Property {
	return wb.name2Property[name]
}

func (wb *WindowBase) hasStyleBits(bits uint32) bool {
	return hasWindowLongBits(wb.hWnd, win.GWL_STYLE, bits)
}

func (wb *WindowBase) hasExtendedStyleBits(bits uint32) bool {
	return hasWindowLongBits(wb.hWnd, win.GWL_EXSTYLE, bits)
}

func hasWindowLongBits(hwnd win.HWND, index int32, bits uint32) bool {
	value := uint32(win.GetWindowLong(hwnd, index))

	return value&bits == bits
}

func (wb *WindowBase) setAndClearStyleBits(set, clear uint32) error {
	return setAndClearWindowLongBits(wb.hWnd, win.GWL_STYLE, set, clear)
}

func (wb *WindowBase) setAndClearExtendedStyleBits(set, clear uint32) error {
	return setAndClearWindowLongBits(wb.hWnd, win.GWL_EXSTYLE, set, clear)
}

func setAndClearWindowLongBits(hwnd win.HWND, index int32, set, clear uint32) error {
	value := uint32(win.GetWindowLong(hwnd, index))
	if value == 0 {
		return lastError("GetWindowLong")
	}

	if newValue := value&^clear | set; newValue != value {
		win.SetLastError(0)
		if win.SetWindowLong(hwnd, index, int32(newValue)) == 0 {
			return lastError("SetWindowLong")
		}
	}

	return nil
}

func (wb *WindowBase) ensureStyleBits(bits uint32, set bool) error {
	return ensureWindowLongBits(wb.hWnd, win.GWL_STYLE, bits, set)
}

func (wb *WindowBase) ensureExtendedStyleBits(bits uint32, set bool) error {
	return ensureWindowLongBits(wb.hWnd, win.GWL_EXSTYLE, bits, set)
}

func ensureWindowLongBits(hwnd win.HWND, index int32, bits uint32, set bool) error {
	var setBits uint32
	var clearBits uint32
	if set {
		setBits = bits
	} else {
		clearBits = bits
	}
	return setAndClearWindowLongBits(hwnd, index, setBits, clearBits)
}

// Handle returns the window handle of the Window.
func (wb *WindowBase) Handle() win.HWND {
	return wb.hWnd
}

// SendMessage sends a message to the window and returns the result.
func (wb *WindowBase) SendMessage(msg uint32, wParam, lParam uintptr) uintptr {
	return win.SendMessage(wb.hWnd, msg, wParam, lParam)
}

// Name returns the name of the *WindowBase.
func (wb *WindowBase) Name() string {
	return wb.name
}

// SetName sets the name of the *WindowBase.
func (wb *WindowBase) SetName(name string) {
	wb.name = name
}

func (wb *WindowBase) writePath(buf *bytes.Buffer) {
	hWndParent := win.GetAncestor(wb.hWnd, win.GA_PARENT)
	if pwi := windowFromHandle(hWndParent); pwi != nil {
		if sv, ok := pwi.(*ScrollView); ok {
			pwi = sv.Parent()
		}
		pwi.AsWindowBase().writePath(buf)
		buf.WriteByte('/')
	}

	buf.WriteString(wb.name)
}

func (wb *WindowBase) path() string {
	buf := bytes.NewBuffer(nil)

	wb.writePath(buf)

	return buf.String()
}

// WindowBase simply returns the receiver.
func (wb *WindowBase) AsWindowBase() *WindowBase {
	return wb
}

// AddDisposable adds a Disposable resource that should be disposed of
// together with this Window.
func (wb *WindowBase) AddDisposable(d Disposable) {
	wb.disposables = append(wb.disposables, d)
}

// Dispose releases the operating system resources, associated with the
// *WindowBase.
//
// If a user closes a *MainWindow or *Dialog, it is automatically released.
// Also, if a Container is disposed of, all its descendants will be released
// as well.
func (wb *WindowBase) Dispose() {
	for _, d := range wb.disposables {
		d.Dispose()
	}

	if wb.background != nil {
		wb.background.detachWindow(wb)
	}

	hWnd := wb.hWnd
	if hWnd != 0 {
		wb.disposingPublisher.Publish()

		switch w := wb.window.(type) {
		case *ToolTip:
		case Widget:
			globalToolTip.RemoveTool(w)
		}

		wb.hWnd = 0
		if _, ok := hwnd2WindowBase[hWnd]; ok {
			win.DestroyWindow(hWnd)
		}
	}

	if cm := wb.contextMenu; cm != nil {
		cm.actions.Clear()
		cm.Dispose()
	}

	for _, p := range wb.name2Property {
		p.SetSource(nil)
	}
}

// Disposing returns an Event that is published when the Window is disposed
// of.
func (wb *WindowBase) Disposing() *Event {
	return wb.disposingPublisher.Event()
}

// IsDisposed returns if the *WindowBase has been disposed of.
func (wb *WindowBase) IsDisposed() bool {
	return wb.hWnd == 0
}

// ContextMenu returns the context menu of the *WindowBase.
//
// By default this is nil.
func (wb *WindowBase) ContextMenu() *Menu {
	return wb.contextMenu
}

// SetContextMenu sets the context menu of the *WindowBase.
func (wb *WindowBase) SetContextMenu(value *Menu) {
	wb.contextMenu = value
}

// ShortcutActions returns the list of actions that will be triggered if their
// shortcut is pressed when this window or one of its descendants has the
// keyboard focus.
func (wb *WindowBase) ShortcutActions() *ActionList {
	if wb.shortcutActions == nil {
		wb.shortcutActions = newActionList(wb)
	}

	return wb.shortcutActions
}

// Background returns the background Brush of the *WindowBase.
//
// By default this is nil.
func (wb *WindowBase) Background() Brush {
	return wb.background
}

// SetBackground sets the background Brush of the *WindowBase.
func (wb *WindowBase) SetBackground(background Brush) {
	if wb.background != nil {
		wb.background.detachWindow(wb)
	}

	wb.background = background

	if background != nil {
		background.attachWindow(wb)
	}

	wb.Invalidate()

	// Sliders need some extra encouragement...
	walkDescendants(wb, func(w Window) bool {
		if s, ok := w.(*Slider); ok {
			s.SetRange(s.MinValue(), s.MaxValue()+1)
			s.SetRange(s.MinValue(), s.MaxValue()-1)
		}

		return true
	})
}

// Cursor returns the Cursor of the *WindowBase.
//
// By default this is nil.
func (wb *WindowBase) Cursor() Cursor {
	return wb.cursor
}

// SetCursor sets the Cursor of the *WindowBase.
func (wb *WindowBase) SetCursor(value Cursor) {
	wb.cursor = value
}

// DoubleBuffering returns whether double buffering of the
// drawing is enabled, which may help reduce flicker.
func (wb *WindowBase) DoubleBuffering() bool {
	return wb.hasExtendedStyleBits(win.WS_EX_COMPOSITED)
}

// SetDoubleBuffering enables or disables double buffering of the
// drawing, which may help reduce flicker.
func (wb *WindowBase) SetDoubleBuffering(enabled bool) error {
	return wb.ensureExtendedStyleBits(win.WS_EX_COMPOSITED, enabled)
}

type ApplySysColorser interface {
	ApplySysColors()
}

func (wb *WindowBase) ApplySysColors() {
	wb.Invalidate()
}

// DPI returns the current DPI value of the WindowBase.
func (wb *WindowBase) DPI() int {
	return int(win.GetDpiForWindow(wb.hWnd))
}

type ApplyDPIer interface {
	ApplyDPI(dpi int)
}

func (wb *WindowBase) ApplyDPI(dpi int) {
	if af, ok := wb.window.(applyFonter); ok {
		af.applyFont(wb.window.Font())
	}
}

func (wb *WindowBase) IntFrom96DPI(value int) int {
	return IntFrom96DPI(value, wb.DPI())
}

func (wb *WindowBase) IntTo96DPI(value int) int {
	return IntTo96DPI(value, wb.DPI())
}

func (wb *WindowBase) MarginsFrom96DPI(value Margins) Margins {
	return MarginsFrom96DPI(value, wb.DPI())
}

func (wb *WindowBase) MarginsTo96DPI(value Margins) Margins {
	return MarginsTo96DPI(value, wb.DPI())
}

func (wb *WindowBase) PointFrom96DPI(value Point) Point {
	return PointFrom96DPI(value, wb.DPI())
}

func (wb *WindowBase) PointTo96DPI(value Point) Point {
	return PointTo96DPI(value, wb.DPI())
}

func (wb *WindowBase) RectangleFrom96DPI(value Rectangle) Rectangle {
	return RectangleFrom96DPI(value, wb.DPI())
}

func (wb *WindowBase) RectangleTo96DPI(value Rectangle) Rectangle {
	return RectangleTo96DPI(value, wb.DPI())
}

func (wb *WindowBase) SizeFrom96DPI(value Size) Size {
	return SizeFrom96DPI(value, wb.DPI())
}

func (wb *WindowBase) SizeTo96DPI(value Size) Size {
	return SizeTo96DPI(value, wb.DPI())
}

// Enabled returns if the *WindowBase is enabled for user interaction.
func (wb *WindowBase) Enabled() bool {
	return wb.enabled
}

// SetEnabled sets if the *WindowBase is enabled for user interaction.
func (wb *WindowBase) SetEnabled(enabled bool) {
	wb.enabled = enabled

	wb.window.(applyEnableder).applyEnabled(wb.window.Enabled())

	if widget, ok := wb.window.(Widget); ok {
		widget.AsWidgetBase().invalidateBorderInParent()
	}

	wb.enabledChangedPublisher.Publish()
}

type applyEnableder interface {
	applyEnabled(enabled bool)
}

func (wb *WindowBase) applyEnabled(enabled bool) {
	setWindowEnabled(wb.hWnd, enabled)
}

func setWindowEnabled(hwnd win.HWND, enabled bool) {
	win.EnableWindow(hwnd, enabled)

	win.UpdateWindow(hwnd)
}

// Font returns the *Font of the *WindowBase.
//
// By default this is a MS Shell Dlg 2, 8 point font.
func (wb *WindowBase) Font() *Font {
	if wb.font != nil {
		return wb.font
	}

	return defaultFont
}

// SetFont sets the *Font of the *WindowBase.
func (wb *WindowBase) SetFont(font *Font) {
	if font != wb.font {
		wb.font = font

		wb.window.(applyFonter).applyFont(font)
	}
}

type applyFonter interface {
	applyFont(font *Font)
}

type ApplyFonter interface {
	ApplyFont(font *Font)
}

func (wb *WindowBase) applyFont(font *Font) {
	if hFont := font.handleForDPI(wb.DPI()); hFont != wb.hFont {
		wb.hFont = hFont

		setWindowFont(wb.hWnd, hFont)
	}

	if af, ok := wb.window.(ApplyFonter); ok {
		af.ApplyFont(font)
	}
}

func SetWindowFont(hwnd win.HWND, font *Font) {
	dpi := int(win.GetDpiForWindow(hwnd))
	setWindowFont(hwnd, font.handleForDPI(dpi))
}

func setWindowFont(hwnd win.HWND, hFont win.HFONT) {
	win.SendMessage(hwnd, win.WM_SETFONT, uintptr(hFont), 1)

	if window := windowFromHandle(hwnd); window != nil {
		if widget, ok := window.(Widget); ok {
			widget.AsWidgetBase().updateParentLayoutWithReset(false)
		}
	}
}

// Suspended returns if the *WindowBase is suspended for layout and repainting
// purposes.
func (wb *WindowBase) Suspended() bool {
	return wb.suspended
}

// SetSuspended sets if the *WindowBase is suspended for layout and repainting
// purposes.
//
// You should call SetSuspended(true), before doing a batch of modifications
// that would cause multiple layout or drawing updates. Remember to call
// SetSuspended(false) afterwards, which will update the *WindowBase
// accordingly.
func (wb *WindowBase) SetSuspended(suspend bool) {
	if suspend == wb.suspended {
		return
	}

	var wParam int
	if suspend {
		wParam = 0
	} else {
		wParam = 1
	}

	wb.SendMessage(win.WM_SETREDRAW, uintptr(wParam), 0)

	wb.suspended = suspend

	if !suspend {
		wb.Invalidate()
	}
}

// Invalidate schedules a full repaint of the *WindowBase.
func (wb *WindowBase) Invalidate() error {
	if !win.InvalidateRect(wb.hWnd, nil, true) {
		return newError("InvalidateRect failed")
	}

	return nil
}

func (wb *WindowBase) text() string {
	return windowText(wb.hWnd)
}

func (wb *WindowBase) setText(text string) error {
	if err := setWindowText(wb.hWnd, text); err != nil {
		return err
	}

	if wb.calcTextSizeInfoPrev != nil {
		wb.calcTextSizeInfoPrev.font.family = ""
		wb.calcTextSizeInfoPrev.text = text
	}

	return nil
}

func windowText(hwnd win.HWND) string {
	textLength := win.SendMessage(hwnd, win.WM_GETTEXTLENGTH, 0, 0)
	buf := make([]uint16, textLength+1)
	win.SendMessage(hwnd, win.WM_GETTEXT, uintptr(textLength+1), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}

func setWindowText(hwnd win.HWND, text string) error {
	if win.TRUE != win.SendMessage(hwnd, win.WM_SETTEXT, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text)))) {
		return newError("WM_SETTEXT failed")
	}

	return nil
}

func (wb *WindowBase) RestoreState() (err error) {
	wb.ForEachDescendant(func(widget Widget) bool {
		if persistable, ok := widget.(Persistable); ok && persistable.Persistent() {
			if err = persistable.RestoreState(); err != nil {
				return false
			}
		}

		if _, ok := widget.(Container); ok {
			return false
		}

		return true
	})

	return
}

func (wb *WindowBase) SaveState() (err error) {
	wb.ForEachDescendant(func(widget Widget) bool {
		if persistable, ok := widget.(Persistable); ok && persistable.Persistent() {
			if err = persistable.SaveState(); err != nil {
				return false
			}
		}

		if _, ok := widget.(Container); ok {
			return false
		}

		return true
	})

	return
}

func forEachDescendant(hwnd win.HWND, lParam uintptr) uintptr {
	if window := windowFromHandle(hwnd); window != nil && forEachDescendantCallback(window.(Widget)) {
		return 1
	}

	return 0
}

var (
	forEachDescendantCallbackPtr = syscall.NewCallback(forEachDescendant)
	forEachDescendantCallback    func(widget Widget) bool
)

func (wb *WindowBase) ForEachDescendant(f func(widget Widget) bool) {
	forEachDescendantCallback = f
	defer func() {
		forEachDescendantCallback = nil
	}()

	win.EnumChildWindows(wb.hWnd, forEachDescendantCallbackPtr, 0)
}

// Visible returns if the *WindowBase is visible.
func (wb *WindowBase) Visible() bool {
	return win.IsWindowVisible(wb.hWnd)
}

// SetVisible sets if the *WindowBase is visible.
func (wb *WindowBase) SetVisible(visible bool) {
	old := wb.Visible()

	setWindowVisible(wb.hWnd, visible)

	wb.visible = visible

	walkDescendants(wb.window, func(w Window) bool {
		w.AsWindowBase().visibleChangedPublisher.Publish()

		return true
	})

	if visible == old {
		return
	}

	if widget, ok := wb.window.(Widget); ok {
		wb := widget.AsWidgetBase()
		wb.invalidateBorderInParent()
		wb.updateParentLayoutWithReset(true)
	}

	wb.visibleChangedPublisher.Publish()
}

// VisibleChanged returns an Event that you can attach to for handling
// visible changed events for the Window.
func (wb *WindowBase) VisibleChanged() *Event {
	return wb.visibleChangedPublisher.Event()
}

func setWindowVisible(hwnd win.HWND, visible bool) {
	var cmd int32
	if visible {
		cmd = win.SW_SHOWNA
	} else {
		cmd = win.SW_HIDE
	}
	win.ShowWindow(hwnd, cmd)
}

// BringToTop moves the *WindowBase to the top of the keyboard focus order.
func (wb *WindowBase) BringToTop() error {
	if !win.SetWindowPos(wb.hWnd, win.HWND_TOP, 0, 0, 0, 0, win.SWP_NOACTIVATE|win.SWP_NOMOVE|win.SWP_NOSIZE) {
		return lastError("SetWindowPos")
	}

	return nil
}

// Bounds returns the outer bounding box Rectangle of the *WindowBase, including
// decorations.
//
// The coordinates are relative to the screen.
func (wb *WindowBase) Bounds() Rectangle {
	return wb.RectangleTo96DPI(wb.BoundsPixels())
}

// SetBounds sets the outer bounding box Rectangle of the *WindowBase,
// including decorations.
//
// For a Form, like *MainWindow or *Dialog, the Rectangle is in screen
// coordinates, for a child Window the coordinates are relative to its parent.
func (wb *WindowBase) SetBounds(bounds Rectangle) error {
	return wb.SetBoundsPixels(wb.RectangleFrom96DPI(bounds))
}

// BoundsPixels returns the outer bounding box Rectangle of the *WindowBase, including
// decorations.
//
// The coordinates are relative to the screen.
func (wb *WindowBase) BoundsPixels() Rectangle {
	var r win.RECT

	if !win.GetWindowRect(wb.hWnd, &r) {
		lastError("GetWindowRect")
		return Rectangle{}
	}

	return Rectangle{
		int(r.Left),
		int(r.Top),
		int(r.Right - r.Left),
		int(r.Bottom - r.Top),
	}
}

// SetBoundsPixels sets the outer bounding box Rectangle of the *WindowBase,
// including decorations.
//
// For a Form, like *MainWindow or *Dialog, the Rectangle is in screen
// coordinates, for a child Window the coordinates are relative to its parent.
func (wb *WindowBase) SetBoundsPixels(bounds Rectangle) error {
	if !win.MoveWindow(
		wb.hWnd,
		int32(bounds.X),
		int32(bounds.Y),
		int32(bounds.Width),
		int32(bounds.Height),
		true) {

		return lastError("MoveWindow")
	}

	return nil
}

// MinSize returns the minimum allowed outer Size for the *WindowBase, including
// decorations.
//
// For child windows, this is only relevant when the parent of the *WindowBase
// has a Layout. RootWidgets, like *MainWindow and *Dialog, also honor this.
func (wb *WindowBase) MinSize() Size {
	return wb.minSize
}

// MinSizePixels returns the minimum allowed outer Size for the *WindowBase, including
// decorations.
//
// For child windows, this is only relevant when the parent of the *WindowBase
// has a Layout. RootWidgets, like *MainWindow and *Dialog, also honor this.
func (wb *WindowBase) MinSizePixels() Size {
	return wb.SizeFrom96DPI(wb.minSize)
}

// MaxSize returns the maximum allowed outer Size for the *WindowBase, including
// decorations.
//
// For child windows, this is only relevant when the parent of the *WindowBase
// has a Layout. RootWidgets, like *MainWindow and *Dialog, also honor this.
func (wb *WindowBase) MaxSize() Size {
	return wb.maxSize
}

// MaxSizePixels returns the maximum allowed outer Size for the *WindowBase, including
// decorations.
//
// For child windows, this is only relevant when the parent of the *WindowBase
// has a Layout. RootWidgets, like *MainWindow and *Dialog, also honor this.
func (wb *WindowBase) MaxSizePixels() Size {
	return wb.SizeFrom96DPI(wb.maxSize)
}

// SetMinMaxSize sets the minimum and maximum outer Size of the *WindowBase,
// including decorations.
//
// Use walk.Size{} to make the respective limit be ignored.
func (wb *WindowBase) SetMinMaxSize(min, max Size) error {
	if min.Width < 0 || min.Height < 0 {
		return newError("min must be positive")
	}
	if max.Width > 0 && max.Width < min.Width ||
		max.Height > 0 && max.Height < min.Height {
		return newError("max must be greater as or equal to min")
	}
	wb.minSize = min
	wb.maxSize = max
	return nil
}

// SetMinMaxSizePixels sets the minimum and maximum outer Size of the *WindowBase,
// including decorations.
//
// Use walk.Size{} to make the respective limit be ignored.
func (wb *WindowBase) SetMinMaxSizePixels(min, max Size) error {
	return wb.SetMinMaxSize(wb.SizeTo96DPI(min), wb.SizeTo96DPI(max))
}

type fontInfoAndDPI struct {
	fontInfo
	dpi int
}

var (
	dialogBaseUnitsUTF16StringPtr  = syscall.StringToUTF16Ptr("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	fontInfoAndDPI2DialogBaseUnits = make(map[fontInfoAndDPI]Size)
)

func (wb *WindowBase) dialogBaseUnits() Size {
	// The window may use a font different from that in WindowBase,
	// like e.g. NumberEdit does, so we try to use the right one.
	font := wb.window.Font()
	fi := fontInfoAndDPI{
		fontInfo: fontInfo{
			family:    font.Family(),
			pointSize: font.PointSize(),
			style:     font.Style(),
		},
		dpi: wb.DPI()}
	if s, ok := fontInfoAndDPI2DialogBaseUnits[fi]; ok {
		return s
	}

	hdc := win.GetDC(wb.hWnd)
	defer win.ReleaseDC(wb.hWnd, hdc)

	hFont := font.handleForDPI(wb.DPI())
	hFontOld := win.SelectObject(hdc, win.HGDIOBJ(hFont))
	defer win.SelectObject(hdc, win.HGDIOBJ(hFontOld))

	var tm win.TEXTMETRIC
	if !win.GetTextMetrics(hdc, &tm) {
		newError("GetTextMetrics failed")
	}

	var size win.SIZE
	if !win.GetTextExtentPoint32(
		hdc,
		dialogBaseUnitsUTF16StringPtr,
		52,
		&size) {
		newError("GetTextExtentPoint32 failed")
	}

	s := Size{int((size.CX/26 + 1) / 2), int(tm.TmHeight)}

	fontInfoAndDPI2DialogBaseUnits[fi] = s

	return s
}

func (wb *WindowBase) dialogBaseUnitsToPixels(dlus Size) (pixels Size) {
	base := wb.dialogBaseUnits()

	return Size{
		int(win.MulDiv(int32(dlus.Width), int32(base.Width), 4)),
		int(win.MulDiv(int32(dlus.Height), int32(base.Height), 8)),
	}
}

func (wb *WindowBase) calculateTextSizeImpl(text string) Size {
	return wb.calculateTextSizeImplForWidth(text, 0)
}

func (wb *WindowBase) calculateTextSizeImplForWidth(text string, width int) Size {
	font := wb.window.Font()

	dpi := wb.DPI()

	if wb.calcTextSizeInfoPrev != nil &&
		width == wb.calcTextSizeInfoPrev.width &&
		font.family == wb.calcTextSizeInfoPrev.font.family &&
		font.pointSize == wb.calcTextSizeInfoPrev.font.pointSize &&
		font.style == wb.calcTextSizeInfoPrev.font.style &&
		text == wb.calcTextSizeInfoPrev.text &&
		dpi == wb.calcTextSizeInfoPrev.dpi {
		return wb.calcTextSizeInfoPrev.size
	}

	if wb.calcTextSizeInfoPrev != nil && dpi != wb.calcTextSizeInfoPrev.dpi {
		width = int(float64(width) * float64(dpi) / float64(wb.calcTextSizeInfoPrev.dpi))
	}

	var size Size
	if width > 0 {
		canvas, err := wb.CreateCanvas()
		if err != nil {
			return size
		}
		defer canvas.Dispose()

		bounds, _, err := canvas.measureTextForDPI(text, font, Rectangle{Width: width, Height: 9999999}, 0, dpi)
		if err != nil {
			return size
		}

		size = bounds.Size()
	} else {
		hdc := win.GetDC(wb.hWnd)
		if hdc == 0 {
			newError("GetDC failed")
			return Size{}
		}
		defer win.ReleaseDC(wb.hWnd, hdc)

		hFontOld := win.SelectObject(hdc, win.HGDIOBJ(font.handleForDPI(dpi)))
		defer win.SelectObject(hdc, hFontOld)

		lines := strings.Split(text, "\n")

		for _, line := range lines {
			var s win.SIZE
			str := syscall.StringToUTF16(strings.TrimRight(line, "\r "))

			if !win.GetTextExtentPoint32(hdc, &str[0], int32(len(str)-1), &s) {
				newError("GetTextExtentPoint32 failed")
				return Size{}
			}

			size.Width = maxi(size.Width, int(s.CX))
			size.Height += int(s.CY)
		}
	}

	if wb.calcTextSizeInfoPrev == nil {
		wb.calcTextSizeInfoPrev = new(calcTextSizeInfo)
	}

	wb.calcTextSizeInfoPrev.width = width
	wb.calcTextSizeInfoPrev.font.family = font.family
	wb.calcTextSizeInfoPrev.font.pointSize = font.pointSize
	wb.calcTextSizeInfoPrev.font.style = font.style
	wb.calcTextSizeInfoPrev.text = text
	wb.calcTextSizeInfoPrev.size = size
	wb.calcTextSizeInfoPrev.dpi = dpi

	return size
}

func (wb *WindowBase) calculateTextSize() Size {
	return wb.calculateTextSizeForWidth(0)
}

func (wb *WindowBase) calculateTextSizeForWidth(width int) Size {
	var text string
	if wb.calcTextSizeInfoPrev != nil {
		// setText copied the new text here for us.
		text = wb.calcTextSizeInfoPrev.text
	}

	if text == "" {
		text = wb.text()
	}

	return wb.calculateTextSizeImplForWidth(text, width)
}

// Size returns the outer Size of the *WindowBase, including decorations.
func (wb *WindowBase) Size() Size {
	return wb.SizeTo96DPI(wb.SizePixels())
}

// SizePixels returns the outer Size of the *WindowBase, including decorations.
func (wb *WindowBase) SizePixels() Size {
	return wb.window.BoundsPixels().Size()
}

// SetSize sets the outer Size of the *WindowBase, including decorations.
func (wb *WindowBase) SetSize(size Size) error {
	return wb.SetSizePixels(wb.SizeFrom96DPI(size))
}

// SetSizePixels sets the outer Size of the *WindowBase, including decorations.
func (wb *WindowBase) SetSizePixels(size Size) error {
	bounds := wb.window.BoundsPixels()

	return wb.SetBoundsPixels(bounds.SetSize(size))
}

// X returns the x coordinate of the *WindowBase, relative to the screen for
// RootWidgets like *MainWindow or *Dialog and relative to the parent for
// child Windows.
func (wb *WindowBase) X() int {
	return wb.IntTo96DPI(wb.XPixels())
}

// XPixels returns the x coordinate of the *WindowBase, relative to the screen for
// RootWidgets like *MainWindow or *Dialog and relative to the parent for
// child Windows.
func (wb *WindowBase) XPixels() int {
	return wb.window.BoundsPixels().X
}

// SetX sets the x coordinate of the *WindowBase, relative to the screen for
// RootWidgets like *MainWindow or *Dialog and relative to the parent for
// child Windows.
func (wb *WindowBase) SetX(value int) error {
	return wb.SetXPixels(wb.IntFrom96DPI(value))
}

// SetXPixels sets the x coordinate of the *WindowBase, relative to the screen for
// RootWidgets like *MainWindow or *Dialog and relative to the parent for
// child Windows.
func (wb *WindowBase) SetXPixels(value int) error {
	bounds := wb.window.BoundsPixels()
	bounds.X = value

	return wb.SetBoundsPixels(bounds)
}

// Y returns the y coordinate of the *WindowBase, relative to the screen for
// RootWidgets like *MainWindow or *Dialog and relative to the parent for
// child Windows.
func (wb *WindowBase) Y() int {
	return wb.IntTo96DPI(wb.YPixels())
}

// YPixels returns the y coordinate of the *WindowBase, relative to the screen for
// RootWidgets like *MainWindow or *Dialog and relative to the parent for
// child Windows.
func (wb *WindowBase) YPixels() int {
	return wb.window.BoundsPixels().Y
}

// SetY sets the y coordinate of the *WindowBase, relative to the screen for
// RootWidgets like *MainWindow or *Dialog and relative to the parent for
// child Windows.
func (wb *WindowBase) SetY(value int) error {
	return wb.SetYPixels(wb.IntFrom96DPI(value))
}

// SetYPixels sets the y coordinate of the *WindowBase, relative to the screen for
// RootWidgets like *MainWindow or *Dialog and relative to the parent for
// child Windows.
func (wb *WindowBase) SetYPixels(value int) error {
	bounds := wb.window.BoundsPixels()
	bounds.Y = value

	return wb.SetBoundsPixels(bounds)
}

// Width returns the outer width of the *WindowBase, including decorations.
func (wb *WindowBase) Width() int {
	return wb.IntTo96DPI(wb.WidthPixels())
}

// WidthPixels returns the outer width of the *WindowBase, including decorations.
func (wb *WindowBase) WidthPixels() int {
	return wb.window.BoundsPixels().Width
}

// SetWidth sets the outer width of the *WindowBase, including decorations.
func (wb *WindowBase) SetWidth(value int) error {
	return wb.SetWidthPixels(wb.IntFrom96DPI(value))
}

// SetWidthPixels sets the outer width of the *WindowBase, including decorations.
func (wb *WindowBase) SetWidthPixels(value int) error {
	bounds := wb.window.BoundsPixels()
	bounds.Width = value

	return wb.SetBoundsPixels(bounds)
}

// Height returns the outer height of the *WindowBase, including decorations.
func (wb *WindowBase) Height() int {
	return wb.IntTo96DPI(wb.HeightPixels())
}

// HeightPixels returns the outer height of the *WindowBase, including decorations.
func (wb *WindowBase) HeightPixels() int {
	return wb.window.BoundsPixels().Height
}

// SetHeight sets the outer height of the *WindowBase, including decorations.
func (wb *WindowBase) SetHeight(value int) error {
	return wb.SetHeightPixels(wb.IntFrom96DPI(value))
}

// SetHeightPixels sets the outer height of the *WindowBase, including decorations.
func (wb *WindowBase) SetHeightPixels(value int) error {
	bounds := wb.window.BoundsPixels()
	bounds.Height = value

	return wb.SetBoundsPixels(bounds)
}

func windowClientBounds(hwnd win.HWND) Rectangle {
	var r win.RECT

	if !win.GetClientRect(hwnd, &r) {
		lastError("GetClientRect")
		return Rectangle{}
	}

	return Rectangle{
		int(r.Left),
		int(r.Top),
		int(r.Right - r.Left),
		int(r.Bottom - r.Top),
	}
}

// ClientBounds returns the inner bounding box Rectangle of the *WindowBase,
// excluding decorations.
func (wb *WindowBase) ClientBounds() Rectangle {
	return wb.RectangleTo96DPI(wb.ClientBoundsPixels())
}

// ClientBoundsPixels returns the inner bounding box Rectangle of the *WindowBase,
// excluding decorations.
func (wb *WindowBase) ClientBoundsPixels() Rectangle {
	return windowClientBounds(wb.hWnd)
}

func (wb *WindowBase) sizeFromClientSizePixels(clientSize Size) Size {
	window := wb.window
	s := window.SizePixels()
	cs := window.ClientBoundsPixels().Size()
	ncs := Size{s.Width - cs.Width, s.Height - cs.Height}

	return Size{clientSize.Width + ncs.Width, clientSize.Height + ncs.Height}
}

func (wb *WindowBase) clientSizeFromSizePixels(size Size) Size {
	window := wb.window
	s := window.SizePixels()
	cs := window.ClientBoundsPixels().Size()
	ncs := Size{s.Width - cs.Width, s.Height - cs.Height}

	return Size{size.Width - ncs.Width, size.Height - ncs.Height}
}

// SetClientSize sets the Size of the inner bounding box of the *WindowBase,
// excluding decorations.
func (wb *WindowBase) SetClientSize(value Size) error {
	return wb.SetClientSizePixels(wb.SizeFrom96DPI(value))
}

// SetClientSizePixels sets the Size of the inner bounding box of the *WindowBase,
// excluding decorations.
func (wb *WindowBase) SetClientSizePixels(value Size) error {
	return wb.SetSizePixels(wb.sizeFromClientSizePixels(value))
}

// RightToLeftReading returns whether the reading order of the Window
// is from right to left.
func (wb *WindowBase) RightToLeftReading() bool {
	return wb.hasExtendedStyleBits(win.WS_EX_RTLREADING)
}

// SetRightToLeftReading sets whether the reading order of the Window
// is from right to left.
func (wb *WindowBase) SetRightToLeftReading(rtl bool) error {
	return wb.ensureExtendedStyleBits(win.WS_EX_RTLREADING, rtl)
}

// Screenshot returns an image of the window.
func (wb *WindowBase) Screenshot() (*image.RGBA, error) {
	bmp, err := NewBitmapFromWindow(wb)
	if err != nil {
		return nil, err
	}
	defer bmp.Dispose()

	return bmp.ToImage()
}

// FocusedWindow returns the Window that has the keyboard input focus.
func FocusedWindow() Window {
	return windowFromHandle(win.GetFocus())
}

// Focused returns whether the Window has the keyboard input focus.
func (wb *WindowBase) Focused() bool {
	return wb.hWnd == win.GetFocus()
}

// SetFocus sets the keyboard input focus to the *WindowBase.
func (wb *WindowBase) SetFocus() error {
	if win.SetFocus(wb.hWnd) == 0 {
		return lastError("SetFocus")
	}

	return nil
}

// FocusedChanged returns an Event that you can attach to for handling focus
// change events for the WindowBase.
func (wb *WindowBase) FocusedChanged() *Event {
	return wb.focusedChangedPublisher.Event()
}

// CreateCanvas creates and returns a *Canvas that can be used to draw
// inside the ClientBoundsPixels of the *WindowBase.
//
// Remember to call the Dispose method on the canvas to release resources,
// when you no longer need it.
func (wb *WindowBase) CreateCanvas() (*Canvas, error) {
	return newCanvasFromWindow(wb.window)
}

func (wb *WindowBase) setTheme(appName string) error {
	if hr := win.SetWindowTheme(wb.hWnd, syscall.StringToUTF16Ptr(appName), nil); win.FAILED(hr) {
		return errorFromHRESULT("SetWindowTheme", hr)
	}

	return nil
}

// KeyDown returns a *KeyEvent that you can attach to for handling key down
// events for the *WindowBase.
func (wb *WindowBase) KeyDown() *KeyEvent {
	return wb.keyDownPublisher.Event()
}

// KeyPress returns a *KeyEvent that you can attach to for handling key press
// events for the *WindowBase.
func (wb *WindowBase) KeyPress() *KeyEvent {
	return wb.keyPressPublisher.Event()
}

// KeyUp returns a *KeyEvent that you can attach to for handling key up
// events for the *WindowBase.
func (wb *WindowBase) KeyUp() *KeyEvent {
	return wb.keyUpPublisher.Event()
}

// DropFiles returns a *DropFilesEvent that you can attach to for handling
// drop file events for the *WindowBase.
func (wb *WindowBase) DropFiles() *DropFilesEvent {
	return wb.dropFilesPublisher.Event(wb.hWnd)
}

// MouseDown returns a *MouseEvent that you can attach to for handling
// mouse down events for the *WindowBase.
func (wb *WindowBase) MouseDown() *MouseEvent {
	return wb.mouseDownPublisher.Event()
}

// MouseMove returns a *MouseEvent that you can attach to for handling
// mouse move events for the *WindowBase.
func (wb *WindowBase) MouseMove() *MouseEvent {
	return wb.mouseMovePublisher.Event()
}

// MouseUp returns a *MouseEvent that you can attach to for handling
// mouse up events for the *WindowBase.
func (wb *WindowBase) MouseUp() *MouseEvent {
	return wb.mouseUpPublisher.Event()
}

func (wb *WindowBase) MouseWheel() *MouseEvent {
	return wb.mouseWheelPublisher.Event()
}

func (wb *WindowBase) publishMouseEvent(publisher *MouseEventPublisher, msg uint32, wParam, lParam uintptr) {
	x := int(win.GET_X_LPARAM(lParam))
	y := int(win.GET_Y_LPARAM(lParam))

	var button MouseButton
	switch msg {
	case win.WM_LBUTTONUP:
		button = LeftButton

	case win.WM_RBUTTONUP:
		button = RightButton

	case win.WM_MBUTTONUP:
		button = MiddleButton

	default:
		button = MouseButton(wParam&win.MK_LBUTTON | wParam&win.MK_RBUTTON | wParam&win.MK_MBUTTON)
	}

	publisher.Publish(x, y, button)
}

func (wb *WindowBase) publishMouseWheelEvent(publisher *MouseEventPublisher, wParam, lParam uintptr) {
	x := int(win.GET_X_LPARAM(lParam))
	y := int(win.GET_Y_LPARAM(lParam))
	button := MouseButton(uint32(wParam))

	publisher.Publish(x, y, button)
}

// SizeChanged returns an *Event that you can attach to for handling size
// changed events for the *WindowBase.
func (wb *WindowBase) SizeChanged() *Event {
	return wb.sizeChangedPublisher.Event()
}

// BoundsChanged returns an *Event that you can attach to for handling bounds
// changed events for the *WindowBase.
func (wb *WindowBase) BoundsChanged() *Event {
	return wb.boundsChangedPublisher.Event()
}

// Synchronize enqueues func f to be called some time later by the main
// goroutine from inside a message loop.
func (wb *WindowBase) Synchronize(f func()) {
	synchronize(f)

	win.PostMessage(wb.hWnd, syncMsgId, 0, 0)
}

func (wb *WindowBase) ReadState() (string, error) {
	settings := appSingleton.settings
	if settings == nil {
		return "", newError("App().Settings() must not be nil")
	}

	state, _ := settings.Get(wb.path())
	return state, nil
}

func (wb *WindowBase) WriteState(state string) error {
	settings := appSingleton.settings
	if settings == nil {
		return newError("App().Settings() must not be nil")
	}

	p := wb.path()
	if strings.HasPrefix(p, "/") ||
		strings.HasSuffix(p, "/") ||
		strings.Contains(p, "//") {

		return nil
	}

	return settings.PutExpiring(p, state)
}

func windowFromHandle(hwnd win.HWND) Window {
	if wb := hwnd2WindowBase[hwnd]; wb != nil {
		return wb.window
	}

	return nil
}

func defaultWndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) (result uintptr) {
	defer func() {
		if len(appSingleton.panickingPublisher.event.handlers) > 0 {
			var err error
			if x := recover(); x != nil {
				if e, ok := x.(error); ok {
					err = wrapErrorNoPanic(e)
				} else {
					err = newErrorNoPanic(fmt.Sprint(x))
				}
			}
			if err != nil {
				appSingleton.panickingPublisher.Publish(err)
			}
		}
	}()

	if msg == notifyIconMessageId {
		return notifyIconWndProc(hwnd, msg, wParam, lParam)
	}

	wi := windowFromHandle(hwnd)
	if wi == nil {
		return win.DefWindowProc(hwnd, msg, wParam, lParam)
	}

	result = wi.WndProc(hwnd, msg, wParam, lParam)

	return
}

type menuer interface {
	Menu() *Menu
}

func menuContainsAction(menu *Menu, action *Action) bool {
	if menu.Actions().Contains(action) {
		return true
	}

	for _, a := range menu.actions.actions {
		if a.menu != nil && menuContainsAction(a.menu, action) {
			return true
		}
	}

	return false
}

func (wb *WindowBase) handleKeyDown(wParam, lParam uintptr) {
	key := Key(wParam)

	if uint32(lParam)>>30 == 0 {
		wb.keyDownPublisher.Publish(key)

		// Using TranslateAccelerators refused to work, so we handle them
		// ourselves, at least for now.
		shortcut := Shortcut{ModifiersDown(), key}
		if action, ok := shortcut2Action[shortcut]; ok {
			if action.Visible() && action.Enabled() {
				window := wb.window

				if w, ok := window.(Widget); ok {
					window = ancestor(w)
				}

				if m, ok := window.(menuer); ok && menuContainsAction(m.Menu(), action) {
					action.raiseTriggered()
				}
			}
		}
	}

	switch key {
	case KeyAlt, KeyControl, KeyShift:
		// nop

	default:
		wb.keyPressPublisher.Publish(key)
	}
}

func (wb *WindowBase) handleKeyUp(wParam, lParam uintptr) {
	wb.keyUpPublisher.Publish(Key(wParam))
}

func (wb *WindowBase) backgroundEffective() (Brush, Window) {
	wnd := wb.window
	bg := wnd.Background()

	if widget, ok := wb.window.(Widget); ok {
		for bg == nullBrushSingleton && widget != nil {
			if hwndParent := win.GetParent(widget.Handle()); hwndParent != 0 {
				if parent := windowFromHandle(hwndParent); parent != nil {
					wnd = parent
					bg = parent.Background()

					widget, _ = parent.(Widget)
				} else {
					break
				}
			} else {
				break
			}
		}
	}

	if bg != nil {
		if pwb, ok := bg.(perWindowBrush); ok {
			bg = pwb.delegateForWindow(wnd.AsWindowBase())
		}
	}

	return bg, wnd
}

func (wb *WindowBase) prepareDCForBackground(hdc win.HDC, hwnd win.HWND, brushWnd Window) {
	win.SetBkMode(hdc, win.TRANSPARENT)

	var bgRC win.RECT
	win.GetWindowRect(brushWnd.Handle(), &bgRC)

	var rc win.RECT
	win.GetWindowRect(hwnd, &rc)

	win.SetBrushOrgEx(hdc, bgRC.Left-rc.Left, bgRC.Top-rc.Top, nil)
}

func (wb *WindowBase) handleWMCTLCOLOR(wParam, lParam uintptr) uintptr {
	hwnd := win.HWND(lParam)
	hdc := win.HDC(wParam)

	type TextColorer interface {
		TextColor() Color
	}

	wnd := windowFromHandle(hwnd)
	if wnd == nil {
		switch windowFromHandle(win.GetParent(hwnd)).(type) {
		case *ComboBox:
			// nop
			return 0
		}

		wnd = wb
	} else if tc, ok := wnd.(TextColorer); ok {
		color := tc.TextColor()
		if color == 0 {
			color = Color(win.GetSysColor(win.COLOR_WINDOWTEXT))
		}
		win.SetTextColor(hdc, win.COLORREF(color))
	}

	if bg, wnd := wnd.AsWindowBase().backgroundEffective(); bg != nil {
		wb.prepareDCForBackground(hdc, hwnd, wnd)

		type Colorer interface {
			Color() Color
		}

		if c, ok := bg.(Colorer); ok {
			win.SetBkColor(hdc, win.COLORREF(c.Color()))
		}

		return uintptr(bg.handle())
	}

	switch wnd.(type) {
	case *LineEdit, *numberLineEdit, *TextEdit:
		type ReadOnlyer interface {
			ReadOnly() bool
		}

		var sysColor int
		if ro, ok := wnd.(ReadOnlyer); ok && ro.ReadOnly() {
			sysColor = win.COLOR_BTNFACE
		} else {
			sysColor = win.COLOR_WINDOW
		}

		win.SetBkColor(hdc, win.COLORREF(win.GetSysColor(sysColor)))

		return uintptr(win.GetSysColorBrush(sysColor))
	}

	return 0
}

// WndProc is the window procedure of the window.
//
// When implementing your own WndProc to add or modify behavior, call the
// WndProc of the embedded window for messages you don't handle yourself.
func (wb *WindowBase) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	window := windowFromHandle(hwnd)

	switch msg {
	case win.WM_ERASEBKGND:
		if _, ok := window.(Widget); !ok {
			return 0
		}

		bg, wnd := wb.backgroundEffective()
		if bg == nil {
			break
		}

		hdc := win.HDC(wParam)

		canvas, err := newCanvasFromHDC(hdc)
		if err != nil {
			break
		}
		defer canvas.Dispose()

		wb.prepareDCForBackground(hdc, hwnd, wnd)

		if err := canvas.FillRectangle(bg, wb.ClientBoundsPixels()); err != nil {
			break
		}

		return 1

	case win.WM_HSCROLL, win.WM_VSCROLL:
		if window := windowFromHandle(win.HWND(lParam)); window != nil {
			// The window that sent the notification shall handle it itself.
			return window.WndProc(hwnd, msg, wParam, lParam)
		}

	case win.WM_LBUTTONDOWN, win.WM_MBUTTONDOWN, win.WM_RBUTTONDOWN:
		if msg == win.WM_LBUTTONDOWN && wb.origWndProcPtr == 0 {
			// Only call SetCapture if this is no subclassed control.
			// (Otherwise e.g. WM_COMMAND(BN_CLICKED) would no longer
			// be generated for PushButton.)
			win.SetCapture(wb.hWnd)
		}
		wb.publishMouseEvent(&wb.mouseDownPublisher, msg, wParam, lParam)

	case win.WM_LBUTTONUP, win.WM_MBUTTONUP, win.WM_RBUTTONUP:
		if msg == win.WM_LBUTTONUP && wb.origWndProcPtr == 0 {
			// See WM_LBUTTONDOWN for why we require origWndProcPtr == 0 here.
			if !win.ReleaseCapture() {
				lastError("ReleaseCapture")
			}
		}
		wb.publishMouseEvent(&wb.mouseUpPublisher, msg, wParam, lParam)

	case win.WM_MOUSEMOVE:
		wb.publishMouseEvent(&wb.mouseMovePublisher, msg, wParam, lParam)

	case win.WM_MOUSEWHEEL:
		wb.publishMouseWheelEvent(&wb.mouseWheelPublisher, wParam, lParam)

	case win.WM_SETFOCUS, win.WM_KILLFOCUS:
		switch wnd := wb.window.(type) {
		case *splitterHandle:
			// nop

		case Widget:
			parent := wnd.Parent()
			if parent == nil {
				hwndParent := win.GetParent(wnd.Handle())
				for parent == nil && hwndParent != 0 {
					hwndParent = win.GetParent(hwndParent)
					if wnd := windowFromHandle(hwndParent); wnd != nil {
						parent, _ = wnd.(Container)
					}
				}
			}
			wnd.AsWidgetBase().invalidateBorderInParent()
		}

		wb.focusedChangedPublisher.Publish()

	case win.WM_SETCURSOR:
		if wb.cursor != nil {
			win.SetCursor(wb.cursor.handle())
			return 0
		}

	case win.WM_CONTEXTMENU:
		sourceWindow := windowFromHandle(win.HWND(wParam))
		if sourceWindow == nil {
			break
		}

		x := win.GET_X_LPARAM(lParam)
		y := win.GET_Y_LPARAM(lParam)

		contextMenu := sourceWindow.ContextMenu()

		var handle win.HWND
		if widget, ok := sourceWindow.(Widget); ok {
			if form := ancestor(widget); form != nil {
				handle = form.Handle()
			}
		}

		if handle == 0 {
			handle = sourceWindow.Handle()
		}

		if contextMenu != nil {
			contextMenu.updateItemsWithImageForWindow(wb.window)

			win.TrackPopupMenuEx(
				contextMenu.hMenu,
				win.TPM_NOANIMATION,
				x,
				y,
				handle,
				nil)
			return 0
		}

	case win.WM_KEYDOWN:
		wb.handleKeyDown(wParam, lParam)

	case win.WM_KEYUP:
		wb.handleKeyUp(wParam, lParam)

	case win.WM_DROPFILES:
		wb.dropFilesPublisher.Publish(win.HDROP(wParam))

	case win.WM_SIZE, win.WM_SIZING:
		if msg == win.WM_SIZE {
			if widget, ok := wb.window.(Widget); ok {
				widget.AsWidgetBase().invalidateBorderInParent()
			}
		}

		wb.sizeChangedPublisher.Publish()

	case win.WM_WINDOWPOSCHANGED:
		wb.boundsChangedPublisher.Publish()

	case win.WM_THEMECHANGED:
		wb.window.(ApplySysColorser).ApplySysColors()

	case win.WM_DESTROY:
		if wb.origWndProcPtr != 0 {
			// As we subclass all windows of system classes, we prevented the
			// clean-up code in the WM_NCDESTROY handlers of some windows from
			// being called. To fix this, we restore the original window
			// procedure here.
			win.SetWindowLongPtr(wb.hWnd, win.GWLP_WNDPROC, wb.origWndProcPtr)
		}

		delete(hwnd2WindowBase, hwnd)

		wb.window.Dispose()
		wb.hWnd = 0
	}

	if window != nil {
		if wndProc := window.AsWindowBase().origWndProcPtr; wndProc != 0 {
			return win.CallWindowProc(wndProc, hwnd, msg, wParam, lParam)
		}
	}

	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}
