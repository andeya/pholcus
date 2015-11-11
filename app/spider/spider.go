package spider

import (
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
)

const (
	USE = util.USE_KEYWORD // 若使用Keyword，则Keyword初始值必须为USE
)

// 蜘蛛规则
type Spider struct {
	Id   int    // 所在SpiderList的下标编号，系统自动分配
	Name string // 必须保证全局唯一
	*RuleTree

	//以下为可选成员
	Description  string
	Pausetime    [2]uint // 暂停区间Pausetime[0]~Pausetime[0]+Pausetime[1]
	EnableCookie bool    // 控制所有请求是否使用cookie记录
	MaxPage      int     // UI传参而来，可在涉及采集页数控制时使用
	Keyword      string  // 如需使用必须附初始值为常量USE

	proxys    []string // 代理服务器列表 example='localhost:80'
	currProxy int      // 当前服务器索引

	// 命名空间相对于数据库名，不依赖具体数据内容，可选
	Namespace func(*Spider) string
	// 子命名空间相对于表名，可依赖具体数据内容，可选
	SubNamespace func(self *Spider, dataCell map[string]interface{}) string
}

// 添加自身到蜘蛛菜单
func (self *Spider) Register() {
	Menu.Add(self)
}

// 返回采集语义字段
func (self *Spider) IndexOutFeild(ruleName string, index int) (feild string) {
	rule := self.GetRule(ruleName)
	if rule == nil {
		return "？？？"
	}
	if len(rule.OutFeild)-1 < index {
		logs.Log.Error("蜘蛛规则 %s - %s 不存在索引为 %v 的输出字段", self.GetName(), ruleName, index)
		return "？？？"
	}
	return rule.OutFeild[index]
}

// 返回采集语义字段的索引位置，不存在时返回-1
func (self *Spider) FindOutFeild(ruleName string, feild string) (index int) {
	rule := self.GetRule(ruleName)
	if rule == nil {
		return -1
	}
	for i, key := range rule.OutFeild {
		if feild == key {
			return i
		}
	}
	return -1
}

// 为指定Rule动态追加采集语义字段，并返回索引位置
// 已存在时返回原来索引位置
func (self *Spider) AddOutFeild(ruleName string, feild string) (index int) {
	for i, v := range self.GetRule(ruleName).OutFeild {
		if v == feild {
			return i
		}
	}
	self.GetRule(ruleName).AddOutFeild(feild)
	return len(self.GetRule(ruleName).OutFeild) - 1
}

// 设置代理服务器列表
func (self *Spider) SetProxys(proxys []string) {
	self.proxys = proxys
	self.currProxy = len(proxys) - 1
}

// 添加代理服务器
func (self *Spider) AddProxys(proxy ...string) {
	self.proxys = append(self.proxys, proxy...)
	self.currProxy += len(proxy) - 1
}

// 获取代理服务器列表
func (self *Spider) GetProxys() []string {
	return self.proxys
}

// 获取下一个代理服务器
func (self *Spider) GetOneProxy() string {
	self.currProxy++
	if self.currProxy > len(self.proxys)-1 {
		self.currProxy = 0
	}
	return self.proxys[self.currProxy]
}

// 获取蜘蛛名称
func (self *Spider) GetName() string {
	return self.Name
}

// 获取蜘蛛描述
func (self *Spider) GetDescription() string {
	return self.Description
}

// 获取蜘蛛ID
func (self *Spider) GetId() int {
	return self.Id
}

// 获取自定义输入
func (self *Spider) GetKeyword() string {
	return self.Keyword
}

// 设置自定义输入
func (self *Spider) SetKeyword(keyword string) {
	self.Keyword = keyword
}

// 获取采集的最大页数
func (self *Spider) GetMaxPage() int {
	return self.MaxPage
}

// 设置采集的最大页数
func (self *Spider) SetMaxPage(max int) {
	self.MaxPage = max
}

// 控制所有请求是否使用cookie
func (self *Spider) GetEnableCookie() bool {
	return self.EnableCookie
}

// 返回规则树
func (self *Spider) GetRules() map[string]*Rule {
	return self.RuleTree.Trunk
}

// 返回指定规则
func (self *Spider) GetRule(ruleName string) *Rule {
	rule, ok := self.RuleTree.Trunk[ruleName]
	if !ok {
		logs.Log.Error("蜘蛛 %s 不存在规则名 %s", self.GetName(), ruleName)
	}
	return rule
}

// 自定义暂停时间 pause[0]~(pause[0]+pause[1])，优先级高于外部传参
// 当且仅当runtime[0]为true时可覆盖现有值
func (self *Spider) SetPausetime(pause [2]uint, runtime ...bool) {
	if self.Pausetime == [2]uint{} || len(runtime) > 0 && runtime[0] {
		self.Pausetime = pause
	}
}

// 开始执行蜘蛛
func (self *Spider) Start() {
	self.RuleTree.Root(NewContext(self, nil))
}

// 返回一个自身复制品
func (self *Spider) Gost() *Spider {
	gost := &Spider{}
	gost.Id = self.Id
	gost.Name = self.Name

	gost.RuleTree = &RuleTree{
		Root:  self.Root,
		Trunk: make(map[string]*Rule, len(self.RuleTree.Trunk)),
	}
	for k, v := range self.RuleTree.Trunk {
		gost.RuleTree.Trunk[k] = new(Rule)

		gost.RuleTree.Trunk[k].OutFeild = make([]string, len(v.OutFeild))
		copy(gost.RuleTree.Trunk[k].OutFeild, v.OutFeild)

		gost.RuleTree.Trunk[k].ParseFunc = v.ParseFunc
		gost.RuleTree.Trunk[k].AidFunc = v.AidFunc
	}

	gost.Description = self.Description
	gost.Pausetime = self.Pausetime
	gost.EnableCookie = self.EnableCookie
	gost.MaxPage = self.MaxPage
	gost.Keyword = self.Keyword

	gost.Namespace = self.Namespace
	gost.SubNamespace = self.SubNamespace

	gost.proxys = make([]string, len(self.proxys))
	copy(gost.proxys, self.proxys)

	gost.currProxy = self.currProxy

	return gost
}

//采集规则树
type RuleTree struct {
	// 执行入口（树根）
	Root func(*Context)
	// 执行解析过程（树干）
	Trunk map[string]*Rule
}

// 返回指定规则
func (self *RuleTree) GetRule(ruleName string) *Rule {
	rule, ok := self.Trunk[ruleName]
	if !ok {
		logs.Log.Error("不存在规则名 %s", ruleName)
	}
	return rule
}

// 采集规则单元
type Rule struct {
	//输出字段
	OutFeild []string
	// 内容解析函数
	ParseFunc func(*Context)
	// 通用辅助函数
	AidFunc func(*Context, map[string]interface{}) interface{}
}

// 获取全部输出字段
func (self *Rule) GetOutFeild() []string {
	return self.OutFeild
}

// 追加输出字段
func (self *Rule) AddOutFeild(feild string) {
	self.OutFeild = append(self.OutFeild, feild)
}
