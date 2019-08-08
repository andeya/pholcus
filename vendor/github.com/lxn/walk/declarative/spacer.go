// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type HSpacer struct {
	// Window

	MaxSize Size
	MinSize Size
	Name    string

	// Widget

	Column        int
	ColumnSpan    int
	Row           int
	RowSpan       int
	StretchFactor int

	// Spacer

	GreedyLocallyOnly bool
	Size              int
}

func (hs HSpacer) Create(builder *Builder) (err error) {
	var flags walk.LayoutFlags
	if hs.Size == 0 {
		flags = walk.ShrinkableHorz | walk.GrowableHorz | walk.GreedyHorz
	}

	var w *walk.Spacer
	if w, err = walk.NewSpacerWithCfg(builder.Parent(), &walk.SpacerCfg{
		LayoutFlags:       flags,
		SizeHint:          walk.Size{Width: hs.Size},
		GreedyLocallyOnly: hs.GreedyLocallyOnly,
	}); err != nil {
		return
	}

	return builder.InitWidget(hs, w, nil)
}

type VSpacer struct {
	// Window

	MaxSize Size
	MinSize Size
	Name    string

	// Widget

	Column        int
	ColumnSpan    int
	Row           int
	RowSpan       int
	StretchFactor int

	// Spacer

	GreedyLocallyOnly bool
	Size              int
}

func (vs VSpacer) Create(builder *Builder) (err error) {
	var flags walk.LayoutFlags
	if vs.Size == 0 {
		flags = walk.ShrinkableVert | walk.GrowableVert | walk.GreedyVert
	}

	var w *walk.Spacer
	if w, err = walk.NewSpacerWithCfg(builder.Parent(), &walk.SpacerCfg{
		LayoutFlags:       flags,
		SizeHint:          walk.Size{Height: vs.Size},
		GreedyLocallyOnly: vs.GreedyLocallyOnly,
	}); err != nil {
		return
	}

	return builder.InitWidget(vs, w, nil)
}
