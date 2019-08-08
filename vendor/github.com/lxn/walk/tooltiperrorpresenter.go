// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

var ValidationErrorEffect WidgetGraphicsEffect

type ToolTipErrorPresenter struct {
	toolTip                     *ToolTip
	curWidget                   Widget
	widget2error                map[Widget]error
	trackedBoundsChangedHandles map[Window]int
	form                        Form
	formActivatingHandle        int
	formDeactivatingHandle      int
}

func NewToolTipErrorPresenter() (*ToolTipErrorPresenter, error) {
	tt, err := newToolTip(win.TTS_BALLOON)
	if err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			tt.Dispose()
		}
	}()

	succeeded = true

	return &ToolTipErrorPresenter{
		toolTip:                     tt,
		widget2error:                make(map[Widget]error),
		trackedBoundsChangedHandles: make(map[Window]int),
		formActivatingHandle:        -1,
		formDeactivatingHandle:      -1,
	}, nil
}

func (ttep *ToolTipErrorPresenter) Dispose() {
	if ttep.toolTip != nil {
		ttep.untrack()
		ttep.toolTip.Dispose()
		ttep.toolTip = nil
		if ttep.form != nil {
			ttep.form.AsFormBase().activatingPublisher.event.Detach(ttep.formActivatingHandle)
			ttep.form.AsFormBase().deactivatingPublisher.event.Detach(ttep.formDeactivatingHandle)
			ttep.form = nil
		}
	}
}

func (ttep *ToolTipErrorPresenter) PresentError(err error, widget Widget) {
	if ttep.toolTip == nil {
		return
	}

	if err == nil && widget == ttep.curWidget {
		ttep.untrack()
	}

	if err == nil {
		ttep.toolTip.RemoveTool(widget)
		delete(ttep.widget2error, widget)
	} else {
		ttep.toolTip.addTrackedTool(widget)
		ttep.widget2error[widget] = err
	}

	var found bool
	if widget != nil {
		walkDescendants(widget.Form().AsFormBase().clientComposite, func(w Window) bool {
			wt := w.(Widget)

			if !found {
				if e, ok := ttep.widget2error[wt]; ok {
					err, widget, found = e, wt, true
				}
			}

			if !found && wt == ttep.curWidget || wt != widget || err == nil {
				wt.GraphicsEffects().Remove(ValidationErrorEffect)
			}

			return true
		})
	}

	if found {
		if widget != ttep.curWidget {
			ttep.untrack()
		}

		if ve, ok := err.(*ValidationError); ok {
			ttep.toolTip.SetErrorTitle(ve.title)
			ttep.toolTip.SetText(widget, ve.message)
		} else {
			ttep.toolTip.SetErrorTitle(tr("Invalid Input"))
			ttep.toolTip.SetText(widget, err.Error())
		}

		if widget != ttep.curWidget {
			ttep.track(widget)

			if effects := widget.GraphicsEffects(); !effects.Contains(ValidationErrorEffect) {
				effects.Add(ValidationErrorEffect)
			}
		}
	}
}

func (ttep *ToolTipErrorPresenter) track(widget Widget) {
	var wnd Window

	wnd = widget

	for wnd != nil {
		handle := wnd.AsWindowBase().boundsChangedPublisher.event.Attach(func() {
			ttep.toolTip.track(widget)
		})

		ttep.trackedBoundsChangedHandles[wnd] = handle

		if ttep.form == nil {
			ttep.form = widget.Form()
			ttep.formActivatingHandle = ttep.form.AsFormBase().activatingPublisher.event.Attach(func() {
				ttep.toolTip.track(widget)
			})
			ttep.formDeactivatingHandle = ttep.form.AsFormBase().deactivatingPublisher.event.Attach(func() {
				ttep.toolTip.track(widget)
			})
		}

		if w, ok := wnd.(Widget); ok {
			if parent := w.Parent(); parent != nil {
				wnd = parent
			}
		} else {
			break
		}
	}

	ttep.toolTip.track(widget)

	ttep.curWidget = widget
}

func (ttep *ToolTipErrorPresenter) untrack() {
	if ttep.curWidget == nil {
		return
	}

	ttep.toolTip.untrack(ttep.curWidget)

	for wnd, handle := range ttep.trackedBoundsChangedHandles {
		wnd.AsWindowBase().boundsChangedPublisher.event.Detach(handle)
		delete(ttep.trackedBoundsChangedHandles, wnd)
	}

	ttep.curWidget = nil
}
