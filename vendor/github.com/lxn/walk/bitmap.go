// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"image"
	"syscall"
	"unsafe"
)

import (
	"github.com/lxn/win"
)

func withCompatibleDC(f func(hdc win.HDC) error) error {
	hdc := win.CreateCompatibleDC(0)
	if hdc == 0 {
		return newError("CreateCompatibleDC failed")
	}
	defer win.DeleteDC(hdc)

	return f(hdc)
}

func hPackedDIBFromHBITMAP(hBmp win.HBITMAP) (win.HGLOBAL, error) {
	var dib win.DIBSECTION
	if win.GetObject(win.HGDIOBJ(hBmp), unsafe.Sizeof(dib), unsafe.Pointer(&dib)) == 0 {
		return 0, newError("GetObject failed")
	}

	bmihSize := uintptr(unsafe.Sizeof(dib.DsBmih))
	pixelsSize := uintptr(
		int32(dib.DsBmih.BiBitCount) * dib.DsBmih.BiWidth * dib.DsBmih.BiHeight)

	totalSize := bmihSize + pixelsSize

	hPackedDIB := win.GlobalAlloc(win.GHND, totalSize)
	dest := win.GlobalLock(hPackedDIB)
	defer win.GlobalUnlock(hPackedDIB)

	src := unsafe.Pointer(&dib.DsBmih)

	win.MoveMemory(dest, src, bmihSize)

	dest = unsafe.Pointer(uintptr(dest) + bmihSize)
	src = unsafe.Pointer(uintptr(src) + bmihSize)

	win.MoveMemory(dest, src, pixelsSize)

	return hPackedDIB, nil
}

func hBitmapFromImage(im image.Image) (win.HBITMAP, error) {
	var bi win.BITMAPV5HEADER
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(im.Bounds().Dx())
	bi.BiHeight = -int32(im.Bounds().Dy())
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = win.BI_BITFIELDS
	// The following mask specification specifies a supported 32 BPP
	// alpha format for Windows XP.
	bi.BV4RedMask = 0x00FF0000
	bi.BV4GreenMask = 0x0000FF00
	bi.BV4BlueMask = 0x000000FF
	bi.BV4AlphaMask = 0xFF000000

	hdc := win.GetDC(0)
	defer win.ReleaseDC(0, hdc)

	var lpBits unsafe.Pointer

	// Create the DIB section with an alpha channel.
	hBitmap := win.CreateDIBSection(hdc, &bi.BITMAPINFOHEADER, win.DIB_RGB_COLORS, &lpBits, 0, 0)
	switch hBitmap {
	case 0, win.ERROR_INVALID_PARAMETER:
		return 0, newError("CreateDIBSection failed")
	}

	// Fill the image
	bitmap_array := (*[1 << 30]byte)(unsafe.Pointer(lpBits))
	i := 0
	for y := im.Bounds().Min.Y; y != im.Bounds().Max.Y; y++ {
		for x := im.Bounds().Min.X; x != im.Bounds().Max.X; x++ {
			r, g, b, a := im.At(x, y).RGBA()
			bitmap_array[i+3] = byte(a >> 8)
			bitmap_array[i+2] = byte(r >> 8)
			bitmap_array[i+1] = byte(g >> 8)
			bitmap_array[i+0] = byte(b >> 8)
			i += 4
		}
	}

	return hBitmap, nil
}

func hBitmapFromWindow(window Window) (win.HBITMAP, error) {
	hdcMem := win.CreateCompatibleDC(0)
	if hdcMem == 0 {
		return 0, newError("CreateCompatibleDC failed")
	}
	defer win.DeleteDC(hdcMem)

	var r win.RECT
	if !win.GetWindowRect(window.Handle(), &r) {
		return 0, newError("GetWindowRect failed")
	}

	hdc := win.GetDC(window.Handle())
	width, height := r.Right-r.Left, r.Bottom-r.Top
	hBmp := win.CreateCompatibleBitmap(hdc, width, height)
	win.ReleaseDC(window.Handle(), hdc)

	hOld := win.SelectObject(hdcMem, win.HGDIOBJ(hBmp))
	flags := win.PRF_CHILDREN | win.PRF_CLIENT | win.PRF_ERASEBKGND | win.PRF_NONCLIENT | win.PRF_OWNED
	window.SendMessage(win.WM_PRINT, uintptr(hdcMem), uintptr(flags))

	win.SelectObject(hdcMem, hOld)
	return hBmp, nil
}

type Bitmap struct {
	hBmp       win.HBITMAP
	hPackedDIB win.HGLOBAL
	size       Size
}

