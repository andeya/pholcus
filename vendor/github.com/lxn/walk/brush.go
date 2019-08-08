// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"unsafe"

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

type SystemColor int

const (
	SysColor3DDkShadow              SystemColor = win.COLOR_3DDKSHADOW
	SysColor3DFace                  SystemColor = win.COLOR_3DFACE
	SysColor3DHighlight             SystemColor = win.COLOR_3DHIGHLIGHT
	SysColor3DLight                 SystemColor = win.COLOR_3DLIGHT
	SysColor3DShadow                SystemColor = win.COLOR_3DSHADOW
	SysColorActiveBorder            SystemColor = win.COLOR_ACTIVEBORDER
	SysColorActiveCaption           SystemColor = win.COLOR_ACTIVECAPTION
	SysColorAppWorkspace            SystemColor = win.COLOR_APPWORKSPACE
	SysColorBackground              SystemColor = win.COLOR_BACKGROUND
	SysColorDesktop                 SystemColor = win.COLOR_DESKTOP
	SysColorBtnFace                 SystemColor = win.COLOR_BTNFACE
	SysColorBtnHighlight            SystemColor = win.COLOR_BTNHIGHLIGHT
	SysColorBtnShadow               SystemColor = win.COLOR_BTNSHADOW
	SysColorBtnText                 SystemColor = win.COLOR_BTNTEXT
	SysColorCaptionText             SystemColor = win.COLOR_CAPTIONTEXT
	SysColorGrayText                SystemColor = win.COLOR_GRAYTEXT
	SysColorHighlight               SystemColor = win.COLOR_HIGHLIGHT
	SysColorHighlightText           SystemColor = win.COLOR_HIGHLIGHTTEXT
	SysColorInactiveBorder          SystemColor = win.COLOR_INACTIVEBORDER
	SysColorInactiveCaption         SystemColor = win.COLOR_INACTIVECAPTION
	SysColorInactiveCaptionText     SystemColor = win.COLOR_INACTIVECAPTIONTEXT
	SysColorInfoBk                  SystemColor = win.COLOR_INFOBK
	SysColorInfoText                SystemColor = win.COLOR_INFOTEXT
	SysColorMenu                    SystemColor = win.COLOR_MENU
	SysColorMenuText                SystemColor = win.COLOR_MENUTEXT
	SysColorScrollBar               SystemColor = win.COLOR_SCROLLBAR
	SysColorWindow                  SystemColor = win.COLOR_WINDOW
	SysColorWindowFrame             SystemColor = win.COLOR_WINDOWFRAME
	SysColorWindowText              SystemColor = win.COLOR_WINDOWTEXT
	SysColorHotLight                SystemColor = win.COLOR_HOTLIGHT
	SysColorGradientActiveCaption   SystemColor = win.COLOR_GRADIENTACTIVECAPTION
	SysColorGradientInactiveCaption SystemColor = win.COLOR_GRADIENTINACTIVECAPTION
)

type Brush interface {
	Dispose()
	handle() win.HBRUSH
	logbrush() *win.LOGBRUSH
	attachWindow(wb *WindowBase)
	detachWindow(wb *WindowBase)
	simple() bool
}

type perWindowBrush interface {
	Brush
	delegateForWindow(wb *WindowBase) Brush
}

type windowBrushInfo struct {
	SizeChangedHandle int
	Delegate          *BitmapBrush
}

type brushBase struct {
	hBrush  win.HBRUSH
	wb2info map[*WindowBase]*windowBrushInfo
}

func (bb *brushBase) Dispose() {
	if bb.hBrush != 0 {
		win.DeleteObject(win.HGDIOBJ(bb.hBrush))

		bb.hBrush = 0
	}
}

func (bb *brushBase) handle() win.HBRUSH {
	return bb.hBrush
}

func (bb *brushBase) attachWindow(wb *WindowBase) {
	if wb == nil {
		return
	}

	if bb.wb2info == nil {
		bb.wb2info = make(map[*WindowBase]*windowBrushInfo)
	}

	bb.wb2info[wb] = nil
}

func (bb *brushBase) detachWindow(wb *WindowBase) {
	if bb.wb2info == nil || wb == nil {
		return
	}

	delete(bb.wb2info, wb)

	if len(bb.wb2info) == 0 {
		bb.Dispose()
	}
}

type nullBrush struct {
	brushBase
}

