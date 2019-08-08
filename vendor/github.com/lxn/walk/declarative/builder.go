// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"fmt"
	"reflect"
	"regexp"
)

import (
	"log"
	"strings"

	"github.com/lxn/walk"
	"gopkg.in/Knetic/govaluate.v3"
)

var (
	conditionsByName = make(map[string]walk.Condition)
	propertyRE       = regexp.MustCompile("[A-Za-z]+[0-9A-Za-z]*(\\.[A-Za-z]+[0-9A-Za-z]*)+")
)

func MustRegisterCondition(name string, condition walk.Condition) {
	if name == "" {
		panic(`name == ""`)
	}
	if condition == nil {
		panic("condition == nil")
	}
	if _, ok := conditionsByName[name]; ok {
		panic("name already registered")
	}

	conditionsByName[name] = condition
}

type declWidget struct {
	d Widget
	w walk.Window
}

type Builder struct {
	level                    int
	rows                     int
	columns                  int
	row                      int
	col                      int
	widgetValue              reflect.Value
	parent                   walk.Container
	declWidgets              []declWidget
	name2Window              map[string]walk.Window
	name2DataBinder          map[string]*walk.DataBinder
	deferredFuncs            []func() error
	knownCompositeConditions map[string]walk.Condition
	expressions              map[string]walk.Expression
	functions                map[string]govaluate.ExpressionFunction
}

func NewBuilder(parent walk.Container) *Builder {
	return &Builder{
		parent:                   parent,
		name2Window:              make(map[string]walk.Window),
		name2DataBinder:          make(map[string]*walk.DataBinder),
		knownCompositeConditions: make(map[string]walk.Condition),
		expressions:              make(map[string]walk.Expression),
		functions:                make(map[string]govaluate.ExpressionFunction),
	}
}

func (b *Builder) Parent() walk.Container {
	return b.parent
}

func (b *Builder) Defer(f func() error) {
	b.deferredFuncs = append(b.deferredFuncs, f)
}

func (b *Builder) deferBuildMenuActions(menu *walk.Menu, items []MenuItem) {
	if len(items) > 0 {
		b.Defer(func() error {
			for _, item := range items {
				if _, err := item.createAction(b, menu); err != nil {
					return err
				}
			}

			return nil
		})
	}
}

func (b *Builder) deferBuildActions(actionList *walk.ActionList, items []MenuItem) {
	if len(items) > 0 {
		b.Defer(func() error {
			for _, item := range items {
				action, err := item.createAction(b, nil)
				if err != nil {
					return err
				}
				if err := actionList.Add(action); err != nil {
					return err
				}
			}

			return nil
		})
	}
}

