// 蜘蛛，采集规则。
package spider

import (
	"github.com/henrylee2cn/pholcus/crawl/downloader/context"
	"github.com/henrylee2cn/pholcus/crawl/scheduler"
)

const (
	// 如需使用Keyword，则用需用CAN_ADD初始化
	CAN_ADD = " " //注意必须为空格
)

type Spider struct {
	Id          int //所在SpiderList的下标编号
	Name        string
	Description string
	Pausetime   [2]uint //暂停区间Pausetime[0]~Pausetime[0]+Pausetime[1]
	*RuleTree
	//以下为可选成员
	MaxPage int
	Keyword string
	Depth   int
	Proxy   string //代理服务器 example='localhost:80'
}

func (self *Spider) Start(sp *Spider) {
	sp.RuleTree.Root(sp)
}

func (self *Spider) AddMenu() {
	Menu.Add(self)
}

func (self *Spider) GetName() string {
	return self.Name
}

func (self *Spider) GetDescription() string {
	return self.Description
}

func (self *Spider) GetId() int {
	return self.Id
}

func (self *Spider) GetKeyword() string {
	return self.Keyword
}

func (self *Spider) GetMaxPage() int {
	return self.MaxPage
}

func (self *Spider) SetMaxPage(max int) {
	self.MaxPage = max
}

func (self *Spider) GetRules() map[string]*Rule {
	return self.RuleTree.Nodes
}

func (self *Spider) SetPausetime(a, b uint) {
	self.Pausetime = [2]uint{a, b}
}

// 根据响应流运行指定解析规则，不推荐在规则中使用
func (self *Spider) GoRule(resp *context.Response) {
	self.RuleTree.Nodes[resp.GetRuleName()].ParseFunc(self, resp)
}

// 用指定规则解析响应流
func (self *Spider) CallRule(ruleName string, resp *context.Response) {
	resp.SetRuleName(ruleName)
	self.GoRule(resp)
}

// 调用指定规则下辅助函数
func (self *Spider) AidRule(ruleName string, aid map[string]interface{}) interface{} {
	rule := self.RuleTree.Nodes[ruleName]
	return rule.AidFunc(self, aid)
}

// 获取任务规则采集语义字段
func (self *Spider) GetOutFeild(resp *context.Response, index int) string {
	return self.RuleTree.Nodes[resp.GetRuleName()].OutFeild[index]
}

// 获取任意规则采集语义字段
func (self *Spider) ShowOutFeild(ruleName string, index int) string {
	return self.RuleTree.Nodes[ruleName].OutFeild[index]
}

func (self *Spider) LoopAddQueue(loop [2]int, urlFn func(int) []string, param map[string]interface{}) {
	for ; loop[0] < loop[1]; loop[0]++ {
		urls := urlFn(loop[0])
		self.BulkAddQueue(urls, param)
	}
}

func (self *Spider) BulkAddQueue(urls []string, param map[string]interface{}) {
	for _, url := range urls {
		param["url"] = url
		self.AddQueue(param)
	}
}

func (self *Spider) AddQueue(param map[string]interface{}) {
	req := self.NewRequest(param)
	scheduler.Sdl.Push(req)
}

// 生成请求
// param全部参数列表
// req := &Request{
// 	url:           param["url"].(string),     //必填
// 	referer:        "",                        //若有必填
// 	rule:          param["rule"].(string),    //必填
// 	spider:        param["spider"].(string),  //自动填写
// 	method:        param["method"].(string),  //可默认
// 	header:        param["header"],//可默认
// 	cookies:       param["cookies"].([]*http.Cookie),//可默认
// 	postData:      param["postData"].(url.Values),//可默认
// 	canOutsource:  param["canOutsource"].(bool),//可默认
// 	checkRedirect: param["checkRedirect"].(func(req *http.Request, via []*http.Request) error),//可默认
// 	temp:          param["temp"].(map[string]interface{}),//可默认
// }

func (self *Spider) NewRequest(param map[string]interface{}) *context.Request {
	param["spider"] = self.GetName()
	req := context.NewRequest(param)
	req.SetSpiderId(self.GetId())
	return req
}

//采集规则树
type RuleTree struct {
	Spread []string //作为服务器时的请求分发点
	Root   func(*Spider)
	Nodes  map[string]*Rule
}

// 采集规则单元
type Rule struct {
	OutFeild []string //注意：有无字段语义和是否输出数据必须保持一致
	// 内容解析函数
	ParseFunc func(*Spider, *context.Response)
	// 通用辅助函数
	AidFunc func(*Spider, map[string]interface{}) interface{}
}

func (self *Rule) GetOutFeild() []string {
	return self.OutFeild
}
