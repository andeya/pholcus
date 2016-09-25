// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var forceUpdate *bool = flag.Bool("force", false, "forces code generation for up-to-date files")
var translatable *bool = flag.Bool("tr", false, "adds calls to a user provided 'func tr(source string, context ...string) string' that returns a translation of the source argument, using provided context args for disambiguation")

type String struct {
	Text         string `xml:"string"`
	Comment      string `xml:"comment,attr"`
	ExtraComment string `xml:"extracomment,attr"`
}

type UI struct {
	Class         string        `xml:"class"`
	Widget        Widget        `xml:"widget"`
	CustomWidgets CustomWidgets `xml:"customwidgets"`
	TabStops      []string      `xml:"tabstops>tabstop"`
}

type Widget struct {
	Class     string       `xml:"class,attr"`
	Name      string       `xml:"name,attr"`
	Attribute []*Attribute `xml:"attribute"`
	Property  []*Property  `xml:"property"`
	Layout    *Layout      `xml:"layout"`
	Widget    []*Widget    `xml:"widget"`
	AddAction []*AddAction `xml:"addaction"`
	Action    []*Action    `xml:"action"`
	ignored   bool
}

type Layout struct {
	Class    string      `xml:"class,attr"`
	Name     string      `xml:"name,attr"`
	Stretch  string      `xml:"stretch,attr"`
	Property []*Property `xml:"property"`
	Item     []*Item     `xml:"item"`
	ignored  bool
}

type Item struct {
	Row     string  `xml:"row,attr"`
	Column  string  `xml:"column,attr"`
	RowSpan string  `xml:"rowspan,attr"`
	ColSpan string  `xml:"colspan,attr"`
	Widget  *Widget `xml:"widget"`
	Spacer  *Spacer `xml:"spacer"`
}

type Spacer struct {
	Name     string      `xml:"name,attr"`
	Property []*Property `xml:"property"`
}

type AddAction struct {
	Name string `xml:"name,attr"`
}

type Action struct {
	Name     string      `xml:"name,attr"`
	Property []*Property `xml:"property"`
}

type Attribute struct {
	Name string `xml:"name,attr"`
	String
}

type Property struct {
	Name   string    `xml:"name,attr"`
	Bool   bool      `xml:"bool"`
	Enum   string    `xml:"enum"`
	Font   *Font     `xml:"font"`
	Number float64   `xml:"number"`
	Rect   Rectangle `xml:"rect"`
	Set    string    `xml:"set"`
	Size   Size      `xml:"size"`
	String
}

type Font struct {
	Family    string `xml:"family"`
	PointSize int    `xml:"pointsize"`
	Italic    bool   `xml:"italic"`
	Bold      bool   `xml:"bold"`
	Underline bool   `xml:"underline"`
	StrikeOut bool   `xml:"strikeout"`
}

type Rectangle struct {
	X      int `xml:"x"`
	Y      int `xml:"y"`
	Width  int `xml:"width"`
	Height int `xml:"height"`
}

type Size struct {
	Width  int `xml:"width"`
	Height int `xml:"height"`
}

type CustomWidgets struct {
	CustomWidget []*CustomWidget `xml:"customwidget"`
}

type CustomWidget struct {
	Class   string `xml:"class"`
	Extends string `xml:"extends"`
}