func (b *Builder) InitWidget(d Widget, w walk.Window, customInit func() error) error {
	oldWidgetValue := b.widgetValue
	b.widgetValue = reflect.ValueOf(d)
	b.level++
	defer func() {
		b.widgetValue = oldWidgetValue
		b.level--
	}()

	var succeeded bool
	defer func() {
		if !succeeded {
			w.Dispose()
		}
	}()

	b.declWidgets = append(b.declWidgets, declWidget{d, w})

	// Widget
	if name := b.string("Name"); name != "" {
		w.SetName(name)
		b.name2Window[name] = w
	}

	if val := b.widgetValue.FieldByName("Background"); val.IsValid() {
		if brush := val.Interface(); brush != nil {
			bg, err := brush.(Brush).Create()
			if err != nil {
				return err
			}

			w.SetBackground(bg)
		}
	}

	if val := b.widgetValue.FieldByName("Font"); val.IsValid() {
		if f, err := val.Interface().(Font).Create(); err != nil {
			return err
		} else if f != nil {
			w.SetFont(f)
		}
	}

	if err := w.SetMinMaxSize(b.size("MinSize").toW(), b.size("MaxSize").toW()); err != nil {
		return err
	}

	if contextMenuItems := b.menuItems("ContextMenuItems"); len(contextMenuItems) > 0 {
		cm, err := walk.NewMenu()
		if err != nil {
			return err
		}

		b.deferBuildMenuActions(cm, contextMenuItems)

		w.SetContextMenu(cm)
	}

	if handler := b.eventHandler("OnBoundsChanged"); handler != nil {
		w.BoundsChanged().Attach(handler)
	}

	if handler := b.keyEventHandler("OnKeyDown"); handler != nil {
		w.KeyDown().Attach(handler)
	}

	if handler := b.keyEventHandler("OnKeyPress"); handler != nil {
		w.KeyPress().Attach(handler)
	}

	if handler := b.keyEventHandler("OnKeyUp"); handler != nil {
		w.KeyUp().Attach(handler)
	}

	if handler := b.mouseEventHandler("OnMouseDown"); handler != nil {
		w.MouseDown().Attach(handler)
	}

	if handler := b.mouseEventHandler("OnMouseMove"); handler != nil {
		w.MouseMove().Attach(handler)
	}

	if handler := b.mouseEventHandler("OnMouseUp"); handler != nil {
		w.MouseUp().Attach(handler)
	}

	if handler := b.eventHandler("OnSizeChanged"); handler != nil {
		w.SizeChanged().Attach(handler)
	}

	if db := b.bool("DoubleBuffering"); db {
		w.SetDoubleBuffering(true)
	}

	if rtl := b.bool("RightToLeftReading"); rtl {
		w.SetRightToLeftReading(true)
	}

	row := b.int("Row")
	rowSpan := b.int("RowSpan")
	column := b.int("Column")
	columnSpan := b.int("ColumnSpan")

	rowBackup := row
	columnBackup := column

	if widget, ok := w.(walk.Widget); ok {
		if alignment := b.alignment(); alignment != AlignHVDefault {
			if err := widget.SetAlignment(walk.Alignment2D(alignment)); err != nil {
				return err
			}
		}

		if err := widget.SetAlwaysConsumeSpace(b.bool("AlwaysConsumeSpace")); err != nil {
			return err
		}

		if field := b.widgetValue.FieldByName("GraphicsEffects"); field.IsValid() {
			for _, effect := range field.Interface().([]walk.WidgetGraphicsEffect) {
				widget.GraphicsEffects().Add(effect)
			}
		}

		if p := widget.Parent(); p != nil {
			type SetStretchFactorer interface {
				SetStretchFactor(widget walk.Widget, factor int) error
			}

			stretchFactor := b.int("StretchFactor")

			if stretchFactor < 1 {
				stretchFactor = 1
			}

			switch l := p.Layout().(type) {
			case SetStretchFactorer:
				if err := l.SetStretchFactor(widget, stretchFactor); err != nil {
					return err
				}

			case *walk.GridLayout:
				csf := l.ColumnStretchFactor(column)
				if csf < stretchFactor {
					csf = stretchFactor
				}
				l.SetColumnStretchFactor(column, csf)

				rsf := l.RowStretchFactor(row)
				if rsf < stretchFactor {
					rsf = stretchFactor
				}
				l.SetRowStretchFactor(row, rsf)

				if rowSpan < 1 {
					rowSpan = 1
				}
				if columnSpan < 1 {
					columnSpan = 1
				}

				if b.rows > 0 && column == 0 && row == 0 {
					if b.row+rowSpan > b.rows {
						b.col++
						b.row = 0
					}

					column = b.col
					row = b.row

					b.row += rowSpan
				}

				if b.columns > 0 && row == 0 && column == 0 {
					if b.col+columnSpan > b.columns {
						b.row++
						b.col = 0
					}

					row = b.row
					column = b.col

					b.col += columnSpan
				}

				r := walk.Rectangle{column, row, columnSpan, rowSpan}

				if err := l.SetRange(widget, r); err != nil {
					return err
				}
			}
		}
	}

	oldParent := b.parent

	// Container
	var db *walk.DataBinder
	if wc, ok := w.(walk.Container); ok {
		var layout Layout
		if val := b.widgetValue.FieldByName("Layout"); val.IsValid() {
			layout, _ = val.Interface().(Layout)
		}

		if layout != nil {
			l, err := layout.Create()
			if err != nil {
				return err
			}

			if err := wc.SetLayout(l); err != nil {
				return err
			}
		}

		type DelegateContainerer interface {
			DelegateContainer() walk.Container
		}

		if dc, ok := wc.(DelegateContainerer); ok {
			if parent := dc.DelegateContainer(); parent != nil {
				b.parent = parent
			} else {
				b.parent = wc
			}
		} else {
			b.parent = wc
		}
		defer func() {
			b.parent = oldParent
		}()

		if layout != nil {
			if g, ok := layout.(Grid); ok {
				rowBackup = b.row
				columnBackup = b.col

				rows := b.rows
				columns := b.columns
				defer func() {
					b.rows, b.columns, b.row, b.col = rows, columns, rowBackup+rowSpan, columnBackup+columnSpan
				}()

				b.rows = g.Rows
				b.columns = g.Columns
				b.row = 0
				b.col = 0
			}
		}

		if val := b.widgetValue.FieldByName("Children"); val.IsValid() {
			for _, child := range val.Interface().([]Widget) {
				if err := child.Create(b); err != nil {
					return err
				}
			}
		}

		dataBinder := b.widgetValue.FieldByName("DataBinder").Interface().(DataBinder)

		if dataBinder.AssignTo != nil || dataBinder.DataSource != nil {
			if dataB, err := dataBinder.create(); err != nil {
				return err
			} else {
				db = dataB

				b.name2DataBinder[dataBinder.Name] = db

				if ep := db.ErrorPresenter(); ep != nil {
					if dep, ok := ep.(walk.Disposable); ok {
						wc.AddDisposable(dep)
					}
				}
			}
		}
	}

	// Custom
	if customInit != nil {
		if err := customInit(); err != nil {
			return err
		}
	}

	b.parent = oldParent

	if b.level == 1 {
		if err := b.initProperties(); err != nil {
			return err
		}
	}

	// Call Reset on DataBinder after customInit, so a Dialog gets a chance to first
	// wire up its DefaultButton to the CanSubmitChanged event of a DataBinder.
	if db != nil {
		if wc, ok := w.(walk.Container); ok {
			b.Defer(func() error {
				// FIXME: Currently SetDataBinder must be called after initProperties.
				wc.SetDataBinder(db)

				if db.DataSource() == nil {
					return nil
				}

				return db.Reset()
			})
		}
	}

	if b.level == 1 {
		for _, f := range b.deferredFuncs {
			if err := f(); err != nil {
				return err
			}
		}
	}

	succeeded = true

	return nil
}

