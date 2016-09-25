// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	errValidationFailed = errors.New("validation failed")
)

type ErrorPresenter interface {
	PresentError(err error, widget Widget)
}

type DataBinder struct {
	dataSource                interface{}
	boundWidgets              []Widget
	properties                []Property
	property2Widget           map[Property]Widget
	property2ChangedHandle    map[Property]int
	errorPresenter            ErrorPresenter
	canSubmitChangedPublisher EventPublisher
	submittedPublisher        EventPublisher
	autoSubmit                bool
	canSubmit                 bool
	inReset                   bool
	dirty                     bool
}

func NewDataBinder() *DataBinder {
	return new(DataBinder)
}

func (db *DataBinder) AutoSubmit() bool {
	return db.autoSubmit
}

func (db *DataBinder) SetAutoSubmit(autoSubmit bool) {
	db.autoSubmit = autoSubmit
}

func (db *DataBinder) Submitted() *Event {
	return db.submittedPublisher.Event()
}

func (db *DataBinder) DataSource() interface{} {
	return db.dataSource
}

func (db *DataBinder) SetDataSource(dataSource interface{}) error {
	if t := reflect.TypeOf(dataSource); t == nil ||
		t.Kind() != reflect.Ptr ||
		t.Elem().Kind() != reflect.Struct {

		return newError("dataSource must be pointer to struct")
	}

	db.dataSource = dataSource

	return nil
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

				if db.autoSubmit {
					if prop.Get() == nil {
						return
					}

					v := reflect.ValueOf(db.dataSource)
					field := db.fieldBoundToProperty(v, prop)
					if field == nil {
						return
					}

					if err := db.submitProperty(prop, field); err != nil {
						return
					}

					db.submittedPublisher.Publish()
				} else {
					if !db.inReset {
						db.validateProperties()
					}
				}
			})
		}
	}
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
				return newError(fmt.Sprintf("Field '%s': Can't convert %t to float64.", prop.Source().(string), field.Get()))
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

	return nil
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
		// This happens e.g. if CurrentIndex() of a ComboBox returns -1.
		// FIXME: Should we handle this differently?
		return nil
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
		field := db.fieldBoundToProperty(dsv, prop)

		if err := f(prop, field); err != nil {
			return err
		}
	}

	return nil
}

func (db *DataBinder) fieldBoundToProperty(v reflect.Value, prop Property) DataField {
	source := prop.Source().(string)

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
}

func dataFieldFromPath(root reflect.Value, path string) (DataField, error) {
	v, err := reflectValueFromPath(root, path)
	if err != nil {
		return nil, err
	}

	// convert to DataField
	if i, ok := v.Interface().(DataField); ok {
		return i, nil
	}

	return reflectField(v), nil
}

func reflectValueFromPath(root reflect.Value, path string) (reflect.Value, error) {
	v := root

	for path != "" {
		var name string
		name, path = nextPathPart(path)

		var p reflect.Value
		for v.Kind() == reflect.Ptr {
			p = v
			v = v.Elem()
		}

		// Try as field first.
		var f reflect.Value
		if v.Kind() == reflect.Struct {
			f = v.FieldByName(name)
		}
		if f.IsValid() {
			v = f
		} else {
			// No field, so let's see if we got a method.
			var m reflect.Value
			if p.IsValid() {
				// Try pointer receiver first.
				m = p.MethodByName(name)
			}

			if !m.IsValid() {
				// No pointer, try directly.
				m = v.MethodByName(name)
			}
			if !m.IsValid() {
				return v, fmt.Errorf("bad member: '%s'", path, ".")
			}

			// We assume it takes no args and returns one mandatory value plus
			// maybe an error.
			rvs := m.Call(nil)
			switch len(rvs) {
			case 1:
				v = rvs[0]

			case 2:
				rv2 := rvs[1].Interface()
				if err, ok := rv2.(error); ok {
					return v, err
				} else if rv2 != nil {
					return v, fmt.Errorf("Second method return value must implement error.")
				}

				v = rvs[0]

			default:
				return v, fmt.Errorf("Method must return a value plus optionally an error: %s", name)
			}
		}
	}

	return v, nil
}

func nextPathPart(p string) (next, remaining string) {
	for i, r := range p {
		if r == '.' {
			return p[:i], p[i+1:]
		}
	}
	return p, ""
}

type reflectField reflect.Value

func (f reflectField) CanSet() bool {
	return reflect.Value(f).CanSet()
}

func (f reflectField) Get() interface{} {
	return reflect.Value(f).Interface()
}

func (f reflectField) Set(value interface{}) error {
	v := reflect.Value(f)
	if f64, ok := value.(float64); ok {
		switch v.Kind() {
		case reflect.Float32, reflect.Float64:
			v.SetFloat(f64)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v.SetInt(int64(f64))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			v.SetUint(uint64(f64))

		default:
			return newError(fmt.Sprintf("Can't convert float64 to %s.", v.Type().Name()))
		}

		return nil
	}

	v.Set(reflect.ValueOf(value))

	return nil
}
