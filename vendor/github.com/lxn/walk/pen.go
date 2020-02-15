// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type PenStyle int

// Pen styles
const (
	PenSolid       PenStyle = win.PS_SOLID
	PenDash        PenStyle = win.PS_DASH
	PenDot         PenStyle = win.PS_DOT
	PenDashDot     PenStyle = win.PS_DASHDOT
	PenDashDotDot  PenStyle = win.PS_DASHDOTDOT
	PenNull        PenStyle = win.PS_NULL
	PenInsideFrame PenStyle = win.PS_INSIDEFRAME
	PenUserStyle   PenStyle = win.PS_USERSTYLE
	PenAlternate   PenStyle = win.PS_ALTERNATE
)

// Pen cap styles (geometric pens only)
const (
	PenCapRound  PenStyle = win.PS_ENDCAP_ROUND
	PenCapSquare PenStyle = win.PS_ENDCAP_SQUARE
	PenCapFlat   PenStyle = win.PS_ENDCAP_FLAT
)

// Pen join styles (geometric pens only)
const (
	PenJoinBevel PenStyle = win.PS_JOIN_BEVEL
	PenJoinMiter PenStyle = win.PS_JOIN_MITER
	PenJoinRound PenStyle = win.PS_JOIN_ROUND
)

type Pen interface {
	handle() win.HPEN
	Dispose()
	Style() PenStyle
	Width() int
}

type nullPen struct {
	hPen win.HPEN
}

func newNullPen() *nullPen {
	lb := &win.LOGBRUSH{LbStyle: win.BS_NULL}

	hPen := win.ExtCreatePen(win.PS_COSMETIC|win.PS_NULL, 1, lb, 0, nil)
	if hPen == 0 {
		panic("failed to create null brush")
	}

	return &nullPen{hPen: hPen}
}

func (p *nullPen) Dispose() {
	if p.hPen != 0 {
		win.DeleteObject(win.HGDIOBJ(p.hPen))

		p.hPen = 0
	}
}

func (p *nullPen) handle() win.HPEN {
	return p.hPen
}

func (p *nullPen) Style() PenStyle {
	return PenNull
}

func (p *nullPen) Width() int {
	return 0
}

var nullPenSingleton Pen = newNullPen()

func NullPen() Pen {
	return nullPenSingleton
}

type CosmeticPen struct {
	hPen  win.HPEN
	style PenStyle
	color Color
}

func NewCosmeticPen(style PenStyle, color Color) (*CosmeticPen, error) {
	lb := &win.LOGBRUSH{LbStyle: win.BS_SOLID, LbColor: win.COLORREF(color)}

	style |= win.PS_COSMETIC

	hPen := win.ExtCreatePen(uint32(style), 1, lb, 0, nil)
	if hPen == 0 {
		return nil, newError("ExtCreatePen failed")
	}

	return &CosmeticPen{hPen: hPen, style: style, color: color}, nil
}

func (p *CosmeticPen) Dispose() {
	if p.hPen != 0 {
		win.DeleteObject(win.HGDIOBJ(p.hPen))

		p.hPen = 0
	}
}

func (p *CosmeticPen) handle() win.HPEN {
	return p.hPen
}

func (p *CosmeticPen) Style() PenStyle {
	return p.style
}

func (p *CosmeticPen) Color() Color {
	return p.color
}

func (p *CosmeticPen) Width() int {
	return 1
}

type GeometricPen struct {
	hPen  win.HPEN
	style PenStyle
	brush Brush
	width int
}

func NewGeometricPen(style PenStyle, width int, brush Brush) (*GeometricPen, error) {
	if brush == nil {
		return nil, newError("brush cannot be nil")
	}

	style |= win.PS_GEOMETRIC

	hPen := win.ExtCreatePen(uint32(style), uint32(width), brush.logbrush(), 0, nil)
	if hPen == 0 {
		return nil, newError("ExtCreatePen failed")
	}

	return &GeometricPen{
		hPen:  hPen,
		style: style,
		width: width,
		brush: brush,
	}, nil
}

func (p *GeometricPen) Dispose() {
	if p.hPen != 0 {
		win.DeleteObject(win.HGDIOBJ(p.hPen))

		p.hPen = 0
	}
}

func (p *GeometricPen) handle() win.HPEN {
	return p.hPen
}

func (p *GeometricPen) Style() PenStyle {
	return p.style
}

func (p *GeometricPen) Width() int {
	return p.width
}

func (p *GeometricPen) Brush() Brush {
	return p.brush
}
