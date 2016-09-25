// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

const splitterHandleWindowClass = `\o/ Walk_SplitterHandle_Class \o/`

func init() {
	MustRegisterWindowClass(splitterHandleWindowClass)
}

type splitterHandle struct {
	WidgetBase
}

func newSplitterHandle(splitter *Splitter) (*splitterHandle, error) {
	if splitter == nil {
		return nil, newError("splitter cannot be nil")
	}

	sh := new(splitterHandle)
	sh.parent = splitter

	if err := InitWindow(
		sh,
		splitter,
		splitterHandleWindowClass,
		win.WS_CHILD|win.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	if err := sh.setAndClearStyleBits(0, win.WS_CLIPSIBLINGS); err != nil {
		return nil, err
	}

	return sh, nil
}

func (sh *splitterHandle) LayoutFlags() LayoutFlags {
	splitter := sh.Parent().(*Splitter)
	if splitter.Orientation() == Horizontal {
		return ShrinkableVert | GrowableVert | GreedyVert
	}

	return ShrinkableHorz | GrowableHorz | GreedyHorz
}

func (sh *splitterHandle) MinSizeHint() Size {
	return sh.SizeHint()
}

func (sh *splitterHandle) SizeHint() Size {
	splitter := sh.Parent().(*Splitter)
	handleWidth := splitter.HandleWidth()
	var size Size

	if splitter.Orientation() == Horizontal {
		size.Width = handleWidth
	} else {
		size.Height = handleWidth
	}

	return size
}
