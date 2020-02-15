// Copyright 2010 The Walk Authorc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

// DrawText format flags
type DrawTextFormat uint

const (
	TextTop                  DrawTextFormat = win.DT_TOP
	TextLeft                 DrawTextFormat = win.DT_LEFT
	TextCenter               DrawTextFormat = win.DT_CENTER
	TextRight                DrawTextFormat = win.DT_RIGHT
	TextVCenter              DrawTextFormat = win.DT_VCENTER
	TextBottom               DrawTextFormat = win.DT_BOTTOM
	TextWordbreak            DrawTextFormat = win.DT_WORDBREAK
	TextSingleLine           DrawTextFormat = win.DT_SINGLELINE
	TextExpandTabs           DrawTextFormat = win.DT_EXPANDTABS
	TextTabstop              DrawTextFormat = win.DT_TABSTOP
	TextNoClip               DrawTextFormat = win.DT_NOCLIP
	TextExternalLeading      DrawTextFormat = win.DT_EXTERNALLEADING
	TextCalcRect             DrawTextFormat = win.DT_CALCRECT
	TextNoPrefix             DrawTextFormat = win.DT_NOPREFIX
	TextInternal             DrawTextFormat = win.DT_INTERNAL
	TextEditControl          DrawTextFormat = win.DT_EDITCONTROL
	TextPathEllipsis         DrawTextFormat = win.DT_PATH_ELLIPSIS
	TextEndEllipsis          DrawTextFormat = win.DT_END_ELLIPSIS
	TextModifyString         DrawTextFormat = win.DT_MODIFYSTRING
	TextRTLReading           DrawTextFormat = win.DT_RTLREADING
	TextWordEllipsis         DrawTextFormat = win.DT_WORD_ELLIPSIS
	TextNoFullWidthCharBreak DrawTextFormat = win.DT_NOFULLWIDTHCHARBREAK
	TextHidePrefix           DrawTextFormat = win.DT_HIDEPREFIX
	TextPrefixOnly           DrawTextFormat = win.DT_PREFIXONLY
)

var gM = syscall.StringToUTF16Ptr("gM")

type Canvas struct {
	hdc                 win.HDC
	window              Window
	dpix                int
	dpiy                int
	bitmap              *Bitmap
	recordingMetafile   *Metafile
	measureTextMetafile *Metafile
	doNotDispose        bool
}

func NewCanvasFromImage(image Image) (*Canvas, error) {
	switch img := image.(type) {
	case *Bitmap:
		hdc := win.CreateCompatibleDC(0)
		if hdc == 0 {
			return nil, newError("CreateCompatibleDC failed")
		}
		succeeded := false

		defer func() {
			if !succeeded {
				win.DeleteDC(hdc)
			}
		}()

		if win.SelectObject(hdc, win.HGDIOBJ(img.hBmp)) == 0 {
			return nil, newError("SelectObject failed")
		}

		succeeded = true

		return (&Canvas{hdc: hdc, bitmap: img, dpix: img.dpi, dpiy: img.dpi}).init()

	case *Metafile:
		c, err := newCanvasFromHDC(img.hdc)
		if err != nil {
			return nil, err
		}

		c.recordingMetafile = img

		return c, nil
	}

	return nil, newError("unsupported image type")
}

func newCanvasFromWindow(window Window) (*Canvas, error) {
	hdc := win.GetDC(window.Handle())
	if hdc == 0 {
		return nil, newError("GetDC failed")
	}

	return (&Canvas{hdc: hdc, window: window}).init()
}

func newCanvasFromHDC(hdc win.HDC) (*Canvas, error) {
	if hdc == 0 {
		return nil, newError("invalid hdc")
	}

	return (&Canvas{hdc: hdc, doNotDispose: true}).init()
}

func (c *Canvas) init() (*Canvas, error) {
	if c.dpix == 0 || c.dpiy == 0 {
		c.dpix = dpiForHDC(c.hdc)
		c.dpiy = c.dpix
	}

	if win.SetBkMode(c.hdc, win.TRANSPARENT) == 0 {
		return nil, newError("SetBkMode failed")
	}

	switch win.SetStretchBltMode(c.hdc, win.HALFTONE) {
	case 0, win.ERROR_INVALID_PARAMETER:
		return nil, newError("SetStretchBltMode failed")
	}

	if !win.SetBrushOrgEx(c.hdc, 0, 0, nil) {
		return nil, newError("SetBrushOrgEx failed")
	}

	return c, nil
}

