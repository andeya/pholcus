// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type TextLabel struct {
	static
	textChangedPublisher EventPublisher
}

func NewTextLabel(parent Container) (*TextLabel, error) {
	return NewTextLabelWithStyle(parent, 0)
}

func NewTextLabelWithStyle(parent Container, style uint32) (*TextLabel, error) {
	tl := new(TextLabel)

	if err := tl.init(tl, parent); err != nil {
		return nil, err
	}

	tl.textAlignment = AlignHNearVNear

	tl.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return tl.Text()
		},
		func(v interface{}) error {
			return tl.SetText(assertStringOr(v, ""))
		},
		tl.textChangedPublisher.Event()))

	return tl, nil
}

func (tl *TextLabel) asStatic() *static {
	return &tl.static
}

func (*TextLabel) LayoutFlags() LayoutFlags {
	return GrowableHorz | GrowableVert
}

func (tl *TextLabel) MinSizeHint() Size {
	if tl.minSize.Width > 0 {
		return tl.calculateTextSizeForWidth(tl.minSize.Width)
	}

	return tl.calculateTextSizeForWidth(0)
}

func (tl *TextLabel) SizeHint() Size {
	return tl.calculateTextSizeForWidth(tl.WidthPixels())
}

func (tl *TextLabel) HeightForWidth(width int) int {
	return tl.calculateTextSizeForWidth(width).Height
}

func (tl *TextLabel) TextAlignment() Alignment2D {
	return tl.textAlignment
}

func (tl *TextLabel) SetTextAlignment(alignment Alignment2D) error {
	if alignment == AlignHVDefault {
		alignment = AlignHNearVNear
	}

	return tl.setTextAlignment(alignment)
}

func (tl *TextLabel) Text() string {
	return tl.text()
}

func (tl *TextLabel) SetText(text string) error {
	if changed, err := tl.setText(text); err != nil {
		return err
	} else if !changed {
		return nil
	}

	tl.textChangedPublisher.Publish()

	return nil
}
