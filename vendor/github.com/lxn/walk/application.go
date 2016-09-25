// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"time"
)

import (
	"github.com/lxn/win"
)

type Settings interface {
	Get(key string) (string, bool)
	Timestamp(key string) (time.Time, bool)
	Put(key, value string) error
	PutExpiring(key, value string) error
	Remove(key string) error
	ExpireDuration() time.Duration
	SetExpireDuration(expireDuration time.Duration)
	Load() error
	Save() error
}

type Persistable interface {
	Persistent() bool
	SetPersistent(value bool)
	SaveState() error
	RestoreState() error
}

type Application struct {
	organizationName   string
	productName        string
	settings           Settings
	exiting            bool
	exitCode           int
	panickingPublisher ErrorEventPublisher
}

var appSingleton *Application = new(Application)

func App() *Application {
	return appSingleton
}

func (app *Application) OrganizationName() string {
	return app.organizationName
}

func (app *Application) SetOrganizationName(value string) {
	app.organizationName = value
}

func (app *Application) ProductName() string {
	return app.productName
}

func (app *Application) SetProductName(value string) {
	app.productName = value
}

func (app *Application) Settings() Settings {
	return app.settings
}

func (app *Application) SetSettings(value Settings) {
	app.settings = value
}

func (app *Application) Exit(exitCode int) {
	app.exiting = true
	app.exitCode = exitCode
	win.PostQuitMessage(int32(exitCode))
}

func (app *Application) ExitCode() int {
	return app.exitCode
}

func (app *Application) Panicking() *ErrorEvent {
	return app.panickingPublisher.Event()
}
