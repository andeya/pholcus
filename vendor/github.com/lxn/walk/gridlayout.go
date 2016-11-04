// Copyright 2011 The Walk Authors. All rights reserved.
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

type gridLayoutCell struct {
	row    int
	column int
	widget Widget
}

type gridLayoutSection struct {
	greedyNonSpacerCount int
	greedySpacerCount    int
}

type gridLayoutWidgetInfo struct {
	cell        *gridLayoutCell
	spanHorz    int
	spanVert    int
	minSize     Size
	minSizeHint Size
}

type GridLayout struct {
	container            Container
	margins              Margins
	spacing              int
	resetNeeded          bool
	rowStretchFactors    []int
	columnStretchFactors []int
	widget2Info          map[Widget]*gridLayoutWidgetInfo
	cells                [][]gridLayoutCell
}

func NewGridLayout() *GridLayout {
	l := &GridLayout{
		widget2Info: make(map[Widget]*gridLayoutWidgetInfo),
	}

	return l
}

func (l *GridLayout) Container() Container {
	return l.container
}

func (l *GridLayout) SetContainer(value Container) {
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

func (l *GridLayout) Margins() Margins {
	return l.margins
}

func (l *GridLayout) SetMargins(value Margins) error {
	if value.HNear < 0 || value.VNear < 0 || value.HFar < 0 || value.VFar < 0 {
		return newError("margins must be positive")
	}

	l.margins = value

	return nil
}

func (l *GridLayout) Spacing() int {
	return l.spacing
}

func (l *GridLayout) SetSpacing(value int) error {
	if value != l.spacing {
		if value < 0 {
			return newError("spacing cannot be negative")
		}

		l.spacing = value

		l.Update(false)
	}

	return nil
}

func (l *GridLayout) sufficientStretchFactors(stretchFactors []int, required int) []int {
	oldLen := len(stretchFactors)
	if oldLen < required {
		if cap(stretchFactors) < required {
			temp := make([]int, required, maxi(required, len(stretchFactors)*2))
			copy(temp, stretchFactors)
			stretchFactors = temp
		} else {
			stretchFactors = stretchFactors[:required]
		}

		for i := oldLen; i < len(stretchFactors); i++ {
			stretchFactors[i] = 1
		}
	}

	return stretchFactors
}

func (l *GridLayout) ensureSufficientSize(rows, columns int) {
	l.rowStretchFactors = l.sufficientStretchFactors(l.rowStretchFactors, rows)
	l.columnStretchFactors = l.sufficientStretchFactors(l.columnStretchFactors, columns)

	if len(l.cells) < len(l.rowStretchFactors) {
		if cap(l.cells) < cap(l.rowStretchFactors) {
			temp := make([][]gridLayoutCell, len(l.rowStretchFactors), cap(l.rowStretchFactors))
			copy(temp, l.cells)
			l.cells = temp
		} else {
			l.cells = l.cells[:len(l.rowStretchFactors)]
		}
	}

	for i := 0; i < len(l.cells); i++ {
		if len(l.cells[i]) < len(l.columnStretchFactors) {
			if cap(l.cells[i]) < cap(l.columnStretchFactors) {
				temp := make([]gridLayoutCell, len(l.columnStretchFactors))
				copy(temp, l.cells[i])
				l.cells[i] = temp
			} else {
				l.cells[i] = l.cells[i][:len(l.columnStretchFactors)]
			}
		}
	}

	// FIXME: Not sure if this works.
	for widget, info := range l.widget2Info {
		l.widget2Info[widget].cell = &l.cells[info.cell.row][info.cell.column]
	}
}

func (l *GridLayout) RowStretchFactor(row int) int {
	if row < 0 {
		// FIXME: Should we rather return an error?
		return -1
	}

	if row >= len(l.rowStretchFactors) {
		return 1
	}

	return l.rowStretchFactors[row]
}

func (l *GridLayout) SetRowStretchFactor(row, factor int) error {
	if row < 0 {
		return newError("row must be >= 0")
	}

	if factor != l.RowStretchFactor(row) {
		if l.container == nil {
			return newError("container required")
		}
		if factor < 1 {
			return newError("factor must be >= 1")
		}

		l.ensureSufficientSize(row+1, len(l.columnStretchFactors))

		l.rowStretchFactors[row] = factor

		l.Update(false)
	}

	return nil
}

func (l *GridLayout) ColumnStretchFactor(column int) int {
	if column < 0 {
		// FIXME: Should we rather return an error?
		return -1
	}

	if column >= len(l.columnStretchFactors) {
		return 1
	}

	return l.columnStretchFactors[column]
}

func (l *GridLayout) SetColumnStretchFactor(column, factor int) error {
	if column < 0 {
		return newError("column must be >= 0")
	}

	if factor != l.ColumnStretchFactor(column) {
		if l.container == nil {
			return newError("container required")
		}
		if factor < 1 {
			return newError("factor must be >= 1")
		}

		l.ensureSufficientSize(len(l.rowStretchFactors), column+1)

		l.columnStretchFactors[column] = factor

		l.Update(false)
	}

	return nil
}

func rangeFromGridLayoutWidgetInfo(info *gridLayoutWidgetInfo) Rectangle {
	return Rectangle{
		X:      info.cell.column,
		Y:      info.cell.row,
		Width:  info.spanHorz,
		Height: info.spanVert,
	}
}

func (l *GridLayout) setWidgetOnCells(widget Widget, r Rectangle) {
	for row := r.Y; row < r.Y+r.Height; row++ {
		for col := r.X; col < r.X+r.Width; col++ {
			l.cells[row][col].widget = widget
		}
	}
}

func (l *GridLayout) Range(widget Widget) (r Rectangle, ok bool) {
	if widget == nil {
		return Rectangle{}, false
	}

	info := l.widget2Info[widget]

	if info == nil ||
		l.container == nil ||
		!l.container.Children().containsHandle(widget.Handle()) {
		return Rectangle{}, false
	}

	return rangeFromGridLayoutWidgetInfo(info), true
}

func (l *GridLayout) SetRange(widget Widget, r Rectangle) error {
	if widget == nil {
		return newError("widget required")
	}
	if l.container == nil {
		return newError("container required")
	}
	if !l.container.Children().containsHandle(widget.Handle()) {
		return newError("widget must be child of container")
	}
	if r.X < 0 || r.Y < 0 {
		return newError("range.X and range.Y must be >= 0")
	}
	if r.Width < 0 || r.Height < 0 {
		return newError("range.Width and range.Height must be > 1")
	}

	info := l.widget2Info[widget]
	if info == nil {
		info = new(gridLayoutWidgetInfo)
	} else {
		l.setWidgetOnCells(nil, rangeFromGridLayoutWidgetInfo(info))
	}

	l.ensureSufficientSize(r.Y+r.Height, r.X+r.Width)

	cell := &l.cells[r.Y][r.X]
	cell.row = r.Y
	cell.column = r.X

	if info.cell == nil {
		// We have to do this _after_ calling ensureSufficientSize().
		l.widget2Info[widget] = info
	}

	info.cell = cell
	info.spanHorz = r.Width
	info.spanVert = r.Height

	l.setWidgetOnCells(widget, r)

	return nil
}

func (l *GridLayout) cleanup() {
	// Make sure only children of our container occupy the precious cells.
	children := l.container.Children()
	for widget, info := range l.widget2Info {
		if !children.containsHandle(widget.Handle()) {
			l.setWidgetOnCells(nil, rangeFromGridLayoutWidgetInfo(info))
			delete(l.widget2Info, widget)
		}
	}
}

func (l *GridLayout) stretchFactorsTotal(stretchFactors []int) int {
	total := 0

	for _, v := range stretchFactors {
		total += maxi(1, v)
	}

	return total
}

func (l *GridLayout) LayoutFlags() LayoutFlags {
	if l.container == nil {
		return 0
	}

	var flags LayoutFlags

	children := l.container.Children()
	count := children.Len()
	if count == 0 {
		return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert
	} else {
		for i := 0; i < count; i++ {
			widget := children.At(i)

			if shouldLayoutWidget(widget) {
				flags |= widget.LayoutFlags()
			}
		}
	}

	return flags
}

func (l *GridLayout) MinSize() Size {
	if l.container == nil || len(l.cells) == 0 {
		return Size{}
	}

	widths := make([]int, len(l.cells[0]))
	heights := make([]int, len(l.cells))

	widget2MinSize := make(map[Widget]Size)

	for widget, _ := range l.widget2Info {
		if !shouldLayoutWidget(widget) {
			continue
		}

		widget2MinSize[widget] = minSizeEffective(widget)
	}

	for row := 0; row < len(heights); row++ {
		for col := 0; col < len(widths); col++ {
			widget := l.cells[row][col].widget

			if !shouldLayoutWidget(widget) {
				continue
			}

			min := widget2MinSize[widget]
			info := l.widget2Info[widget]

			if info.spanHorz == 1 {
				widths[col] = maxi(widths[col], min.Width)
			}
			if info.spanVert == 1 {
				heights[row] = maxi(heights[row], min.Height)
			}
		}
	}

	width := l.margins.HNear + l.margins.HFar
	height := l.margins.VNear + l.margins.VFar

	for i, w := range widths {
		if w > 0 {
			if i > 0 {
				width += l.spacing
			}
			width += w
		}
	}
	for i, h := range heights {
		if h > 0 {
			if i > 0 {
				height += l.spacing
			}
			height += h
		}
	}

	return Size{width, height}
}

type gridLayoutSectionInfo struct {
	index              int
	minSize            int
	maxSize            int
	stretch            int
	hasGreedyNonSpacer bool
	hasGreedySpacer    bool
}

type gridLayoutSectionInfoList []gridLayoutSectionInfo

func (l gridLayoutSectionInfoList) Len() int {
	return len(l)
}

func (l gridLayoutSectionInfoList) Less(i, j int) bool {
	if l[i].hasGreedyNonSpacer == l[j].hasGreedyNonSpacer {
		if l[i].hasGreedySpacer == l[j].hasGreedySpacer {
			minDiff := l[i].minSize - l[j].minSize

			if minDiff == 0 {
				return l[i].maxSize/l[i].stretch < l[j].maxSize/l[j].stretch
			}

			return minDiff > 0
		}

		return l[i].hasGreedySpacer
	}

	return l[i].hasGreedyNonSpacer
}

func (l gridLayoutSectionInfoList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l *GridLayout) Update(reset bool) error {
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

		l.cleanup()
	}

	widths := l.sectionSizes(Horizontal)
	heights := l.sectionSizes(Vertical)

	hdwp := win.BeginDeferWindowPos(int32(l.container.Children().Len()))
	if hdwp == 0 {
		return lastError("BeginDeferWindowPos")
	}

	for widget, info := range l.widget2Info {
		if !shouldLayoutWidget(widget) {
			continue
		}

		x := l.margins.HNear
		for i := 0; i < info.cell.column; i++ {
			if w := widths[i]; w > 0 {
				x += w + l.spacing
			}
		}

		y := l.margins.VNear
		for i := 0; i < info.cell.row; i++ {
			if h := heights[i]; h > 0 {
				y += h + l.spacing
			}
		}

		w := 0
		for i := info.cell.column; i < info.cell.column+info.spanHorz; i++ {
			w += widths[i]
			if i > info.cell.column {
				w += l.spacing
			}
		}

		h := 0
		for i := info.cell.row; i < info.cell.row+info.spanVert; i++ {
			h += heights[i]
			if i > info.cell.row {
				h += l.spacing
			}
		}

		if lf := widget.LayoutFlags(); lf&GrowableHorz == 0 || lf&GrowableVert == 0 {
			hint := widget.SizeHint()

			if lf&GrowableHorz == 0 {
				w = hint.Width
			}
			if lf&GrowableVert == 0 {
				h = hint.Height
			}
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
	}

	if !win.EndDeferWindowPos(hdwp) {
		return lastError("EndDeferWindowPos")
	}

	return nil
}

func (l *GridLayout) sectionSizes(orientation Orientation) []int {
	var stretchFactors []int
	if orientation == Horizontal {
		stretchFactors = l.columnStretchFactors
	} else {
		stretchFactors = l.rowStretchFactors
	}

	var sectionCountWithGreedyNonSpacer int
	var sectionCountWithGreedySpacer int
	var stretchFactorsTotal [3]int
	var minSizesRemaining int
	minSizes := make([]int, len(stretchFactors))
	maxSizes := make([]int, len(stretchFactors))
	sizes := make([]int, len(stretchFactors))
	sortedSections := gridLayoutSectionInfoList(make([]gridLayoutSectionInfo, len(stretchFactors)))

	for i := 0; i < len(stretchFactors); i++ {
		var otherAxisCount int
		if orientation == Horizontal {
			otherAxisCount = len(l.rowStretchFactors)
		} else {
			otherAxisCount = len(l.columnStretchFactors)
		}

		for j := 0; j < otherAxisCount; j++ {
			var widget Widget
			if orientation == Horizontal {
				widget = l.cells[j][i].widget
			} else {
				widget = l.cells[i][j].widget
			}

			if !shouldLayoutWidget(widget) {
				continue
			}

			info := l.widget2Info[widget]
			flags := widget.LayoutFlags()

			min := widget.MinSize()
			max := widget.MaxSize()
			minHint := widget.MinSizeHint()
			pref := widget.SizeHint()

			if orientation == Horizontal {
				if info.spanHorz == 1 {
					minSizes[i] = maxi(minSizes[i], maxi(min.Width, minHint.Width))
				}

				if max.Width > 0 {
					maxSizes[i] = maxi(maxSizes[i], max.Width)
				} else if pref.Width > 0 && flags&GrowableHorz == 0 {
					maxSizes[i] = maxi(maxSizes[i], pref.Width)
				} else {
					maxSizes[i] = 32768
				}

				if info.spanHorz == 1 && flags&GreedyHorz > 0 {
					if _, isSpacer := widget.(*Spacer); isSpacer {
						sortedSections[i].hasGreedySpacer = true
					} else {
						sortedSections[i].hasGreedyNonSpacer = true
					}
				}
			} else {
				if info.spanVert == 1 {
					minSizes[i] = maxi(minSizes[i], maxi(min.Height, minHint.Height))
				}

				if max.Height > 0 {
					maxSizes[i] = maxi(maxSizes[i], max.Height)
				} else if pref.Height > 0 && flags&GrowableVert == 0 {
					maxSizes[i] = maxi(maxSizes[i], pref.Height)
				} else {
					maxSizes[i] = 32768
				}

				if info.spanVert == 1 && flags&GreedyVert > 0 {
					if _, isSpacer := widget.(*Spacer); isSpacer {
						sortedSections[i].hasGreedySpacer = true
					} else {
						sortedSections[i].hasGreedyNonSpacer = true
					}
				}
			}
		}

		sortedSections[i].index = i
		sortedSections[i].minSize = minSizes[i]
		sortedSections[i].maxSize = maxSizes[i]
		sortedSections[i].stretch = maxi(1, stretchFactors[i])

		minSizesRemaining += minSizes[i]

		if sortedSections[i].hasGreedyNonSpacer {
			sectionCountWithGreedyNonSpacer++
			stretchFactorsTotal[0] += stretchFactors[i]
		} else if sortedSections[i].hasGreedySpacer {
			sectionCountWithGreedySpacer++
			stretchFactorsTotal[1] += stretchFactors[i]
		} else {
			stretchFactorsTotal[2] += stretchFactors[i]
		}
	}

	sort.Stable(sortedSections)

	cb := l.container.ClientBounds()
	var space int
	if orientation == Horizontal {
		space = cb.Width - l.margins.HNear - l.margins.HFar
	} else {
		space = cb.Height - l.margins.VNear - l.margins.VFar
	}

	var spacingRemaining int
	for _, max := range maxSizes {
		if max > 0 {
			spacingRemaining += l.spacing
		}
	}
	if spacingRemaining > 0 {
		spacingRemaining -= l.spacing
	}

	offsets := [3]int{0, sectionCountWithGreedyNonSpacer, sectionCountWithGreedyNonSpacer + sectionCountWithGreedySpacer}
	counts := [3]int{sectionCountWithGreedyNonSpacer, sectionCountWithGreedySpacer, len(stretchFactors) - sectionCountWithGreedyNonSpacer - sectionCountWithGreedySpacer}

	for i := 0; i < 3; i++ {
		stretchFactorsRemaining := stretchFactorsTotal[i]

		for j := 0; j < counts[i]; j++ {
			info := sortedSections[offsets[i]+j]
			k := info.index

			stretch := stretchFactors[k]
			min := info.minSize
			max := info.maxSize
			size := min

			if min < max {
				excessSpace := float64(space - minSizesRemaining - spacingRemaining)

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

			space -= (size + l.spacing)
			spacingRemaining -= l.spacing
		}
	}

	return sizes
}
