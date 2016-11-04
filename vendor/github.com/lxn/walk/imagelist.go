// Copyright 2010 The Walk Authors. All rights reserved.
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

type ImageList struct {
	hIml      win.HIMAGELIST
	maskColor Color
}

func NewImageList(imageSize Size, maskColor Color) (*ImageList, error) {
	hIml := win.ImageList_Create(
		int32(imageSize.Width),
		int32(imageSize.Height),
		win.ILC_MASK|win.ILC_COLOR32,
		8,
		8)
	if hIml == 0 {
		return nil, newError("ImageList_Create failed")
	}

	return &ImageList{hIml: hIml, maskColor: maskColor}, nil
}

func (il *ImageList) Handle() win.HIMAGELIST {
	return il.hIml
}

func (il *ImageList) Add(bitmap, maskBitmap *Bitmap) (int, error) {
	if bitmap == nil {
		return 0, newError("bitmap cannot be nil")
	}

	var maskHandle win.HBITMAP
	if maskBitmap != nil {
		maskHandle = maskBitmap.handle()
	}

	index := int(win.ImageList_Add(il.hIml, bitmap.handle(), maskHandle))
	if index == -1 {
		return 0, newError("ImageList_Add failed")
	}

	return index, nil
}

func (il *ImageList) AddMasked(bitmap *Bitmap) (int32, error) {
	if bitmap == nil {
		return 0, newError("bitmap cannot be nil")
	}

	index := win.ImageList_AddMasked(
		il.hIml,
		bitmap.handle(),
		win.COLORREF(il.maskColor))
	if index == -1 {
		return 0, newError("ImageList_AddMasked failed")
	}

	return index, nil
}

func (il *ImageList) Dispose() {
	if il.hIml != 0 {
		win.ImageList_Destroy(il.hIml)
		il.hIml = 0
	}
}

func (il *ImageList) MaskColor() Color {
	return il.maskColor
}

func imageListForImage(image interface{}) (hIml win.HIMAGELIST, isSysIml bool, err error) {
	if filePath, ok := image.(string); ok {
		_, hIml = iconIndexAndHImlForFilePath(filePath)
		isSysIml = hIml != 0
	} else {
		w, h := win.GetSystemMetrics(win.SM_CXSMICON), win.GetSystemMetrics(win.SM_CYSMICON)

		hIml = win.ImageList_Create(w, h, win.ILC_MASK|win.ILC_COLOR32, 8, 8)
		if hIml == 0 {
			return 0, false, newError("ImageList_Create failed")
		}
	}

	return
}

func iconIndexAndHImlForFilePath(filePath string) (int32, win.HIMAGELIST) {
	var shfi win.SHFILEINFO

	if hIml := win.HIMAGELIST(win.SHGetFileInfo(
		syscall.StringToUTF16Ptr(filePath),
		0,
		&shfi,
		uint32(unsafe.Sizeof(shfi)),
		win.SHGFI_SYSICONINDEX|win.SHGFI_SMALLICON)); hIml != 0 {

		return shfi.IIcon, hIml
	}

	return -1, 0
}

func imageIndexMaybeAdd(image interface{}, hIml win.HIMAGELIST, isSysIml bool, imageUintptr2Index map[uintptr]int32, filePath2IconIndex map[string]int32) int32 {
	if !isSysIml {
		return imageIndexAddIfNotExists(image, hIml, imageUintptr2Index)
	} else if filePath, ok := image.(string); ok {
		if iIcon, ok := filePath2IconIndex[filePath]; ok {
			return iIcon
		}

		if iIcon, _ := iconIndexAndHImlForFilePath(filePath); iIcon != -1 {
			filePath2IconIndex[filePath] = iIcon
			return iIcon
		}
	}

	return -1
}

func imageIndexAddIfNotExists(image interface{}, hIml win.HIMAGELIST, imageUintptr2Index map[uintptr]int32) int32 {
	imageIndex := int32(-1)

	if image != nil {
		var ptr uintptr
		switch img := image.(type) {
		case *Bitmap:
			ptr = uintptr(unsafe.Pointer(img))

		case *Icon:
			ptr = uintptr(unsafe.Pointer(img))
		}

		if ptr == 0 {
			return -1
		}

		if imageIndex, ok := imageUintptr2Index[ptr]; ok {
			return imageIndex
		}

		switch img := image.(type) {
		case *Bitmap:
			imageIndex = win.ImageList_AddMasked(hIml, img.hBmp, 0)

		case *Icon:
			imageIndex = win.ImageList_ReplaceIcon(hIml, -1, img.hIcon)
		}

		if imageIndex > -1 {
			imageUintptr2Index[ptr] = imageIndex
		}
	}

	return imageIndex
}
