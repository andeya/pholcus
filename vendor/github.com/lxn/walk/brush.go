// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type HatchStyle int

const (
	HatchHorizontal       HatchStyle = win.HS_HORIZONTAL
	HatchVertical         HatchStyle = win.HS_VERTICAL
	HatchForwardDiagonal  HatchStyle = win.HS_FDIAGONAL
	HatchBackwardDiagonal HatchStyle = win.HS_BDIAGONAL
	HatchCross            HatchStyle = win.HS_CROSS
	HatchDiagonalCross    HatchStyle = win.HS_DIAGCROSS
)

type Brush interface {
	Dispose()
	handle() win.HBRUSH
	logbrush() *win.LOGBRUSH
}

type nullBrush struct {
	hBrush win.HBRUSH
}

func newNullBrush() *nullBrush {
	lb := &win.LOGBRUSH{LbStyle: win.BS_NULL}

	hBrush := win.CreateBrushIndirect(lb)
	if hBrush == 0 {
		panic("failed to create null brush")
	}

	return &nullBrush{hBrush: hBrush}
}

func (b *nullBrush) Dispose() {
	if b.hBrush != 0 {
		win.DeleteObject(win.HGDIOBJ(b.hBrush))

		b.hBrush = 0
	}
}

func (b *nullBrush) handle() win.HBRUSH {
	return b.hBrush
}

func (b *nullBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{LbStyle: win.BS_NULL}
}

var nullBrushSingleton Brush = newNullBrush()

func NullBrush() Brush {
	return nullBrushSingleton
}

type SystemColorBrush struct {
	hBrush     win.HBRUSH
	colorIndex int
}

func NewSystemColorBrush(colorIndex int) (*SystemColorBrush, error) {
	hBrush := win.GetSysColorBrush(colorIndex)
	if hBrush == 0 {
		return nil, newError("GetSysColorBrush failed")
	}

	return &SystemColorBrush{hBrush, colorIndex}, nil
}

func (b *SystemColorBrush) ColorIndex() int {
	return b.colorIndex
}

func (b *SystemColorBrush) Dispose() {
	// nop
}

func (b *SystemColorBrush) handle() win.HBRUSH {
	return b.hBrush
}

func (b *SystemColorBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{
		LbStyle: win.BS_SOLID,
		LbColor: win.COLORREF(win.GetSysColor(b.colorIndex)),
	}
}

type SolidColorBrush struct {
	hBrush win.HBRUSH
	color  Color
}

func NewSolidColorBrush(color Color) (*SolidColorBrush, error) {
	lb := &win.LOGBRUSH{LbStyle: win.BS_SOLID, LbColor: win.COLORREF(color)}

	hBrush := win.CreateBrushIndirect(lb)
	if hBrush == 0 {
		return nil, newError("CreateBrushIndirect failed")
	}

	return &SolidColorBrush{hBrush: hBrush, color: color}, nil
}

func (b *SolidColorBrush) Color() Color {
	return b.color
}

func (b *SolidColorBrush) Dispose() {
	if b.hBrush != 0 {
		win.DeleteObject(win.HGDIOBJ(b.hBrush))

		b.hBrush = 0
	}
}

func (b *SolidColorBrush) handle() win.HBRUSH {
	return b.hBrush
}

func (b *SolidColorBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{LbStyle: win.BS_SOLID, LbColor: win.COLORREF(b.color)}
}

type HatchBrush struct {
	hBrush win.HBRUSH
	color  Color
	style  HatchStyle
}

func NewHatchBrush(color Color, style HatchStyle) (*HatchBrush, error) {
	lb := &win.LOGBRUSH{LbStyle: win.BS_HATCHED, LbColor: win.COLORREF(color), LbHatch: uintptr(style)}

	hBrush := win.CreateBrushIndirect(lb)
	if hBrush == 0 {
		return nil, newError("CreateBrushIndirect failed")
	}

	return &HatchBrush{hBrush: hBrush, color: color, style: style}, nil
}

func (b *HatchBrush) Color() Color {
	return b.color
}

func (b *HatchBrush) Dispose() {
	if b.hBrush != 0 {
		win.DeleteObject(win.HGDIOBJ(b.hBrush))

		b.hBrush = 0
	}
}

func (b *HatchBrush) handle() win.HBRUSH {
	return b.hBrush
}

func (b *HatchBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{LbStyle: win.BS_HATCHED, LbColor: win.COLORREF(b.color), LbHatch: uintptr(b.style)}
}

func (b *HatchBrush) Style() HatchStyle {
	return b.style
}

type BitmapBrush struct {
	hBrush win.HBRUSH
	bitmap *Bitmap
}

func NewBitmapBrush(bitmap *Bitmap) (*BitmapBrush, error) {
	if bitmap == nil {
		return nil, newError("bitmap cannot be nil")
	}

	lb := &win.LOGBRUSH{LbStyle: win.BS_DIBPATTERN, LbColor: win.DIB_RGB_COLORS, LbHatch: uintptr(bitmap.hPackedDIB)}

	hBrush := win.CreateBrushIndirect(lb)
	if hBrush == 0 {
		return nil, newError("CreateBrushIndirect failed")
	}

	return &BitmapBrush{hBrush: hBrush, bitmap: bitmap}, nil
}

func (b *BitmapBrush) Dispose() {
	if b.hBrush != 0 {
		win.DeleteObject(win.HGDIOBJ(b.hBrush))

		b.hBrush = 0
	}
}

func (b *BitmapBrush) handle() win.HBRUSH {
	return b.hBrush
}

func (b *BitmapBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{LbStyle: win.BS_DIBPATTERN, LbColor: win.DIB_RGB_COLORS, LbHatch: uintptr(b.bitmap.hPackedDIB)}
}

func (b *BitmapBrush) Bitmap() *Bitmap {
	return b.bitmap
}
