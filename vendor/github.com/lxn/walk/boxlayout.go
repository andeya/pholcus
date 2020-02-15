// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"sort"

	"github.com/lxn/win"
)

type Orientation byte

const (
	Horizontal Orientation = iota
	Vertical
)

type BoxLayout struct {
	LayoutBase
	orientation        Orientation
	hwnd2StretchFactor map[win.HWND]int
}

func newBoxLayout(orientation Orientation) *BoxLayout {
	l := &BoxLayout{
		LayoutBase: LayoutBase{
			sizeAndDPI2MinSize: make(map[sizeAndDPI]Size),
			margins96dpi:       Margins{9, 9, 9, 9},
			spacing96dpi:       6,
		},
		orientation:        orientation,
		hwnd2StretchFactor: make(map[win.HWND]int),
	}
	l.layout = l

	return l
}

func NewHBoxLayout() *BoxLayout {
	return newBoxLayout(Horizontal)
}

func NewVBoxLayout() *BoxLayout {
	return newBoxLayout(Vertical)
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
	index      int
	minSize    int
	maxSize    int
	stretch    int
	greedy     bool
	widgetBase *WidgetBase
}

type widgetInfoList []widgetInfo

func (l widgetInfoList) Len() int {
	return len(l)
}

func (l widgetInfoList) Less(i, j int) bool {
	_, iIsSpacer := l[i].widgetBase.window.(*Spacer)
	_, jIsSpacer := l[j].widgetBase.window.(*Spacer)

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

func (l *BoxLayout) LayoutFlags() LayoutFlags {
	if l.container == nil {
		return 0
	}

	return boxLayoutFlags(l.orientation, l.container.Children())
}

func (l *BoxLayout) MinSize() Size {
	if l.container == nil {
		return Size{}
	}

	return l.MinSizeForSize(l.container.ClientBoundsPixels().Size())
}

func (l *BoxLayout) MinSizeForSize(size Size) Size {
	if l.container == nil {
		return Size{}
	}

	dpi := l.container.DPI()

	if min, ok := l.sizeAndDPI2MinSize[sizeAndDPI{size, dpi}]; ok {
		return min
	}

	bounds := Rectangle{Width: size.Width, Height: size.Height}

	items, err := boxLayoutItems(widgetsToLayout(l.Container().Children()), l.orientation, l.alignment, bounds, l.margins, l.spacing, l.hwnd2StretchFactor)
	if err != nil {
		return Size{}
	}

	s := Size{l.margins.HNear + l.margins.HFar, l.margins.VNear + l.margins.VFar}

	var maxSecondary int

	for _, item := range items {
		min := minSizeEffective(item.widget)

		if hfw, ok := item.widget.(HeightForWidther); ok {
			item.bounds.Height = hfw.HeightForWidth(item.bounds.Width)
		} else {
			item.bounds.Height = min.Height
		}
		item.bounds.Width = min.Width

		if l.orientation == Horizontal {
			maxSecondary = maxi(maxSecondary, item.bounds.Height)

			s.Width += item.bounds.Width
		} else {
			maxSecondary = maxi(maxSecondary, item.bounds.Width)

			s.Height += item.bounds.Height
		}
	}

	if l.orientation == Horizontal {
		s.Width += (len(items) - 1) * l.spacing
		s.Height += maxSecondary
	} else {
		s.Height += (len(items) - 1) * l.spacing
		s.Width += maxSecondary
	}

	if s.Width > 0 && s.Height > 0 {
		l.sizeAndDPI2MinSize[sizeAndDPI{size, dpi}] = s
	}

	return s
}

func (l *BoxLayout) Update(reset bool) error {
	if l.container == nil {
		return nil
	}

	l.sizeAndDPI2MinSize = make(map[sizeAndDPI]Size)

	if reset {
		l.resetNeeded = true
	}

	if l.container.Suspended() {
		return nil
	}

	if widget, ok := l.container.(Widget); ok {
		if form := widget.Form(); form == nil || form.Suspended() {
			return nil
		}
	}

	if !performingScheduledLayouts && scheduleLayout(l) {
		return nil
	}

	l.dirty = false

	if l.resetNeeded {
		l.resetNeeded = false

		l.cleanupStretchFactors()
	}

	ifContainerIsScrollViewDoCoolSpecialLayoutStuff(l)

	items, err := boxLayoutItems(widgetsToLayout(l.Container().Children()), l.orientation, l.alignment, l.container.ClientBoundsPixels(), l.margins, l.spacing, l.hwnd2StretchFactor)
	if err != nil {
		return err
	}

	if err := applyLayoutResults(l.container, items); err != nil {
		return err
	}

	return nil
}

func boxLayoutFlags(orientation Orientation, children *WidgetList) LayoutFlags {
	var flags LayoutFlags
	var hasNonShrinkableHorz bool
	var hasNonShrinkableVert bool

	count := children.Len()
	if count == 0 {
		return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert
	} else {
		for i := 0; i < count; i++ {
			widget := children.At(i)

			if _, ok := widget.(*splitterHandle); ok || !shouldLayoutWidget(widget) {
				continue
			}

			if s, ok := widget.(*Spacer); ok {
				if s.greedyLocallyOnly {
					continue
				}
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

	if orientation == Horizontal {
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

func boxLayoutItems(widgets []Widget, orientation Orientation, alignment Alignment2D, bounds Rectangle, margins Margins, spacing int, hwnd2StretchFactor map[win.HWND]int) ([]layoutResultItem, error) {
	if len(widgets) == 0 {
		return nil, nil
	}

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
		sf := hwnd2StretchFactor[widget.Handle()]
		if sf == 0 {
			sf = 1
		}
		stretchFactors[i] = sf

		flags := widget.LayoutFlags()

		max := widget.MaxSizePixels()
		pref := widget.SizeHint()

		if orientation == Horizontal {
			growable2[i] = flags&GrowableVert > 0

			minSizes[i] = minSizeEffective(widget).Width

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

			if hfw, ok := widget.(HeightForWidther); ok {
				minSizes[i] = hfw.HeightForWidth(bounds.Width - margins.HNear - margins.HFar)
			} else {
				minSizes[i] = minSizeEffective(widget).Height
			}

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
		sortedWidgetInfo[i].widgetBase = widget.AsWidgetBase()

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

	var start1, start2, space1, space2 int
	if orientation == Horizontal {
		start1 = bounds.X + margins.HNear
		start2 = bounds.Y + margins.VNear
		space1 = bounds.Width - margins.HNear - margins.HFar
		space2 = bounds.Height - margins.VNear - margins.VFar
	} else {
		start1 = bounds.Y + margins.VNear
		start2 = bounds.X + margins.HNear
		space1 = bounds.Height - margins.VNear - margins.VFar
		space2 = bounds.Width - margins.HNear - margins.HFar
	}

	spacingRemaining := spacing * (len(widgets) - 1)

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
			space1 -= (size + spacing)
			spacingRemaining -= spacing
		}
	}

	results := make([]layoutResultItem, 0, len(widgets))

	excessTotal := space1 - minSizesRemaining - spacingRemaining
	excessShare := excessTotal / len(widgets)
	halfExcessShare := excessTotal / (len(widgets) * 2)
	p1 := start1
	for i, widget := range widgets {
		s1 := sizes[i]

		var s2 int
		if growable2[i] {
			s2 = space2
		} else {
			s2 = prefSizes2[i]
		}

		align := widget.Alignment()
		if align == AlignHVDefault {
			align = alignment
		}

		var x, y, w, h, p2 int
		if orientation == Horizontal {
			switch align {
			case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
				// nop

			case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
				p1 += excessShare

			default:
				p1 += halfExcessShare
			}

			switch align {
			case AlignHNearVNear, AlignHCenterVNear, AlignHFarVNear:
				p2 = start2

			case AlignHNearVFar, AlignHCenterVFar, AlignHFarVFar:
				p2 = start2 + space2 - s2

			default:
				p2 = start2 + (space2-s2)/2
			}

			x, y, w, h = p1, p2, s1, s2
		} else {
			switch align {
			case AlignHNearVNear, AlignHCenterVNear, AlignHFarVNear:
				// nop

			case AlignHNearVFar, AlignHCenterVFar, AlignHFarVFar:
				p1 += excessShare

			default:
				p1 += halfExcessShare
			}

			switch align {
			case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
				p2 = start2

			case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
				p2 = start2 + space2 - s2

			default:
				p2 = start2 + (space2-s2)/2
			}

			x, y, w, h = p2, p1, s2, s1
		}

		if orientation == Horizontal {
			switch align {
			case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
				p1 += excessShare

			case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
				// nop

			default:
				p1 += halfExcessShare
			}

		} else {
			switch align {
			case AlignHNearVNear, AlignHCenterVNear, AlignHFarVNear:
				p1 += excessShare

			case AlignHNearVFar, AlignHCenterVFar, AlignHFarVFar:
				// nop

			default:
				p1 += halfExcessShare
			}
		}

		p1 += s1 + spacing

		results = append(results, layoutResultItem{widget: widget, bounds: Rectangle{X: x, Y: y, Width: w, Height: h}})
	}

	return results, nil
}
