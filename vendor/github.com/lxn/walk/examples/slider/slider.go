// Copyright 2016 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var slv, slh *walk.Slider
	var maxEdit, minEdit, valueEdit *walk.NumberEdit

	data := struct{ Min, Max, Value int }{0, 100, 30}

	MainWindow{
		Title:   "Walk Slider Example",
		MinSize: Size{320, 240},
		Layout:  HBox{},
		Children: []Widget{
			Slider{
				AssignTo:    &slv,
				MinValue:    data.Min,
				MaxValue:    data.Max,
				Value:       data.Value,
				Orientation: Vertical,
				OnValueChanged: func() {
					data.Value = slv.Value()
					valueEdit.SetValue(float64(data.Value))

				},
			},
			Composite{
				Layout:        Grid{Columns: 3},
				StretchFactor: 4,
				Children: []Widget{
					Label{Text: "Min value"},
					Label{Text: "Value"},
					Label{Text: "Max value"},
					NumberEdit{
						AssignTo: &minEdit,
						Value:    float64(data.Min),
						OnValueChanged: func() {
							data.Min = int(minEdit.Value())
							slh.SetRange(data.Min, data.Max)
							slv.SetRange(data.Min, data.Max)
						},
					},
					NumberEdit{
						AssignTo: &valueEdit,
						Value:    float64(data.Value),
						OnValueChanged: func() {
							data.Value = int(valueEdit.Value())
							slh.SetValue(data.Value)
							slv.SetValue(data.Value)
						},
					},
					NumberEdit{
						AssignTo: &maxEdit,
						Value:    float64(data.Max),
						OnValueChanged: func() {
							data.Max = int(maxEdit.Value())
							slh.SetRange(data.Min, data.Max)
							slv.SetRange(data.Min, data.Max)
						},
					},
					Slider{
						ColumnSpan: 3,
						AssignTo:   &slh,
						MinValue:   data.Min,
						MaxValue:   data.Max,
						Value:      data.Value,
						OnValueChanged: func() {
							data.Value = slh.Value()
							valueEdit.SetValue(float64(data.Value))
						},
					},
					VSpacer{},
					PushButton{
						ColumnSpan: 3,
						Text:       "Print state",
						OnClicked: func() {
							log.Printf("H: < %d | %d | %d >\n", slh.MinValue(), slh.Value(), slh.MaxValue())
							log.Printf("V: < %d | %d | %d >\n", slv.MinValue(), slv.Value(), slv.MaxValue())
						},
					},
				},
			},
		},
	}.Run()
}
