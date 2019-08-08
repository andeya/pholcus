// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
	"syscall"
)

// FontMemResource represents a font resource loaded into memory from
// the application's resources.
type FontMemResource struct {
	hFontResource win.HANDLE
}

func newFontMemResource(resourceName *uint16) (*FontMemResource, error) {
	hModule := win.HMODULE(win.GetModuleHandle(nil))
	if hModule == win.HMODULE(0) {
		return nil, lastError("GetModuleHandle")
	}

	hres := win.FindResource(hModule, resourceName, win.MAKEINTRESOURCE(8) /*RT_FONT*/)
	if hres == win.HRSRC(0) {
		return nil, lastError("FindResource")
	}

	size := win.SizeofResource(hModule, hres)
	if size == 0 {
		return nil, lastError("SizeofResource")
	}

	hResLoad := win.LoadResource(hModule, hres)
	if hResLoad == win.HGLOBAL(0) {
		return nil, lastError("LoadResource")
	}

	ptr := win.LockResource(hResLoad)
	if ptr == 0 {
		return nil, lastError("LockResource")
	}

	numFonts := uint32(0)
	hFontResource := win.AddFontMemResourceEx(ptr, size, nil, &numFonts)

	if hFontResource == win.HANDLE(0) || numFonts == 0 {
		return nil, lastError("AddFontMemResource")
	}

	return &FontMemResource{hFontResource: hFontResource}, nil
}

// NewFontMemResourceByName function loads a font resource from the executable's resources
// using the resource name.
// The font must be embedded into resources using corresponding operator in the
// application's RC script.
func NewFontMemResourceByName(name string) (*FontMemResource, error) {
	lpstr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	return newFontMemResource(lpstr)
}

// NewFontMemResourceById function loads a font resource from the executable's resources
// using the resource ID.
// The font must be embedded into resources using corresponding operator in the
// application's RC script.
func NewFontMemResourceById(id int) (*FontMemResource, error) {
	return newFontMemResource(win.MAKEINTRESOURCE(uintptr(id)))
}

// Dispose removes the font resource from memory
func (fmr *FontMemResource) Dispose() {
	if fmr.hFontResource != 0 {
		win.RemoveFontMemResourceEx(fmr.hFontResource)
		fmr.hFontResource = 0
	}
}