func (b *Builder) alignment() Alignment2D {
	fieldValue := b.widgetValue.FieldByName("Alignment")

	if fieldValue.IsValid() {
		return fieldValue.Interface().(Alignment2D)
	}

	return AlignHVDefault
}

func (b *Builder) bool(fieldName string) bool {
	fieldValue := b.widgetValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		return fieldValue.Interface().(bool)
	}

	return false
}

func (b *Builder) eventHandler(fieldName string) walk.EventHandler {
	fieldValue := b.widgetValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		return fieldValue.Interface().(walk.EventHandler)
	}

	return nil
}

func (b *Builder) float64(fieldName string) float64 {
	fieldValue := b.widgetValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		return fieldValue.Interface().(float64)
	}

	return 0
}

func (b *Builder) int(fieldName string) int {
	fieldValue := b.widgetValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		return fieldValue.Interface().(int)
	}

	return 0
}

func (b *Builder) keyEventHandler(fieldName string) walk.KeyEventHandler {
	fieldValue := b.widgetValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		return fieldValue.Interface().(walk.KeyEventHandler)
	}

	return nil
}

func (b *Builder) menuItems(fieldName string) []MenuItem {
	fieldValue := b.widgetValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		return fieldValue.Interface().([]MenuItem)
	}

	return nil
}

func (b *Builder) mouseEventHandler(fieldName string) walk.MouseEventHandler {
	fieldValue := b.widgetValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		return fieldValue.Interface().(walk.MouseEventHandler)
	}

	return nil
}

func (b *Builder) size(fieldName string) Size {
	fieldValue := b.widgetValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		return fieldValue.Interface().(Size)
	}

	return Size{}
}

func (b *Builder) string(fieldName string) string {
	fieldValue := b.widgetValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		return fieldValue.Interface().(string)
	}

	return ""
}

