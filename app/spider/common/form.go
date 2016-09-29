package common

import (
	"net/url"
	"strings"

	"github.com/henrylee2cn/pholcus/common/goquery"

	"github.com/henrylee2cn/pholcus/app/downloader/request"
	. "github.com/henrylee2cn/pholcus/app/spider"
)

// Form is the default form element.
type Form struct {
	ctx       *Context
	rule      string
	selection *goquery.Selection
	method    string
	action    string
	fields    url.Values
	buttons   url.Values
}

// NewForm creates and returns a *Form type.
func NewForm(ctx *Context, rule string, u string, form *goquery.Selection, schemeAndHost ...string) *Form {
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
func (self *Form) Method() string {
	return self.method
}

// Action returns the form action URL.
// The URL will always be absolute.
func (self *Form) Action() string {
	return self.action
}

// Input sets the value of a form field.
func (self *Form) Input(name, value string) *Form {
	if _, ok := self.fields[name]; ok {
		self.fields.Set(name, value)
	}
	return self
}

// Input sets the value of a form field.
func (self *Form) Inputs(kv map[string]string) *Form {
	for k, v := range kv {
		if _, ok := self.fields[k]; ok {
			self.fields.Set(k, v)
		}
	}
	return self
}

// Submit submits the form.
// Clicks the first button in the form, or submits the form without using
// any button when the form does not contain any buttons.
func (self *Form) Submit() bool {
	if len(self.buttons) > 0 {
		for name := range self.buttons {
			return self.Click(name)
		}
	}
	return self.send("", "")
}

// Click submits the form by clicking the button with the given name.
func (self *Form) Click(button string) bool {
	if _, ok := self.buttons[button]; !ok {
		return false
	}
	return self.send(button, self.buttons[button][0])
}

// Dom returns the inner *goquery.Selection.
func (self *Form) Dom() *goquery.Selection {
	return self.selection
}

// send submits the form.
func (self *Form) send(buttonName, buttonValue string) bool {

	values := make(url.Values, len(self.fields)+1)
	for name, vals := range self.fields {
		values[name] = vals
	}
	if buttonName != "" {
		values.Set(buttonName, buttonValue)
	}
	valsStr := values.Encode()
	if self.Method() == "GET" {
		self.ctx.AddQueue(&request.Request{
			Rule:   self.rule,
			Url:    self.Action() + "?" + valsStr,
			Method: self.Method(),
		})
		return true
	} else {
		enctype, _ := self.selection.Attr("enctype")
		if enctype == "multipart/form-data" {
			self.ctx.AddQueue(&request.Request{
				Rule:     self.rule,
				Url:      self.Action(),
				PostData: valsStr,
				Method:   "POST-M",
			})
			return true
		}
		self.ctx.AddQueue(&request.Request{
			Rule:     self.rule,
			Url:      self.Action(),
			PostData: valsStr,
			Method:   self.Method(),
		})
		return true
	}

	return false
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
		name, ok := s.Attr("name")
		if ok {
			typ, ok := s.Attr("type")
			if ok || s.Is("textarea") {
				if typ == "submit" {
					val, ok := s.Attr("value")
					if ok {
						buttons.Add(name, val)
					} else {
						buttons.Add(name, "")
					}
				} else {
					val, ok := s.Attr("value")
					if !ok {
						val = ""
					}
					fields.Add(name, val)
				}
			}
		}
	})

	return fields, buttons
}

func formAttributes(u string, form *goquery.Selection, schemeAndHost ...string) (string, string) {
	method, ok := form.Attr("method")
	if !ok {
		method = "GET"
	}
	action, ok := form.Attr("action")
	if !ok {
		action = u
	}
	if action, ok = MakeUrl(action, schemeAndHost...); !ok {
		return "", ""
	}

	return strings.ToUpper(method), action
}
