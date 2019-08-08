// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/lxn/walk"
)

type WebView struct {
	// Window

	Background         Brush
	ContextMenuItems   []MenuItem
	DoubleBuffering    bool
	Enabled            Property
	Font               Font
	MaxSize            Size
	MinSize            Size
	Name               string
	OnBoundsChanged    walk.EventHandler
	OnKeyDown          walk.KeyEventHandler
	OnKeyPress         walk.KeyEventHandler
	OnKeyUp            walk.KeyEventHandler
	OnMouseDown        walk.MouseEventHandler
	OnMouseMove        walk.MouseEventHandler
	OnMouseUp          walk.MouseEventHandler
	OnSizeChanged      walk.EventHandler
	Persistent         bool
	RightToLeftReading bool
	ToolTipText        Property
	Visible            Property

	// Widget

	Alignment          Alignment2D
	AlwaysConsumeSpace bool
	Column             int
	ColumnSpan         int
	GraphicsEffects    []walk.WidgetGraphicsEffect
	Row                int
	RowSpan            int
	StretchFactor      int

	// WebView

	AssignTo                          **walk.WebView
	NativeContextMenuEnabled          Property
	OnBrowserVisibleChanged           walk.EventHandler
	OnCanGoBackChanged                walk.EventHandler
	OnCanGoForwardChanged             walk.EventHandler
	OnDocumentCompleted               walk.StringEventHandler
	OnDocumentTitleChanged            walk.EventHandler
	OnDownloaded                      walk.EventHandler
	OnDownloading                     walk.EventHandler
	OnNativeContextMenuEnabledChanged walk.EventHandler
	OnNavigated                       walk.StringEventHandler
	OnNavigatedError                  walk.WebViewNavigatedErrorEventHandler
	OnNavigating                      walk.WebViewNavigatingEventHandler
	OnNewWindow                       walk.WebViewNewWindowEventHandler
	OnProgressChanged                 walk.EventHandler
	OnQuitting                        walk.EventHandler
	OnShortcutsEnabledChanged         walk.EventHandler
	OnStatusBarVisibleChanged         walk.EventHandler
	OnStatusTextChanged               walk.EventHandler
	OnTheaterModeChanged              walk.EventHandler
	OnToolBarEnabledChanged           walk.EventHandler
	OnToolBarVisibleChanged           walk.EventHandler
	OnURLChanged                      walk.EventHandler
	OnWindowClosing                   walk.WebViewWindowClosingEventHandler
	ShortcutsEnabled                  Property
	URL                               Property
}

func (wv WebView) Create(builder *Builder) error {
	w, err := walk.NewWebView(builder.Parent())
	if err != nil {
		return err
	}

	if wv.AssignTo != nil {
		*wv.AssignTo = w
	}

	return builder.InitWidget(wv, w, func() error {
		if wv.OnBrowserVisibleChanged != nil {
			w.BrowserVisibleChanged().Attach(wv.OnBrowserVisibleChanged)
		}
		if wv.OnCanGoBackChanged != nil {
			w.CanGoBackChanged().Attach(wv.OnCanGoBackChanged)
		}
		if wv.OnCanGoForwardChanged != nil {
			w.CanGoForwardChanged().Attach(wv.OnCanGoForwardChanged)
		}
		if wv.OnDocumentCompleted != nil {
			w.DocumentCompleted().Attach(wv.OnDocumentCompleted)
		}
		if wv.OnDocumentTitleChanged != nil {
			w.DocumentTitleChanged().Attach(wv.OnDocumentTitleChanged)
		}
		if wv.OnDownloaded != nil {
			w.Downloaded().Attach(wv.OnDownloaded)
		}
		if wv.OnDownloading != nil {
			w.Downloading().Attach(wv.OnDownloading)
		}
		if wv.OnNativeContextMenuEnabledChanged != nil {
			w.NativeContextMenuEnabledChanged().Attach(wv.OnNativeContextMenuEnabledChanged)
		}
		if wv.OnNavigated != nil {
			w.Navigated().Attach(wv.OnNavigated)
		}
		if wv.OnNavigatedError != nil {
			w.NavigatedError().Attach(wv.OnNavigatedError)
		}
		if wv.OnNavigating != nil {
			w.Navigating().Attach(wv.OnNavigating)
		}
		if wv.OnNewWindow != nil {
			w.NewWindow().Attach(wv.OnNewWindow)
		}
		if wv.OnProgressChanged != nil {
			w.ProgressChanged().Attach(wv.OnProgressChanged)
		}
		if wv.OnURLChanged != nil {
			w.URLChanged().Attach(wv.OnURLChanged)
		}
		if wv.OnShortcutsEnabledChanged != nil {
			w.ShortcutsEnabledChanged().Attach(wv.OnShortcutsEnabledChanged)
		}
		if wv.OnStatusBarVisibleChanged != nil {
			w.StatusBarVisibleChanged().Attach(wv.OnStatusBarVisibleChanged)
		}
		if wv.OnStatusTextChanged != nil {
			w.StatusTextChanged().Attach(wv.OnStatusTextChanged)
		}
		if wv.OnTheaterModeChanged != nil {
			w.TheaterModeChanged().Attach(wv.OnTheaterModeChanged)
		}
		if wv.OnToolBarEnabledChanged != nil {
			w.ToolBarEnabledChanged().Attach(wv.OnToolBarEnabledChanged)
		}
		if wv.OnToolBarVisibleChanged != nil {
			w.ToolBarVisibleChanged().Attach(wv.OnToolBarVisibleChanged)
		}
		if wv.OnQuitting != nil {
			w.Quitting().Attach(wv.OnQuitting)
		}
		if wv.OnWindowClosing != nil {
			w.WindowClosing().Attach(wv.OnWindowClosing)
		}

		return nil
	})
}
