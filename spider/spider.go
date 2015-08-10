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
	MaxPage   int
	Keyword   string
	UseCookie bool
	Proxy     string //代理服务器 example='localhost:80'
}

// 开始执行
func (self *Spider) Start(sp *Spider) {
	sp.RuleTree.Root(sp)
}

func (self *Spider) Gost() *Spider {
	gost := &Spider{}
	gost.Id = self.Id
	gost.Name = self.Name
	gost.Description = self.Description
	gost.Pausetime = self.Pausetime
	gost.MaxPage = self.MaxPage
	gost.Keyword = self.Keyword
	gost.UseCookie = self.UseCookie
	gost.Proxy = self.Proxy
	gost.RuleTree = &RuleTree{
		Root:  self.Root,
		Nodes: map[string]*Rule{},
	}

	for k, v := range self.RuleTree.Nodes {
		nv := *v
		gost.RuleTree.Nodes[k] = &nv
		gost.RuleTree.Nodes[k].OutFeild = self.RuleTree.Nodes[k].OutFeild
	}
	return gost
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

func (self *Spider) SetKeyword(keyword string) {
	self.Keyword = keyword
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

func (self *Spider) SetPausetime(pause [2]uint) {
	self.Pausetime = pause
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

// 获取当前规则采集语义字段
func (self *Spider) GetOutFeild(resp *context.Response, index int) string {
	return self.RuleTree.Nodes[resp.GetRuleName()].OutFeild[index]
}

// 获取指定规则采集语义字段
func (self *Spider) ShowOutFeild(ruleName string, index int) string {
	return self.RuleTree.Nodes[ruleName].OutFeild[index]
}

// 为指定规则动态追加采集语义字段，速度不如静态字段快
func (self *Spider) AddOutFeild(ruleName string, feild string) {
	for _, v := range self.RuleTree.Nodes[ruleName].OutFeild {
		if v == feild {
			return
		}
	}
	self.RuleTree.Nodes[ruleName].AddOutFeild(feild)
}

func (self *Spider) LoopAddQueue(loop [2]int, urlFn func(int) []string, param map[string]interface{}) {
	for ; loop[0] < loop[1]; loop[0]++ {
		urls := urlFn(loop[0])
		self.BulkAddQueue(urls, param)
	}
}

func (self *Spider) BulkAddQueue(urls []string, param map[string]interface{}) {
	for _, url := range urls {
		param["Url"] = url
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
// 	Url:           param["Url"].(string),     //必填
// 	Referer:        "",                       //为输出字段，根据需要选填
// 	Rule:          param["Rule"].(string),    //必填
// 	Spider:        param["Spider"].(string),  //自动填写
// 	Method:        param["Method"].(string),  //默认为GET
// 	Header:        param["Header"],//可默认
// 	Cookies:       param["Cookies"].([]*http.Cookie),//默认为空
// 	PostData:      param["PostData"].(url.Values),//post方法时用，默认为空
// 	CheckRedirect: param["CheckRedirect"].(func(req *http.Request, via []*http.Request) error),//默认为空
// 	Temp:          param["Temp"].(map[string]interface{}),//默认为空
// 	Priority:      param["Priority"].(int),//队列优先级，默认为0
// }

func (self *Spider) NewRequest(param map[string]interface{}) *context.Request {
	param["Spider"] = self.GetName()
	req := context.NewRequest(param)
	req.SetSpiderId(self.GetId())
	return req
}

//采集规则树
type RuleTree struct {
	// 执行入口
	Root func(*Spider)
	// 执行过程
	Nodes map[string]*Rule
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

func (self *Rule) AddOutFeild(feild string) {
	self.OutFeild = append(self.OutFeild, feild)
}
