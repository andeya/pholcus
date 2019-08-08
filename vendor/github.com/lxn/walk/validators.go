// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"errors"
	"fmt"
	"math"
	"regexp"
)

type Validator interface {
	Validate(v interface{}) error
}

type ValidationError struct {
	title   string
	message string
}

func NewValidationError(title, message string) *ValidationError {
	return &ValidationError{title: title, message: message}
}

func (ve *ValidationError) Title() string {
	return ve.title
}

func (ve *ValidationError) Message() string {
	return ve.message
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("%s - %s", ve.title, ve.message)
}

type RangeValidator struct {
	min float64
	max float64
}

func NewRangeValidator(min, max float64) (*RangeValidator, error) {
	if max < min {
		return nil, errors.New("max < min")
	}

	return &RangeValidator{min: min, max: max}, nil
}

func (rv *RangeValidator) Min() float64 {
	return rv.min
}

func (rv *RangeValidator) Max() float64 {
	return rv.max
}

func (rv *RangeValidator) Reset(min, max float64) error {
	if max < min {
		return errors.New("max < min")
	}

	rv.min, rv.max = min, max

	return nil
}

func (rv *RangeValidator) Validate(v interface{}) error {
	f64 := v.(float64)

	if f64 < rv.min || f64 > rv.max {
		var msg string
		if math.Abs(rv.min-math.Floor(rv.min)) < math.SmallestNonzeroFloat64 &&
			math.Abs(rv.max-math.Floor(rv.max)) < math.SmallestNonzeroFloat64 {

			msg = fmt.Sprintf(tr("Please enter a number from %.f to %.f.", "walk"),
				rv.min, rv.max)
		} else {
			msg = fmt.Sprintf(tr("Please enter a number from %s to %s.", "walk"),
				FormatFloatGrouped(rv.min, 2), FormatFloatGrouped(rv.max, 2))
		}

		return NewValidationError(tr("Number out of allowed range", "walk"), msg)
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
		return NewValidationError(
			tr("Selection Required", "walk"),
			tr("Please select one of the provided options.", "walk"))
	}

	return nil
}
