// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"strings"
)

type NumberLabel struct {
	static
	decimals                 int
	decimalsChangedPublisher EventPublisher
	suffix                   string
	suffixChangedPublisher   EventPublisher
	value                    float64
	valueChangedPublisher    EventPublisher
}

func NewNumberLabel(parent Container) (*NumberLabel, error) {
	nl := new(NumberLabel)

	if err := nl.init(nl, parent); err != nil {
		return nil, err
	}

	nl.SetTextAlignment(AlignFar)
	if _, err := nl.updateText(); err != nil {
		return nil, err
	}

	nl.MustRegisterProperty("Decimals", NewProperty(
		func() interface{} {
			return nl.Decimals()
		},
		func(v interface{}) error {
			return nl.SetDecimals(assertIntOr(v, 0))
		},
		nl.decimalsChangedPublisher.Event()))

	nl.MustRegisterProperty("Suffix", NewProperty(
		func() interface{} {
			return nl.Suffix()
		},
		func(v interface{}) error {
			return nl.SetSuffix(assertStringOr(v, ""))
		},
		nl.suffixChangedPublisher.Event()))

	nl.MustRegisterProperty("Value", NewProperty(
		func() interface{} {
			return nl.Value()
		},
		func(v interface{}) error {
			return nl.SetValue(assertFloat64Or(v, 0.0))
		},
		nl.valueChangedPublisher.Event()))

	return nl, nil
}

func (nl *NumberLabel) asStatic() *static {
	return &nl.static
}

func (nl *NumberLabel) TextAlignment() Alignment1D {
	return nl.textAlignment1D()
}

func (nl *NumberLabel) SetTextAlignment(alignment Alignment1D) error {
	if alignment == AlignDefault {
		alignment = AlignFar
	}

	return nl.setTextAlignment1D(alignment)
}

func (nl *NumberLabel) Decimals() int {
	return nl.decimals
}

func (nl *NumberLabel) SetDecimals(decimals int) error {
	if decimals == nl.decimals {
		return nil
	}

	old := nl.decimals

	nl.decimals = decimals

	if _, err := nl.updateText(); err != nil {
		nl.decimals = old
		return err
	}

	nl.decimalsChangedPublisher.Publish()

	return nil
}

func (nl *NumberLabel) Suffix() string {
	return nl.suffix
}

func (nl *NumberLabel) SetSuffix(suffix string) error {
	if suffix == nl.suffix {
		return nil
	}

	old := nl.suffix

	nl.suffix = suffix

	if _, err := nl.updateText(); err != nil {
		nl.suffix = old
		return err
	}

	nl.suffixChangedPublisher.Publish()

	return nil
}

func (nl *NumberLabel) Value() float64 {
	return nl.value
}

func (nl *NumberLabel) SetValue(value float64) error {
	if value == nl.value {
		return nil
	}

	old := nl.value

	nl.value = value

	if _, err := nl.updateText(); err != nil {
		nl.value = old
		return err
	}

	nl.valueChangedPublisher.Publish()

	return nil
}

func (nl *NumberLabel) updateText() (changed bool, err error) {
	var sb strings.Builder

	sb.WriteString(FormatFloatGrouped(nl.value, nl.decimals))

	if nl.suffix != "" {
		sb.WriteString(nl.suffix)
	}

	return nl.setText(sb.String())
}