func newBitmapFromHBITMAP(hBmp win.HBITMAP) (bmp *Bitmap, err error) {
	var dib win.DIBSECTION
	if win.GetObject(win.HGDIOBJ(hBmp), unsafe.Sizeof(dib), unsafe.Pointer(&dib)) == 0 {
		return nil, newError("GetObject failed")
	}

	bmih := &dib.DsBmih

	bmihSize := uintptr(unsafe.Sizeof(*bmih))
	pixelsSize := uintptr(int32(bmih.BiBitCount)*bmih.BiWidth*bmih.BiHeight) / 8

	totalSize := uintptr(bmihSize + pixelsSize)

	hPackedDIB := win.GlobalAlloc(win.GHND, totalSize)
	dest := win.GlobalLock(hPackedDIB)
	defer win.GlobalUnlock(hPackedDIB)

	src := unsafe.Pointer(&dib.DsBmih)

	win.MoveMemory(dest, src, bmihSize)

	dest = unsafe.Pointer(uintptr(dest) + bmihSize)
	src = dib.DsBm.BmBits

	win.MoveMemory(dest, src, pixelsSize)

	return &Bitmap{
		hBmp:       hBmp,
		hPackedDIB: hPackedDIB,
		size: Size{
			int(bmih.BiWidth),
			int(bmih.BiHeight),
		},
	}, nil
}

func NewBitmap(size Size) (bmp *Bitmap, err error) {
	var hdr win.BITMAPINFOHEADER
	hdr.BiSize = uint32(unsafe.Sizeof(hdr))
	hdr.BiBitCount = 24
	hdr.BiCompression = win.BI_RGB
	hdr.BiPlanes = 1
	hdr.BiWidth = int32(size.Width)
	hdr.BiHeight = int32(size.Height)

	err = withCompatibleDC(func(hdc win.HDC) error {
		hBmp := win.CreateDIBSection(hdc, &hdr, win.DIB_RGB_COLORS, nil, 0, 0)
		switch hBmp {
		case 0, win.ERROR_INVALID_PARAMETER:
			return newError("CreateDIBSection failed")
		}

		bmp, err = newBitmapFromHBITMAP(hBmp)
		return err
	})

	return
}

func NewBitmapFromFile(filePath string) (*Bitmap, error) {
	var si win.GdiplusStartupInput
	si.GdiplusVersion = 1
	if status := win.GdiplusStartup(&si, nil); status != win.Ok {
		return nil, newError(fmt.Sprintf("GdiplusStartup failed with status '%s'", status))
	}
	defer win.GdiplusShutdown()

	var gpBmp *win.GpBitmap
	if status := win.GdipCreateBitmapFromFile(syscall.StringToUTF16Ptr(filePath), &gpBmp); status != win.Ok {
		return nil, newError(fmt.Sprintf("GdipCreateBitmapFromFile failed with status '%s' for file '%s'", status, filePath))
	}
	defer win.GdipDisposeImage((*win.GpImage)(gpBmp))

	var hBmp win.HBITMAP
	if status := win.GdipCreateHBITMAPFromBitmap(gpBmp, &hBmp, 0); status != win.Ok {
		return nil, newError(fmt.Sprintf("GdipCreateHBITMAPFromBitmap failed with status '%s' for file '%s'", status, filePath))
	}

	return newBitmapFromHBITMAP(hBmp)
}

func NewBitmapFromImage(im image.Image) (*Bitmap, error) {
	hBmp, err := hBitmapFromImage(im)
	if err != nil {
		return nil, err
	}
	return newBitmapFromHBITMAP(hBmp)
}

func (bmp *Bitmap) withSelectedIntoMemDC(f func(hdcMem win.HDC) error) error {
	return withCompatibleDC(func(hdcMem win.HDC) error {
		hBmpOld := win.SelectObject(hdcMem, win.HGDIOBJ(bmp.hBmp))
		if hBmpOld == 0 {
			return newError("SelectObject failed")
		}
		defer win.SelectObject(hdcMem, hBmpOld)

		return f(hdcMem)
	})
}

func (bmp *Bitmap) draw(hdc win.HDC, location Point) error {
	return bmp.withSelectedIntoMemDC(func(hdcMem win.HDC) error {
		size := bmp.Size()

		if !win.BitBlt(
			hdc,
			int32(location.X),
			int32(location.Y),
			int32(size.Width),
			int32(size.Height),
			hdcMem,
			0,
			0,
			win.SRCCOPY) {

			return lastError("BitBlt")
		}

		return nil
	})
}

func (bmp *Bitmap) drawStretched(hdc win.HDC, bounds Rectangle) error {
	return bmp.withSelectedIntoMemDC(func(hdcMem win.HDC) error {
		size := bmp.Size()

		if !win.StretchBlt(
			hdc,
			int32(bounds.X),
			int32(bounds.Y),
			int32(bounds.Width),
			int32(bounds.Height),
			hdcMem,
			0,
			0,
			int32(size.Width),
			int32(size.Height),
			win.SRCCOPY) {

			return newError("StretchBlt failed")
		}

		return nil
	})
}

func (bmp *Bitmap) handle() win.HBITMAP {
	return bmp.hBmp
}

func (bmp *Bitmap) Dispose() {
	if bmp.hBmp != 0 {
		win.DeleteObject(win.HGDIOBJ(bmp.hBmp))

		win.GlobalUnlock(bmp.hPackedDIB)
		win.GlobalFree(bmp.hPackedDIB)

		bmp.hPackedDIB = 0
		bmp.hBmp = 0
	}
}

func (bmp *Bitmap) Size() Size {
	return bmp.size
}
