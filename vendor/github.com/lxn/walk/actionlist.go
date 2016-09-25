// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type actionListObserver interface {
	onInsertedAction(action *Action) error
	onRemovingAction(action *Action) error
	onClearingActions() error
}

type ActionList struct {
	actions  []*Action
	observer actionListObserver
}

func newActionList(observer actionListObserver) *ActionList {
	if observer == nil {
		panic("observer == nil")
	}

	return &ActionList{observer: observer}
}

func (l *ActionList) Add(action *Action) error {
	return l.Insert(len(l.actions), action)
}

func (l *ActionList) AddMenu(menu *Menu) (*Action, error) {
	return l.InsertMenu(len(l.actions), menu)
}

func (l *ActionList) At(index int) *Action {
	return l.actions[index]
}

func (l *ActionList) Clear() error {
	if err := l.observer.onClearingActions(); err != nil {
		return err
	}

	for _, a := range l.actions {
		a.release()
	}

	l.actions = l.actions[:0]

	return nil
}

func (l *ActionList) Contains(action *Action) bool {
	return l.Index(action) > -1
}

func (l *ActionList) Index(action *Action) int {
	for i, a := range l.actions {
		if a == action {
			return i
		}
	}

	return -1
}

func (l *ActionList) indexInObserver(action *Action) int {
	var idx int

	for _, a := range l.actions {
		if a == action {
			return idx
		}
		if a.Visible() {
			idx++
		}
	}

	return -1
}

func (l *ActionList) Insert(index int, action *Action) error {
	l.actions = append(l.actions, nil)
	copy(l.actions[index+1:], l.actions[index:])
	l.actions[index] = action

	if err := l.observer.onInsertedAction(action); err != nil {
		l.actions = append(l.actions[:index], l.actions[index+1:]...)

		return err
	}

	action.addRef()

	if action.Visible() {
		return l.updateSeparatorVisibility()
	}

	return nil
}

func (l *ActionList) InsertMenu(index int, menu *Menu) (*Action, error) {
	action := NewAction()
	action.menu = menu

	if err := l.Insert(index, action); err != nil {
		return nil, err
	}

	return action, nil
}

func (l *ActionList) Len() int {
	return len(l.actions)
}

func (l *ActionList) Remove(action *Action) error {
	index := l.Index(action)
	if index == -1 {
		return nil
	}

	return l.RemoveAt(index)
}

func (l *ActionList) RemoveAt(index int) error {
	action := l.actions[index]
	if action.Visible() {
		if err := l.observer.onRemovingAction(action); err != nil {
			return err
		}
	}

	action.release()

	l.actions = append(l.actions[:index], l.actions[index+1:]...)

	if action.Visible() {
		return l.updateSeparatorVisibility()
	}

	return nil
}

func (l *ActionList) updateSeparatorVisibility() error {
	var hasCurVisAct bool
	var curVisSep *Action

	for _, a := range l.actions {
		if visible := a.Visible(); a.IsSeparator() {
			toggle := visible != hasCurVisAct

			if toggle {
				visible = !visible
				if err := a.SetVisible(visible); err != nil {
					return err
				}
			}

			if visible {
				curVisSep = a
			}

			hasCurVisAct = false
		} else if visible {
			hasCurVisAct = true
		}
	}

	if !hasCurVisAct && curVisSep != nil {
		return curVisSep.SetVisible(false)
	}

	return nil
}