func (c *Canvas) Dispose() {
	if !c.doNotDispose && c.hdc != 0 {
		if c.bitmap != nil {
			win.DeleteDC(c.hdc)
			c.bitmap.postProcess()
		} else {
			win.ReleaseDC(c.window.Handle(), c.hdc)
		}

		c.hdc = 0
	}

	if c.recordingMetafile != nil {
		c.recordingMetafile.ensureFinished()
		c.recordingMetafile = nil
	}

	if c.measureTextMetafile != nil {
		c.measureTextMetafile.Dispose()
		c.measureTextMetafile = nil
	}
}

func (c *Canvas) DPI() int {
	if c.window != nil {
		return c.window.DPI()
	}

	return c.dpix
}

func (c *Canvas) withGdiObj(handle win.HGDIOBJ, f func() error) error {
	oldHandle := win.SelectObject(c.hdc, handle)
	if oldHandle == 0 {
		return newError("SelectObject failed")
	}
	defer win.SelectObject(c.hdc, oldHandle)

	return f()
}

func (c *Canvas) withBrush(brush Brush, f func() error) error {
	return c.withGdiObj(win.HGDIOBJ(brush.handle()), f)
}

func (c *Canvas) withFontAndTextColor(font *Font, color Color, f func() error) error {
	return c.withGdiObj(win.HGDIOBJ(font.handleForDPI(c.DPI())), func() error {
		oldColor := win.SetTextColor(c.hdc, win.COLORREF(color))
		if oldColor == win.CLR_INVALID {
			return newError("SetTextColor failed")
		}
		defer func() {
			win.SetTextColor(c.hdc, oldColor)
		}()

		return f()
	})
}

func (c *Canvas) HDC() win.HDC {
	return c.hdc
}

func (c *Canvas) Bounds() Rectangle {
	return RectangleTo96DPI(Rectangle{
		Width:  int(win.GetDeviceCaps(c.hdc, win.HORZRES)),
		Height: int(win.GetDeviceCaps(c.hdc, win.VERTRES)),
	}, c.DPI())
}

func (c *Canvas) withPen(pen Pen, f func() error) error {
	return c.withGdiObj(win.HGDIOBJ(pen.handle()), f)
}

func (c *Canvas) withBrushAndPen(brush Brush, pen Pen, f func() error) error {
	return c.withBrush(brush, func() error {
		return c.withPen(pen, f)
	})
}

func (c *Canvas) ellipse(brush Brush, pen Pen, bounds Rectangle, sizeCorrection int) error {
	return c.withBrushAndPen(brush, pen, func() error {
		bounds = RectangleFrom96DPI(bounds, c.DPI())

		if !win.Ellipse(
			c.hdc,
			int32(bounds.X),
			int32(bounds.Y),
			int32(bounds.X+bounds.Width+sizeCorrection),
			int32(bounds.Y+bounds.Height+sizeCorrection)) {

			return newError("Ellipse failed")
		}

		return nil
	})
}

func (c *Canvas) DrawEllipse(pen Pen, bounds Rectangle) error {
	return c.ellipse(nullBrushSingleton, pen, bounds, 0)
}

func (c *Canvas) FillEllipse(brush Brush, bounds Rectangle) error {
	return c.ellipse(brush, nullPenSingleton, bounds, 1)
}

func (c *Canvas) DrawImage(image Image, location Point) error {
	if image == nil {
		return newError("image cannot be nil")
	}

	location = PointFrom96DPI(location, c.DPI())

	return image.draw(c.hdc, location)
}

func (c *Canvas) DrawImageStretched(image Image, bounds Rectangle) error {
	if image == nil {
		return newError("image cannot be nil")
	}

	bounds = RectangleFrom96DPI(bounds, c.DPI())

	if dsoc, ok := image.(interface {
		drawStretchedOnCanvas(canvas *Canvas, bounds Rectangle) error
	}); ok {
		return dsoc.drawStretchedOnCanvas(c, bounds)
	}

	return image.drawStretched(c.hdc, bounds)
}

func (c *Canvas) DrawBitmapWithOpacity(bmp *Bitmap, bounds Rectangle, opacity byte) error {
	if bmp == nil {
		return newError("bmp cannot be nil")
	}

	bounds = RectangleFrom96DPI(bounds, c.DPI())

	return bmp.alphaBlend(c.hdc, bounds, opacity)
}

func (c *Canvas) DrawBitmapPart(bmp *Bitmap, dst, src Rectangle) error {
	return c.DrawBitmapPartWithOpacity(bmp, dst, src, 0xff)
}

func (c *Canvas) DrawBitmapPartWithOpacity(bmp *Bitmap, dst, src Rectangle, opacity byte) error {
	if bmp == nil {
		return newError("bmp cannot be nil")
	}

	dst = RectangleFrom96DPI(dst, c.DPI())
	src = RectangleFrom96DPI(src, c.DPI())

	return bmp.alphaBlendPart(c.hdc, dst, src, opacity)
}