func newNullBrush() *nullBrush {
	lb := &win.LOGBRUSH{LbStyle: win.BS_NULL}

	hBrush := win.CreateBrushIndirect(lb)
	if hBrush == 0 {
		panic("failed to create null brush")
	}

	return &nullBrush{brushBase: brushBase{hBrush: hBrush}}
}

func (b *nullBrush) Dispose() {
	if b == nullBrushSingleton {
		return
	}

	b.brushBase.Dispose()
}

func (*nullBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{LbStyle: win.BS_NULL}
}

func (*nullBrush) simple() bool {
	return true
}

var nullBrushSingleton Brush = newNullBrush()

func NullBrush() Brush {
	return nullBrushSingleton
}

type SystemColorBrush struct {
	brushBase
	sysColor SystemColor
}

var sysColorBtnFaceBrush, _ = NewSystemColorBrush(SysColorBtnFace)

func NewSystemColorBrush(sysColor SystemColor) (*SystemColorBrush, error) {
	hBrush := win.GetSysColorBrush(int(sysColor))
	if hBrush == 0 {
		return nil, newError("GetSysColorBrush failed")
	}

	return &SystemColorBrush{brushBase: brushBase{hBrush: hBrush}, sysColor: sysColor}, nil
}

func (b *SystemColorBrush) Color() Color {
	return Color(win.GetSysColor(int(b.sysColor)))
}

func (b *SystemColorBrush) SystemColor() SystemColor {
	return b.sysColor
}

func (*SystemColorBrush) Dispose() {
	// nop
}

func (b *SystemColorBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{
		LbStyle: win.BS_SOLID,
		LbColor: win.COLORREF(win.GetSysColor(int(b.sysColor))),
	}
}

func (*SystemColorBrush) simple() bool {
	return true
}

var sysColorBtnFaceBrushSingleton, _ = NewSystemColorBrush(SysColorBtnFace)

type SolidColorBrush struct {
	brushBase
	color Color
}

func NewSolidColorBrush(color Color) (*SolidColorBrush, error) {
	lb := &win.LOGBRUSH{LbStyle: win.BS_SOLID, LbColor: win.COLORREF(color)}

	hBrush := win.CreateBrushIndirect(lb)
	if hBrush == 0 {
		return nil, newError("CreateBrushIndirect failed")
	}

	return &SolidColorBrush{brushBase: brushBase{hBrush: hBrush}, color: color}, nil
}

func (b *SolidColorBrush) Color() Color {
	return b.color
}

func (b *SolidColorBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{LbStyle: win.BS_SOLID, LbColor: win.COLORREF(b.color)}
}

func (*SolidColorBrush) simple() bool {
	return true
}

type HatchBrush struct {
	brushBase
	color Color
	style HatchStyle
}

func NewHatchBrush(color Color, style HatchStyle) (*HatchBrush, error) {
	lb := &win.LOGBRUSH{LbStyle: win.BS_HATCHED, LbColor: win.COLORREF(color), LbHatch: uintptr(style)}

	hBrush := win.CreateBrushIndirect(lb)
	if hBrush == 0 {
		return nil, newError("CreateBrushIndirect failed")
	}

	return &HatchBrush{brushBase: brushBase{hBrush: hBrush}, color: color, style: style}, nil
}

func (b *HatchBrush) Color() Color {
	return b.color
}

func (b *HatchBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{LbStyle: win.BS_HATCHED, LbColor: win.COLORREF(b.color), LbHatch: uintptr(b.style)}
}

func (b *HatchBrush) Style() HatchStyle {
	return b.style
}

func (b *HatchBrush) simple() bool {
	return false
}

type BitmapBrush struct {
	brushBase
	bitmap *Bitmap
}

func NewBitmapBrush(bitmap *Bitmap) (*BitmapBrush, error) {
	if bitmap == nil {
		return nil, newError("bitmap cannot be nil")
	}

	hBrush := win.CreatePatternBrush(bitmap.hBmp)
	if hBrush == 0 {
		return nil, newError("CreatePatternBrush failed")
	}

	return &BitmapBrush{brushBase: brushBase{hBrush: hBrush}, bitmap: bitmap}, nil
}

func (b *BitmapBrush) logbrush() *win.LOGBRUSH {
	return &win.LOGBRUSH{LbStyle: win.BS_DIBPATTERN, LbColor: win.DIB_RGB_COLORS, LbHatch: uintptr(b.bitmap.hPackedDIB)}
}

func (b *BitmapBrush) Bitmap() *Bitmap {
	return b.bitmap
}

func (b *BitmapBrush) simple() bool {
	return false
}

