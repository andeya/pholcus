// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"errors"
	"fmt"
)

import (
	"github.com/lxn/walk"
)

type Shortcut struct {
	Modifiers walk.Modifiers
	Key       walk.Key
}

type Action struct {
	AssignTo    **walk.Action
	Text        string
	Image       interface{}
	Enabled     Property
	Visible     Property
	Shortcut    Shortcut
	OnTriggered walk.EventHandler
	Checkable   bool
}

func (a Action) createAction(builder *Builder, menu *walk.Menu) (*walk.Action, error) {
	action := walk.NewAction()

	if err := action.SetText(a.Text); err != nil {
		return nil, err
	}
	if err := setActionImage(action, a.Image); err != nil {
		return nil, err
	}
	if err := action.SetCheckable(a.Checkable); err != nil {
		return nil, err
	}

	if a.Enabled != nil {
		if b, ok := a.Enabled.(bool); ok {
			if err := action.SetEnabled(b); err != nil {
				return nil, err
			}
		} else if s := builder.conditionOrProperty(a.Enabled); s != nil {
			if c, ok := s.(walk.Condition); ok {
				action.SetEnabledCondition(c)
			} else {
				return nil, fmt.Errorf("value of invalid type bound to Action.Enabled: %T", s)
			}
		}
	}
	if a.Visible != nil {
		if b, ok := a.Visible.(bool); ok {
			if err := action.SetVisible(b); err != nil {
				return nil, err
			}
		} else if s := builder.conditionOrProperty(a.Visible); s != nil {
			if c, ok := s.(walk.Condition); ok {
				action.SetVisibleCondition(c)
			} else {
				return nil, fmt.Errorf("value of invalid type bound to Action.Visible: %T", s)
			}
		}
	}

	s := a.Shortcut
	if err := action.SetShortcut(walk.Shortcut{s.Modifiers, s.Key}); err != nil {
		return nil, err
	}

	if a.OnTriggered != nil {
		action.Triggered().Attach(a.OnTriggered)
	}

	if menu != nil {
		if err := menu.Actions().Add(action); err != nil {
			return nil, err
		}
	}

	if a.AssignTo != nil {
		*a.AssignTo = action
	}

	return action, nil
}

type ActionRef struct {
	Action **walk.Action
}

func (ar ActionRef) createAction(builder *Builder, menu *walk.Menu) (*walk.Action, error) {
	if menu != nil {
		if err := menu.Actions().Add(*ar.Action); err != nil {
			return nil, err
		}
	}

	return *ar.Action, nil
}

type Menu struct {
	AssignTo       **walk.Menu
	AssignActionTo **walk.Action
	Text           string
	Image          interface{}
	Items          []MenuItem
	OnTriggered    walk.EventHandler
}

func (m Menu) createAction(builder *Builder, menu *walk.Menu) (*walk.Action, error) {
	subMenu, err := walk.NewMenu()
	if err != nil {
		return nil, err
	}

	var action *walk.Action
	if menu == nil {
		action = walk.NewMenuAction(subMenu)
	} else if action, err = menu.Actions().AddMenu(subMenu); err != nil {
		return nil, err
	}

	if err := action.SetText(m.Text); err != nil {
		return nil, err
	}
	if err := setActionImage(action, m.Image); err != nil {
		return nil, err
	}

	for _, item := range m.Items {
		if _, err := item.createAction(builder, subMenu); err != nil {
			return nil, err
		}
	}

	if m.OnTriggered != nil {
		action.Triggered().Attach(m.OnTriggered)
	}

	if m.AssignActionTo != nil {
		*m.AssignActionTo = action
	}
	if m.AssignTo != nil {
		*m.AssignTo = subMenu
	}

	return action, nil
}

type Separator struct {
}

func (s Separator) createAction(builder *Builder, menu *walk.Menu) (*walk.Action, error) {
	action := walk.NewSeparatorAction()

	if menu != nil {
		if err := menu.Actions().Add(action); err != nil {
			return nil, err
		}
	}

	return action, nil
}

func addToActionList(list *walk.ActionList, actions []*walk.Action) error {
	for _, a := range actions {
		if err := list.Add(a); err != nil {
			return err
		}
	}

	return nil
}

func setActionImage(action *walk.Action, image interface{}) (err error) {
	var img walk.Image

	switch image := image.(type) {
	case nil:
		return nil

	case *walk.Bitmap:
		img = image

	case string:
		if img, err = imageFromFile(image); err != nil {
			return
		}
	}

	if bmp, ok := img.(*walk.Bitmap); ok {
		return action.SetImage(bmp)
	}

	return errors.New("invalid type for Image")
}
