// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type FlowLayout struct {
	LayoutBase
	hwnd2StretchFactor map[win.HWND]int
}

type flowLayoutSection struct {
	items            []flowLayoutSectionItem
	primarySpaceLeft int
	secondaryMinSize int
}

type flowLayoutSectionItem struct {
	widget  Widget
	minSize Size
}

func NewFlowLayout() *FlowLayout {
	l := &FlowLayout{
		LayoutBase: LayoutBase{
			sizeAndDPI2MinSize: make(map[sizeAndDPI]Size),
			margins96dpi:       Margins{9, 9, 9, 9},
			spacing96dpi:       6,
		},
		hwnd2StretchFactor: make(map[win.HWND]int),
	}
	l.layout = l

	return l
}

func (l *FlowLayout) StretchFactor(widget Widget) int {
	if factor, ok := l.hwnd2StretchFactor[widget.Handle()]; ok {
		return factor
	}

	return 1
}

func (l *FlowLayout) SetStretchFactor(widget Widget, factor int) error {
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

func (l *FlowLayout) cleanupStretchFactors() {
	widgets := l.container.Children()

	for handle, _ := range l.hwnd2StretchFactor {
		if !widgets.containsHandle(handle) {
			delete(l.hwnd2StretchFactor, handle)
		}
	}
}

func (l *FlowLayout) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (l *FlowLayout) MinSize() Size {
	if l.container == nil {
		return Size{}
	}

	return l.MinSizeForSize(l.container.ClientBoundsPixels().Size())
}

func (l *FlowLayout) MinSizeForSize(size Size) Size {
	if l.container == nil {
		return Size{}
	}

	dpi := l.container.DPI()

	if min, ok := l.sizeAndDPI2MinSize[sizeAndDPI{size, dpi}]; ok {
		return min
	}

	bounds := Rectangle{Width: size.Width}

	sections := l.sectionsForPrimarySize(size.Width)

	var s Size
	var maxPrimary int

	for i, section := range sections {
		var widgets []Widget
		var sectionMinWidth int
		for _, item := range section.items {
			widgets = append(widgets, item.widget)

			sectionMinWidth += item.minSize.Width
		}
		sectionMinWidth += (len(section.items) - 1) * l.spacing
		maxPrimary = maxi(maxPrimary, sectionMinWidth)

		bounds.Height = section.secondaryMinSize

		margins := l.margins
		if i > 0 {
			margins.VNear = 0
		}
		if i < len(sections)-1 {
			margins.VFar = 0
		}

		layoutItems, err := boxLayoutItems(widgets, Horizontal, l.alignment, bounds, margins, l.spacing, l.hwnd2StretchFactor)
		if err != nil {
			return Size{}
		}

		var maxSecondary int

		for _, item := range layoutItems {
			min := minSizeEffective(item.widget)

			if hfw, ok := item.widget.(HeightForWidther); ok {
				item.bounds.Height = hfw.HeightForWidth(item.bounds.Width)
			} else {
				item.bounds.Height = min.Height
			}

			maxSecondary = maxi(maxSecondary, item.bounds.Height)
		}

		s.Height += maxSecondary

		bounds.Y += maxSecondary + l.spacing
	}

	s.Width = maxPrimary

	s.Width += l.margins.HNear + l.margins.HFar
	s.Height += l.margins.VNear + l.margins.VFar + (len(sections)-1)*l.spacing

	if s.Width > 0 && s.Height > 0 {
		l.sizeAndDPI2MinSize[sizeAndDPI{size, dpi}] = s
	}

	return s
}

func (l *FlowLayout) Update(reset bool) error {
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

	if !performingScheduledLayouts && scheduleLayout(l) {
		return nil
	}

	l.dirty = false

	if l.resetNeeded {
		l.resetNeeded = false

		l.cleanupStretchFactors()
	}

	ifContainerIsScrollViewDoCoolSpecialLayoutStuff(l)

	bounds := l.container.ClientBoundsPixels()
	sections := l.sectionsForPrimarySize(bounds.Width)

	for i, section := range sections {
		var widgets []Widget
		for _, item := range section.items {
			widgets = append(widgets, item.widget)
		}

		bounds.Height = section.secondaryMinSize

		margins := l.margins
		if i > 0 {
			margins.VNear = 0
		}
		if i < len(sections)-1 {
			margins.VFar = 0
		}

		layoutItems, err := boxLayoutItems(widgets, Horizontal, l.alignment, bounds, margins, l.spacing, l.hwnd2StretchFactor)
		if err != nil {
			return err
		}

		var maxSecondary int

		for _, li := range layoutItems {
			if hfw, ok := li.widget.(HeightForWidther); ok {
				li.bounds.Height = hfw.HeightForWidth(li.bounds.Width)
			} else {
				li.bounds.Height = minSizeEffective(li.widget).Height
			}

			maxSecondary = maxi(maxSecondary, li.bounds.Height)
		}

		bounds.Height = maxSecondary + margins.VNear + margins.VFar

		if layoutItems, err = boxLayoutItems(widgets, Horizontal, l.alignment, bounds, margins, l.spacing, l.hwnd2StretchFactor); err != nil {
			return err
		}

		if err := applyLayoutResults(l.container, layoutItems); err != nil {
			return err
		}

		bounds.Y += bounds.Height + l.spacing
	}

	return nil
}

func (l *FlowLayout) sectionsForPrimarySize(primarySize int) []flowLayoutSection {
	children := l.container.Children()
	count := children.Len()

	var sections []flowLayoutSection

	section := flowLayoutSection{
		primarySpaceLeft: primarySize - l.margins.HNear - l.margins.HFar,
	}

	addSection := func() {
		sections = append(sections, section)
		section.items = nil
		section.primarySpaceLeft = primarySize - l.margins.HNear - l.margins.HFar
		section.secondaryMinSize = 0
	}

	for i := 0; i < count; i++ {
		var item flowLayoutSectionItem

		item.widget = children.At(i)

		if !shouldLayoutWidget(item.widget) {
			continue
		}

		item.minSize = minSizeEffective(item.widget)

		addItem := func() {
			section.items = append(section.items, item)
			if len(section.items) > 1 {
				section.primarySpaceLeft -= l.spacing
			}
			section.primarySpaceLeft -= item.minSize.Width

			section.secondaryMinSize = maxi(section.secondaryMinSize, item.minSize.Height)
		}

		if section.primarySpaceLeft < item.minSize.Width && len(section.items) == 0 {
			addItem()
			addSection()
		} else if section.primarySpaceLeft < l.spacing+item.minSize.Width && len(section.items) > 0 {
			addSection()
			addItem()
		} else {
			addItem()
		}
	}

	if len(section.items) > 0 {
		addSection()
	}

	if len(sections) > 0 {
		sections[0].secondaryMinSize += l.margins.VNear
		sections[len(sections)-1].secondaryMinSize += l.margins.VFar
	}

	return sections
}
