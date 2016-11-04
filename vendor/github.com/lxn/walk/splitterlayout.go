// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type splitterLayout struct {
	container          Container
	orientation        Orientation
	fractions          []float64
	hwnd2StretchFactor map[win.HWND]int
	resetNeeded        bool
}

func newSplitterLayout(orientation Orientation) *splitterLayout {
	return &splitterLayout{
		orientation:        orientation,
		hwnd2StretchFactor: make(map[win.HWND]int),
	}
}

func (l *splitterLayout) Container() Container {
	return l.container
}

func (l *splitterLayout) SetContainer(value Container) {
	if value != l.container {
		if l.container != nil {
			l.container.SetLayout(nil)
		}

		l.container = value

		if value != nil && value.Layout() != Layout(l) {
			value.SetLayout(l)

			l.Update(true)
		}
	}
}

func (l *splitterLayout) Margins() Margins {
	return Margins{}
}

func (l *splitterLayout) SetMargins(value Margins) error {
	return newError("not supported")
}

func (l *splitterLayout) Spacing() int {
	return 0
}

func (l *splitterLayout) SetSpacing(value int) error {
	return newError("not supported")
}

func (l *splitterLayout) Orientation() Orientation {
	return l.orientation
}

func (l *splitterLayout) SetOrientation(value Orientation) error {
	if value != l.orientation {
		switch value {
		case Horizontal, Vertical:

		default:
			return newError("invalid Orientation value")
		}

		l.orientation = value

		l.Update(false)
	}

	return nil
}

func (l *splitterLayout) Fractions() []float64 {
	return l.fractions
}

func (l *splitterLayout) SetFractions(fractions []float64) error {
	l.fractions = fractions

	return l.Update(false)
}

func (l *splitterLayout) StretchFactor(widget Widget) int {
	if factor, ok := l.hwnd2StretchFactor[widget.Handle()]; ok {
		return factor
	}

	return 1
}

func (l *splitterLayout) SetStretchFactor(widget Widget, factor int) error {
	if factor != l.StretchFactor(widget) {
		if l.container == nil {
			return newError("container required")
		}

		handle := widget.Handle()

		if !l.container.Children().containsHandle(handle) {
			return newError("unknown widget")
		}
		if factor < 1 {
			return newError("factor must be >= 1")
		}

		l.hwnd2StretchFactor[handle] = factor

		l.Update(false)
	}

	return nil
}

func (l *splitterLayout) cleanupStretchFactors() {
	widgets := l.container.Children()

	for handle, _ := range l.hwnd2StretchFactor {
		if !widgets.containsHandle(handle) {
			delete(l.hwnd2StretchFactor, handle)
		}
	}
}

func (l *splitterLayout) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (l *splitterLayout) MinSize() Size {
	var s Size

	for _, widget := range l.container.Children().items {
		cur := minSizeEffective(widget)

		if l.orientation == Horizontal {
			s.Width += cur.Width
			s.Height = maxi(s.Height, cur.Height)
		} else {
			s.Height += cur.Height
			s.Width = maxi(s.Width, cur.Width)
		}
	}

	return s
}

func (l *splitterLayout) spaceForRegularWidgets() int {
	splitter := l.container.(*Splitter)
	cb := splitter.ClientBounds().Size()

	var space int
	if l.orientation == Horizontal {
		space = cb.Width
	} else {
		space = cb.Height
	}

	return space - (splitter.Children().Len()/2)*splitter.handleWidth
}

func (l *splitterLayout) reset() {
	l.cleanupStretchFactors()

	children := l.container.Children()
	regularCount := children.Len()/2 + children.Len()%2

	if cap(l.fractions) < regularCount {
		temp := make([]float64, regularCount)
		copy(temp, l.fractions)
		l.fractions = temp
	}

	l.fractions = l.fractions[:regularCount]

	if regularCount == 0 {
		return
	}

	stretchTotal := 0
	for i := children.Len() - 1; i >= 0; i-- {
		if i%2 == 1 {
			continue
		}

		child := children.At(i)

		stretchTotal += l.StretchFactor(child)
	}

	j := len(l.fractions) - 1
	for i := children.Len() - 1; i >= 0; i-- {
		if i%2 == 1 {
			continue
		}

		child := children.At(i)

		l.fractions[j] = float64(l.StretchFactor(child)) / float64(stretchTotal)
		j--
	}
}

func (l *splitterLayout) Update(reset bool) error {
	if l.container == nil {
		return newError("container required")
	}

	if reset {
		l.resetNeeded = true
	}

	if l.container.Suspended() {
		return nil
	}

	if l.resetNeeded {
		l.resetNeeded = false

		l.reset()
	}

	widgets := l.container.Children().items
	splitter := l.container.(*Splitter)
	handleWidth := splitter.HandleWidth()
	sizes := make([]int, len(widgets))
	cb := splitter.ClientBounds()
	space1 := l.spaceForRegularWidgets()

	var space2 int
	if l.orientation == Horizontal {
		space2 = cb.Height
	} else {
		space2 = cb.Width
	}

	for i := range widgets {
		j := i/2 + i%2

		if i%2 == 0 {
			sizes[i] = int(float64(space1) * l.fractions[j])
		} else {
			sizes[i] = handleWidth
		}
	}

	hdwp := win.BeginDeferWindowPos(int32(len(widgets)))
	if hdwp == 0 {
		return lastError("BeginDeferWindowPos")
	}

	p1 := 0
	for i, widget := range widgets {
		var s1 int
		if i == len(widgets)-1 {
			s1 = space1 + len(widgets)/2*handleWidth - p1
		} else {
			s1 = sizes[i]
		}

		var x, y, w, h int
		if l.orientation == Horizontal {
			x, y, w, h = p1, 0, s1, space2
		} else {
			x, y, w, h = 0, p1, space2, s1
		}

		if hdwp = win.DeferWindowPos(
			hdwp,
			widget.Handle(),
			0,
			int32(x),
			int32(y),
			int32(w),
			int32(h),
			win.SWP_NOACTIVATE|win.SWP_NOOWNERZORDER|win.SWP_NOZORDER); hdwp == 0 {

			return lastError("DeferWindowPos")
		}

		p1 += s1
	}

	if !win.EndDeferWindowPos(hdwp) {
		return lastError("EndDeferWindowPos")
	}

	return nil
}