func (b *Builder) initProperties() error {
	for _, dw := range b.declWidgets {
		d, w := dw.d, dw.w

		sv := reflect.ValueOf(d)
		st := sv.Type()
		if st.Kind() != reflect.Struct {
			panic("d must be a struct value")
		}

		wb := w.AsWindowBase()

		fieldCount := st.NumField()
		for i := 0; i < fieldCount; i++ {
			sf := st.Field(i)

			prop := wb.Property(sf.Name)

			switch val := sv.Field(i).Interface().(type) {
			case nil:
				// nop

			case bindData:
				if prop == nil {
					panic(sf.Name + " is not a property")
				}

				src := b.conditionOrProperty(val)

				if src == nil {
					// No luck so far, so we assume the expression refers to
					// something in the data source.
					src = val.expression

					if val.validator != nil {
						validator, err := val.validator.Create()
						if err != nil {
							return err
						}
						if err := prop.SetValidator(validator); err != nil {
							return err
						}
					}
				}

				if err := prop.SetSource(src); err != nil {
					return err
				}

			case walk.Condition:
				if prop == nil {
					panic(sf.Name + " is not a property")
				}

				if err := prop.SetSource(val); err != nil {
					return err
				}

			default:
				if prop == nil {
					continue
				}

				if err := prop.Set(val); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (b *Builder) conditionOrProperty(data Property) interface{} {
	switch val := data.(type) {
	case bindData:
		if val.expression == "" {
			return nil
		}

		e := &expression{
			text:           val.expression,
			subExprsByPath: subExpressions(make(map[string]walk.Expression)),
		}

		var singleExpr walk.Expression

		text := propertyRE.ReplaceAllStringFunc(val.expression, func(s string) string {
			if _, ok := e.subExprsByPath[s]; !ok {
				parts := strings.Split(s, ".")

				if w, ok := b.name2Window[parts[0]]; ok {
					if prop := w.AsWindowBase().Property(parts[1]); prop != nil {
						if len(s) == len(val.expression) {
							singleExpr = prop
							return ""
						}

						if len(parts) == 2 {
							e.addSubExpression(s, prop)
						} else {
							e.addSubExpression(s, walk.NewReflectExpression(prop, s[len(parts[0])+len(parts[1])+2:]))
						}
					} else {
						panic(fmt.Errorf(`invalid sub expression: "%s"`, s))
					}
				} else if db, ok := b.name2DataBinder[parts[0]]; ok {
					e.addSubExpression(s, db.Expression(s[len(parts[0])+1:]))
				} else if expr, ok := b.expressions[parts[0]]; ok {
					e.addSubExpression(s, walk.NewReflectExpression(expr, s[len(parts[0])+1:]))
				}
			}

			return strings.Replace(s, ".", "\\.", -1)
		})

		if singleExpr != nil {
			return singleExpr
		}

		expr, err := govaluate.NewEvaluableExpressionWithFunctions(text, b.functions)
		if err != nil {
			panic(fmt.Errorf(`invalid expression "%s": %s`, e.text, err.Error()))
		}

		for _, token := range expr.Tokens() {
			if token.Kind == govaluate.VARIABLE {
				name := token.Value.(string)
				if c, ok := conditionsByName[name]; ok {
					e.addSubExpression(name, c)
				}
				if x, ok := b.expressions[name]; ok {
					e.addSubExpression(name, x)
				}
			}
		}

		e.expr = expr

		if _, err := e.expr.Eval(e.subExprsByPath); err != nil {
			// We hope for the best and leave it to a DataBinder...
			return nil
		}

		if _, ok := e.Value().(bool); ok {
			return &boolExpression{expression: e}
		}

		return e

	case walk.Expression:
		return val
	}

	return nil
}

type expression struct {
	expr                   *govaluate.EvaluableExpression
	text                   string
	subExprsByPath         subExpressions
	subExprsChangedHandles []int
	changedPublisher       walk.EventPublisher
	lastReportedValue      interface{}
}

type subExpressions map[string]walk.Expression

func (se subExpressions) Get(name string) (interface{}, error) {
	if sub, ok := se[name]; ok {
		return sub.Value(), nil
	}

	return nil, fmt.Errorf(`invalid sub expression: "%s"`, name)
}

func (e *expression) String() string {
	return e.text
}

func (e *expression) Value() interface{} {
	val, err := e.expr.Eval(e.subExprsByPath)
	if err != nil {
		log.Printf(`walk - failed to evaluate expression "%s": %s`, e.text, err.Error())
	}

	e.lastReportedValue = val

	return val
}

func (e *expression) Changed() *walk.Event {
	return e.changedPublisher.Event()
}

func (e *expression) addSubExpression(path string, subExpr walk.Expression) {
	e.subExprsByPath[path] = subExpr

	handle := subExpr.Changed().Attach(func() {
		last := e.lastReportedValue
		if v := e.Value(); v != last {
			e.changedPublisher.Publish()
		}
	})
	e.subExprsChangedHandles = append(e.subExprsChangedHandles, handle)
}

type boolExpression struct {
	*expression
}

func (be *boolExpression) Satisfied() bool {
	satisfied, ok := be.Value().(bool)
	return ok && satisfied
}
