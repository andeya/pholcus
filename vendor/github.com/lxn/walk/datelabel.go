// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"time"
)

type DateLabel struct {
	static
	date                   time.Time
	dateChangedPublisher   EventPublisher
	format                 string
	formatChangedPublisher EventPublisher
}

func NewDateLabel(parent Container) (*DateLabel, error) {
	dl := new(DateLabel)

	if err := dl.init(dl, parent); err != nil {
		return nil, err
	}

	dl.SetTextAlignment(AlignFar)
	if _, err := dl.updateText(); err != nil {
		return nil, err
	}

	dl.MustRegisterProperty("Date", NewProperty(
		func() interface{} {
			return dl.Date()
		},
		func(v interface{}) error {
			return dl.SetDate(assertTimeOr(v, time.Time{}))
		},
		dl.dateChangedPublisher.Event()))

	dl.MustRegisterProperty("Format", NewProperty(
		func() interface{} {
			return dl.Format()
		},
		func(v interface{}) error {
			return dl.SetFormat(assertStringOr(v, ""))
		},
		dl.formatChangedPublisher.Event()))

	return dl, nil
}

func (dl *DateLabel) asStatic() *static {
	return &dl.static
}

func (dl *DateLabel) TextAlignment() Alignment1D {
	return dl.textAlignment1D()
}

func (dl *DateLabel) SetTextAlignment(alignment Alignment1D) error {
	if alignment == AlignDefault {
		alignment = AlignNear
	}

	return dl.setTextAlignment1D(alignment)
}

func (dl *DateLabel) Date() time.Time {
	return dl.date
}

func (dl *DateLabel) SetDate(date time.Time) error {
	if date == dl.date {
		return nil
	}

	old := dl.date

	dl.date = date

	if _, err := dl.updateText(); err != nil {
		dl.date = old
		return err
	}

	dl.dateChangedPublisher.Publish()

	return nil
}

func (dl *DateLabel) Format() string {
	return dl.format
}

func (dl *DateLabel) SetFormat(format string) error {
	if format == dl.format {
		return nil
	}

	old := dl.format

	dl.format = format

	if _, err := dl.updateText(); err != nil {
		dl.format = old
		return err
	}

	dl.formatChangedPublisher.Publish()

	return nil
}

func (dl *DateLabel) updateText() (changed bool, err error) {
	return dl.setText(dl.date.Format(dl.format))
}
