// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"strings"

	"github.com/lxn/win"
)

type Image interface {
	draw(hdc win.HDC, location Point) error
	drawStretched(hdc win.HDC, bounds Rectangle) error
	Dispose()
	Size() Size
}

func NewImageFromFile(filePath string) (Image, error) {
	if strings.HasSuffix(filePath, ".ico") {
		return NewIconFromFile(filePath)
	} else if strings.HasSuffix(filePath, ".emf") {
		return NewMetafileFromFile(filePath)
	}

	return NewBitmapFromFile(filePath)
}

type PaintFuncImage struct {
	size96dpi Size
	paint     func(canvas *Canvas, bounds Rectangle) error
	dispose   func()
}

func NewPaintFuncImage(size Size, paint func(canvas *Canvas, bounds Rectangle) error) *PaintFuncImage {
	return &PaintFuncImage{size96dpi: size, paint: paint}
}

func NewPaintFuncImageWithDispose(size Size, paint func(canvas *Canvas, bounds Rectangle) error, dispose func()) *PaintFuncImage {
	return &PaintFuncImage{size96dpi: size, paint: paint, dispose: dispose}
}

func (pfi *PaintFuncImage) draw(hdc win.HDC, location Point) error {
	dpi := dpiForHDC(hdc)
	size := SizeFrom96DPI(pfi.size96dpi, dpi)

	return pfi.drawStretched(hdc, Rectangle{location.X, location.Y, size.Width, size.Height})
}

func (pfi *PaintFuncImage) drawStretched(hdc win.HDC, bounds Rectangle) error {
	canvas, err := newCanvasFromHDC(hdc)
	if err != nil {
		return err
	}
	defer canvas.Dispose()

	return pfi.drawStretchedOnCanvas(canvas, bounds)
}

func (pfi *PaintFuncImage) drawStretchedOnCanvas(canvas *Canvas, bounds Rectangle) error {
	bounds = RectangleTo96DPI(bounds, canvas.dpix)

	return pfi.paint(canvas, bounds)
}

func (pfi *PaintFuncImage) Dispose() {
	if pfi.dispose != nil {
		pfi.dispose()
		pfi.dispose = nil
	}
}

func (pfi *PaintFuncImage) Size() Size {
	return pfi.size96dpi
}
