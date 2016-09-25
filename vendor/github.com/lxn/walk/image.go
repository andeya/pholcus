// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"strings"
)

import (
	"github.com/lxn/win"
)

type Image interface {
	draw(hdc win.HDC, location Point) error
	drawStretched(hdc win.HDC, bounds Rectangle) error
	Dispose()
	Size() Size
}

func NewImageFromFile(filePath string) (Image, error) {
	if strings.HasSuffix(filePath, ".ico") {
		return NewIconFromFile(filePath)
	} else if strings.HasSuffix(filePath, ".emf") {
		return NewMetafileFromFile(filePath)
	}

	return NewBitmapFromFile(filePath)
}
