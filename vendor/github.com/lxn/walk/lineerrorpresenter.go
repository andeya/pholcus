// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"bytes"
	"strings"
)

import (
	"github.com/lxn/win"
)

const lineErrorPresenterWindowClass = `\o/ Walk_LineErrorPresenter_Class \o/`

func init() {
	MustRegisterWindowClass(lineErrorPresenterWindowClass)

	lineErrorPresenterBackground, _ = NewSolidColorBrush(RGB(255, 128, 128))
}

var lineErrorPresenterBackground Brush

type LineErrorPresenter struct {
	WidgetBase
	composite    *Composite
	label        *Label
	curWidget    Widget
	widget2error map[Widget]error
}

func NewLineErrorPresenter(parent Container) (*LineErrorPresenter, error) {
	lep := &LineErrorPresenter{widget2error: make(map[Widget]error)}

	if err := InitWidget(
		lep,
		parent,
		lineErrorPresenterWindowClass,
		win.WS_VISIBLE,
		win.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			lep.Dispose()
		}
	}()

	var err error

	if lep.composite, err = newCompositeWithStyle(lep, 0); err != nil {
		return nil, err
	}

	l := NewGridLayout()
	l.SetMargins(Margins{2, 2, 2, 2})

	if err = lep.composite.SetLayout(l); err != nil {
		return nil, err
	}

	if lep.label, err = NewLabel(lep.composite); err != nil {
		return nil, err
	}

	l.SetRange(lep.label, Rectangle{0, 0, 1, 1})

	focusCurWidget := func(x, y int, button MouseButton) {
		widget := lep.curWidget

		if button == LeftButton && widget != nil {
			widget.SetFocus()

			if textSel, ok := widget.(textSelectable); ok {
				textSel.SetTextSelection(0, -1)
			}
		}
	}

	lep.MouseDown().Attach(focusCurWidget)
	lep.composite.MouseDown().Attach(focusCurWidget)
	lep.label.MouseDown().Attach(focusCurWidget)

	succeeded = true

	return lep, nil
}

func (*LineErrorPresenter) LayoutFlags() LayoutFlags {
	return GrowableHorz | GreedyHorz
}

func (lep *LineErrorPresenter) MinSizeHint() Size {
	if lep.label == nil {
		return Size{}
	}

	text := lep.label.Text()
	if text == "" {
		text = "gM"
	}

	s := lep.label.calculateTextSizeImpl(text)

	return Size{s.Width + 8, s.Height + 8}
}

func (lep *LineErrorPresenter) SizeHint() Size {
	return lep.MinSizeHint()
}

func (lep *LineErrorPresenter) applyEnabled(enabled bool) {
	lep.WidgetBase.applyEnabled(enabled)

	lep.composite.applyEnabled(enabled)
}

func (lep *LineErrorPresenter) applyFont(font *Font) {
	lep.WidgetBase.applyFont(font)

	if lep.composite == nil {
		return
	}

	// We have to call SetFont instead of applyFont here, because
	// LineErrorPresenter does not implement Container.
	lep.composite.SetFont(font)
}

func (lep *LineErrorPresenter) PresentError(err error, widget Widget) {
	if err == nil {
		delete(lep.widget2error, widget)
	} else {
		lep.widget2error[widget] = err
	}

	var found bool
	walkDescendants(ancestor(widget).AsFormBase().clientComposite, func(w Window) bool {
		if found {
			return false
		}

		wt := w.(Widget)

		if e, ok := lep.widget2error[wt]; ok {
			err, widget, found = e, wt, true
		}

		return !found
	})

	if err != nil {
		lep.curWidget = widget
	} else {
		lep.curWidget = nil
	}

	var msg string
	var background Brush

	if err != nil {
		background = lineErrorPresenterBackground

		var labelText string
		if widget != nil {
			parent := widget.Parent()
			if parent != nil {
				children := parent.Children()

				i := children.Index(widget)
				if i > 0 {
					prev := children.At(i - 1)

					if label, ok := prev.(*Label); ok {
						labelText = label.Text()
					}
				}
			}
		}

		buf := new(bytes.Buffer)
		buf.WriteString(labelText)
		if labelText != "" && !strings.HasSuffix(labelText, ":") {
			buf.WriteString(":")
		}
		if labelText != "" {
			buf.WriteString(" ")
		}
		buf.WriteString(err.Error())

		msg = buf.String()
	}

	lep.SetBackground(background)
	lep.label.SetText(msg)

	if form := ancestor(lep); form != nil && form.Handle() != lep.hWnd {
		form.SetBounds(form.Bounds())
	}
}

func (lep *LineErrorPresenter) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_SIZE, win.WM_SIZING:
		if lep.composite != nil {
			b := lep.ClientBounds()
			lep.composite.SetBounds(Rectangle{b.X + 2, b.Y + 2, b.Width - 4, b.Height - 4})
		}
	}

	return lep.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
