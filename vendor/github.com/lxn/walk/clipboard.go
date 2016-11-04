// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

const clipboardWindowClass = `\o/ Walk_Clipboard_Class \o/`

func init() {
	MustRegisterWindowClass(clipboardWindowClass)

	hwnd := win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr(clipboardWindowClass),
		nil,
		0,
		0,
		0,
		0,
		0,
		win.HWND_MESSAGE,
		0,
		0,
		nil)

	if hwnd == 0 {
		panic("failed to create clipboard window")
	}

	if !win.AddClipboardFormatListener(hwnd) {
		lastError("AddClipboardFormatListener")
	}

	clipboard.hwnd = hwnd
}

func clipboardWndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case win.WM_CLIPBOARDUPDATE:
		clipboard.contentsChangedPublisher.Publish()
		return 0
	}

	return win.DefWindowProc(hwnd, msg, wp, lp)
}

var clipboard ClipboardService

// Clipboard returns an object that provides access to the system clipboard.
func Clipboard() *ClipboardService {
	return &clipboard
}

// ClipboardService provides access to the system clipboard.
type ClipboardService struct {
	hwnd                     win.HWND
	contentsChangedPublisher EventPublisher
}

// ContentsChanged returns an Event that you can attach to for handling
// clipboard content changes.
func (c *ClipboardService) ContentsChanged() *Event {
	return c.contentsChangedPublisher.Event()
}

// Clear clears the contents of the clipboard.
func (c *ClipboardService) Clear() error {
	return c.withOpenClipboard(func() error {
		if !win.EmptyClipboard() {
			return lastError("EmptyClipboard")
		}

		return nil
	})
}

// ContainsText returns whether the clipboard currently contains text data.
func (c *ClipboardService) ContainsText() (available bool, err error) {
	err = c.withOpenClipboard(func() error {
		available = win.IsClipboardFormatAvailable(win.CF_UNICODETEXT)

		return nil
	})

	return
}

// Text returns the current text data of the clipboard.
func (c *ClipboardService) Text() (text string, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_UNICODETEXT))
		if hMem == 0 {
			return lastError("GetClipboardData")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)

		text = win.UTF16PtrToString((*uint16)(p))

		return nil
	})

	return
}

// SetText sets the current text data of the clipboard.
func (c *ClipboardService) SetText(s string) error {
	return c.withOpenClipboard(func() error {
		utf16, err := syscall.UTF16FromString(s)
		if err != nil {
			return err
		}

		hMem := win.GlobalAlloc(win.GMEM_MOVEABLE, uintptr(len(utf16)*2))
		if hMem == 0 {
			return lastError("GlobalAlloc")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}

		win.MoveMemory(p, unsafe.Pointer(&utf16[0]), uintptr(len(utf16)*2))

		win.GlobalUnlock(hMem)

		if 0 == win.SetClipboardData(win.CF_UNICODETEXT, win.HANDLE(hMem)) {
			// We need to free hMem.
			defer win.GlobalFree(hMem)

			return lastError("SetClipboardData")
		}

		// The system now owns the memory referred to by hMem.

		return nil
	})
}

func (c *ClipboardService) withOpenClipboard(f func() error) error {
	if !win.OpenClipboard(c.hwnd) {
		return lastError("OpenClipboard")
	}
	defer win.CloseClipboard()

	return f()
}
