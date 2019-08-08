// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type Separator struct {
	WidgetBase
	vertical bool
}

func NewHSeparator(parent Container) (*Separator, error) {
	return newSeparator(parent, false)
}

func NewVSeparator(parent Container) (*Separator, error) {
	return newSeparator(parent, true)
}

func newSeparator(parent Container, vertical bool) (*Separator, error) {
	s := &Separator{vertical: vertical}

	if err := InitWidget(
		s,
		parent,
		"STATIC",
		win.WS_VISIBLE|win.SS_ETCHEDHORZ,
		0); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Separator) LayoutFlags() LayoutFlags {
	if s.vertical {
		return GrowableHorz | GreedyHorz
	} else {
		return GrowableVert | GreedyVert
	}
}

func (s *Separator) MinSizeHint() Size {
	return Size{2, 2}
}

func (s *Separator) SizeHint() Size {
	return s.MinSizeHint()
}
