// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"errors"
)

var (
	ErrInvalidType = errors.New("invalid type")
)

func LogErrors() bool {
	return logErrors
}

func SetLogErrors(v bool) {
	logErrors = v
}

func PanicOnError() bool {
	return panicOnError
}

func SetPanicOnError(v bool) {
	panicOnError = v
}

func TranslationFunc() TranslationFunction {
	return translation
}

func SetTranslationFunc(f TranslationFunction) {
	translation = f
}

type TranslationFunction func(source string, context ...string) string

var translation TranslationFunction

func tr(source string, context ...string) string {
	if translation == nil {
		return source
	}

	return translation(source, context...)
}

type Disposable interface {
	Dispose()
}

type Disposables struct {
	items []Disposable
	done  bool
}

func (d *Disposables) Add(item Disposable) {
	d.items = append(d.items, item)
}

func (d *Disposables) Spare() {
	d.done = true
}

func (d *Disposables) Treat() {
	if d.done {
		return
	}

	for _, item := range d.items {
		item.Dispose()
	}

	d.done = true
}
