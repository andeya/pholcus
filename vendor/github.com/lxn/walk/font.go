// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"syscall"
)

import (
	"github.com/lxn/win"
)

type FontStyle byte

// Font style flags
const (
	FontBold      FontStyle = 0x01
	FontItalic    FontStyle = 0x02
	FontUnderline FontStyle = 0x04
	FontStrikeOut FontStyle = 0x08
)

var (
	screenDPIX  int
	screenDPIY  int
	defaultFont *Font
	knownFonts  = make(map[fontInfo]*Font)
)

func init() {
	// Retrieve screen DPI
	hDC := win.GetDC(0)
	defer win.ReleaseDC(0, hDC)
	screenDPIX = int(win.GetDeviceCaps(hDC, win.LOGPIXELSX))
	screenDPIY = int(win.GetDeviceCaps(hDC, win.LOGPIXELSY))

	// Initialize default font
	var err error
	if defaultFont, err = NewFont("MS Shell Dlg 2", 8, 0); err != nil {
		panic("failed to create default font")
	}
}

type fontInfo struct {
	family    string
	pointSize int
	style     FontStyle
}

// Font represents a typographic typeface that is used for text drawing
// operations and on many GUI widgets.
type Font struct {
	dpi2hFont map[int]win.HFONT
	family    string
	pointSize int
	style     FontStyle
}

// NewFont returns a new Font with the specified attributes.
func NewFont(family string, pointSize int, style FontStyle) (*Font, error) {
	if style > FontBold|FontItalic|FontUnderline|FontStrikeOut {
		return nil, newError("invalid style")
	}

	fi := fontInfo{
		family:    family,
		pointSize: pointSize,
		style:     style,
	}

	if font, ok := knownFonts[fi]; ok {
		return font, nil
	}

	font := &Font{
		family:    family,
		pointSize: pointSize,
		style:     style,
		dpi2hFont: make(map[int]win.HFONT),
	}

	knownFonts[fi] = font

	return font, nil
}

func newFontFromLOGFONT(lf *win.LOGFONT, dpi int) (*Font, error) {
	if lf == nil {
		return nil, newError("lf cannot be nil")
	}

	family := win.UTF16PtrToString(&lf.LfFaceName[0])
	pointSize := int(win.MulDiv(lf.LfHeight, 72, int32(dpi)))
	if pointSize < 0 {
		pointSize = -pointSize
	}

	var style FontStyle
	if lf.LfWeight > win.FW_NORMAL {
		style |= FontBold
	}
	if lf.LfItalic == win.TRUE {
		style |= FontItalic
	}
	if lf.LfUnderline == win.TRUE {
		style |= FontUnderline
	}
	if lf.LfStrikeOut == win.TRUE {
		style |= FontStrikeOut
	}

	return NewFont(family, pointSize, style)
}

func (f *Font) createForDPI(dpi int) win.HFONT {
	var lf win.LOGFONT

	lf.LfHeight = -win.MulDiv(int32(f.pointSize), int32(dpi), 72)
	if f.style&FontBold > 0 {
		lf.LfWeight = win.FW_BOLD
	} else {
		lf.LfWeight = win.FW_NORMAL
	}
	if f.style&FontItalic > 0 {
		lf.LfItalic = 1
	}
	if f.style&FontUnderline > 0 {
		lf.LfUnderline = 1
	}
	if f.style&FontStrikeOut > 0 {
		lf.LfStrikeOut = 1
	}
	lf.LfCharSet = win.DEFAULT_CHARSET
	lf.LfOutPrecision = win.OUT_TT_PRECIS
	lf.LfClipPrecision = win.CLIP_DEFAULT_PRECIS
	lf.LfQuality = win.CLEARTYPE_QUALITY
	lf.LfPitchAndFamily = win.VARIABLE_PITCH | win.FF_SWISS

	src := syscall.StringToUTF16(f.family)
	dest := lf.LfFaceName[:]
	copy(dest, src)

	return win.CreateFontIndirect(&lf)
}

// Bold returns if text drawn using the Font appears with
// greater weight than normal.
func (f *Font) Bold() bool {
	return f.style&FontBold > 0
}

// Dispose releases the os resources that were allocated for the Font.
//
// The Font can no longer be used for drawing operations or with GUI widgets
// after calling this method. It is safe to call Dispose multiple times.
func (f *Font) Dispose() {
	for dpi, hFont := range f.dpi2hFont {
		if dpi != 0 {
			win.DeleteObject(win.HGDIOBJ(hFont))
		}

		delete(f.dpi2hFont, dpi)
	}
}

// Family returns the family name of the Font.
func (f *Font) Family() string {
	return f.family
}

// Italic returns if text drawn using the Font appears slanted.
func (f *Font) Italic() bool {
	return f.style&FontItalic > 0
}

// HandleForDPI returns the os resource handle of the font for the specified
// DPI value.
//
// A value of 0 returns a HFONT suitable for the screen.
func (f *Font) handleForDPI(dpi int) win.HFONT {
	if len(f.dpi2hFont) == 0 {
		hFont := f.createForDPI(screenDPIY)
		if hFont == 0 {
			return 0
		}

		f.dpi2hFont[screenDPIY] = hFont
		f.dpi2hFont[0] = hFont // Make HFONT for screen easier accessible.
	}

	hFont := f.dpi2hFont[dpi]
	if hFont == 0 {
		hFont = f.createForDPI(dpi)
		f.dpi2hFont[dpi] = hFont
	}

	return hFont
}

// StrikeOut returns if text drawn using the Font appears striked out.
func (f *Font) StrikeOut() bool {
	return f.style&FontStrikeOut > 0
}

// Style returns the combination of style flags of the Font.
func (f *Font) Style() FontStyle {
	return f.style
}

// Underline returns if text drawn using the font appears underlined.
func (f *Font) Underline() bool {
	return f.style&FontUnderline > 0
}

// PointSize returns the size of the Font in point units.
func (f *Font) PointSize() int {
	return f.pointSize
}
