// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	errValidationFailed = errors.New("validation failed")
)

type ErrorPresenter interface {
	PresentError(err error, widget Widget)
}

type DataBinder struct {
	dataSource                 interface{}
	boundWidgets               []Widget
	properties                 []Property
	property2Widget            map[Property]Widget
	property2ChangedHandle     map[Property]int
	rootExpression             Expression
	path2Expression            map[string]Expression
	errorPresenter             ErrorPresenter
	dataSourceChangedPublisher EventPublisher
	canSubmitChangedPublisher  EventPublisher
	submittedPublisher         EventPublisher
	resetPublisher             EventPublisher
	autoSubmitDelay            time.Duration
	autoSubmitTimer            *time.Timer
	autoSubmit                 bool
	autoSubmitSuspended        bool
	canSubmit                  bool
	inReset                    bool
	dirty                      bool
}

func NewDataBinder() *DataBinder {
	db := new(DataBinder)

	db.rootExpression = &dataBinderRootExpression{db}

	return db
}

func (db *DataBinder) AutoSubmit() bool {
	return db.autoSubmit
}

func (db *DataBinder) SetAutoSubmit(autoSubmit bool) {
	db.autoSubmit = autoSubmit
	if autoSubmit {
		db.canSubmit = true
	}
}

func (db *DataBinder) AutoSubmitDelay() time.Duration {
	return db.autoSubmitDelay
}

func (db *DataBinder) SetAutoSubmitDelay(delay time.Duration) {
	db.autoSubmitDelay = delay
}

func (db *DataBinder) AutoSubmitSuspended() bool {
	return db.autoSubmitSuspended
}

func (db *DataBinder) SetAutoSubmitSuspended(suspended bool) {
	if suspended == db.autoSubmitSuspended {
		return
	}

	db.autoSubmitSuspended = suspended

	if suspended {
		if db.autoSubmitTimer != nil {
			db.autoSubmitTimer.Stop()
		}
	} else {
		db.Submit()
	}
}

func (db *DataBinder) Submitted() *Event {
	return db.submittedPublisher.Event()
}

func (db *DataBinder) DataSource() interface{} {
	return db.dataSource
}

func (db *DataBinder) SetDataSource(dataSource interface{}) error {
	if kind := reflect.ValueOf(dataSource).Kind(); kind != reflect.Func && kind != reflect.Map && kind != reflect.Slice &&
		kind == reflect.ValueOf(db.dataSource).Kind() && dataSource == db.dataSource {
		return nil
	}

	if dataSource != nil {
		if t := reflect.TypeOf(dataSource); t.Kind() != reflect.Map && (t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct) {
			return newError("dataSource must be pointer to struct or map[string]interface{}")
		}
	}

	db.dataSource = dataSource

	db.dataSourceChangedPublisher.Publish()

	return nil
}

type dataBinderRootExpression struct {
	db *DataBinder
}

func (dbre *dataBinderRootExpression) Value() interface{} {
	return dbre.db.dataSource
}

func (dbre *dataBinderRootExpression) Changed() *Event {
	return dbre.db.resetPublisher.Event()
}

func (db *DataBinder) DataSourceChanged() *Event {
	return db.dataSourceChangedPublisher.Event()
}

func (db *DataBinder) BoundWidgets() []Widget {
	return db.boundWidgets
}

func (db *DataBinder) SetBoundWidgets(boundWidgets []Widget) {
	for prop, handle := range db.property2ChangedHandle {
		prop.Changed().Detach(handle)
	}

	db.boundWidgets = boundWidgets

	db.property2Widget = make(map[Property]Widget)
	db.property2ChangedHandle = make(map[Property]int)

	for _, widget := range boundWidgets {
		widget := widget

		for _, prop := range widget.AsWindowBase().name2Property {
			prop := prop
			if _, ok := prop.Source().(string); !ok {
				continue
			}

			db.properties = append(db.properties, prop)
			db.property2Widget[prop] = widget

			db.property2ChangedHandle[prop] = prop.Changed().Attach(func() {
				db.dirty = true

				if db.autoSubmit && !db.autoSubmitSuspended {
					if db.autoSubmitDelay > 0 {
						if db.autoSubmitTimer == nil {
							db.autoSubmitTimer = time.AfterFunc(db.autoSubmitDelay, func() {
								synchronize(func() {
									db.Submit()
								})
							})
						} else {
							db.autoSubmitTimer.Reset(db.autoSubmitDelay)
						}
					} else {
						v := reflect.ValueOf(db.dataSource)
						field := db.fieldBoundToProperty(v, prop)
						if field == nil {
							return
						}

						if err := db.submitProperty(prop, field); err != nil {
							return
						}

						db.submittedPublisher.Publish()
					}
				} else {
					if !db.inReset {
						db.validateProperties()
					}
				}
			})
		}
	}
}

