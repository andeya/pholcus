// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"sort"
)

type gridLayoutCell struct {
	row        int
	column     int
	widgetBase *WidgetBase
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
	LayoutBase
	rowStretchFactors    []int
	columnStretchFactors []int
	widgetBase2Info      map[*WidgetBase]*gridLayoutWidgetInfo
	cells                [][]gridLayoutCell
}

func NewGridLayout() *GridLayout {
	l := &GridLayout{
		LayoutBase: LayoutBase{
			sizeAndDPI2MinSize: make(map[sizeAndDPI]Size),
			margins96dpi:       Margins{9, 9, 9, 9},
			spacing96dpi:       6,
		},
		widgetBase2Info: make(map[*WidgetBase]*gridLayoutWidgetInfo),
	}
	l.layout = l

	return l
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
	for wb, info := range l.widgetBase2Info {
		l.widgetBase2Info[wb].cell = &l.cells[info.cell.row][info.cell.column]
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
	var wb *WidgetBase
	if widget != nil {
		wb = widget.AsWidgetBase()
	}

	for row := r.Y; row < r.Y+r.Height; row++ {
		for col := r.X; col < r.X+r.Width; col++ {
			l.cells[row][col].widgetBase = wb
		}
	}
}

func (l *GridLayout) Range(widget Widget) (r Rectangle, ok bool) {
	if widget == nil {
		return Rectangle{}, false
	}

	info := l.widgetBase2Info[widget.AsWidgetBase()]

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
	if r.Width < 1 || r.Height < 1 {
		return newError("range.Width and range.Height must be >= 1")
	}

	wb := widget.AsWidgetBase()

	info := l.widgetBase2Info[wb]
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
		l.widgetBase2Info[wb] = info
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
	for wb, info := range l.widgetBase2Info {
		if !children.containsHandle(wb.window.Handle()) {
			l.setWidgetOnCells(nil, rangeFromGridLayoutWidgetInfo(info))
			delete(l.widgetBase2Info, wb)
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

			if s, ok := widget.(*Spacer); ok && s.greedyLocallyOnly || !shouldLayoutWidget(widget) {
				continue
			}

			wf := widget.LayoutFlags()

			if wf&GreedyHorz != 0 && widget.MaxSizePixels().Width > 0 {
				wf &^= GreedyHorz
			}
			if wf&GreedyVert != 0 && widget.MaxSizePixels().Height > 0 {
				wf &^= GreedyVert
			}

			flags |= wf
		}
	}

	return flags
}

func (l *GridLayout) MinSize() Size {
	if l.container == nil || len(l.cells) == 0 {
		return Size{}
	}

	return l.MinSizeForSize(l.container.ClientBoundsPixels().Size())
}

func (l *GridLayout) MinSizeForSize(size Size) Size {
	if l.container == nil || len(l.cells) == 0 {
		return Size{}
	}

	dpi := l.container.DPI()

	if min, ok := l.sizeAndDPI2MinSize[sizeAndDPI{size, dpi}]; ok {
		return min
	}

	ws := make([]int, len(l.cells[0]))

	widget2MinSize := make(map[Widget]Size)

	for wb, _ := range l.widgetBase2Info {
		widget := wb.window.(Widget)

		if !shouldLayoutWidget(widget) {
			continue
		}

		widget2MinSize[widget] = minSizeEffective(widget)
	}

	for row := 0; row < len(l.cells); row++ {
		for col := 0; col < len(ws); col++ {
			wb := l.cells[row][col].widgetBase
			if wb == nil {
				continue
			}

			widget := wb.window.(Widget)

			if !shouldLayoutWidget(widget) {
				continue
			}

			min := widget2MinSize[widget]
			info := l.widgetBase2Info[wb]

			if info.spanHorz == 1 {
				ws[col] = maxi(ws[col], min.Width)
			}
		}
	}

	widths := l.sectionSizesForSpace(Horizontal, size.Width, nil)
	heights := l.sectionSizesForSpace(Vertical, size.Height, widths)

	for row := range heights {
		var maxHeight int
		for col := range widths {
			wb := l.cells[row][col].widgetBase
			if wb == nil {
				continue
			}

			widget := wb.window.(Widget)

			if !shouldLayoutWidget(widget) {
				continue
			}

			if info := l.widgetBase2Info[wb]; info.spanVert == 1 {
				if hfw, ok := widget.(HeightForWidther); ok {
					maxHeight = maxi(maxHeight, hfw.HeightForWidth(l.spannedWidth(info, widths)))
				} else {
					maxHeight = maxi(maxHeight, widget2MinSize[widget].Height)
				}
			}
		}
		heights[row] = maxHeight
	}

	width := l.margins.HNear + l.margins.HFar
	height := l.margins.VNear + l.margins.VFar

	for i, w := range ws {
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

	if width > 0 && height > 0 {
		l.sizeAndDPI2MinSize[sizeAndDPI{size, dpi}] = Size{width, height}
	}

	return Size{width, height}
}

func (l *GridLayout) spannedWidth(info *gridLayoutWidgetInfo, widths []int) int {
	width := 0

	for i := info.cell.column; i < info.cell.column+info.spanHorz; i++ {
		if w := widths[i]; w > 0 {
			width += w
			if i > info.cell.column {
				width += l.spacing
			}
		}
	}

	return width
}

func (l *GridLayout) spannedHeight(info *gridLayoutWidgetInfo, heights []int) int {
	height := 0

	for i := info.cell.row; i < info.cell.row+info.spanVert; i++ {
		if h := heights[i]; h > 0 {
			height += h
			if i > info.cell.row {
				height += l.spacing
			}
		}
	}

	return height
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

		l.cleanup()
	}

	ifContainerIsScrollViewDoCoolSpecialLayoutStuff(l)

	cb := l.container.ClientBoundsPixels()

	widths := l.sectionSizesForSpace(Horizontal, cb.Width, nil)
	heights := l.sectionSizesForSpace(Vertical, cb.Height, widths)

	items := make([]layoutResultItem, 0, len(l.widgetBase2Info))

	for wb, info := range l.widgetBase2Info {
		widget := wb.window.(Widget)

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

		width := l.spannedWidth(info, widths)
		height := l.spannedHeight(info, heights)

		w := width
		h := height

		if lf := widget.LayoutFlags(); lf&GrowableHorz == 0 || lf&GrowableVert == 0 {
			s := widget.SizeHint()
			max := widget.MaxSizePixels()
			if max.Width > 0 && s.Width > max.Width {
				s.Width = max.Width
			}
			if max.Height > 0 && s.Height > max.Height {
				s.Height = max.Height
			}

			if lf&GrowableHorz == 0 {
				w = s.Width
			}
			if lf&GrowableVert == 0 {
				h = s.Height
			}
		}

		alignment := widget.Alignment()
		if alignment == AlignHVDefault {
			alignment = l.alignment
		}

		if w != width {
			switch alignment {
			case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
				x += (width - w) / 2

			case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
				x += width - w
			}
		}

		if h != height {
			switch alignment {
			case AlignHNearVCenter, AlignHCenterVCenter, AlignHFarVCenter:
				y += (height - h) / 2

			case AlignHNearVFar, AlignHCenterVFar, AlignHFarVFar:
				y += height - h
			}
		}

		items = append(items, layoutResultItem{widget: widget, bounds: Rectangle{X: x, Y: y, Width: w, Height: h}})
	}

	if err := applyLayoutResults(l.container, items); err != nil {
		return err
	}

	return nil
}

func (l *GridLayout) sectionSizesForSpace(orientation Orientation, space int, widths []int) []int {
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
			var wb *WidgetBase
			if orientation == Horizontal {
				wb = l.cells[j][i].widgetBase
			} else {
				wb = l.cells[i][j].widgetBase
			}

			if wb == nil {
				continue
			}

			widget := wb.window.(Widget)

			if !shouldLayoutWidget(widget) {
				continue
			}

			info := l.widgetBase2Info[wb]
			flags := widget.LayoutFlags()

			max := widget.MaxSizePixels()
			pref := widget.SizeHint()

			if orientation == Horizontal {
				if info.spanHorz == 1 {
					minSizes[i] = maxi(minSizes[i], minSizeEffective(widget).Width)
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
					if hfw, ok := widget.(HeightForWidther); ok {
						minSizes[i] = maxi(minSizes[i], hfw.HeightForWidth(l.spannedWidth(info, widths)))
					} else {
						minSizes[i] = maxi(minSizes[i], minSizeEffective(widget).Height)
					}
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

	if orientation == Horizontal {
		space -= l.margins.HNear + l.margins.HFar
	} else {
		space -= l.margins.VNear + l.margins.VFar
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
