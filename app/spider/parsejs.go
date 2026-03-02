package spider

import (
	"encoding/xml"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/robertkrimen/otto"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
)

var scriptTagRe = regexp.MustCompile(`(?s)(<Script[^>]*>)(.*?)(</Script>)`)

// evalScript executes JS and returns Result with Catch.
func evalScript(vm *otto.Otto, script string) (r result.Result[otto.Value]) {
	defer r.Catch()
	val, err := vm.Eval(script)
	result.Ret(val, err).Unwrap()
	return result.Ok(val)
}

// SpiderModle is the XML model for dynamic (JavaScript-based) spider rules.
type (
	SpiderModle struct {
		Name            string      `xml:"Name"`
		Description     string      `xml:"Description"`
		Pausetime       int64       `xml:"Pausetime"`
		EnableLimit     bool        `xml:"EnableLimit"`
		EnableKeyin     bool        `xml:"EnableKeyin"`
		EnableCookie    bool        `xml:"EnableCookie"`
		NotDefaultField bool        `xml:"NotDefaultField"`
		Namespace       string      `xml:"Namespace>Script"`
		SubNamespace    string      `xml:"SubNamespace>Script"`
		Root            string      `xml:"Root>Script"`
		Trunk           []RuleModle `xml:"Rule"`
	}
	// RuleModle is the XML model for a single dynamic rule node.
	RuleModle struct {
		Name      string `xml:"name,attr"`
		ParseFunc string `xml:"ParseFunc>Script"`
		AidFunc   string `xml:"AidFunc>Script"`
	}
)

func init() {
	for _, _m := range getSpiderModles() {
		m := _m
		var sp = &Spider{
			Name:            m.Name,
			Description:     m.Description,
			Pausetime:       m.Pausetime,
			EnableCookie:    m.EnableCookie,
			NotDefaultField: m.NotDefaultField,
			RuleTree:        &RuleTree{Trunk: map[string]*Rule{}},
		}
		if m.EnableLimit {
			sp.Limit = LIMIT
		}
		if m.EnableKeyin {
			sp.Keyin = KEYIN
		}

		if m.Namespace != "" {
			sp.Namespace = func(self *Spider) string {
				vm := otto.New()
				vm.Set("self", self)
				r := evalScript(vm, m.Namespace)
				if r.IsErr() {
					logs.Log.Error(" *     dynamic rule [Namespace]: %v\n", r.UnwrapErr())
					return ""
				}
				s, _ := r.Unwrap().ToString()
				return s
			}
		}

		if m.SubNamespace != "" {
			sp.SubNamespace = func(self *Spider, dataCell map[string]interface{}) string {
				vm := otto.New()
				vm.Set("self", self)
				vm.Set("dataCell", dataCell)
				r := evalScript(vm, m.SubNamespace)
				if r.IsErr() {
					logs.Log.Error(" *     dynamic rule [SubNamespace]: %v\n", r.UnwrapErr())
					return ""
				}
				s, _ := r.Unwrap().ToString()
				return s
			}
		}

		sp.RuleTree.Root = func(ctx *Context) {
			vm := otto.New()
			vm.Set("ctx", ctx)
			r := evalScript(vm, m.Root)
			if r.IsErr() {
				logs.Log.Error(" *     dynamic rule [Root]: %v\n", r.UnwrapErr())
			}
		}

		for _, rule := range m.Trunk {
			r := new(Rule)
			r.ParseFunc = func(parse string) func(*Context) {
				return func(ctx *Context) {
					vm := otto.New()
					vm.Set("ctx", ctx)
					ev := evalScript(vm, parse)
					if ev.IsErr() {
						logs.Log.Error(" *     dynamic rule [ParseFunc]: %v\n", ev.UnwrapErr())
					}
				}
			}(rule.ParseFunc)

			r.AidFunc = func(parse string) func(*Context, map[string]interface{}) interface{} {
				return func(ctx *Context, aid map[string]interface{}) interface{} {
					vm := otto.New()
					vm.Set("ctx", ctx)
					vm.Set("aid", aid)
					ev := evalScript(vm, parse)
					if ev.IsErr() {
						logs.Log.Error(" *     dynamic rule [AidFunc]: %v\n", ev.UnwrapErr())
						return nil
					}
					return ev.Unwrap()
				}
			}(rule.AidFunc)
			sp.RuleTree.Trunk[rule.Name] = r
		}
		sp.Register()
	}
}

// wrapScriptCDATA wraps <Script> tag content in CDATA sections if not already wrapped,
// allowing users to write <, >, & etc. in scripts without manual escaping.
func wrapScriptCDATA(data []byte) []byte {
	return scriptTagRe.ReplaceAllFunc(data, func(match []byte) []byte {
		parts := scriptTagRe.FindSubmatch(match)
		open, body, close_ := parts[1], parts[2], parts[3]
		trimmed := strings.TrimSpace(string(body))
		if strings.HasPrefix(trimmed, "<![CDATA[") {
			return match
		}
		var buf []byte
		buf = append(buf, open...)
		buf = append(buf, "<![CDATA["...)
		buf = append(buf, body...)
		buf = append(buf, "]]>"...)
		buf = append(buf, close_...)
		return buf
	})
}

// getSpiderModles loads all dynamic spider rule files from the configured directory.
func getSpiderModles() (ms []*SpiderModle) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error("panic recovered (dynamic rule parsing): %v\n%s", p, debug.Stack())
		}
	}()
	files, _ := filepath.Glob(path.Join(config.SPIDER_DIR, "*"+config.SPIDER_EXT))
	oldFiles, _ := filepath.Glob(path.Join(config.SPIDER_DIR, "*"+config.SPIDER_EXT_OLD))
	files = append(oldFiles, files...)
	for _, filename := range files {
		b, err := os.ReadFile(filename)
		if err != nil {
			log.Printf("[E] dynamic rule [%s]: %v\n", filename, err)
			continue
		}
		b = wrapScriptCDATA(b)
		var m SpiderModle
		err = xml.Unmarshal(b, &m)
		if err != nil {
			log.Printf("[E] dynamic rule [%s]: %v\n", filename, err)
			continue
		}
		ms = append(ms, &m)
	}
	return
}
