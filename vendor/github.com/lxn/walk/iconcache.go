// Copyright 2019 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

var iconCache = NewIconCache()

type IconCache struct {
	imageAndDPI2Bitmap map[imageAndDPI]*Bitmap
	imageAndDPI2Icon   map[imageAndDPI]*Icon
}

type imageAndDPI struct {
	image Image
	dpi   int
}

func NewIconCache() *IconCache {
	return &IconCache{
		imageAndDPI2Bitmap: make(map[imageAndDPI]*Bitmap),
		imageAndDPI2Icon:   make(map[imageAndDPI]*Icon),
	}
}

func (ic *IconCache) Clear() {
	for key, bmp := range ic.imageAndDPI2Bitmap {
		bmp.Dispose()
		delete(ic.imageAndDPI2Bitmap, key)
	}
	for key, ico := range ic.imageAndDPI2Icon {
		ico.Dispose()
		delete(ic.imageAndDPI2Icon, key)
	}
}

func (ic *IconCache) Dispose() {
	ic.Clear()
}

func (ic *IconCache) Bitmap(image Image, dpi int) (*Bitmap, error) {
	key := imageAndDPI{image, dpi}

	if bmp, ok := ic.imageAndDPI2Bitmap[key]; ok {
		return bmp, nil
	}

	size := SizeFrom96DPI(image.Size(), dpi)

	bmp, err := NewBitmapFromImageWithSize(image, size)
	if err != nil {
		return nil, err
	}

	ic.imageAndDPI2Bitmap[key] = bmp

	return bmp, nil
}

func (ic *IconCache) Icon(image Image, dpi int) (*Icon, error) {
	key := imageAndDPI{image, dpi}

	if ico, ok := ic.imageAndDPI2Icon[key]; ok {
		return ico, nil
	}

	if ico, ok := image.(*Icon); ok {
		if ico.handleForDPI(dpi) != 0 {
			ic.imageAndDPI2Icon[key] = ico
			return ico, nil
		}
	}

	ico, err := newIconFromImageForDPI(image, dpi)
	if err != nil {
		return nil, err
	}

	ic.imageAndDPI2Icon[key] = ico

	return ico, nil
}
