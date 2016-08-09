package spider

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"

	"github.com/robertkrimen/otto"

	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

// 蜘蛛规则解释器模型
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
	RuleModle struct {
		Name      string `xml:"name,attr"`
		ParseFunc string `xml:"ParseFunc>Script"`
		AidFunc   string `xml:"AidFunc>Script"`
	}
)

func init() {
	for _, _m := range getSpiderModles() {
		m := _m //保证闭包变量
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
				val, err := vm.Eval(m.Namespace)
				if err != nil {
					logs.Log.Error(" *     动态规则  [Namespace]: %v\n", err)
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
					logs.Log.Error(" *     动态规则  [SubNamespace]: %v\n", err)
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
				logs.Log.Error(" *     动态规则  [Root]: %v\n", err)
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
						logs.Log.Error(" *     动态规则  [ParseFunc]: %v\n", err)
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
						logs.Log.Error(" *     动态规则  [AidFunc]: %v\n", err)
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
			log.Printf("[E] HTML动态规则解析: %v\n", p)
		}
	}()
	files, _ := filepath.Glob(path.Join(config.SPIDER_DIR, "*"+config.SPIDER_EXT))
	for _, filename := range files {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("[E] HTML动态规则[%s]: %v\n", filename, err)
			continue
		}
		var m SpiderModle
		err = xml.Unmarshal(b, &m)
		if err != nil {
			log.Printf("[E] HTML动态规则[%s]: %v\n", filename, err)
			continue
		}
		ms = append(ms, &m)
	}
	return
}
