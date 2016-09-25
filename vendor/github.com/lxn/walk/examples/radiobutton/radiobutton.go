// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

import (
	. "github.com/lxn/walk/declarative"
)

type Foo struct {
	Bar string
	Baz int
}

func main() {
	foo := &Foo{"b", 0}

	MainWindow{
		Title:   "Walk RadioButton Example",
		MinSize: Size{320, 240},
		Layout:  VBox{},
		DataBinder: DataBinder{
			DataSource: foo,
			AutoSubmit: true,
			OnSubmitted: func() {
				fmt.Println(foo)
			},
		},
		Children: []Widget{
			// RadioButtonGroup is needed for data binding only.
			RadioButtonGroup{
				DataMember: "Bar",
				Buttons: []RadioButton{
					RadioButton{
						Name:  "aRB",
						Text:  "A",
						Value: "a",
					},
					RadioButton{
						Name:  "bRB",
						Text:  "B",
						Value: "b",
					},
					RadioButton{
						Name:  "cRB",
						Text:  "C",
						Value: "c",
					},
				},
			},
			Label{
				Text:    "A",
				Enabled: Bind("aRB.Checked"),
			},
			Label{
				Text:    "B",
				Enabled: Bind("bRB.Checked"),
			},
			Label{
				Text:    "C",
				Enabled: Bind("cRB.Checked"),
			},
			RadioButtonGroup{
				DataMember: "Baz",
				Buttons: []RadioButton{
					RadioButton{
						Name:  "oneRB",
						Text:  "1",
						Value: 1,
					},
					RadioButton{
						Name:  "twoRB",
						Text:  "2",
						Value: 2,
					},
					RadioButton{
						Name:  "threeRB",
						Text:  "3",
						Value: 3,
					},
				},
			},
			Label{
				Text:    "1",
				Enabled: Bind("oneRB.Checked"),
			},
			Label{
				Text:    "2",
				Enabled: Bind("twoRB.Checked"),
			},
			Label{
				Text:    "3",
				Enabled: Bind("threeRB.Checked"),
			},
		},
	}.Run()
}
