// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"
)

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var mw *walk.MainWindow
	var windowIcon *walk.Icon
	counter := 0

	if _, err := (MainWindow{
		AssignTo: &mw,
		Title:    "Walk Image Icon Example",
		Layout:   HBox{},
		Children: []Widget{
			HSpacer{},
			PushButton{
				Text: "Push me",
				OnClicked: func() {
					ic, err := walk.NewIconFromImage(makeDigitImage(counter))
					if err != nil {
						return
					}
					counter++
					mw.SetIcon(ic)
					if windowIcon != nil {
						windowIcon.Dispose()
					}
					windowIcon = ic
				},
			},
			HSpacer{},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}

//  A
// F B
//  G
// E C
//  D
var hexdigits = []int{0x3F, 0x06, 0x5B, 0x4F, 0x66, 0x6D, 0x7D, 0x07, 0x7F, 0x6F, 0x77, 0x7C, 0x39, 0x5E, 0x79, 0x71}

//0x7E, 0x30, 0x6D, 0x79, 0x33, 0x5B, 0x5F, 0x70, 0x7F, 0x7B, 0x77, 0x1F, 0x4E, 0x3D, 0x4F, 0x47

type seg struct {
	sx, sy int
	dx, dy int
}

var segments = []seg{
	{0, 0, 1, 0},
	{1, 0, 0, 1},
	{1, 1, 0, 1},
	{0, 2, 1, 0},
	{0, 1, 0, 1},
	{0, 0, 0, 1},
	{0, 1, 1, 0},
}

func digit(im draw.Image, col color.Color, x, y, size, digit int) {
	n := hexdigits[digit]
	for _, s := range segments {
		if n&1 != 0 {
			xx, yy := x+s.sx*size, y+s.sy*size
			for i := 0; i <= size; i++ {
				im.Set(xx, yy, col)
				xx += s.dx
				yy += s.dy
			}
		}
		n >>= 1
	}
}

func makeDigitImage(n int) image.Image {
	im := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for p := 11; p >= 0; p -= 5 {
		digit(im, color.Black, p, 5, 3, n%10)
		n /= 10
	}
	return im
}
