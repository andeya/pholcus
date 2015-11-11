package spider

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/app/scheduler"
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
	EnableCookie bool    // 是否启用cookie记录
	MaxPage      int     // UI传参而来，可在涉及采集页数控制时使用
	Keyword      string  // 如需使用必须附初始值为常量USE

	proxys    []string // 代理服务器列表 example='localhost:80'
	currProxy int      // 当前服务器索引

	// 命名空间相对于数据库名，不依赖具体数据内容，可选
	Namespace func(*Spider) string
	// 子命名空间相对于表名，可依赖具体数据内容，可选
	SubNamespace func(self *Spider, dataCell map[string]interface{}) string
}

// 生成并添加请求至队列
// Request.Url与Request.Rule必须设置
// Request.Spider无需手动设置(由系统自动设置)
// Request.EnableCookie在Spider字段中统一设置，规则请求中指定的无效
// 以下字段有默认值，可不设置:
// Request.Method默认为GET方法;
// Request.DialTimeout默认为常量context.DefaultDialTimeout，小于0时不限制等待响应时长;
// Request.ConnTimeout默认为常量context.DefaultConnTimeout，小于0时不限制下载超时;
// Request.TryTimes默认为常量context.DefaultTryTimes，小于0时不限制失败重载次数;
// Request.RedirectTimes默认不限制重定向次数，小于0时可禁止重定向跳转;
// Request.RetryPause默认为常量context.DefaultRetryPause;
// Request.DownloaderID指定下载器ID，0为默认的Surf高并发下载器，功能完备，1为PhantomJS下载器，特点破防力强，速度慢，低并发。
func (self *Spider) AddQueue(req *context.Request) {
	req.
		SetSpiderName(self.Name).
		SetSpiderId(self.GetId()).
		SetEnableCookie(self.EnableCookie).
		Prepare()
	scheduler.Sdl.Push(req)
}

// 批量url生成请求，并添加至队列
func (self *Spider) BulkAddQueue(urls []string, req *context.Request) {
	for _, url := range urls {
		req.SetUrl(url)
		self.AddQueue(req)
	}
}

// 输出文本结果
// item允许的类型为map[int]interface{}或map[string]interface{}
func (self *Spider) Output(ruleName string, resp *context.Response, item interface{}) {
	resp.SetRuleName(ruleName)
	switch item2 := item.(type) {
	case map[int]interface{}:
		resp.AddItem(self.CreatItem(ruleName, item2))
	case map[string]interface{}:
		resp.AddItem(item2)
	}
}

// 输出文件结果
func (self *Spider) FileOutput(resp *context.Response, name ...string) {
	resp.AddFile(name...)
}

// 生成文本结果
func (self *Spider) CreatItem(ruleName string, item map[int]interface{}) map[string]interface{} {
	var item2 = make(map[string]interface{}, len(item))
	for k, v := range item {
		item2[self.IndexOutFeild(ruleName, k)] = v
	}
	return item2
}

// 添加自身到蜘蛛菜单
func (self *Spider) Register() {
	Menu.Add(self)
}

// 调用指定Rule下辅助函数AidFunc()
func (self *Spider) Aid(ruleName string, aid map[string]interface{}) interface{} {
	return self.GetRule(ruleName).AidFunc(self, aid)
}

// 指定ruleName时，调用相应ParseFunc()解析响应流
// 未指定ruleName时或ruleName为空时，调用Root()
func (self *Spider) Parse(resp *context.Response, ruleName ...string) {
	if len(ruleName) == 0 || ruleName[0] == "" {
		if resp != nil {
			resp.SetRuleName("")
		}
		self.RuleTree.Root(self, resp)
		return
	}

	resp.SetRuleName(ruleName[0])
	self.GetRule(ruleName[0]).ParseFunc(self, resp)
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

// 设置是否启用cookie
func (self *Spider) SetEnableCookie(enableCookie bool) {
	self.EnableCookie = enableCookie
}

// 开始执行蜘蛛
func (self *Spider) Start(sp *Spider) {
	sp.RuleTree.Root(sp, nil)
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
	Root func(*Spider, *context.Response)
	// 执行解析过程（树干）
	Trunk map[string]*Rule
}

// 采集规则单元
type Rule struct {
	//输出字段
	OutFeild []string
	// 内容解析函数
	ParseFunc func(*Spider, *context.Response)
	// 通用辅助函数
	AidFunc func(*Spider, map[string]interface{}) interface{}
}

// 获取全部输出字段
func (self *Rule) GetOutFeild() []string {
	return self.OutFeild
}

// 追加输出字段
func (self *Rule) AddOutFeild(feild string) {
	self.OutFeild = append(self.OutFeild, feild)
}
