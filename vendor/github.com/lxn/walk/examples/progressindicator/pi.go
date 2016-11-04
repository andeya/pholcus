// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"time"
)

import (
	"github.com/lxn/walk"
)

func main() {
	if _, err := RunMyDialog(nil); err != nil {
		log.Fatal(err)
	}
}

type MyDialog struct {
	*walk.Dialog
	ui myDialogUI
}

func (dlg *MyDialog) setState(state walk.PIState) {
	if err := dlg.ProgressIndicator().SetState(state); err != nil {
		log.Print(err)
	}
}

func RunMyDialog(owner walk.Form) (int, error) {
	dlg := new(MyDialog)
	if err := dlg.init(owner); err != nil {
		return 0, err
	}

	dlg.ui.indeterminateBtn.Clicked().Attach(func() {
		fmt.Println("SetState indeterminate")
		dlg.setState(walk.PIIndeterminate)
	})
	dlg.ui.noProgressBtn.Clicked().Attach(func() {
		fmt.Println("SetState noprogress")
		dlg.setState(walk.PINoProgress)
	})

	dlg.ui.normalBtn.Clicked().Attach(func() {
		fmt.Println("SetState normal")
		dlg.setState(walk.PINormal)
	})

	dlg.ui.errBtn.Clicked().Attach(func() {
		fmt.Println("SetState error")
		dlg.setState(walk.PIError)
	})

	dlg.ui.pausedBtn.Clicked().Attach(func() {
		fmt.Println("SetState paused")
		dlg.setState(walk.PIPaused)
	})

	dlg.ui.startBtn.Clicked().Attach(func() {
		go func() {
			dlg.ProgressIndicator().SetTotal(100)
			var i uint32
			for i = 0; i < 100; i++ {
				fmt.Println("SetProgress", i)
				time.Sleep(100 * time.Millisecond)
				if err := dlg.ProgressIndicator().SetCompleted(i); err != nil {
					log.Print(err)
				}
			}
		}()
	})

	return dlg.Run(), nil
}
