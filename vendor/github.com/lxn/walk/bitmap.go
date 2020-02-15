// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"fmt"
	"image"
	"image/color"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

type Bitmap struct {
	hBmp       win.HBITMAP
	hPackedDIB win.HGLOBAL
	size       Size
	dpi        int
}

func NewBitmap(size Size) (*Bitmap, error) {
	return newBitmap(size, false)
}

func NewBitmapWithTransparentPixels(size Size) (*Bitmap, error) {
	return newBitmap(size, true)
}

func newBitmap(size Size, transparent bool) (bmp *Bitmap, err error) {
	err = withCompatibleDC(func(hdc win.HDC) error {
		bufSize := size.Width * size.Height * 4

		var hdr win.BITMAPINFOHEADER
		hdr.BiSize = uint32(unsafe.Sizeof(hdr))
		hdr.BiBitCount = 32
		hdr.BiCompression = win.BI_RGB
		hdr.BiPlanes = 1
		hdr.BiWidth = int32(size.Width)
		hdr.BiHeight = int32(size.Height)
		hdr.BiSizeImage = uint32(bufSize)

		var bitsPtr unsafe.Pointer

		hBmp := win.CreateDIBSection(hdc, &hdr, win.DIB_RGB_COLORS, &bitsPtr, 0, 0)
		switch hBmp {
		case 0, win.ERROR_INVALID_PARAMETER:
			return newError("CreateDIBSection failed")
		}

		if transparent {
			win.GdiFlush()

			bits := (*[1 << 24]byte)(bitsPtr)

			for i := 0; i < bufSize; i += 4 {
				// Mark pixel as not drawn to by GDI.
				bits[i+3] = 0x01
			}
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

func NewBitmapFromResource(name string) (*Bitmap, error) {
	return newBitmapFromResource(syscall.StringToUTF16Ptr(name))
}

func NewBitmapFromResourceId(id int) (*Bitmap, error) {
	return newBitmapFromResource(win.MAKEINTRESOURCE(uintptr(id)))
}

func newBitmapFromResource(res *uint16) (bm *Bitmap, err error) {
	hInst := win.GetModuleHandle(nil)
	if hInst == 0 {
		err = lastError("GetModuleHandle")
		return
	}

	if hBmp := win.LoadImage(hInst, res, win.IMAGE_BITMAP, 0, 0, win.LR_CREATEDIBSECTION); hBmp == 0 {
		err = lastError("LoadImage")
	} else {
		bm, err = newBitmapFromHBITMAP(win.HBITMAP(hBmp))
	}

	return
}

func NewBitmapFromImageWithSize(image Image, size Size) (*Bitmap, error) {
	var disposables Disposables
	defer disposables.Treat()

	bmp, err := NewBitmapWithTransparentPixels(size)
	if err != nil {
		return nil, err
	}
	disposables.Add(bmp)

	dpi := int(float64(size.Width) / float64(image.Size().Width) * 96.0)

	canvas, err := NewCanvasFromImage(bmp)
	if err != nil {
		return nil, err
	}
	defer canvas.Dispose()

	canvas.dpix = dpi
	canvas.dpiy = dpi

	size = SizeTo96DPI(size, dpi)

	if err := canvas.DrawImageStretched(image, Rectangle{0, 0, size.Width, size.Height}); err != nil {
		return nil, err
	}

	disposables.Spare()

	return bmp, nil
}

func NewBitmapFromWindow(window Window) (*Bitmap, error) {
	hBmp, err := hBitmapFromWindow(window)
	if err != nil {
		return nil, err
	}

	return newBitmapFromHBITMAP(hBmp)
}

func NewBitmapFromIcon(icon *Icon, size Size) (*Bitmap, error) {
	hBmp, err := hBitmapFromIcon(icon, size)
	if err != nil {
		return nil, err
	}

	return newBitmapFromHBITMAP(hBmp)
}

func (bmp *Bitmap) ToImage() (*image.RGBA, error) {
	var bi win.BITMAPINFO
	bi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bi.BmiHeader))
	hdc := win.GetDC(0)
	if ret := win.GetDIBits(hdc, bmp.hBmp, 0, 0, nil, &bi, win.DIB_RGB_COLORS); ret == 0 {
		return nil, newError("GetDIBits get bitmapinfo failed")
	}

	buf := make([]byte, bi.BmiHeader.BiSizeImage)
	bi.BmiHeader.BiCompression = win.BI_RGB
	if ret := win.GetDIBits(hdc, bmp.hBmp, 0, uint32(bi.BmiHeader.BiHeight), &buf[0], &bi, win.DIB_RGB_COLORS); ret == 0 {
		return nil, newError("GetDIBits failed")
	}

	width := int(bi.BmiHeader.BiWidth)
	height := int(bi.BmiHeader.BiHeight)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	n := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			a := buf[n+3]
			r := buf[n+2]
			g := buf[n+1]
			b := buf[n+0]
			n += int(bi.BmiHeader.BiBitCount) / 8
			img.Set(x, height-y-1, color.RGBA{r, g, b, a})
		}
	}

	return img, nil
}

