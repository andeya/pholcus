// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type Font struct {
	Family    string
	PointSize int
	Bold      bool
	Italic    bool
	Underline bool
	StrikeOut bool
}

func (f Font) Create() (*walk.Font, error) {
	if f.Family == "" && f.PointSize == 0 {
		return nil, nil
	}

	var fs walk.FontStyle

	if f.Bold {
		fs |= walk.FontBold
	}
	if f.Italic {
		fs |= walk.FontItalic
	}
	if f.Underline {
		fs |= walk.FontUnderline
	}
	if f.StrikeOut {
		fs |= walk.FontStrikeOut
	}

	return walk.NewFont(f.Family, f.PointSize, fs)
}
