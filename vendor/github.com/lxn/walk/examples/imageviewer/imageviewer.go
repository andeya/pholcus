// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"path"
	"strings"
)

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	mw := new(MyMainWindow)
	var openAction *walk.Action

	if _, err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "Walk Image Viewer Example",
		MenuItems: []MenuItem{
			Menu{
				Text: "&File",
				Items: []MenuItem{
					Action{
						AssignTo:    &openAction,
						Text:        "&Open",
						Image:       "../img/open.png",
						OnTriggered: mw.openAction_Triggered,
					},
					Separator{},
					Action{
						Text:        "Exit",
						OnTriggered: func() { mw.Close() },
					},
				},
			},
			Menu{
				Text: "&Help",
				Items: []MenuItem{
					Action{
						Text:        "About",
						OnTriggered: mw.aboutAction_Triggered,
					},
				},
			},
		},
		ToolBarItems: []MenuItem{
			ActionRef{&openAction},
		},
		MinSize: Size{320, 240},
		Size:    Size{800, 600},
		Layout:  VBox{MarginsZero: true},
		Children: []Widget{
			TabWidget{
				AssignTo: &mw.tabWidget,
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}

type MyMainWindow struct {
	*walk.MainWindow
	tabWidget    *walk.TabWidget
	prevFilePath string
}

func (mw *MyMainWindow) openAction_Triggered() {
	if err := mw.openImage(); err != nil {
		log.Print(err)
	}
}

func (mw *MyMainWindow) openImage() error {
	dlg := new(walk.FileDialog)

	dlg.FilePath = mw.prevFilePath
	dlg.Filter = "Image Files (*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff)|*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff"
	dlg.Title = "Select an Image"

	if ok, err := dlg.ShowOpen(mw); err != nil {
		return err
	} else if !ok {
		return nil
	}

	mw.prevFilePath = dlg.FilePath

	img, err := walk.NewImageFromFile(dlg.FilePath)
	if err != nil {
		return err
	}

	var succeeded bool
	defer func() {
		if !succeeded {
			img.Dispose()
		}
	}()

	page, err := walk.NewTabPage()
	if err != nil {
		return err
	}

	if page.SetTitle(path.Base(strings.Replace(dlg.FilePath, "\\", "/", -1))); err != nil {
		return err
	}
	page.SetLayout(walk.NewHBoxLayout())

	defer func() {
		if !succeeded {
			page.Dispose()
		}
	}()

	imageView, err := walk.NewImageView(page)
	if err != nil {
		return err
	}

	defer func() {
		if !succeeded {
			imageView.Dispose()
		}
	}()

	if err := imageView.SetImage(img); err != nil {
		return err
	}

	if err := mw.tabWidget.Pages().Add(page); err != nil {
		return err
	}

	if err := mw.tabWidget.SetCurrentIndex(mw.tabWidget.Pages().Len() - 1); err != nil {
		return err
	}

	succeeded = true

	return nil
}

func (mw *MyMainWindow) aboutAction_Triggered() {
	walk.MsgBox(mw, "About", "Walk Image Viewer Example", walk.MsgBoxIconInformation)
}
