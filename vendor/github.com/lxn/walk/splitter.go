// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"bytes"
	"log"
	"strconv"
	"strings"
)

import (
	"github.com/lxn/win"
)

const splitterWindowClass = `\o/ Walk_Splitter_Class \o/`

var splitterHandleDraggingBrush *SolidColorBrush

func init() {
	MustRegisterWindowClass(splitterWindowClass)

	splitterHandleDraggingBrush, _ = NewSolidColorBrush(Color(win.GetSysColor(win.COLOR_BTNSHADOW)))
}

type Splitter struct {
	ContainerBase
	handleWidth   int
	mouseDownPos  Point
	draggedHandle *splitterHandle
	persistent    bool
}

func newSplitter(parent Container, orientation Orientation) (*Splitter, error) {
	layout := newSplitterLayout(Horizontal)
	s := &Splitter{
		ContainerBase: ContainerBase{
			layout: layout,
		},
		handleWidth: 4,
	}
	s.children = newWidgetList(s)
	layout.container = s

	if err := InitWidget(
		s,
		parent,
		splitterWindowClass,
		win.WS_VISIBLE,
		win.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	var succeeded bool
	defer func() {
		if !succeeded {
			s.Dispose()
		}
	}()

	if err := s.setOrientation(orientation); err != nil {
		return nil, err
	}

	s.SetPersistent(true)

	succeeded = true

	return s, nil
}

func NewHSplitter(parent Container) (*Splitter, error) {
	return newSplitter(parent, Horizontal)
}

func NewVSplitter(parent Container) (*Splitter, error) {
	return newSplitter(parent, Vertical)
}

func (s *Splitter) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (s *Splitter) SizeHint() Size {
	return Size{100, 100}
}

func (s *Splitter) SetLayout(value Layout) error {
	return newError("not supported")
}

func (s *Splitter) HandleWidth() int {
	return s.handleWidth
}

func (s *Splitter) SetHandleWidth(value int) error {
	if value == s.handleWidth {
		return nil
	}

	if value < 1 {
		return newError("invalid handle width")
	}

	s.handleWidth = value

	return s.layout.Update(false)
}

func (s *Splitter) Orientation() Orientation {
	layout := s.layout.(*splitterLayout)
	return layout.Orientation()
}

func (s *Splitter) setOrientation(value Orientation) error {
	var cursor Cursor
	if value == Horizontal {
		cursor = CursorSizeWE()
	} else {
		cursor = CursorSizeNS()
	}

	for i, w := range s.Children().items {
		if i%2 == 1 {
			w.SetCursor(cursor)
		}
	}

	layout := s.layout.(*splitterLayout)
	return layout.SetOrientation(value)
}

func (s *Splitter) Persistent() bool {
	return s.persistent
}

func (s *Splitter) SetPersistent(value bool) {
	s.persistent = value
}

func (s *Splitter) SaveState() error {
	buf := bytes.NewBuffer(nil)

	count := s.children.Len()
	layout := s.Layout().(*splitterLayout)

	for i := 0; i < count; i += 2 {
		if i > 0 {
			buf.WriteString(" ")
		}

		buf.WriteString(strconv.FormatFloat(layout.fractions[i/2], 'f', -1, 64))
	}

	s.putState(buf.String())

	for _, widget := range s.children.items {
		if persistable, ok := widget.(Persistable); ok {
			if err := persistable.SaveState(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Splitter) RestoreState() error {
	childCount := s.children.Len()/2 + 1
	if childCount == 0 {
		return nil
	}

	state, err := s.getState()
	if err != nil {
		return err
	}
	if state == "" {
		return nil
	}

	fractionStrs := strings.Split(state, " ")

	// FIXME: Solve this in a better way.
	if len(fractionStrs) != childCount {
		log.Print("*Splitter.RestoreState: failed due to unexpected child count (FIXME!)")
		return nil
	}

	layout := s.layout.(*splitterLayout)

	s.SetSuspended(true)
	defer s.SetSuspended(false)

	var fractionsTotal float64
	var fractions []float64

	for i, widget := range s.children.items {
		if i%2 == 0 {
			fraction, err := strconv.ParseFloat(fractionStrs[i/2+i%2], 64)
			if err != nil {
				return err
			}

			fractionsTotal += fraction
			fractions = append(fractions, fraction)
		}

		if persistable, ok := widget.(Persistable); ok {
			if err := persistable.RestoreState(); err != nil {
				return err
			}
		}
	}

	for i := range fractions {
		fractions[i] = fractions[i] / fractionsTotal
	}

	return layout.SetFractions(fractions)
}

func (s *Splitter) onInsertingWidget(index int, widget Widget) (err error) {
	return s.ContainerBase.onInsertingWidget(index, widget)
}

func (s *Splitter) onInsertedWidget(index int, widget Widget) (err error) {
	_, isHandle := widget.(*splitterHandle)
	if isHandle {
		if s.Orientation() == Horizontal {
			widget.SetCursor(CursorSizeWE())
		} else {
			widget.SetCursor(CursorSizeNS())
		}
	} else if s.children.Len()%2 == 0 {
		defer func() {
			if err != nil {
				return
			}

			var handle *splitterHandle
			handle, err = newSplitterHandle(s)
			if err != nil {
				return
			}

			var handleIndex int
			if index == 0 {
				handleIndex = 1
			} else {
				handleIndex = index
			}
			err = s.children.Insert(handleIndex, handle)
			if err == nil {
				// FIXME: These handlers will be leaked, if widgets get removed.
				handle.MouseDown().Attach(func(x, y int, button MouseButton) {
					if button != LeftButton {
						return
					}

					s.draggedHandle = handle
					s.mouseDownPos = Point{x, y}
					handle.SetBackground(splitterHandleDraggingBrush)
				})

				handle.MouseMove().Attach(func(x, y int, button MouseButton) {
					if s.draggedHandle == nil {
						return
					}

					handleIndex := s.children.Index(s.draggedHandle)

					prev := s.children.At(handleIndex - 1)
					bp := prev.Bounds()
					msep := minSizeEffective(prev)

					next := s.children.At(handleIndex + 1)
					bn := next.Bounds()
					msen := minSizeEffective(next)

					if s.Orientation() == Horizontal {
						xh := s.draggedHandle.X()

						xnew := xh + x - s.mouseDownPos.X
						if xnew < bp.X+msep.Width {
							xnew = bp.X + msep.Width
						} else if xnew >= bn.X+bn.Width-msen.Width-s.handleWidth {
							xnew = bn.X + bn.Width - msen.Width - s.handleWidth
						}

						if e := s.draggedHandle.SetX(xnew); e != nil {
							return
						}
					} else {
						yh := s.draggedHandle.Y()

						ynew := yh + y - s.mouseDownPos.Y
						if ynew < bp.Y+msep.Height {
							ynew = bp.Y + msep.Height
						} else if ynew >= bn.Y+bn.Height-msen.Height-s.handleWidth {
							ynew = bn.Y + bn.Height - msen.Height - s.handleWidth
						}

						if e := s.draggedHandle.SetY(ynew); e != nil {
							return
						}
					}
				})

				handle.MouseUp().Attach(func(x, y int, button MouseButton) {
					if s.draggedHandle == nil {
						return
					}

					dragHandle := s.draggedHandle
					s.draggedHandle = nil
					dragHandle.SetBackground(nil)

					handleIndex := s.children.Index(dragHandle)
					prev := s.children.At(handleIndex - 1)
					next := s.children.At(handleIndex + 1)

					prev.SetSuspended(true)
					defer prev.Invalidate()
					defer prev.SetSuspended(false)
					next.SetSuspended(true)
					defer next.Invalidate()
					defer next.SetSuspended(false)

					bh := dragHandle.Bounds()
					bp := prev.Bounds()
					bn := next.Bounds()

					var sizePrev int
					var sizeNext int

					if s.Orientation() == Horizontal {
						bp.Width = bh.X - bp.X
						bn.Width -= (bh.X + bh.Width) - bn.X
						bn.X = bh.X + bh.Width
						sizePrev = bp.Width
						sizeNext = bn.Width
					} else {
						bp.Height = bh.Y - bp.Y
						bn.Height -= (bh.Y + bh.Height) - bn.Y
						bn.Y = bh.Y + bh.Height
						sizePrev = bp.Height
						sizeNext = bn.Height
					}

					if e := prev.SetBounds(bp); e != nil {
						return
					}

					if e := next.SetBounds(bn); e != nil {
						return
					}

					layout := s.Layout().(*splitterLayout)
					space := float64(layout.spaceForRegularWidgets())
					fractions := layout.fractions
					i := handleIndex - 1
					jp := i/2 + i%2
					jn := jp + 1
					fractions[jp] = float64(sizePrev) / space
					fractions[jn] = float64(sizeNext) / space
				})
			}
		}()
	}

	return s.ContainerBase.onInsertedWidget(index, widget)
}

func (s *Splitter) onRemovingWidget(index int, widget Widget) (err error) {
	return s.ContainerBase.onRemovingWidget(index, widget)
}

func (s *Splitter) onRemovedWidget(index int, widget Widget) (err error) {
	_, isHandle := widget.(*splitterHandle)
	if isHandle && s.children.Len()%2 == 1 {
		return newError("cannot remove splitter handle")
	}

	if !isHandle && s.children.Len() > 1 {
		defer func() {
			if err != nil {
				return
			}

			var handleIndex int
			if index == 0 {
				handleIndex = 0
			} else {
				handleIndex = index - 1
			}
			err = s.children.RemoveAt(handleIndex)
		}()
	}

	err = s.ContainerBase.onRemovedWidget(index, widget)
	if isHandle && err == nil {
		widget.Dispose()
	}

	return
}

func (s *Splitter) onClearingWidgets() (err error) {
	panic("not implemented")
}

func (s *Splitter) onClearedWidgets() (err error) {
	panic("not implemented")
}
