// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"image"
	"path/filepath"
	"syscall"
)

import (
	"github.com/lxn/win"
)

// Icon is a bitmap that supports transparency and combining multiple
// variants of an image in different resolutions.
type Icon struct {
	hIcon   win.HICON
	isStock bool
}

func IconApplication() *Icon {
	return &Icon{win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_APPLICATION)), true}
}

func IconError() *Icon {
	return &Icon{win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_ERROR)), true}
}

func IconQuestion() *Icon {
	return &Icon{win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_QUESTION)), true}
}

func IconWarning() *Icon {
	return &Icon{win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_WARNING)), true}
}

func IconInformation() *Icon {
	return &Icon{win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_INFORMATION)), true}
}

func IconWinLogo() *Icon {
	return &Icon{win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_WINLOGO)), true}
}

func IconShield() *Icon {
	return &Icon{win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_SHIELD)), true}
}

// NewIconFromFile returns a new Icon, using the specified icon image file.
func NewIconFromFile(filePath string) (*Icon, error) {
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, wrapError(err)
	}

	hIcon := win.HICON(win.LoadImage(
		0,
		syscall.StringToUTF16Ptr(absFilePath),
		win.IMAGE_ICON,
		0,
		0,
		win.LR_DEFAULTSIZE|win.LR_LOADFROMFILE))
	if hIcon == 0 {
		return nil, lastError("LoadImage")
	}

	return &Icon{hIcon: hIcon}, nil
}

// NewIconFromResource returns a new Icon, using the specified icon resource.
func NewIconFromResource(resName string) (ic *Icon, err error) {
	hInst := win.GetModuleHandle(nil)
	if hInst == 0 {
		err = lastError("GetModuleHandle")
		return
	}
	if hIcon := win.LoadIcon(hInst, syscall.StringToUTF16Ptr(resName)); hIcon == 0 {
		err = lastError("LoadIcon")
	} else {
		ic = &Icon{hIcon: hIcon}
	}
	return
}

func NewIconFromResourceId(id uintptr) (*Icon, error) {
	hInst := win.GetModuleHandle(nil)
	if hInst == 0 {
		return nil, lastError("GetModuleHandle")
	}

	hIcon := win.LoadIcon(hInst, win.MAKEINTRESOURCE(id))
	if hIcon == 0 {
		return nil, lastError("LoadIcon")
	}

	return &Icon{hIcon: hIcon}, nil
}

func NewIconFromImage(im image.Image) (ic *Icon, err error) {
	hIcon, err := createAlphaCursorOrIconFromImage(im, image.Pt(0, 0), true)
	if err != nil {
		return nil, err
	}
	return &Icon{hIcon: hIcon}, nil
}

// Dispose releases the operating system resources associated with the Icon.
func (i *Icon) Dispose() {
	if i.isStock || i.hIcon == 0 {
		return
	}

	win.DestroyIcon(i.hIcon)
	i.hIcon = 0
}

func (i *Icon) draw(hdc win.HDC, location Point) error {
	s := i.Size()

	return i.drawStretched(hdc, Rectangle{location.X, location.Y, s.Width, s.Height})
}

func (i *Icon) drawStretched(hdc win.HDC, bounds Rectangle) error {
	if !win.DrawIconEx(hdc, int32(bounds.X), int32(bounds.Y), i.hIcon, int32(bounds.Width), int32(bounds.Height), 0, 0, win.DI_NORMAL) {
		return lastError("DrawIconEx")
	}

	return nil
}

func (i *Icon) Size() Size {
	return Size{int(win.GetSystemMetrics(win.SM_CXICON)), int(win.GetSystemMetrics(win.SM_CYICON))}
}

// create an Alpha Icon or Cursor from an Image
// http://support.microsoft.com/kb/318876
func createAlphaCursorOrIconFromImage(im image.Image, hotspot image.Point, fIcon bool) (win.HICON, error) {
	hBitmap, err := hBitmapFromImage(im)
	if err != nil {
		return 0, err
	}
	defer win.DeleteObject(win.HGDIOBJ(hBitmap))

	// Create an empty mask bitmap.
	hMonoBitmap := win.CreateBitmap(int32(im.Bounds().Dx()), int32(im.Bounds().Dy()), 1, 1, nil)
	if hMonoBitmap == 0 {
		return 0, newError("CreateBitmap failed")
	}
	defer win.DeleteObject(win.HGDIOBJ(hMonoBitmap))

	var ii win.ICONINFO
	if fIcon {
		ii.FIcon = win.TRUE
	}
	ii.XHotspot = uint32(hotspot.X)
	ii.YHotspot = uint32(hotspot.Y)
	ii.HbmMask = hMonoBitmap
	ii.HbmColor = hBitmap

	// Create the alpha cursor with the alpha DIB section.
	hIconOrCursor := win.CreateIconIndirect(&ii)

	return hIconOrCursor, nil
}
