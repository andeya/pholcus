package common

import (
	"net/url"
	"strings"

	"github.com/andeya/pholcus/common/goquery"

	"github.com/andeya/pholcus/app/downloader/request"
	spider "github.com/andeya/pholcus/app/spider"
)

// Form is the default form element.
type Form struct {
	ctx       *spider.Context
	rule      string
	selection *goquery.Selection
	method    string
	action    string
	fields    url.Values
	buttons   url.Values
}

// NewForm creates and returns a *Form type.
func NewForm(ctx *spider.Context, rule string, u string, form *goquery.Selection, schemeAndHost ...string) *Form {
	fields, buttons := serializeForm(form)
	if len(schemeAndHost) == 0 {
		aurl, _ := url.Parse(u)
		schemeAndHost = append(schemeAndHost, aurl.Scheme+"://"+aurl.Host)
	}
	method, action := formAttributes(u, form, schemeAndHost...)
	if action == "" {
		return nil
	}
	if method == "" {
		method = "GET"
	}
	return &Form{
		ctx:       ctx,
		rule:      rule,
		selection: form,
		method:    method,
		action:    action,
		fields:    fields,
		buttons:   buttons,
	}
}

// Method returns the form method, eg "GET" or "POST" or "POST-M".
func (f *Form) Method() string {
	return f.method
}

// Action returns the form action URL.
// The URL will always be absolute.
func (f *Form) Action() string {
	return f.action
}

// Input sets the value of a form field.
func (f *Form) Input(name, value string) *Form {
	if _, ok := f.fields[name]; ok {
		f.fields.Set(name, value)
	}
	return f
}

// Input sets the value of a form field.
func (f *Form) Inputs(kv map[string]string) *Form {
	for k, v := range kv {
		if _, ok := f.fields[k]; ok {
			f.fields.Set(k, v)
		}
	}
	return f
}

// Submit submits the form.
// Clicks the first button in the form, or submits the form without using
// any button when the form does not contain any buttons.
func (f *Form) Submit() bool {
	if len(f.buttons) > 0 {
		for name := range f.buttons {
			return f.Click(name)
		}
	}
	return f.send("", "")
}

// Click submits the form by clicking the button with the given name.
func (f *Form) Click(button string) bool {
	if _, ok := f.buttons[button]; !ok {
		return false
	}
	return f.send(button, f.buttons[button][0])
}

// Dom returns the inner *goquery.Selection.
func (f *Form) Dom() *goquery.Selection {
	return f.selection
}

// send submits the form.
func (f *Form) send(buttonName, buttonValue string) bool {

	values := make(url.Values, len(f.fields)+1)
	for name, vals := range f.fields {
		values[name] = vals
	}
	if buttonName != "" {
		values.Set(buttonName, buttonValue)
	}
	valsStr := values.Encode()
	if f.Method() == "GET" {
		f.ctx.AddQueue(&request.Request{
			Rule:   f.rule,
			URL:    f.Action() + "?" + valsStr,
			Method: f.Method(),
		})
		return true
	} else {
		enctype := f.selection.Attr("enctype").UnwrapOr("")
		if enctype == "multipart/form-data" {
			f.ctx.AddQueue(&request.Request{
				Rule:     f.rule,
				URL:      f.Action(),
				PostData: valsStr,
				Method:   "POST-M",
			})
			return true
		}
		f.ctx.AddQueue(&request.Request{
			Rule:     f.rule,
			URL:      f.Action(),
			PostData: valsStr,
			Method:   f.Method(),
		})
		return true
	}
}

// Serialize converts the form fields into a url.Values type.
// Returns two url.Value types. The first is the form field values, and the
// second is the form button values.
func serializeForm(sel *goquery.Selection) (url.Values, url.Values) {
	input := sel.Find("input,button,textarea")
	if input.Length() == 0 {
		return url.Values{}, url.Values{}
	}

	fields := make(url.Values)
	buttons := make(url.Values)
	input.Each(func(_ int, s *goquery.Selection) {
		name := s.Attr("name")
		if name.IsSome() {
			typ := s.Attr("type")
			if typ.IsSome() || s.Is("textarea") {
				if typ.UnwrapOr("") == "submit" {
					val := s.Attr("value")
					buttons.Add(name.Unwrap(), val.UnwrapOr(""))
				} else {
					fields.Add(name.Unwrap(), s.Attr("value").UnwrapOr(""))
				}
			}
		}
	})

	return fields, buttons
}

func formAttributes(u string, form *goquery.Selection, schemeAndHost ...string) (string, string) {
	method := form.Attr("method").UnwrapOr("GET")
	action := form.Attr("action").UnwrapOr(u)
	action, ok := MakeUrl(action, schemeAndHost...)
	if !ok {
		return "", ""
	}

	return strings.ToUpper(method), action
}