func (c *Canvas) DrawLine(pen Pen, from, to Point) error {
	dpi := c.DPI()

	from = PointFrom96DPI(from, dpi)
	if !win.MoveToEx(c.hdc, from.X, from.Y, nil) {
		return newError("MoveToEx failed")
	}

	return c.withPen(pen, func() error {
		to = PointFrom96DPI(to, dpi)
		if !win.LineTo(c.hdc, int32(to.X), int32(to.Y)) {
			return newError("LineTo failed")
		}

		return nil
	})
}

func (c *Canvas) DrawPolyline(pen Pen, points []Point) error {
	if len(points) < 1 {
		return nil
	}

	dpi := c.DPI()

	pts := make([]win.POINT, len(points))
	for i, p := range points {
		p = PointFrom96DPI(p, dpi)

		pts[i] = win.POINT{X: int32(p.X), Y: int32(p.Y)}
	}

	return c.withPen(pen, func() error {
		if !win.Polyline(c.hdc, unsafe.Pointer(&pts[0].X), int32(len(pts))) {
			return newError("Polyline failed")
		}

		return nil
	})
}

func (c *Canvas) rectangle(brush Brush, pen Pen, bounds Rectangle, sizeCorrection int) error {
	return c.rectanglePixels(brush, pen, RectangleFrom96DPI(bounds, c.DPI()), sizeCorrection)
}

func (c *Canvas) rectanglePixels(brush Brush, pen Pen, bounds Rectangle, sizeCorrection int) error {
	return c.withBrushAndPen(brush, pen, func() error {
		if !win.Rectangle_(
			c.hdc,
			int32(bounds.X),
			int32(bounds.Y),
			int32(bounds.X+bounds.Width+sizeCorrection),
			int32(bounds.Y+bounds.Height+sizeCorrection)) {

			return newError("Rectangle_ failed")
		}

		return nil
	})
}

func (c *Canvas) DrawRectangle(pen Pen, bounds Rectangle) error {
	return c.rectangle(nullBrushSingleton, pen, bounds, 0)
}

func (c *Canvas) FillRectangle(brush Brush, bounds Rectangle) error {
	return c.rectangle(brush, nullPenSingleton, bounds, 1)
}

func (c *Canvas) fillRectanglePixels(brush Brush, bounds Rectangle) error {
	return c.rectanglePixels(brush, nullPenSingleton, bounds, 1)
}

func (c *Canvas) roundedRectangle(brush Brush, pen Pen, bounds Rectangle, ellipseSize Size, sizeCorrection int) error {
	return c.withBrushAndPen(brush, pen, func() error {
		dpi := c.DPI()

		bounds = RectangleFrom96DPI(bounds, dpi)
		ellipseSize = SizeFrom96DPI(ellipseSize, dpi)

		if !win.RoundRect(
			c.hdc,
			int32(bounds.X),
			int32(bounds.Y),
			int32(bounds.X+bounds.Width+sizeCorrection),
			int32(bounds.Y+bounds.Height+sizeCorrection),
			int32(ellipseSize.Width),
			int32(ellipseSize.Height)) {

			return newError("RoundRect failed")
		}

		return nil
	})
}

func (c *Canvas) DrawRoundedRectangle(pen Pen, bounds Rectangle, ellipseSize Size) error {
	return c.roundedRectangle(nullBrushSingleton, pen, bounds, ellipseSize, 0)
}

func (c *Canvas) FillRoundedRectangle(brush Brush, bounds Rectangle, ellipseSize Size) error {
	return c.roundedRectangle(brush, nullPenSingleton, bounds, ellipseSize, 1)
}

func (c *Canvas) GradientFillRectangle(color1, color2 Color, orientation Orientation, bounds Rectangle) error {
	bounds = RectangleFrom96DPI(bounds, c.DPI())

	vertices := [2]win.TRIVERTEX{
		{
			X:     int32(bounds.X),
			Y:     int32(bounds.Y),
			Red:   uint16(color1.R()) * 256,
			Green: uint16(color1.G()) * 256,
			Blue:  uint16(color1.B()) * 256,
			Alpha: 0,
		}, {
			X:     int32(bounds.X + bounds.Width),
			Y:     int32(bounds.Y + bounds.Height),
			Red:   uint16(color2.R()) * 256,
			Green: uint16(color2.G()) * 256,
			Blue:  uint16(color2.B()) * 256,
			Alpha: 0,
		},
	}

	indices := win.GRADIENT_RECT{
		UpperLeft:  0,
		LowerRight: 1,
	}

	if !win.GradientFill(c.hdc, &vertices[0], 2, unsafe.Pointer(&indices), 1, uint32(orientation)) {
		return newError("GradientFill failed")
	}

	return nil
}

