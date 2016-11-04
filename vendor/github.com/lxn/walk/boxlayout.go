// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"sort"
)

import (
	"github.com/lxn/win"
)

type Orientation byte

const (
	Horizontal Orientation = iota
	Vertical
)

type BoxLayout struct {
	container          Container
	margins            Margins
	spacing            int
	orientation        Orientation
	hwnd2StretchFactor map[win.HWND]int
	resetNeeded        bool
}

func newBoxLayout(orientation Orientation) *BoxLayout {
	return &BoxLayout{
		orientation:        orientation,
		hwnd2StretchFactor: make(map[win.HWND]int),
	}
}

func NewHBoxLayout() *BoxLayout {
	return newBoxLayout(Horizontal)
}

func NewVBoxLayout() *BoxLayout {
	return newBoxLayout(Vertical)
}

func (l *BoxLayout) Container() Container {
	return l.container
}

func (l *BoxLayout) SetContainer(value Container) {
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

func (l *BoxLayout) Margins() Margins {
	return l.margins
}

func (l *BoxLayout) SetMargins(value Margins) error {
	if value.HNear < 0 || value.VNear < 0 || value.HFar < 0 || value.VFar < 0 {
		return newError("margins must be positive")
	}

	l.margins = value

	l.Update(false)

	return nil
}

func (l *BoxLayout) Orientation() Orientation {
	return l.orientation
}

func (l *BoxLayout) SetOrientation(value Orientation) error {
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

func (l *BoxLayout) Spacing() int {
	return l.spacing
}

func (l *BoxLayout) SetSpacing(value int) error {
	if value != l.spacing {
		if value < 0 {
			return newError("spacing cannot be negative")
		}

		l.spacing = value

		l.Update(false)
	}

	return nil
}

func (l *BoxLayout) StretchFactor(widget Widget) int {
	if factor, ok := l.hwnd2StretchFactor[widget.Handle()]; ok {
		return factor
	}

	return 1
}

func (l *BoxLayout) SetStretchFactor(widget Widget, factor int) error {
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

func (l *BoxLayout) cleanupStretchFactors() {
	widgets := l.container.Children()

	for handle, _ := range l.hwnd2StretchFactor {
		if !widgets.containsHandle(handle) {
			delete(l.hwnd2StretchFactor, handle)
		}
	}
}

type widgetInfo struct {
	index   int
	minSize int
	maxSize int
	stretch int
	greedy  bool
	widget  Widget
}

type widgetInfoList []widgetInfo

func (l widgetInfoList) Len() int {
	return len(l)
}

func (l widgetInfoList) Less(i, j int) bool {
	_, iIsSpacer := l[i].widget.(*Spacer)
	_, jIsSpacer := l[j].widget.(*Spacer)

	if l[i].greedy == l[j].greedy {
		if iIsSpacer == jIsSpacer {
			minDiff := l[i].minSize - l[j].minSize

			if minDiff == 0 {
				return l[i].maxSize/l[i].stretch < l[j].maxSize/l[j].stretch
			}

			return minDiff > 0
		}

		return jIsSpacer
	}

	return l[i].greedy
}

func (l widgetInfoList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l *BoxLayout) widgets() []Widget {
	children := l.container.Children()
	widgets := make([]Widget, 0, children.Len())

	for i := 0; i < cap(widgets); i++ {
		widget := children.At(i)

		if !shouldLayoutWidget(widget) {
			continue
		}

		ps := widget.SizeHint()
		if ps.Width == 0 && ps.Height == 0 && widget.LayoutFlags() == 0 {
			continue
		}

		widgets = append(widgets, widget)
	}

	return widgets
}

func (l *BoxLayout) LayoutFlags() LayoutFlags {
	if l.container == nil {
		return 0
	}

	var flags LayoutFlags
	var hasNonShrinkableHorz bool
	var hasNonShrinkableVert bool

	children := l.container.Children()
	count := children.Len()
	if count == 0 {
		return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert
	} else {
		for i := 0; i < count; i++ {
			widget := children.At(i)

			if !shouldLayoutWidget(widget) {
				continue
			}

			f := widget.LayoutFlags()
			flags |= f
			if f&ShrinkableHorz == 0 {
				hasNonShrinkableHorz = true
			}
			if f&ShrinkableVert == 0 {
				hasNonShrinkableVert = true
			}
		}
	}

	if l.orientation == Horizontal {
		flags |= GrowableHorz

		if hasNonShrinkableVert {
			flags &^= ShrinkableVert
		}
	} else {
		flags |= GrowableVert

		if hasNonShrinkableHorz {
			flags &^= ShrinkableHorz
		}
	}

	return flags
}

func (l *BoxLayout) MinSize() Size {
	if l.container == nil {
		return Size{}
	}

	widgets := l.widgets()
	var s Size

	for _, widget := range widgets {
		min := minSizeEffective(widget)

		if l.orientation == Horizontal {
			s.Width += min.Width
			s.Height = maxi(s.Height, min.Height)
		} else {
			s.Height += min.Height
			s.Width = maxi(s.Width, min.Width)
		}
	}

	if l.orientation == Horizontal {
		s.Width += l.spacing * (len(widgets) - 1)
		s.Width += l.margins.HNear + l.margins.HFar
		s.Height += l.margins.VNear + l.margins.VFar
	} else {
		s.Height += l.spacing * (len(widgets) - 1)
		s.Height += l.margins.VNear + l.margins.VFar
		s.Width += l.margins.HNear + l.margins.HFar
	}

	return s
}

func (l *BoxLayout) Update(reset bool) error {
	if l.container == nil {
		return nil
	}

	if reset {
		l.resetNeeded = true
	}

	if l.container.Suspended() {
		return nil
	}

	if l.resetNeeded {
		l.resetNeeded = false

		// Make GC happy.
		l.cleanupStretchFactors()
	}

	// Begin by finding out which widgets we care about.
	widgets := l.widgets()

	// Prepare some useful data.
	var greedyNonSpacerCount int
	var greedySpacerCount int
	var stretchFactorsTotal [3]int
	stretchFactors := make([]int, len(widgets))
	var minSizesRemaining int
	minSizes := make([]int, len(widgets))
	maxSizes := make([]int, len(widgets))
	sizes := make([]int, len(widgets))
	prefSizes2 := make([]int, len(widgets))
	growable2 := make([]bool, len(widgets))
	sortedWidgetInfo := widgetInfoList(make([]widgetInfo, len(widgets)))

	for i, widget := range widgets {
		sf := l.hwnd2StretchFactor[widget.Handle()]
		if sf == 0 {
			sf = 1
		}
		stretchFactors[i] = sf

		flags := widget.LayoutFlags()

		min := widget.MinSize()
		max := widget.MaxSize()
		minHint := widget.MinSizeHint()
		pref := widget.SizeHint()

		if l.orientation == Horizontal {
			growable2[i] = flags&GrowableVert > 0

			minSizes[i] = maxi(min.Width, minHint.Width)

			if max.Width > 0 {
				maxSizes[i] = max.Width
			} else if pref.Width > 0 && flags&GrowableHorz == 0 {
				maxSizes[i] = pref.Width
			} else {
				maxSizes[i] = 32768
			}

			prefSizes2[i] = pref.Height

			sortedWidgetInfo[i].greedy = flags&GreedyHorz > 0
		} else {
			growable2[i] = flags&GrowableHorz > 0

			minSizes[i] = maxi(min.Height, minHint.Height)

			if max.Height > 0 {
				maxSizes[i] = max.Height
			} else if pref.Height > 0 && flags&GrowableVert == 0 {
				maxSizes[i] = pref.Height
			} else {
				maxSizes[i] = 32768
			}

			prefSizes2[i] = pref.Width

			sortedWidgetInfo[i].greedy = flags&GreedyVert > 0
		}

		sortedWidgetInfo[i].index = i
		sortedWidgetInfo[i].minSize = minSizes[i]
		sortedWidgetInfo[i].maxSize = maxSizes[i]
		sortedWidgetInfo[i].stretch = sf
		sortedWidgetInfo[i].widget = widget

		minSizesRemaining += minSizes[i]

		if sortedWidgetInfo[i].greedy {
			if _, isSpacer := widget.(*Spacer); !isSpacer {
				greedyNonSpacerCount++
				stretchFactorsTotal[0] += sf
			} else {
				greedySpacerCount++
				stretchFactorsTotal[1] += sf
			}
		} else {
			stretchFactorsTotal[2] += sf
		}
	}

	sort.Stable(sortedWidgetInfo)

	cb := l.container.ClientBounds()
	var start1, start2, space1, space2 int
	if l.orientation == Horizontal {
		start1 = cb.X + l.margins.HNear
		start2 = cb.Y + l.margins.VNear
		space1 = cb.Width - l.margins.HNear - l.margins.HFar
		space2 = cb.Height - l.margins.VNear - l.margins.VFar
	} else {
		start1 = cb.Y + l.margins.VNear
		start2 = cb.X + l.margins.HNear
		space1 = cb.Height - l.margins.VNear - l.margins.VFar
		space2 = cb.Width - l.margins.HNear - l.margins.HFar
	}

	// Now calculate widget primary axis sizes.
	spacingRemaining := l.spacing * (len(widgets) - 1)

	offsets := [3]int{0, greedyNonSpacerCount, greedyNonSpacerCount + greedySpacerCount}
	counts := [3]int{greedyNonSpacerCount, greedySpacerCount, len(widgets) - greedyNonSpacerCount - greedySpacerCount}

	for i := 0; i < 3; i++ {
		stretchFactorsRemaining := stretchFactorsTotal[i]

		for j := 0; j < counts[i]; j++ {
			info := sortedWidgetInfo[offsets[i]+j]
			k := info.index

			stretch := stretchFactors[k]
			min := info.minSize
			max := info.maxSize
			size := min

			if min < max {
				excessSpace := float64(space1 - minSizesRemaining - spacingRemaining)
				size += int(excessSpace * float64(stretch) / float64(stretchFactorsRemaining))
				if size < min {
					size = min
				} else if size > max {
					size = max
				}
			}

			sizes[k] = size

			minSizesRemaining -= min
			stretchFactorsRemaining -= stretch
			space1 -= (size + l.spacing)
			spacingRemaining -= l.spacing
		}
	}

	// Finally position widgets.
	hdwp := win.BeginDeferWindowPos(int32(len(widgets)))
	if hdwp == 0 {
		return lastError("BeginDeferWindowPos")
	}

	excessTotal := space1 - minSizesRemaining - spacingRemaining
	excessShare := excessTotal / (len(widgets) + 1)
	p1 := start1
	for i, widget := range widgets {
		p1 += excessShare
		s1 := sizes[i]

		var s2 int
		if growable2[i] {
			s2 = space2
		} else {
			s2 = prefSizes2[i]
		}

		p2 := start2 + (space2-s2)/2

		var x, y, w, h int
		if l.orientation == Horizontal {
			x, y, w, h = p1, p2, s1, s2
		} else {
			x, y, w, h = p2, p1, s2, s1
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

		p1 += s1 + l.spacing
	}

	if !win.EndDeferWindowPos(hdwp) {
		return lastError("EndDeferWindowPos")
	}

	return nil
}