func (bmp *Bitmap) postProcess() {
	var bi win.BITMAPINFO
	bi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bi.BmiHeader))
	hdc := win.GetDC(0)
	if ret := win.GetDIBits(hdc, bmp.hBmp, 0, 0, nil, &bi, win.DIB_RGB_COLORS); ret == 0 {
		return
	}

	buf := make([]byte, bi.BmiHeader.BiSizeImage)
	bi.BmiHeader.BiCompression = win.BI_RGB
	if ret := win.GetDIBits(hdc, bmp.hBmp, 0, uint32(bi.BmiHeader.BiHeight), &buf[0], &bi, win.DIB_RGB_COLORS); ret == 0 {
		return
	}

	win.GdiFlush()

	for i := 0; i < len(buf); i += 4 {
		switch buf[i+3] {
		case 0x00:
			// The pixel has been drawn to by GDI, so we make it fully opaque.
			buf[i+3] = 0xff

		case 0x01:
			// The pixel has not been drawn to by GDI, so we make it fully transparent.
			buf[i+3] = 0x00
		}
	}

	if 0 == win.SetDIBits(hdc, bmp.hBmp, 0, uint32(bi.BmiHeader.BiHeight), &buf[0], &bi, win.DIB_RGB_COLORS) {
		return
	}
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

func (bmp *Bitmap) handle() win.HBITMAP {
	return bmp.hBmp
}

func (bmp *Bitmap) draw(hdc win.HDC, location Point) error {
	return bmp.drawStretched(hdc, Rectangle{X: location.X, Y: location.Y, Width: bmp.size.Width, Height: bmp.size.Height})
}

func (bmp *Bitmap) drawStretched(hdc win.HDC, bounds Rectangle) error {
	return bmp.alphaBlend(hdc, bounds, 255)
}

func (bmp *Bitmap) alphaBlend(hdc win.HDC, bounds Rectangle, opacity byte) error {
	return bmp.alphaBlendPart(hdc, bounds, Rectangle{0, 0, bmp.size.Width, bmp.size.Height}, opacity)
}

func (bmp *Bitmap) alphaBlendPart(hdc win.HDC, dst, src Rectangle, opacity byte) error {
	return bmp.withSelectedIntoMemDC(func(hdcMem win.HDC) error {
		if !win.AlphaBlend(
			hdc,
			int32(dst.X),
			int32(dst.Y),
			int32(dst.Width),
			int32(dst.Height),
			hdcMem,
			int32(src.X),
			int32(src.Y),
			int32(src.Width),
			int32(src.Height),
			win.BLENDFUNCTION{AlphaFormat: win.AC_SRC_ALPHA, SourceConstantAlpha: opacity}) {

			return newError("AlphaBlend failed")
		}

		return nil
	})
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

func hBitmapFromIcon(icon *Icon, size Size) (win.HBITMAP, error) {
	hdc := win.GetDC(0)
	defer win.ReleaseDC(0, hdc)

	hdcMem := win.CreateCompatibleDC(hdc)
	if hdcMem == 0 {
		return 0, newError("CreateCompatibleDC failed")
	}
	defer win.DeleteDC(hdcMem)

	var bi win.BITMAPV5HEADER
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(size.Width)
	bi.BiHeight = int32(size.Height)
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = win.BI_RGB
	// The following mask specification specifies a supported 32 BPP
	// alpha format for Windows XP.
	bi.BV4RedMask = 0x00FF0000
	bi.BV4GreenMask = 0x0000FF00
	bi.BV4BlueMask = 0x000000FF
	bi.BV4AlphaMask = 0xFF000000

	hBmp := win.CreateDIBSection(hdcMem, &bi.BITMAPINFOHEADER, win.DIB_RGB_COLORS, nil, 0, 0)
	switch hBmp {
	case 0, win.ERROR_INVALID_PARAMETER:
		return 0, newError("CreateDIBSection failed")
	}

	hOld := win.SelectObject(hdcMem, win.HGDIOBJ(hBmp))
	defer win.SelectObject(hdcMem, hOld)

	err := icon.drawStretched(hdcMem, Rectangle{Width: size.Width, Height: size.Height})
	if err != nil {
		return 0, err
	}

	return hBmp, nil
}

func withCompatibleDC(f func(hdc win.HDC) error) error {
	hdc := win.CreateCompatibleDC(0)
	if hdc == 0 {
		return newError("CreateCompatibleDC failed")
	}
	defer win.DeleteDC(hdc)

	return f(hdc)
}
