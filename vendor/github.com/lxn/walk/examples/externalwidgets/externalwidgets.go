// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
)

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

const myWidgetWindowClass = "MyWidget Class"

func init() {
	walk.MustRegisterWindowClass(myWidgetWindowClass)
}

func main() {
	var mw *walk.MainWindow

	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "Walk External Widgets Example",
		Size:     Size{400, 300},
		Layout:   HBox{},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	for _, name := range []string{"a", "b", "c"} {
		if w, err := NewMyWidget(mw); err != nil {
			log.Fatal(err)
		} else {
			w.SetName(name)
		}
	}

	mpb, err := NewMyPushButton(mw)
	if err != nil {
		log.Fatal(err)
	}
	mpb.SetText("MyPushButton")

	mw.Run()
}

type MyWidget struct {
	walk.WidgetBase
}

func NewMyWidget(parent walk.Container) (*MyWidget, error) {
	w := new(MyWidget)

	if err := walk.InitWidget(
		w,
		parent,
		myWidgetWindowClass,
		win.WS_VISIBLE,
		0); err != nil {

		return nil, err
	}

	bg, err := walk.NewSolidColorBrush(walk.RGB(0, 255, 0))
	if err != nil {
		return nil, err
	}
	w.SetBackground(bg)

	return w, nil
}

func (*MyWidget) MinSizeHint() walk.Size {
	return walk.Size{50, 50}
}

func (w *MyWidget) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_LBUTTONDOWN:
		log.Printf("%s: WM_LBUTTONDOWN", w.Name())
	}

	return w.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

type MyPushButton struct {
	*walk.PushButton
}

func NewMyPushButton(parent walk.Container) (*MyPushButton, error) {
	pb, err := walk.NewPushButton(parent)
	if err != nil {
		return nil, err
	}

	mpb := &MyPushButton{pb}

	if err := walk.InitWrapperWindow(mpb); err != nil {
		return nil, err
	}

	return mpb, nil
}

func (mpb *MyPushButton) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_LBUTTONDOWN:
		log.Printf("%s: WM_LBUTTONDOWN", mpb.Text())
	}

	return mpb.PushButton.WndProc(hwnd, msg, wParam, lParam)
}