func (db *DataBinder) Expression(path string) Expression {
	if db.path2Expression == nil {
		db.path2Expression = make(map[string]Expression)
	}

	if prop, ok := db.path2Expression[path]; ok {
		return prop
	}

	expr := NewReflectExpression(db.rootExpression, path)

	db.path2Expression[path] = expr

	return expr
}

func (db *DataBinder) validateProperties() {
	var hasError bool

	for _, prop := range db.properties {
		validator := prop.Validator()
		if validator == nil {
			continue
		}

		err := validator.Validate(prop.Get())
		if err != nil {
			hasError = true
		}

		if db.errorPresenter != nil {
			widget := db.property2Widget[prop]

			db.errorPresenter.PresentError(err, widget)
		}
	}

	if hasError == db.canSubmit {
		db.canSubmit = !hasError
		db.canSubmitChangedPublisher.Publish()
	}
}

func (db *DataBinder) ErrorPresenter() ErrorPresenter {
	return db.errorPresenter
}

func (db *DataBinder) SetErrorPresenter(ep ErrorPresenter) {
	db.errorPresenter = ep
}

func (db *DataBinder) CanSubmit() bool {
	return db.canSubmit
}

func (db *DataBinder) CanSubmitChanged() *Event {
	return db.canSubmitChangedPublisher.Event()
}

