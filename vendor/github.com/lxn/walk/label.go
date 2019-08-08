// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type Label struct {
	static
	textChangedPublisher EventPublisher
}

func NewLabel(parent Container) (*Label, error) {
	return NewLabelWithStyle(parent, 0)
}

func NewLabelWithStyle(parent Container, style uint32) (*Label, error) {
	l := new(Label)

	if err := l.init(l, parent); err != nil {
		return nil, err
	}

	l.SetTextAlignment(AlignNear)

	l.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return l.Text()
		},
		func(v interface{}) error {
			return l.SetText(assertStringOr(v, ""))
		},
		l.textChangedPublisher.Event()))

	return l, nil
}

func (l *Label) asStatic() *static {
	return &l.static
}

func (l *Label) TextAlignment() Alignment1D {
	return l.textAlignment1D()
}

func (l *Label) SetTextAlignment(alignment Alignment1D) error {
	if alignment == AlignDefault {
		alignment = AlignNear
	}

	return l.setTextAlignment1D(alignment)
}

func (l *Label) Text() string {
	return l.text()
}

func (l *Label) SetText(text string) error {
	if changed, err := l.setText(text); err != nil {
		return err
	} else if !changed {
		return nil
	}

	l.textChangedPublisher.Publish()

	return nil
}
