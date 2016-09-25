// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"errors"
	"fmt"
	"regexp"
)

type Validator interface {
	Validate(v interface{}) error
}

type RangeValidator struct {
	min float64
	max float64
}

func NewRangeValidator(min, max float64) (*RangeValidator, error) {
	if max <= min {
		return nil, errors.New("max <= min")
	}

	return &RangeValidator{min: min, max: max}, nil
}

func (rv *RangeValidator) Min() float64 {
	return rv.min
}

func (rv *RangeValidator) Max() float64 {
	return rv.max
}

func (rv *RangeValidator) Validate(v interface{}) error {
	f64 := v.(float64)

	if f64 < rv.min || f64 > rv.max {
		return errors.New(tr("The number is out of the allowed range.", "walk"))
	}

	return nil
}

type RegexpValidator struct {
	re *regexp.Regexp
}

func NewRegexpValidator(pattern string) (*RegexpValidator, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &RegexpValidator{re}, nil
}

func (rv *RegexpValidator) Pattern() string {
	return rv.re.String()
}

func (rv *RegexpValidator) Validate(v interface{}) error {
	var matched bool

	switch val := v.(type) {
	case string:
		matched = rv.re.MatchString(val)

	case []byte:
		matched = rv.re.Match(val)

	case fmt.Stringer:
		matched = rv.re.MatchString(val.String())

	default:
		panic("Unsupported type")
	}

	if !matched {
		return errors.New(tr("The text does not match the required pattern.", "walk"))
	}

	return nil
}

type selectionRequiredValidator struct {
}

var selectionRequiredValidatorSingleton Validator = selectionRequiredValidator{}

func SelectionRequiredValidator() Validator {
	return selectionRequiredValidatorSingleton
}

func (selectionRequiredValidator) Validate(v interface{}) error {
	if v == nil {
		// For Widgets like ComboBox nil is passed to indicate "no selection".
		return errors.New(tr("A selection is required.", "walk"))
	}

	return nil
}