func (db *DataBinder) Reset() error {
	db.inReset = true
	defer func() {
		db.inReset = false
	}()

	if err := db.forEach(func(prop Property, field DataField) error {
		if f64, ok := prop.Get().(float64); ok {
			switch v := field.Get().(type) {
			case float32:
				f64 = float64(v)

			case float64:
				f64 = v

			case int:
				f64 = float64(v)

			case int8:
				f64 = float64(v)

			case int16:
				f64 = float64(v)

			case int32:
				f64 = float64(v)

			case int64:
				f64 = float64(v)

			case uint:
				f64 = float64(v)

			case uint8:
				f64 = float64(v)

			case uint16:
				f64 = float64(v)

			case uint32:
				f64 = float64(v)

			case uint64:
				f64 = float64(v)

			case uintptr:
				f64 = float64(v)

			default:
				return newError(fmt.Sprintf("Field '%s': Can't convert %T to float64.", prop.Source().(string), field.Get()))
			}

			if err := prop.Set(f64); err != nil {
				return err
			}
		} else {
			if err := prop.Set(field.Get()); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	db.validateProperties()

	db.dirty = false

	db.resetPublisher.Publish()

	return nil
}

func (db *DataBinder) ResetFinished() *Event {
	return db.resetPublisher.Event()
}

func (db *DataBinder) Submit() error {
	if !db.CanSubmit() {
		return errValidationFailed
	}

	if err := db.forEach(func(prop Property, field DataField) error {
		return db.submitProperty(prop, field)
	}); err != nil {
		return err
	}

	db.dirty = false

	db.submittedPublisher.Publish()

	return nil
}

func (db *DataBinder) Dirty() bool {
	return db.dirty
}

func (db *DataBinder) submitProperty(prop Property, field DataField) error {
	if !field.CanSet() {
		// FIXME: handle properly
		return nil
	}

	value := prop.Get()
	if value == nil {
		if _, ok := db.property2Widget[prop].(*RadioButton); ok {
			return nil
		}

		return field.Set(field.Zero())
	}
	if err, ok := value.(error); ok {
		return err
	}

	return field.Set(value)
}

func (db *DataBinder) forEach(f func(prop Property, field DataField) error) error {
	dsv := reflect.ValueOf(db.dataSource)
	if dsv.Kind() == reflect.Ptr && dsv.IsNil() {
		return nil
	}

	for _, prop := range db.properties {
		// if widget := db.property2Widget[prop]; !widget.Visible() {
		// 	continue
		// }

		field := db.fieldBoundToProperty(dsv, prop)
		if field == nil {
			continue
		}

		if err := f(prop, field); err != nil {
			return err
		}
	}

	return nil
}

func (db *DataBinder) fieldBoundToProperty(v reflect.Value, prop Property) DataField {
	if db.dataSource == nil {
		return nilField{prop: prop}
	}

	source, ok := prop.Source().(string)
	if !ok || source == "" {
		return nil
	}

	f, err := dataFieldFromPath(v, source)
	if err != nil {
		panic(fmt.Sprintf("invalid source '%s'", source))
	}

	return f
}

func validateBindingMemberSyntax(member string) error {
	// FIXME
	return nil
}

type DataField interface {
	CanSet() bool
	Get() interface{}
	Set(interface{}) error
	Zero() interface{}
}

func dataFieldFromPath(root reflect.Value, path string) (DataField, error) {
	parent, value, err := reflectValueFromPath(root, path)
	if err != nil {
		return nil, err
	}

	// convert to DataField
	if i, ok := value.Interface().(DataField); ok {
		return i, nil
	}

	return &reflectField{parent: parent, value: value, key: path[strings.LastIndexByte(path, '.')+1:]}, nil
}

func reflectValueFromPath(root reflect.Value, path string) (parent, value reflect.Value, err error) {
	fullPath := path
	value = root

	for path != "" {
		var name string
		name, path = nextPathPart(path)

		var p reflect.Value
		for value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
			p = value
			value = value.Elem()
		}

		switch value.Kind() {
		case reflect.Map:
			parent = value
			value = value.MapIndex(reflect.ValueOf(name))

		case reflect.Struct:
			parent = value

			var fun reflect.Value

			// Try as field first.
			if f := value.FieldByName(name); f.IsValid() {
				switch f.Kind() {
				case reflect.Func:
					fun = f

				case reflect.Interface:
					if fn := f.Elem(); fn.Kind() == reflect.Func {
						fun = fn
					} else {
						value = f
					}

				default:
					value = f
				}
			} else {
				// No field, so let's see if we got a method.
				if p.IsValid() {
					// Try pointer receiver first.
					fun = p.MethodByName(name)
				}

				if !fun.IsValid() {
					// No pointer, try directly.
					fun = value.MethodByName(name)
				}
				if !fun.IsValid() {
					return parent, value, fmt.Errorf("bad member: '%s', path: '%s'", path, fullPath)
				}
			}

			if fun.IsValid() {
				// We assume it takes no args and returns one mandatory value plus
				// maybe an error.
				rvs := fun.Call(nil)
				switch len(rvs) {
				case 1:
					value = rvs[0]

				case 2:
					rv2 := rvs[1].Interface()
					if err, ok := rv2.(error); ok {
						return parent, value, err
					} else if rv2 != nil {
						return parent, value, fmt.Errorf("Second method return value must implement error.")
					}

					value = rvs[0]

				default:
					return parent, value, fmt.Errorf("Method must return a value plus optionally an error: %s", name)
				}
			}
		}
	}

	return parent, value, nil
}

func nextPathPart(p string) (next, remaining string) {
	for i, r := range p {
		if r == '.' {
			return p[:i], p[i+1:]
		}
	}
	return p, ""
}

type nilField struct {
	prop Property
}

func (nilField) CanSet() bool {
	return false
}

func (f nilField) Get() interface{} {
	return f.Zero()
}

func (nilField) Set(interface{}) error {
	return nil
}

func (f nilField) Zero() interface{} {
	return reflect.Zero(reflect.TypeOf(f.prop.Get())).Interface()
}

type reflectField struct {
	parent reflect.Value
	value  reflect.Value
	key    string
}

func (f *reflectField) CanSet() bool {
	if f.parent.IsValid() && f.parent.Kind() == reflect.Map {
		return true
	}

	return f.value.CanSet()
}

func (f *reflectField) Get() interface{} {
	return f.value.Interface()
}

func (f *reflectField) Set(value interface{}) error {
	if f.parent.IsValid() && f.parent.Kind() == reflect.Map {
		f.parent.SetMapIndex(reflect.ValueOf(f.key), reflect.ValueOf(value))
		return nil
	}

	if f64, ok := value.(float64); ok {
		switch f.value.Kind() {
		case reflect.Float32, reflect.Float64:
			f.value.SetFloat(f64)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.value.SetInt(int64(f64))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			f.value.SetUint(uint64(f64))

		default:
			return newError(fmt.Sprintf("Can't convert float64 to %s.", f.value.Type().Name()))
		}

		return nil
	}

	f.value.Set(reflect.ValueOf(value))

	return nil
}

func (f *reflectField) Zero() interface{} {
	return reflect.Zero(f.value.Type()).Interface()
}
