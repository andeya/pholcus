// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"fmt"
)

import (
	"strconv"

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
	Checked     Property
	Enabled     Property
	Visible     Property
	Shortcut    Shortcut
	OnTriggered walk.EventHandler
	Checkable   bool
}

func (a Action) createAction(builder *Builder, menu *walk.Menu) (*walk.Action, error) {
	action := walk.NewAction()

	if a.AssignTo != nil {
		*a.AssignTo = action
	}

	if err := action.SetText(a.Text); err != nil {
		return nil, err
	}
	if err := setActionImage(action, a.Image); err != nil {
		return nil, err
	}

	if err := setActionBoolOrCondition(action.SetChecked, action.SetCheckedCondition, a.Checked, "Action.Checked", builder); err != nil {
		return nil, err
	}
	if err := setActionBoolOrCondition(action.SetEnabled, action.SetEnabledCondition, a.Enabled, "Action.Enabled", builder); err != nil {
		return nil, err
	}
	if err := setActionBoolOrCondition(action.SetVisible, action.SetVisibleCondition, a.Visible, "Action.Visible", builder); err != nil {
		return nil, err
	}

	if err := action.SetCheckable(a.Checkable || action.CheckedCondition() != nil); err != nil {
		return nil, err
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
	Enabled        Property
	Visible        Property
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

	if err := setActionBoolOrCondition(action.SetEnabled, action.SetEnabledCondition, m.Enabled, "Menu.Enabled", builder); err != nil {
		return nil, err
	}
	if err := setActionBoolOrCondition(action.SetVisible, action.SetVisibleCondition, m.Visible, "Menu.Visible", builder); err != nil {
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
	var bm *walk.Bitmap

	switch image := image.(type) {
	case nil:
		return nil

	case *walk.Bitmap:
		bm = image

	case int:
		var err error
		if bm, err = walk.Resources.Bitmap(strconv.Itoa(image)); err != nil {
			return err
		}

	case string:
		if bm, err = walk.Resources.Bitmap(image); err != nil {
			return
		}

	default:
		return walk.ErrInvalidType
	}

	return action.SetImage(bm)
}

func setActionBoolOrCondition(setBool func(bool) error, setCond func(walk.Condition), value Property, path string, builder *Builder) error {
	if value != nil {
		if b, ok := value.(bool); ok {
			if err := setBool(b); err != nil {
				return err
			}
		} else if s := builder.conditionOrProperty(value); s != nil {
			if c, ok := s.(walk.Condition); ok {
				setCond(c)
			} else {
				return fmt.Errorf("value of invalid type bound to %s: %T", path, s)
			}
		}
	}

	return nil
}