type GradientVertex struct {
	X     float64
	Y     float64
	Color Color
}

type GradientTriangle struct {
	Vertex1 int
	Vertex2 int
	Vertex3 int
}

type GradientBrush struct {
	brushBase
	mainDelegate *BitmapBrush
	vertexes     []GradientVertex
	triangles    []GradientTriangle
	absolute     bool
	vertical     bool
}

func NewGradientBrush(vertexes []GradientVertex, triangles []GradientTriangle) (*GradientBrush, error) {
	if len(vertexes) < 3 {
		return nil, newError("at least 3 vertexes are required")
	}

	if len(triangles) < 1 {
		return nil, newError("at least 1 triangle is required")
	}

	var size Size
	for _, v := range vertexes {
		size = maxSize(size, Size{int(v.X), int(v.Y)})
	}

	gb := &GradientBrush{vertexes: vertexes, triangles: triangles, absolute: size.Width > 1 || size.Height > 1}

	if gb.absolute {
		bb, err := gb.create(size)
		if err != nil {
			return nil, err
		}

		gb.mainDelegate = bb
		gb.hBrush = bb.hBrush
	}

	return gb, nil
}

func (b *GradientBrush) logbrush() *win.LOGBRUSH {
	if b.mainDelegate == nil {
		return nil
	}

	return b.mainDelegate.logbrush()
}

func (*GradientBrush) simple() bool {
	return false
}

func (b *GradientBrush) create(size Size) (*BitmapBrush, error) {
	var disposables Disposables
	defer disposables.Treat()

	bitmap, err := NewBitmap(size)
	if err != nil {
		return nil, err
	}
	disposables.Add(bitmap)

	canvas, err := NewCanvasFromImage(bitmap)
	if err != nil {
		return nil, err
	}
	defer canvas.Dispose()

	var scaleX, scaleY float64
	if b.absolute {
		scaleX, scaleY = 1, 1
	} else {
		scaleX, scaleY = float64(size.Width), float64(size.Height)
	}

	vertexes := make([]win.TRIVERTEX, len(b.vertexes))
	for i, src := range b.vertexes {
		dst := &vertexes[i]

		dst.X = int32(src.X * scaleX)
		dst.Y = int32(src.Y * scaleY)
		dst.Red = uint16(src.Color.R()) * 256
		dst.Green = uint16(src.Color.G()) * 256
		dst.Blue = uint16(src.Color.B()) * 256
	}

	triangles := make([]win.GRADIENT_TRIANGLE, len(b.triangles))
	for i, src := range b.triangles {
		dst := &triangles[i]

		dst.Vertex1 = uint32(src.Vertex1)
		dst.Vertex2 = uint32(src.Vertex2)
		dst.Vertex3 = uint32(src.Vertex3)
	}

	if !win.GradientFill(canvas.hdc, &vertexes[0], uint32(len(vertexes)), unsafe.Pointer(&triangles[0]), uint32(len(triangles)), win.GRADIENT_FILL_TRIANGLE) {
		return nil, newError("GradientFill failed")
	}

	disposables.Spare()

	return NewBitmapBrush(bitmap)
}

func (b *GradientBrush) attachWindow(wb *WindowBase) {
	b.brushBase.attachWindow(wb)

	if b.absolute {
		return
	}

	var info *windowBrushInfo

	update := func() {
		if bb, err := b.create(wb.window.ClientBoundsPixels().Size()); err == nil {
			if info.Delegate != nil {
				info.Delegate.bitmap.Dispose()
				info.Delegate.Dispose()
			}

			info.Delegate = bb

			wb.Invalidate()
		}
	}

	info = &windowBrushInfo{
		SizeChangedHandle: wb.SizeChanged().Attach(update),
	}

	update()

	b.wb2info[wb] = info
}

func (b *GradientBrush) detachWindow(wb *WindowBase) {
	if !b.absolute {
		if info, ok := b.wb2info[wb]; ok {
			if info.Delegate != nil {
				info.Delegate.bitmap.Dispose()
				info.Delegate.Dispose()
			}

			wb.SizeChanged().Detach(info.SizeChangedHandle)
		}
	}

	b.brushBase.detachWindow(wb)
}

func (b *GradientBrush) delegateForWindow(wb *WindowBase) Brush {
	if b.absolute {
		return b.mainDelegate
	}

	if info, ok := b.wb2info[wb]; ok && info.Delegate != nil {
		return info.Delegate
	}

	return nil
}
