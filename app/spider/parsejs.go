package spider

import (
	"encoding/xml"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/robertkrimen/otto"

	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

// 蜘蛛规则解释器模型
type (
	SpiderModle struct {
		Name          string      `xml:"Name"`
		Description   string      `xml:"Description"`
		EnableKeyword bool        `xml:"EnableKeyword"`
		EnableCookie  bool        `xml:"EnableCookie"`
		EnableMaxPage bool        `xml:"EnableMaxPage"`
		Pausetime     int64       `xml:"Pausetime"`
		Namespace     string      `xml:"Namespace>Script"`
		SubNamespace  string      `xml:"SubNamespace>Script"`
		Root          string      `xml:"Root>Script"`
		Trunk         []RuleModle `xml:"Rule"`
	}
	RuleModle struct {
		Name      string `xml:"name,attr"`
		ParseFunc string `xml:"ParseFunc>Script"`
		AidFunc   string `xml:"AidFunc>Script"`
	}
)

func init() {
	for _, m := range getSpiderModles() {
		var sp = &Spider{
			Name:         m.Name,
			Description:  m.Description,
			EnableCookie: m.EnableCookie,
			Pausetime:    m.Pausetime,
			RuleTree:     &RuleTree{Trunk: map[string]*Rule{}},
		}
		if m.EnableKeyword {
			sp.Keyword = KEYWORD
		}
		if m.EnableMaxPage {
			sp.MaxPage = MAXPAGE
		}

		if m.Namespace != "" {
			sp.Namespace = func(self *Spider) string {
				vm := otto.New()
				vm.Set("self", self)
				val, err := vm.Eval(m.Namespace)
				if err != nil {
					logs.Log.Error("%v\n", err)
				}
				s, _ := val.ToString()
				return s
			}
		}

		if m.SubNamespace != "" {
			sp.SubNamespace = func(self *Spider, dataCell map[string]interface{}) string {
				vm := otto.New()
				vm.Set("self", self)
				vm.Set("dataCell", dataCell)
				val, err := vm.Eval(m.SubNamespace)
				if err != nil {
					logs.Log.Error("%v\n", err)
				}
				s, _ := val.ToString()
				return s
			}
		}

		sp.RuleTree.Root = func(ctx *Context) {
			vm := otto.New()
			vm.Set("ctx", ctx)
			_, err := vm.Eval(m.Root)
			if err != nil {
				logs.Log.Error("Root: %v\n", err)
			}
		}

		for _, rule := range m.Trunk {
			r := new(Rule)
			r.ParseFunc = func(parse string) func(*Context) {
				return func(ctx *Context) {
					vm := otto.New()
					vm.Set("ctx", ctx)
					_, err := vm.Eval(parse)
					if err != nil {
						logs.Log.Error("ParseFunc: %v\n", err)
					}
				}
			}(rule.ParseFunc)

			r.AidFunc = func(parse string) func(*Context, map[string]interface{}) interface{} {
				return func(ctx *Context, aid map[string]interface{}) interface{} {
					vm := otto.New()
					vm.Set("ctx", ctx)
					vm.Set("aid", aid)
					val, err := vm.Eval(parse)
					if err != nil {
						logs.Log.Error("AidFunc: %v\n", err)
					}
					return val
				}
			}(rule.ParseFunc)
			sp.RuleTree.Trunk[rule.Name] = r
		}
		sp.Register()
	}
}

func getSpiderModles() (ms []*SpiderModle) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error("%v", p)
		}
	}()
	files, _ := filepath.Glob(path.Join(config.DYNAMIC_SPIDER_DIR, "*.pholcus.html"))
	for _, filename := range files {
		b, err := ioutil.ReadFile(filename)
		util.CheckErrPanic(err)
		var m SpiderModle
		err = xml.Unmarshal(b, &m)
		util.CheckErrPanic(err)
		ms = append(ms, &m)
	}
	return
}
