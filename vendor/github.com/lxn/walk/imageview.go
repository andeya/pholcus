// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"math"
	"strconv"

	"github.com/lxn/win"
)

type ImageViewMode int

const (
	ImageViewModeIdeal ImageViewMode = iota
	ImageViewModeCorner
	ImageViewModeCenter
	ImageViewModeShrink
	ImageViewModeZoom
	ImageViewModeStretch
)

type ImageView struct {
	*CustomWidget
	image                  Image
	imageChangedPublisher  EventPublisher
	margin96dpi            int
	marginChangedPublisher EventPublisher
	mode                   ImageViewMode
}

func NewImageView(parent Container) (*ImageView, error) {
	iv := new(ImageView)

	cw, err := NewCustomWidget(parent, 0, func(canvas *Canvas, updateBounds Rectangle) error {
		return iv.drawImage(canvas, updateBounds)
	})
	if err != nil {
		return nil, err
	}

	iv.CustomWidget = cw

	if err := InitWrapperWindow(iv); err != nil {
		iv.Dispose()
		return nil, err
	}

	iv.SetInvalidatesOnResize(true)

	iv.SetBackground(NullBrush())

	iv.MustRegisterProperty("Image", NewProperty(
		func() interface{} {
			return iv.Image()
		},
		func(v interface{}) error {
			var img Image

			switch val := v.(type) {
			case Image:
				img = val

			case int:
				var err error
				if img, err = Resources.Image(strconv.Itoa(val)); err != nil {
					return err
				}

			case string:
				var err error
				if img, err = Resources.Image(val); err != nil {
					return err
				}

			default:
				return ErrInvalidType
			}

			return iv.SetImage(img)
		},
		iv.imageChangedPublisher.Event()))

	iv.MustRegisterProperty("Margin", NewProperty(
		func() interface{} {
			return iv.Margin()
		},
		func(v interface{}) error {
			return iv.SetMargin(assertIntOr(v, 0))
		},
		iv.MarginChanged()))

	return iv, nil
}

func (iv *ImageView) LayoutFlags() LayoutFlags {
	if iv.mode == ImageViewModeIdeal {
		return 0
	}

	return iv.CustomWidget.LayoutFlags()
}

func (iv *ImageView) MinSizeHint() Size {
	if iv.mode == ImageViewModeIdeal {
		return iv.SizeHint()
	}

	s := iv.IntFrom96DPI(iv.margin96dpi)*2 + 1

	return Size{s, s}
}

func (iv *ImageView) SizeHint() Size {
	if iv.mode == ImageViewModeIdeal && iv.image != nil {
		m2 := iv.IntFrom96DPI(iv.margin96dpi) * 2
		s := iv.SizeFrom96DPI(iv.image.Size())
		s.Width += m2
		s.Height += m2
		return s
	}

	return iv.CustomWidget.SizeHint()
}

func (iv *ImageView) Mode() ImageViewMode {
	return iv.mode
}

func (iv *ImageView) SetMode(mode ImageViewMode) {
	if mode == iv.mode {
		return
	}

	iv.mode = mode

	iv.Invalidate()

	iv.updateParentLayout()
}

func (iv *ImageView) applyDPI(dpi int) {
	iv.CustomWidget.ApplyDPI(dpi)

	iv.Invalidate()

	iv.updateParentLayout()
}

func (iv *ImageView) Image() Image {
	return iv.image
}

func (iv *ImageView) SetImage(value Image) error {
	if value == iv.image {
		return nil
	}

	iv.image = value

	_, isMetafile := value.(*Metafile)
	iv.SetClearsBackground(isMetafile)

	err := iv.Invalidate()

	if iv.mode == ImageViewModeIdeal {
		iv.updateParentLayout()
	}

	iv.imageChangedPublisher.Publish()

	return err
}

func (iv *ImageView) ImageChanged() *Event {
	return iv.imageChangedPublisher.Event()
}

func (iv *ImageView) Margin() int {
	return iv.margin96dpi
}

func (iv *ImageView) SetMargin(margin int) error {
	if margin == iv.margin96dpi {
		return nil
	}

	iv.margin96dpi = margin

	err := iv.Invalidate()

	if iv.mode == ImageViewModeIdeal {
		iv.updateParentLayout()
	}

	iv.marginChangedPublisher.Publish()

	return err
}

func (iv *ImageView) MarginChanged() *Event {
	return iv.marginChangedPublisher.Event()
}

func (iv *ImageView) drawImage(canvas *Canvas, _ Rectangle) error {
	if iv.image == nil {
		return nil
	}

	cb := iv.ClientBoundsPixels()

	margin := iv.IntFrom96DPI(iv.margin96dpi)

	cb.Width -= margin * 2
	cb.Height -= margin * 2

	s := iv.SizeFrom96DPI(iv.image.Size())

	switch iv.mode {
	case ImageViewModeShrink, ImageViewModeZoom, ImageViewModeStretch:
		var bounds Rectangle

		if iv.mode == ImageViewModeStretch {
			bounds.X = margin
			bounds.Y = margin
			bounds.Width = cb.Width
			bounds.Height = cb.Height
		} else {
			var scale float64
			if iv.mode == ImageViewModeZoom || s.Width > cb.Width || s.Height > cb.Height {
				sx := float64(cb.Width) / float64(s.Width)
				sy := float64(cb.Height) / float64(s.Height)

				scale = math.Min(sx, sy)
			} else {
				scale = 1.0
			}

			bounds.Width = int(float64(s.Width) * scale)
			bounds.Height = int(float64(s.Height) * scale)
			bounds.X = margin + (cb.Width-bounds.Width)/2
			bounds.Y = margin + (cb.Height-bounds.Height)/2
		}

		return canvas.DrawImageStretched(iv.image, RectangleTo96DPI(bounds, iv.DPI()))

	case ImageViewModeCorner, ImageViewModeCenter:
		win.IntersectClipRect(canvas.hdc, int32(margin), int32(margin), int32(cb.Width+margin), int32(cb.Height+margin))
	}

	var pos Point

	switch iv.mode {
	case ImageViewModeIdeal, ImageViewModeCorner:
		pos.X = margin
		pos.Y = margin

	case ImageViewModeCenter:
		pos.X = margin + (cb.Width-s.Width)/2
		pos.Y = margin + (cb.Height-s.Height)/2
	}

	return canvas.DrawImage(iv.image, PointTo96DPI(pos, iv.DPI()))
}
