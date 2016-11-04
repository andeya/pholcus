// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var textEdit *walk.TextEdit
	MainWindow{
		Title:   "Walk DropFiles Example",
		MinSize: Size{320, 240},
		Layout:  VBox{},
		OnDropFiles: func(files []string) {
			textEdit.SetText(strings.Join(files, "\r\n"))
		},
		Children: []Widget{
			TextEdit{
				AssignTo: &textEdit,
				ReadOnly: true,
				Text:     "Drop files here, from windows explorer...",
			},
		},
	}.Run()
}
