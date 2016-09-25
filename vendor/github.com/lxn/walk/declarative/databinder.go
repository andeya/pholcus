// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type DataBinder struct {
	AssignTo       **walk.DataBinder
	DataSource     interface{}
	ErrorPresenter ErrorPresenter
	AutoSubmit     bool
	OnSubmitted    walk.EventHandler
}

func (db DataBinder) create() (*walk.DataBinder, error) {
	b := walk.NewDataBinder()

	if db.ErrorPresenter != nil {
		ep, err := db.ErrorPresenter.Create()
		if err != nil {
			return nil, err
		}
		b.SetErrorPresenter(ep)
	}

	b.SetDataSource(db.DataSource)

	b.SetAutoSubmit(db.AutoSubmit)

	if db.OnSubmitted != nil {
		b.Submitted().Attach(db.OnSubmitted)
	}

	if db.AssignTo != nil {
		*db.AssignTo = b
	}

	return b, nil
}
