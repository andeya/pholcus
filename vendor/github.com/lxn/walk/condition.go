// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

type Condition interface {
	Expression
	Satisfied() bool
}

type MutableCondition struct {
	satisfied        bool
	changedPublisher EventPublisher
}

func NewMutableCondition() *MutableCondition {
	return new(MutableCondition)
}

func (mc *MutableCondition) Value() interface{} {
	return mc.satisfied
}

func (mc *MutableCondition) Satisfied() bool {
	return mc.satisfied
}

func (mc *MutableCondition) SetSatisfied(satisfied bool) error {
	if satisfied == mc.satisfied {
		return nil
	}

	mc.satisfied = satisfied

	mc.changedPublisher.Publish()

	return nil
}

func (mc *MutableCondition) Changed() *Event {
	return mc.changedPublisher.Event()
}

type DelegateCondition struct {
	satisfied func() bool
	changed   *Event
}

func NewDelegateCondition(satisfied func() bool, changed *Event) *DelegateCondition {
	return &DelegateCondition{satisfied, changed}
}

func (dc *DelegateCondition) Value() interface{} {
	return dc.satisfied()
}

func (dc *DelegateCondition) Satisfied() bool {
	return dc.satisfied()
}

func (dc *DelegateCondition) Changed() *Event {
	return dc.changed
}

type compositeCondition struct {
	items               []Condition
	itemsChangedHandles []int
	changedPublisher    EventPublisher
}

func (cc *compositeCondition) init(items []Condition) {
	cc.items = append(cc.items, items...)

	for _, item := range items {
		handle := item.Changed().Attach(func() {
			cc.changedPublisher.Publish()
		})
		cc.itemsChangedHandles = append(cc.itemsChangedHandles, handle)
	}
}

func (cc *compositeCondition) satisfied(all bool) bool {
	for _, item := range cc.items {
		if all != item.Satisfied() {
			return !all
		}
	}

	return all
}

func (cc *compositeCondition) Changed() *Event {
	return cc.changedPublisher.Event()
}

func (cc *compositeCondition) Dispose() {
	for i, item := range cc.items {
		item.Changed().Detach(cc.itemsChangedHandles[i])
	}
}

type allCondition struct {
	compositeCondition
}

func NewAllCondition(items ...Condition) Condition {
	ac := new(allCondition)

	ac.init(items)

	return ac
}

func (ac *allCondition) Value() interface{} {
	return ac.Satisfied()
}

func (ac *allCondition) Satisfied() bool {
	return ac.satisfied(true)
}

type anyCondition struct {
	compositeCondition
}

func NewAnyCondition(items ...Condition) Condition {
	ac := new(anyCondition)

	ac.init(items)

	return ac
}

func (ac *anyCondition) Value() interface{} {
	return ac.Satisfied()
}

func (ac *anyCondition) Satisfied() bool {
	return ac.satisfied(false)
}

type negatedCondition struct {
	other Condition
}

func NewNegatedCondition(other Condition) Condition {
	return &negatedCondition{other}
}

func (nc *negatedCondition) Value() interface{} {
	return nc.Satisfied()
}

func (nc *negatedCondition) Satisfied() bool {
	return !nc.other.Satisfied()
}

func (nc *negatedCondition) Changed() *Event {
	return nc.other.Changed()
}
