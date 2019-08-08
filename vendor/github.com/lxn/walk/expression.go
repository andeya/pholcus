// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"log"
	"reflect"
)

import _ "gopkg.in/Knetic/govaluate.v3"

type Expression interface {
	Value() interface{}
	Changed() *Event
}

type reflectExpression struct {
	root Expression
	path string
}

func NewReflectExpression(root Expression, path string) Expression {
	return &reflectExpression{root: root, path: path}
}

func (re *reflectExpression) Value() interface{} {
	rootVal := re.root.Value()
	if rootVal == nil {
		return nil
	}

	_, val, err := reflectValueFromPath(reflect.ValueOf(rootVal), re.path)
	if err != nil {
		log.Print("walk - reflectExpression.Value - Error: ", err.Error())
	}

	if !val.IsValid() {
		return nil
	}

	return val.Interface()
}

func (re *reflectExpression) Changed() *Event {
	return re.root.Changed()
}
