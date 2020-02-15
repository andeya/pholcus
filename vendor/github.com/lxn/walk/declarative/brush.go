// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
	"strconv"
)

type TransparentBrush struct {
}

func (TransparentBrush) Create() (walk.Brush, error) {
	return walk.NullBrush(), nil
}

type SolidColorBrush struct {
	Color walk.Color
}

func (scb SolidColorBrush) Create() (walk.Brush, error) {
	return walk.NewSolidColorBrush(scb.Color)
}

type SystemColorBrush struct {
	Color walk.SystemColor
}

func (scb SystemColorBrush) Create() (walk.Brush, error) {
	return walk.NewSystemColorBrush(scb.Color)
}

type BitmapBrush struct {
	Image interface{}
}

func (bb BitmapBrush) Create() (walk.Brush, error) {
	var bmp *walk.Bitmap
	var err error

	switch img := bb.Image.(type) {
	case *walk.Bitmap:
		bmp = img

	case string:
		if bmp, err = walk.Resources.Bitmap(img); err != nil {
			return nil, err
		}

	case int:
		if bmp, err = walk.Resources.Bitmap(strconv.Itoa(img)); err != nil {
			return nil, err
		}

	default:
		return nil, walk.ErrInvalidType
	}

	return walk.NewBitmapBrush(bmp)
}

type GradientBrush struct {
	Vertexes  []walk.GradientVertex
	Triangles []walk.GradientTriangle
}

func (gb GradientBrush) Create() (walk.Brush, error) {
	return walk.NewGradientBrush(gb.Vertexes, gb.Triangles)
}
