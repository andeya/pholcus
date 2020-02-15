// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"bytes"
	"errors"
)

import (
	"github.com/lxn/walk"
)

type RadioButtonGroup struct {
	Buttons    []RadioButton
	DataMember string
	Optional   bool
}

func (rbg RadioButtonGroup) Create(builder *Builder) error {
	if len(rbg.Buttons) == 0 {
		return nil
	}

	var first *walk.RadioButton

	for _, rb := range rbg.Buttons {
		if first == nil {
			if rb.AssignTo == nil {
				rb.AssignTo = &first
			}
		}

		if err := rb.Create(builder); err != nil {
			return err
		}

		if first == nil {
			first = *rb.AssignTo
		}
	}

	parent := builder.Parent()

	builder.Defer(func() error {
		group := first.Group()

		validator := newRadioButtonGroupValidator(group, parent)

		for _, rb := range group.Buttons() {
			prop := rb.AsWindowBase().Property("CheckedValue")

			if err := prop.SetSource(rbg.DataMember); err != nil {
				return err
			}
			if err := prop.SetValidator(validator); err != nil {
				return err
			}
		}

		return nil
	})

	return nil
}

type radioButtonGroupValidator struct {
	group *walk.RadioButtonGroup
	err   error
}

func newRadioButtonGroupValidator(group *walk.RadioButtonGroup, parent walk.Container) *radioButtonGroupValidator {
	b := new(bytes.Buffer)

	if gb, ok := parent.(*walk.GroupBox); ok {
		b.WriteString(gb.Title())
	} else {
		for i, rb := range group.Buttons() {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteString(rb.Text())
		}
	}

	b.WriteString(": ")

	b.WriteString(tr("A selection is required.", "walk"))

	return &radioButtonGroupValidator{group: group, err: errors.New(b.String())}
}

func (rbgv *radioButtonGroupValidator) Validate(v interface{}) error {
	if rbgv.group.CheckedButton() == nil {
		return rbgv.err
	}

	return nil
}