func (c *Canvas) DrawText(text string, font *Font, color Color, bounds Rectangle, format DrawTextFormat) error {
	return c.withFontAndTextColor(font, color, func() error {
		bounds = RectangleFrom96DPI(bounds, c.DPI())

		rect := bounds.toRECT()
		ret := win.DrawTextEx(
			c.hdc,
			syscall.StringToUTF16Ptr(text),
			-1,
			&rect,
			uint32(format)|win.DT_EDITCONTROL,
			nil)
		if ret == 0 {
			return newError("DrawTextEx failed")
		}

		return nil
	})
}

func (c *Canvas) fontHeight(font *Font) (height int, err error) {
	err = c.withFontAndTextColor(font, 0, func() error {
		var size win.SIZE
		if !win.GetTextExtentPoint32(c.hdc, gM, 2, &size) {
			return newError("GetTextExtentPoint32 failed")
		}

		height = int(size.CY)
		if height == 0 {
			return newError("invalid font height")
		}

		return nil
	})

	return
}

func (c *Canvas) measureTextForDPI(text string, font *Font, bounds Rectangle, format DrawTextFormat, dpi int) (boundsMeasured Rectangle, runesFitted int, err error) {
	// HACK: We don't want to actually draw on the Canvas here, but if we use
	// the DT_CALCRECT flag to avoid drawing, DRAWTEXTPARAMc.UiLengthDrawn will
	// not contain a useful value. To work around this, we create an in-memory
	// metafile and draw into that instead.
	if c.measureTextMetafile == nil {
		c.measureTextMetafile, err = NewMetafile(c)
		if err != nil {
			return
		}
	}

	hFont := win.HGDIOBJ(font.handleForDPI(dpi))
	oldHandle := win.SelectObject(c.measureTextMetafile.hdc, hFont)
	if oldHandle == 0 {
		err = newError("SelectObject failed")
		return
	}
	defer win.SelectObject(c.measureTextMetafile.hdc, oldHandle)

	rect := &win.RECT{
		int32(bounds.X),
		int32(bounds.Y),
		int32(bounds.X + bounds.Width),
		int32(bounds.Y + bounds.Height),
	}
	var params win.DRAWTEXTPARAMS
	params.CbSize = uint32(unsafe.Sizeof(params))

	strPtr := syscall.StringToUTF16Ptr(text)
	dtfmt := uint32(format) | win.DT_EDITCONTROL | win.DT_WORDBREAK | win.DT_CALCRECT

	height := win.DrawTextEx(
		c.measureTextMetafile.hdc, strPtr, -1, rect, dtfmt, &params)
	if height == 0 {
		err = newError("DrawTextEx failed")
		return
	}

	boundsMeasured = Rectangle{
		int(rect.Left),
		int(rect.Top),
		int(rect.Right - rect.Left),
		int(height),
	}
	runesFitted = int(params.UiLengthDrawn)

	return
}

func (c *Canvas) MeasureText(text string, font *Font, bounds Rectangle, format DrawTextFormat) (boundsMeasured Rectangle, runesFitted int, err error) {
	// HACK: We don't want to actually draw on the Canvas here, but if we use
	// the DT_CALCRECT flag to avoid drawing, DRAWTEXTPARAMc.UiLengthDrawn will
	// not contain a useful value. To work around this, we create an in-memory
	// metafile and draw into that instead.
	if c.measureTextMetafile == nil {
		c.measureTextMetafile, err = NewMetafile(c)
		if err != nil {
			return
		}
	}

	hFont := win.HGDIOBJ(font.handleForDPI(c.DPI()))
	oldHandle := win.SelectObject(c.measureTextMetafile.hdc, hFont)
	if oldHandle == 0 {
		err = newError("SelectObject failed")
		return
	}
	defer win.SelectObject(c.measureTextMetafile.hdc, oldHandle)

	bounds = RectangleFrom96DPI(bounds, c.DPI())

	rect := &win.RECT{
		int32(bounds.X),
		int32(bounds.Y),
		int32(bounds.X + bounds.Width),
		int32(bounds.Y + bounds.Height),
	}
	var params win.DRAWTEXTPARAMS
	params.CbSize = uint32(unsafe.Sizeof(params))

	strPtr := syscall.StringToUTF16Ptr(text)
	dtfmt := uint32(format) | win.DT_EDITCONTROL | win.DT_WORDBREAK

	height := win.DrawTextEx(
		c.measureTextMetafile.hdc, strPtr, -1, rect, dtfmt, &params)
	if height == 0 {
		err = newError("DrawTextEx failed")
		return
	}

	boundsMeasured = RectangleTo96DPI(Rectangle{
		int(rect.Left),
		int(rect.Top),
		int(rect.Right - rect.Left),
		int(height),
	}, c.DPI())
	runesFitted = int(params.UiLengthDrawn)

	return
}