func trString(str *String) string {
	if str == nil {
		return ""
	}

	if !*translatable {
		return fmt.Sprintf("`%s`", str.Text)
	}

	buf := new(bytes.Buffer)
	buf.WriteString("tr(`")
	buf.WriteString(str.Text)
	buf.WriteString("`")

	if str.Comment != "" {
		buf.WriteString(", `")
		buf.WriteString(str.Comment)
		buf.WriteString("`")
	}

	if str.ExtraComment != "" {
		buf.WriteString(", `")
		buf.WriteString(str.ExtraComment)
		buf.WriteString("`")
	}

	buf.WriteString(")")

	return buf.String()
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parseUI(reader io.Reader) (*UI, error) {
	ui := &UI{}

	if err := xml.NewDecoder(reader).Decode(ui); err != nil {
		return nil, err
	}

	return ui, nil
}

func writeAttribute(buf *bytes.Buffer, attr *Attribute, qualifiedReceiver string) (err error) {
	switch attr.Name {
	case "title":
		buf.WriteString(fmt.Sprintf(
			"if err := %s.SetTitle(%s); err != nil {\nreturn err\n}\n",
			qualifiedReceiver, trString(&attr.String)))

	default:
		fmt.Printf("Ignoring unsupported attribute: '%s'\n", attr.Name)
		return nil
	}

	return nil
}

func writeAttributes(buf *bytes.Buffer, attrs []*Attribute, qualifiedReceiver string) error {
	for _, attr := range attrs {
		if err := writeAttribute(buf, attr, qualifiedReceiver); err != nil {
			return err
		}
	}

	return nil
}

func writeProperty(buf *bytes.Buffer, prop *Property, qualifiedReceiver string, widget *Widget) (err error) {
	if prop.Name == "windowTitle" && widget != nil && widget.Class == "QWidget" {
		return
	}

	switch prop.Name {
	case "decimals":
		buf.WriteString(fmt.Sprintf("if err := %s.SetDecimals(%d); err != nil {\nreturn err\n}\n", qualifiedReceiver, int(prop.Number)))

	case "echoMode":
		switch prop.Enum {
		case "QLineEdit::Normal":
			// nop

		case "QLineEdit::Password":
			buf.WriteString(fmt.Sprintf("%s.SetPasswordMode(true)\n", qualifiedReceiver))

		default:
			fmt.Printf("Ignoring unsupported echoMode: '%s'\n", prop.Enum)
			return nil
		}

	case "enabled":
		buf.WriteString(fmt.Sprintf("%s.SetEnabled(%t)\n", qualifiedReceiver, prop.Bool))

	case "font":
		f := prop.Font
		family := f.Family
		if family == "" {
			family = "MS Shell Dlg 2"
		}
		pointSize := f.PointSize
		if pointSize == 0 {
			pointSize = 8
		}
		buf.WriteString(fmt.Sprintf("if font, err = walk.NewFont(\"%s\", %d, ",
			family, pointSize))
		included := []bool{f.Bold, f.Italic, f.StrikeOut, f.Underline}
		flags := []string{"walk.FontBold", "walk.FontItalic", "walk.FontStrikeOut", "walk.FontUnderline"}
		var includedFlags []string
		for i := 0; i < len(included); i++ {
			if included[i] {
				includedFlags = append(includedFlags, flags[i])
			}
		}
		if len(includedFlags) == 0 {
			buf.WriteString("0")
		} else {
			buf.WriteString(strings.Join(includedFlags, "|"))
		}
		buf.WriteString(`); err != nil {
			return err
			}
			`)
		buf.WriteString(fmt.Sprintf("%s.SetFont(font)\n", qualifiedReceiver))

	case "geometry":
		if qualifiedReceiver == "w" {
			// Only set client size for top level
			buf.WriteString(fmt.Sprintf(
				`if err := %s.SetClientSize(walk.Size{%d, %d}); err != nil {
			return err
			}
			`,
				qualifiedReceiver, prop.Rect.Width, prop.Rect.Height))
		} else {
			buf.WriteString(fmt.Sprintf(
				`if err := %s.SetBounds(walk.Rectangle{%d, %d, %d, %d}); err != nil {
			return err
			}
			`,
				qualifiedReceiver, prop.Rect.X, prop.Rect.Y, prop.Rect.Width, prop.Rect.Height))
		}

	case "maximumSize", "minimumSize":
		// We do these two guys in writeProperties, because we want to map them
		// to a single method call, if both are present.

	case "maxLength":
		buf.WriteString(fmt.Sprintf("%s.SetMaxLength(%d)\n", qualifiedReceiver, int(prop.Number)))

	case "readOnly":
		buf.WriteString(fmt.Sprintf("%s.SetReadOnly(%t)\n", qualifiedReceiver, prop.Bool))

	case "text":
		buf.WriteString(fmt.Sprintf(
			"if err := %s.SetText(%s); err != nil {\nreturn err\n}\n",
			qualifiedReceiver, trString(&prop.String)))

	case "title", "windowTitle":
		buf.WriteString(fmt.Sprintf(
			"if err := %s.SetTitle(%s); err != nil {\nreturn err\n}\n",
			qualifiedReceiver, trString(&prop.String)))

	case "orientation":
		var orientation string
		switch prop.Enum {
		case "Qt::Horizontal":
			orientation = "walk.Horizontal"

		case "Qt::Vertical":
			orientation = "walk.Vertical"

		default:
			return errors.New(fmt.Sprintf("unknown orientation: '%s'", prop.Enum))
		}

		buf.WriteString(fmt.Sprintf(
			`if err := %s.SetOrientation(%s); err != nil {
			return err
			}
			`,
			qualifiedReceiver, orientation))

	default:
		fmt.Printf("Ignoring unsupported property: '%s'\n", prop.Name)
		return nil
	}

	return
}

func writeProperties(buf *bytes.Buffer, props []*Property, qualifiedReceiver string, widget *Widget) error {
	var minSize, maxSize Size
	var hasMinOrMaxSize bool

	for _, prop := range props {
		if err := writeProperty(buf, prop, qualifiedReceiver, widget); err != nil {
			return err
		}

		if prop.Name == "minimumSize" {
			minSize = prop.Size
			hasMinOrMaxSize = true
		}
		if prop.Name == "maximumSize" {
			maxSize = prop.Size
			hasMinOrMaxSize = true
		}
	}

	if hasMinOrMaxSize {
		buf.WriteString(fmt.Sprintf(
			`if err := %s.SetMinMaxSize(walk.Size{%d, %d}, walk.Size{%d, %d}); err != nil {
			return err
			}
			`,
			qualifiedReceiver, minSize.Width, minSize.Height, maxSize.Width, maxSize.Height))
	}

	return nil
}

func writeItemInitializations(buf *bytes.Buffer, items []*Item, parent *Widget, qualifiedParent string, layout string) error {
	for _, item := range items {
		var itemName string

		if item.Spacer != nil {
			itemName = item.Spacer.Name
			name2Prop := make(map[string]*Property)

			for _, prop := range item.Spacer.Property {
				name2Prop[prop.Name] = prop
			}

			orientation := name2Prop["orientation"]
			sizeType := name2Prop["sizeType"]
			sizeHint := name2Prop["sizeHint"]

			var orientStr string
			var fixedStr string
			var secondParamStr string

			if orientation.Enum == "Qt::Horizontal" {
				orientStr = "H"

				if sizeType != nil && sizeType.Enum == "QSizePolicy::Fixed" {
					fixedStr = "Fixed"
					secondParamStr = fmt.Sprintf(", %d", sizeHint.Size.Width)
				}
			} else {
				orientStr = "V"

				if sizeType != nil && sizeType.Enum == "QSizePolicy::Fixed" {
					fixedStr = "Fixed"
					secondParamStr = fmt.Sprintf(", %d", sizeHint.Size.Height)
				}
			}

			if layout == "" {
				buf.WriteString(fmt.Sprintf(
					`
					// anonymous spacer
					if _, err := walk.New%sSpacer%s(%s%s); err != nil {
					return err
					}
					`,
					orientStr, fixedStr, qualifiedParent, secondParamStr))
			} else {
				buf.WriteString(fmt.Sprintf(
					`
					// %s
					%s, err := walk.New%sSpacer%s(%s%s)
					if err != nil {
					return err
					}
					`,
					itemName, itemName, orientStr, fixedStr, qualifiedParent, secondParamStr))
			}
		}

		if item.Widget != nil && !item.Widget.ignored {
			itemName = fmt.Sprintf("w.ui.%s", item.Widget.Name)
			if err := writeWidgetInitialization(buf, item.Widget, parent, qualifiedParent); err != nil {
				return err
			}
		}

		if layout != "" && itemName != "" && item.Row != "" && item.Column != "" {
			if item.ColSpan == "" {
				item.ColSpan = "1"
			}
			if item.RowSpan == "" {
				item.RowSpan = "1"
			}
			buf.WriteString(fmt.Sprintf(
				`				if err := %s.SetRange(%s, walk.Rectangle{%s, %s, %s, %s}); err != nil {
				return err
				}
				`,
				layout, itemName, item.Column, item.Row, item.ColSpan, item.RowSpan))
		}
	}

	return nil
}

func writeLayoutInitialization(buf *bytes.Buffer, layout *Layout, parent *Widget, qualifiedParent string) error {
	var typ string
	switch layout.Class {
	case "QGridLayout":
		typ = "GridLayout"

	case "QHBoxLayout":
		typ = "HBoxLayout"

	case "QVBoxLayout":
		typ = "VBoxLayout"

	default:
		return errors.New(fmt.Sprintf("unsupported layout type: '%s'", layout.Class))
	}

	buf.WriteString(fmt.Sprintf("%s := walk.New%s()\n",
		layout.Name, typ))

	buf.WriteString(fmt.Sprintf(
		`if err := %s.SetLayout(%s); err != nil {
		return err
		}
		`,
		qualifiedParent, layout.Name))

	spacing := 6
	margL, margT, margR, margB := 9, 9, 9, 9

	for _, prop := range layout.Property {
		switch prop.Name {
		case "spacing":
			spacing = int(prop.Number)

		case "leftMargin":
			margL = int(prop.Number)

		case "topMargin":
			margT = int(prop.Number)

		case "rightMargin":
			margR = int(prop.Number)

		case "bottomMargin":
			margB = int(prop.Number)

		case "margin":
			m := int(prop.Number)
			margL, margT, margR, margB = m, m, m, m
		}
	}

	if margL != 0 || margT != 0 || margR != 0 || margB != 0 {
		buf.WriteString(fmt.Sprintf(
			`if err := %s.SetMargins(walk.Margins{%d, %d, %d, %d}); err != nil {
			return err
			}
			`,
			layout.Name, margL, margT, margR, margB))
	}

	if spacing != 0 {
		buf.WriteString(fmt.Sprintf(
			`if err := %s.SetSpacing(%d); err != nil {
			return err
			}
			`,
			layout.Name, spacing))
	}

	var layoutName string
	if typ == "GridLayout" {
		layoutName = layout.Name
	}

	if err := writeItemInitializations(buf, layout.Item, parent, qualifiedParent, layoutName); err != nil {
		return err
	}

	return nil
}

func writeWidgetInitialization(buf *bytes.Buffer, widget *Widget, parent *Widget, qualifiedParent string) error {
	receiver := fmt.Sprintf("w.ui.%s", widget.Name)

	var typ string
	var custom bool
	switch widget.Class {
	case "QCheckBox":
		typ = "CheckBox"

	case "QComboBox":
		typ = "ComboBox"

	case "QDateEdit":
		typ = "DateEdit"

	case "QDoubleSpinBox", "QSpinBox":
		typ = "NumberEdit"

	case "QFrame":
		typ = "Composite"

	case "QGroupBox":
		typ = "GroupBox"

	case "QLabel":
		typ = "Label"

	case "QLineEdit":
		typ = "LineEdit"

	case "QPlainTextEdit", "QTextEdit":
		typ = "TextEdit"

	case "QProgressBar":
		typ = "ProgressBar"

	case "QPushButton":
		typ = "PushButton"

	case "QRadioButton":
		typ = "RadioButton"

	case "QSplitter":
		typ = "Splitter"

	case "QTabWidget":
		typ = "TabWidget"

	case "QTableView", "QTableWidget":
		typ = "TableView"

	case "QToolButton":
		typ = "ToolButton"

	case "QTreeView", "QTreeWidget":
		typ = "TreeView"

	case "QWebView":
		typ = "WebView"

	case "QWidget":
		if parent != nil && parent.Class == "QTabWidget" {
			typ = "TabPage"
		} else {
			typ = "Composite"
		}

	default:
		// FIXME: We assume this is a custom widget in the same package.
		// We also require a func NewFoo(parent) (*Foo, error).
		typ = widget.Class
		custom = true
	}

	if custom {
		buf.WriteString(fmt.Sprintf(
			`
			// %s
			if %s, err = New%s(%s); err != nil {
			return err
			}
			`,
			widget.Name, receiver, typ, qualifiedParent))
	} else {
		if typ == "TabPage" {
			buf.WriteString(fmt.Sprintf(
				`
				// %s
				if %s, err = walk.NewTabPage(); err != nil {
				return err
				}
				`,
				widget.Name, receiver))
		} else {
			buf.WriteString(fmt.Sprintf(
				`
				// %s
				if %s, err = walk.New%s(%s); err != nil {
				return err
				}
				`,
				widget.Name, receiver, typ, qualifiedParent))
		}
	}

	buf.WriteString(fmt.Sprintf("%s.SetName(\"%s\")\n",
		receiver, widget.Name))

	if err := writeAttributes(buf, widget.Attribute, receiver); err != nil {
		return err
	}

	if err := writeProperties(buf, widget.Property, receiver, widget); err != nil {
		return err
	}

	if widget.Layout != nil && !widget.Layout.ignored {
		if err := writeLayoutInitialization(buf, widget.Layout, widget, receiver); err != nil {
			return err
		}
	}

	if typ == "TabPage" {
		buf.WriteString(fmt.Sprintf(
			`if err := %s.Pages().Add(%s); err != nil {
			return err
			}
			`,
			qualifiedParent, receiver))
	}

	return writeWidgetInitializations(buf, widget.Widget, widget, receiver)
}

func writeWidgetInitializations(buf *bytes.Buffer, widgets []*Widget, parent *Widget, qualifiedParent string) error {
	for _, widget := range widgets {
		if widget.ignored || widget.Class == "QMenuBar" || widget.Class == "QStatusBar" {
			continue
		}

		if err := writeWidgetInitialization(buf, widget, parent, qualifiedParent); err != nil {
			return err
		}
	}

	return nil
}

func writeWidgetDecl(buf *bytes.Buffer, widget *Widget, parent *Widget) error {
	var typ string
	switch widget.Class {
	case "QCheckBox":
		typ = "walk.CheckBox"

	case "QComboBox":
		typ = "walk.ComboBox"

	case "QDateEdit":
		typ = "walk.DateEdit"

	case "QDoubleSpinBox", "QSpinBox":
		typ = "walk.NumberEdit"

	case "QFrame":
		typ = "walk.Composite"

	case "QGroupBox":
		typ = "walk.GroupBox"

	case "QLabel":
		typ = "walk.Label"

	case "QLineEdit":
		typ = "walk.LineEdit"

	case "QPlainTextEdit", "QTextEdit":
		typ = "walk.TextEdit"

	case "QProgressBar":
		typ = "walk.ProgressBar"

	case "QPushButton":
		typ = "walk.PushButton"

	case "QRadioButton":
		typ = "walk.RadioButton"

	case "QSplitter":
		typ = "walk.Splitter"

	case "QTabWidget":
		typ = "walk.TabWidget"

	case "QTableView", "QTableWidget":
		typ = "walk.TableView"

	case "QToolButton":
		typ = "walk.ToolButton"

	case "QTreeView", "QTreeWidget":
		typ = "walk.TreeView"

	case "QWebView":
		typ = "walk.WebView"

	case "QWidget":
		if parent != nil && parent.Class == "QTabWidget" {
			typ = "walk.TabPage"
		} else {
			typ = "walk.Composite"
		}

	default:
		// FIXME: For now, we assume this is a custom widget in the same package
		typ = widget.Class
	}

	buf.WriteString(fmt.Sprintf("%s *%s\n", widget.Name, typ))

	if widget.Layout != nil {
		return writeItemDecls(buf, widget.Layout.Item, widget)
	}

	return writeWidgetDecls(buf, widget.Widget, widget)
}

func writeWidgetDecls(buf *bytes.Buffer, widgets []*Widget, parent *Widget) error {
	for _, widget := range widgets {
		switch widget.Class {
		case "QMenuBar", "QStatusBar":
			continue
		}

		if err := writeWidgetDecl(buf, widget, parent); err != nil {
			return err
		}
	}

	return nil
}

func writeItemDecls(buf *bytes.Buffer, items []*Item, parent *Widget) error {
	for _, item := range items {
		if item.Widget == nil {
			continue
		}

		if err := writeWidgetDecl(buf, item.Widget, parent); err != nil {
			return err
		}
	}

	return nil
}

func writeActionDecl(buf *bytes.Buffer, action *Action) error {
	buf.WriteString(action.Name)
	buf.WriteString(" *walk.Action\n")
	return nil
}

func writeActionDecls(buf *bytes.Buffer, actions []*Action) error {
	for _, action := range actions {
		if err := writeActionDecl(buf, action); err != nil {
			return err
		}
	}

	return nil
}

func writeMenuInitialization(buf *bytes.Buffer, menu *Widget, realActions map[string]bool) error {
	var qualifiedParentMenu string

	if menu.Class == "QMenuBar" {
		buf.WriteString("// Menus\n\n")

		qualifiedParentMenu = "w.Menu()"
	} else {
		qualifiedParentMenu = menu.Name
	}

	for _, addAction := range menu.AddAction {
		if realActions[addAction.Name] {
			buf.WriteString("if err := ")
			buf.WriteString(qualifiedParentMenu)
			buf.WriteString(".Actions().Add(w.ui.actions.")
			buf.WriteString(addAction.Name)
			buf.WriteString(`); err != nil {
				return err
			}
			
			`)
		} else {
			for _, submenu := range menu.Widget {
				if submenu.Name != addAction.Name {
					continue
				}

				buf.WriteString("// ")
				buf.WriteString(submenu.Name)
				buf.WriteString("\n")

				buf.WriteString(submenu.Name)
				buf.WriteString(`, err := walk.NewMenu()
					if err != nil {
						return err
					}
					`)
				submenuActionName := submenu.Name + "Action"

				buf.WriteString(submenuActionName)
				buf.WriteString(", err := ")
				buf.WriteString(qualifiedParentMenu)
				buf.WriteString(".Actions().AddMenu(")
				buf.WriteString(submenu.Name)
				buf.WriteString(`)
					if err != nil {
						return err
					}
					`)

				for _, prop := range submenu.Property {
					if prop.Name == "title" {
						buf.WriteString("if err := ")
						buf.WriteString(submenuActionName)
						buf.WriteString(".SetText(")
						buf.WriteString(trString(&prop.String))
						buf.WriteString(`); err != nil {
							return err
						}
						
						`)
						break
					}
				}

				if err := writeMenuInitialization(buf, submenu, realActions); err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func writeActionInitializations(buf *bytes.Buffer, actions []*Action) error {
	buf.WriteString("\n// Actions\n\n")

	for _, action := range actions {
		qualifiedReceiver := "w.ui.actions." + action.Name

		buf.WriteString("// ")
		buf.WriteString(qualifiedReceiver)
		buf.WriteString("\n")

		buf.WriteString(qualifiedReceiver)
		buf.WriteString(" = walk.NewAction()\n")

		if err := writeProperties(buf, action.Property, qualifiedReceiver, nil); err != nil {
			return err
		}

		buf.WriteString("\n")
	}

	return nil
}

func generateUICode(buf *bytes.Buffer, ui *UI) error {
	// Comment, package decl, imports
	buf.WriteString(
		`// This file was created by ui2walk and may be regenerated.
		// DO NOT EDIT OR YOUR MODIFICATIONS WILL BE LOST!

		package main
		
		import (
			"github.com/lxn/walk"
		)
		
		`)

	// Embed the corresponding Walk type.
	var embeddedType string
	switch ui.Widget.Class {
	case "QMainWindow":
		embeddedType = "MainWindow"

	case "QDialog":
		embeddedType = "Dialog"

	case "QWidget":
		embeddedType = "Composite"

	default:
		return errors.New(fmt.Sprintf("Top level '%s' currently not supported.", ui.Widget.Class))
	}

	genTypeBaseName := strings.ToLower(ui.Class[:1]) + ui.Class[1:]

	if len(ui.Widget.Action) > 0 {
		// This struct will contain actions.
		buf.WriteString(fmt.Sprintf("type %sActions struct {\n", genTypeBaseName))

		writeActionDecls(buf, ui.Widget.Action)

		buf.WriteString("}\n\n")
	}

	// Struct containing all descendant widgets.
	buf.WriteString(fmt.Sprintf("type %sUI struct {\n", genTypeBaseName))

	if len(ui.Widget.Action) > 0 {
		buf.WriteString(fmt.Sprintf("actions %sActions\n", genTypeBaseName))
	}

	// Descendant widget decls
	if ui.Widget.Widget != nil {
		if err := writeWidgetDecls(buf, ui.Widget.Widget, &ui.Widget); err != nil {
			return err
		}
	}

	if ui.Widget.Layout != nil {
		if err := writeItemDecls(buf, ui.Widget.Layout.Item, &ui.Widget); err != nil {
			return err
		}
	}

	// end struct
	buf.WriteString("}\n\n")

	// init func
	switch embeddedType {
	case "MainWindow":
		buf.WriteString(fmt.Sprintf(
			`func (w *%s) init() (err error) {
			if w.MainWindow, err = walk.NewMainWindow()`,
			ui.Widget.Name))

	case "Dialog":
		buf.WriteString(fmt.Sprintf(
			`func (w *%s) init(owner walk.Form) (err error) {
			if w.Dialog, err = walk.NewDialog(owner)`,
			ui.Widget.Name))

	case "Composite":
		buf.WriteString(fmt.Sprintf(
			`func (w *%s) init(parent walk.Container) (err error) {
			if w.Composite, err = walk.NewComposite(parent)`,
			ui.Widget.Name))
	}

	buf.WriteString(fmt.Sprintf(`; err != nil {
			return err
			}
			
			succeeded := false
			defer func(){
				if !succeeded {
					w.Dispose()
				}
			}()
			
			var font *walk.Font
			if font == nil {
				font = nil
			}
			
			w.SetName("%s")
			`,
		ui.Widget.Name))

	if embeddedType == "MainWindow" {
		buf.WriteString(fmt.Sprintf(
			`l := walk.NewVBoxLayout()
			if err := l.SetMargins(walk.Margins{0, 0, 0, 0}); err != nil {
				return err
			}
			if err := w.SetLayout(l); err != nil {
				return err
			}
			`))
	}

	if err := writeProperties(buf, ui.Widget.Property, "w", &ui.Widget); err != nil {
		return err
	}

	// Let's see if we find a QMenuBar widget.
	var menuBar *Widget
	for _, widget := range ui.Widget.Widget {
		if widget.Class == "QMenuBar" {
			menuBar = widget
			break
		}
	}

	if len(ui.Widget.Action) > 0 {
		writeActionInitializations(buf, ui.Widget.Action)

		if menuBar != nil {
			realActions := make(map[string]bool)

			for _, action := range ui.Widget.Action {
				realActions[action.Name] = true
			}

			writeMenuInitialization(buf, menuBar, realActions)
		}
	}

	if ui.Widget.Widget != nil {
		if err := writeWidgetInitializations(buf, ui.Widget.Widget, &ui.Widget, "w"); err != nil {
			return err
		}
	}

	if ui.Widget.Layout != nil {
		if err := writeLayoutInitialization(buf, ui.Widget.Layout, &ui.Widget, "w"); err != nil {
			return err
		}
	}

	buf.WriteString("\n// Tab order\n")
	for i := len(ui.TabStops) - 1; i >= 0; i-- {
		buf.WriteString(fmt.Sprintf(`if err = w.ui.%s.BringToTop(); err != nil {
			return err
		}
		`,
			ui.TabStops[i]))
	}

	// end func
	buf.WriteString(`
		succeeded = true
		
		return nil
		}`)

	return nil
}

func generateLogicCode(buf *bytes.Buffer, ui *UI) error {
	// Comment, package decl, imports
	buf.WriteString(
		`package main
		
		import (
			"github.com/lxn/walk"
		)
		
		`)

	// Embed the corresponding Walk type.
	var embeddedType string
	switch ui.Widget.Class {
	case "QMainWindow":
		embeddedType = "MainWindow"

	case "QDialog":
		embeddedType = "Dialog"

	case "QWidget":
		embeddedType = "Composite"

	default:
		return errors.New(fmt.Sprintf("Top level '%s' currently not supported.", ui.Widget.Class))
	}

	buf.WriteString("type ")
	buf.WriteString(ui.Widget.Name)
	buf.WriteString(" struct {\n*walk.")
	buf.WriteString(embeddedType)
	buf.WriteString("\nui ")
	buf.WriteString(strings.ToLower(ui.Class[:1]) + ui.Class[1:])
	buf.WriteString(`UI
	}
	
	`)

	switch embeddedType {
	case "MainWindow":
		buf.WriteString("func run")
		buf.WriteString(ui.Widget.Name)
		buf.WriteString(`() (int, error) {
		mw := new(`)
		buf.WriteString(ui.Widget.Name)
		buf.WriteString(`)
		if err := mw.init(); err != nil {
			return 0, err
		}
		defer mw.Dispose()
		
		// TODO: Do further required setup, e.g. for event handling, here.
		
		mw.Show()
		
		return mw.Run(), nil
		}
		`)

	case "Dialog":
		buf.WriteString("func run")
		buf.WriteString(ui.Widget.Name)
		buf.WriteString(`(owner walk.Form) (int, error) {
		dlg := new(`)
		buf.WriteString(ui.Widget.Name)
		buf.WriteString(`)
		if err := dlg.init(owner); err != nil {
			return 0, err
		}
		
		`)

		if b := findWidget(&ui.Widget, "QPushButton", []string{"accept", "ok"}); b != nil {
			buf.WriteString("if err := dlg.SetDefaultButton(dlg.ui.")
			buf.WriteString(b.Name)
			buf.WriteString(`); err != nil {
			return 0, err
			}
			
			dlg.ui.`)
			buf.WriteString(b.Name)
			buf.WriteString(`.Clicked().Attach(func(){
			dlg.Accept()
			})
			
			`)
		}

		if b := findWidget(&ui.Widget, "QPushButton", []string{"cancel"}); b != nil {
			buf.WriteString("if err := dlg.SetCancelButton(dlg.ui.")
			buf.WriteString(b.Name)
			buf.WriteString(`); err != nil {
			return 0, err
			}
			
			dlg.ui.`)
			buf.WriteString(b.Name)
			buf.WriteString(`.Clicked().Attach(func(){
			dlg.Cancel()
			})
			
			`)
		}

		buf.WriteString(`// TODO: Do further required setup, e.g. for event handling, here.
		
		return dlg.Run(), nil
		}
		`)

	case "Composite":
		buf.WriteString("func new")
		buf.WriteString(ui.Widget.Name)
		buf.WriteString("(parent walk.Container) (*")
		buf.WriteString(ui.Widget.Name)
		buf.WriteString(`, error) {
		c := new(`)
		buf.WriteString(ui.Widget.Name)
		buf.WriteString(`)
		if err := c.init(parent); err != nil {
			return nil, err
		}
		
		// TODO: Do further required setup, e.g. for event handling, here.
		
		return c, nil
		}
		`)
	}

	return nil
}

func findWidget(parent *Widget, class string, nameSubstrs []string) *Widget {
	find := func(widget *Widget) *Widget {
		if widget.Class == class {
			for _, substr := range nameSubstrs {
				if strings.Contains(widget.Name, substr) {
					return widget
				}
			}
		}

		if w := findWidget(widget, class, nameSubstrs); w != nil {
			return w
		}

		return nil
	}

	for _, widget := range parent.Widget {
		if w := find(widget); w != nil {
			return w
		}
	}

	if parent.Layout != nil {
		for _, item := range parent.Layout.Item {
			if item.Widget != nil {
				if w := find(item.Widget); w != nil {
					return w
				}
			}
		}
	}

	return nil
}

func processFile(uiFilePath string) error {
	goLogicFilePath := uiFilePath[:len(uiFilePath)-3] + ".go"
	goUIFilePath := uiFilePath[:len(uiFilePath)-3] + "_ui.go"

	uiFileInfo, err := os.Stat(uiFilePath)
	if err != nil {
		return err
	}

	goUIFileInfo, err := os.Stat(goUIFilePath)
	if !*forceUpdate && err == nil && !uiFileInfo.ModTime().After(goUIFileInfo.ModTime()) {
		// The go file should be up-to-date
		return nil
	}

	fmt.Printf("Processing '%s'\n", uiFilePath)
	defer fmt.Println("")

	uiFile, err := os.Open(uiFilePath)
	if err != nil {
		return err
	}
	defer uiFile.Close()

	reader := bufio.NewReader(uiFile)

	ui, err := parseUI(reader)
	if err != nil {
		return err
	}

	goLogicFile, err := os.OpenFile(goLogicFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err == nil {
		defer goLogicFile.Close()

		buf := new(bytes.Buffer)

		if err := generateLogicCode(buf, ui); err != nil {
			return err
		}

		if _, err := io.Copy(goLogicFile, buf); err != nil {
			return err
		}
		if err := goLogicFile.Close(); err != nil {
			return err
		}
	}

	goUIFile, err := os.Create(goUIFilePath)
	if err != nil {
		return err
	}
	defer goUIFile.Close()

	buf := new(bytes.Buffer)

	if err := generateUICode(buf, ui); err != nil {
		return err
	}

	if _, err := io.Copy(goUIFile, buf); err != nil {
		return err
	}
	if err := goUIFile.Close(); err != nil {
		return err
	}

	dirPath := os.Getenv("GOBIN")
	if dirPath == "" {
		dirPath = filepath.Join(runtime.GOROOT(), "bin")
	}

	gofmtPath := filepath.Join(dirPath, "gofmt.exe")

	args := []string{gofmtPath, "-w", goUIFilePath}
	if goLogicFile != nil {
		args = append(args, goLogicFilePath)
	}

	gofmt, err := os.StartProcess(gofmtPath, args, &os.ProcAttr{Files: []*os.File{nil, nil, os.Stderr}})
	if err != nil {
		return err
	}
	defer gofmt.Release()

	return nil
}

func processDirectory(dirPath string) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		fullPath := path.Join(dirPath, name)

		fi, err := os.Stat(fullPath)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			if err := processDirectory(fullPath); err != nil {
				return err
			}
		} else if !fi.IsDir() && strings.HasSuffix(name, ".ui") {
			if err := processFile(fullPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	flag.Parse()

	cwd, err := os.Getwd()
	logFatal(err)

	logFatal(processDirectory(cwd))
}
