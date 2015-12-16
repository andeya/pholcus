package spider

import (
	"math"

	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/app/scheduler"
	"github.com/henrylee2cn/pholcus/common/util"
)

const (
	KEYWORD = util.USE_KEYWORD // 若使用Keyword，则Keyword初始值必须为USE_KEYWORD
	MAXPAGE = math.MaxInt64    // 如希望在规则中自定义控制MaxPage，则MaxPage初始值必须为MAXPAGE
)

// 蜘蛛规则
type Spider struct {
	Id   int    // 所在SpiderList的下标编号，系统自动分配
	Name string // 必须保证全局唯一
	*RuleTree

	//以下为可选成员
	Description  string
	Pausetime    int64  // 暂停区间(随机: Pausetime/2 ~ Pausetime*2)
	EnableCookie bool   // 控制所有请求是否使用cookie记录
	MaxPage      int64  // 为负值时自动在调度中限制请求数，为正值时在规则中自定义控制
	Keyword      string // 如需使用必须附初始值为常量USE_KEYWORD

	// 命名空间相对于数据库名，不依赖具体数据内容，可选
	Namespace func(*Spider) string
	// 子命名空间相对于表名，可依赖具体数据内容，可选
	SubNamespace func(self *Spider, dataCell map[string]interface{}) string

	// 请求矩阵
	ReqMatrix *scheduler.Matrix
}

//采集规则树
type RuleTree struct {
	// 执行入口（树根）
	Root func(*Context)
	// 执行解析过程（树干）
	Trunk map[string]*Rule
}

// 采集规则单元
type Rule struct {
	// 输出结果的字段名列表
	ItemFields []string
	// 内容解析函数
	ParseFunc func(*Context)
	// 通用辅助函数
	AidFunc func(*Context, map[string]interface{}) interface{}
}

// 添加自身到蜘蛛菜单
func (self *Spider) Register() *Spider {
	Menu.Add(self)
	return self
}

// 指定规则的获取结果的字段名列表
func (self *Spider) GetItemFields(rule *Rule) []string {
	return rule.ItemFields
}

// 返回结果字段名的值
// 不存在时返回空字符串
func (self *Spider) GetItemField(rule *Rule, index int) (field string) {
	if index > len(rule.ItemFields)-1 || index < 0 {
		return ""
	}
	return rule.ItemFields[index]
}

// 返回结果字段名的其索引
// 不存在时索引为-1
func (self *Spider) GetItemFieldIndex(rule *Rule, field string) (index int) {
	for idx, v := range rule.ItemFields {
		if v == field {
			return idx
		}
	}
	return -1
}

// 为指定Rule动态追加结果字段名，并返回索引位置
// 已存在时返回原来索引位置
func (self *Spider) UpsertItemField(rule *Rule, field string) (index int) {
	for i, v := range rule.ItemFields {
		if v == field {
			return i
		}
	}
	rule.ItemFields = append(rule.ItemFields, field)
	return len(rule.ItemFields) - 1
}

// 获取蜘蛛名称
func (self *Spider) GetName() string {
	return self.Name
}

// 安全返回指定规则
func (self *Spider) GetRule(ruleName string) (*Rule, bool) {
	rule, found := self.RuleTree.Trunk[ruleName]
	return rule, found
}

// 返回指定规则
func (self *Spider) MustGetRule(ruleName string) *Rule {
	return self.RuleTree.Trunk[ruleName]
}

// 返回规则树
func (self *Spider) GetRules() map[string]*Rule {
	return self.RuleTree.Trunk
}

// 获取蜘蛛描述
func (self *Spider) GetDescription() string {
	return self.Description
}

// 获取蜘蛛ID
func (self *Spider) GetId() int {
	return self.Id
}

// 设置蜘蛛ID
func (self *Spider) SetId(id int) {
	self.Id = id
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
func (self *Spider) GetMaxPage() int64 {
	return self.MaxPage
}

// 设置采集的最大页数
func (self *Spider) SetMaxPage(max int64) {
	self.MaxPage = max
}

// 控制所有请求是否使用cookie
func (self *Spider) GetEnableCookie() bool {
	return self.EnableCookie
}

// 自定义暂停时间 pause[0]~(pause[0]+pause[1])，优先级高于外部传参
// 当且仅当runtime[0]为true时可覆盖现有值
func (self *Spider) SetPausetime(pause int64, runtime ...bool) {
	if self.Pausetime == 0 || len(runtime) > 0 && runtime[0] {
		self.Pausetime = pause
	}
}

// 开始执行蜘蛛
func (self *Spider) Start() {
	self.RuleTree.Root(NewContext(self, nil))
}

// 返回一个自身复制品
func (self *Spider) Copy() *Spider {
	ghost := &Spider{}
	ghost.Name = self.Name

	ghost.RuleTree = &RuleTree{
		Root:  self.Root,
		Trunk: make(map[string]*Rule, len(self.RuleTree.Trunk)),
	}
	for k, v := range self.RuleTree.Trunk {
		ghost.RuleTree.Trunk[k] = new(Rule)

		ghost.RuleTree.Trunk[k].ItemFields = make([]string, len(v.ItemFields))
		copy(ghost.RuleTree.Trunk[k].ItemFields, v.ItemFields)

		ghost.RuleTree.Trunk[k].ParseFunc = v.ParseFunc
		ghost.RuleTree.Trunk[k].AidFunc = v.AidFunc
	}

	ghost.Description = self.Description
	ghost.Pausetime = self.Pausetime
	ghost.EnableCookie = self.EnableCookie
	ghost.MaxPage = self.MaxPage
	ghost.Keyword = self.Keyword

	ghost.Namespace = self.Namespace
	ghost.SubNamespace = self.SubNamespace

	return ghost
}

func (self *Spider) ReqmatrixInit() *Spider {
	if self.MaxPage < 0 {
		self.ReqMatrix = scheduler.NewMatrix(self.Id, self.MaxPage)
		self.MaxPage = 0
	} else {
		self.ReqMatrix = scheduler.NewMatrix(self.Id, math.MinInt64)
	}

	reqs := scheduler.PullFailure(self.GetName())

	for _, req := range reqs {
		req.SetSpiderId(self.Id)
	}
	self.ReqMatrix.SetFailures(reqs)

	return self
}

func (self *Spider) ReqmatrixSetFailure(req *context.Request) bool {
	return self.ReqMatrix.SetFailure(req)
}

func (self *Spider) ReqmatrixPush(req *context.Request) {
	self.ReqMatrix.Push(req)
}

func (self *Spider) ReqmatrixPull() *context.Request {
	return self.ReqMatrix.Pull()
}

func (self *Spider) ReqmatrixUse() {
	self.ReqMatrix.Use()
}

func (self *Spider) ReqmatrixFree() {
	self.ReqMatrix.Free()
}

func (self *Spider) ReqmatrixCanStop() bool {
	return self.ReqMatrix.CanStop()
}
