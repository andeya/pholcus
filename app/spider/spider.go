package spider

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/app/scheduler"
	"github.com/henrylee2cn/pholcus/logs"
)

const (
	USE = "\r\t\n" //若使用Keyword，则Keyword初始值必须为USE
)

// 蜘蛛规则
type Spider struct {
	Id   int    // 所在SpiderList的下标编号，系统自动分配
	Name string // 必须保证全局唯一
	*RuleTree

	//以下为可选成员
	Description  string
	Pausetime    [2]uint // 暂停区间Pausetime[0]~Pausetime[0]+Pausetime[1]
	MaxPage      int     // UI传参而来，可在涉及采集页数控制时使用
	Keyword      string  // 如需使用必须附初始值为常量USE
	EnableCookie bool    //是否启用cookie记录

	proxys    []string // 代理服务器列表 example='localhost:80'
	currProxy int      //当前服务器索引
}

// 生成并添加请求至队列
// Request.Url与Request.Rule必须设置
// Request.Spider无需手动设置(由系统自动设置)
// Request.EnableCookie在Spider字段中统一设置，规则请求中指定的无效
// 其他字段可选，其中Request.Deadline<0时不限制下载超时，Request.RedirectTimes<0时可禁止重定向跳转，Request.RedirectTimes==0时不限制重定向次数
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

// 调用指定Rule下解析函数ParseFunc()，解析响应流
func (self *Spider) Parse(ruleName string, resp *context.Response) {
	resp.SetRuleName(ruleName)
	self.ExecParse(resp)
}

// 调用指定Rule下辅助函数AidFunc()
func (self *Spider) Aid(ruleName string, aid map[string]interface{}) interface{} {
	rule := self.RuleTree.Trunk[ruleName]
	return rule.AidFunc(self, aid)
}

// 获取Rule采集语义字段
// respOrRuleName接受*Response或string两种类型，为*Response类型时指定当前Rule
func (self *Spider) OutFeild(respOrRuleName interface{}, index int) string {
	var ruleName string
	switch rn := respOrRuleName.(type) {
	case *context.Response:
		ruleName = rn.GetRuleName()
	case string:
		ruleName = rn
	default:
		logs.Log.Error("error：参数 %v 的类型应为*Response或string！", respOrRuleName)
		return ""
	}
	return self.RuleTree.Trunk[ruleName].OutFeild[index]
}

// 为指定Rule动态追加采集语义字段，速度不如静态字段快
// respOrRuleName接受*Response或string两种类型，为*Response类型时指定当前Rule
func (self *Spider) AddOutFeild(respOrRuleName interface{}, feild string) {
	var ruleName string
	switch rn := respOrRuleName.(type) {
	case *context.Response:
		ruleName = rn.GetRuleName()
	case string:
		ruleName = rn
	default:
		logs.Log.Error("error：参数 %v 的类型应为*Response或string！", respOrRuleName)
		return
	}
	for _, v := range self.RuleTree.Trunk[ruleName].OutFeild {
		if v == feild {
			return
		}
	}
	self.RuleTree.Trunk[ruleName].AddOutFeild(feild)
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
	sp.RuleTree.Root(sp)
}

// 返回一个自身的复制品
func (self *Spider) Gost() *Spider {
	gost := &Spider{}
	gost.Id = self.Id
	gost.Name = self.Name
	gost.Description = self.Description
	gost.Pausetime = self.Pausetime
	gost.EnableCookie = self.EnableCookie
	gost.MaxPage = self.MaxPage
	gost.Keyword = self.Keyword
	gost.proxys = self.proxys
	gost.RuleTree = &RuleTree{
		Root:  self.Root,
		Trunk: map[string]*Rule{},
	}

	for k, v := range self.RuleTree.Trunk {
		nv := *v
		gost.RuleTree.Trunk[k] = &nv
		gost.RuleTree.Trunk[k].OutFeild = self.RuleTree.Trunk[k].OutFeild
	}
	return gost
}

// 根据响应流运行指定解析Rule，仅用于crawl模块，Rule中请使用Parse()代替
func (self *Spider) ExecParse(resp *context.Response) {
	self.RuleTree.Trunk[resp.GetRuleName()].ParseFunc(self, resp)
}

// 添加自身到蜘蛛菜单
func (self *Spider) Register() {
	Menu.Add(self)
}

//采集规则树
type RuleTree struct {
	// 执行入口（树根）
	Root func(*Spider)
	// 执行解析过程（树干）
	Trunk map[string]*Rule
}

// 采集规则单元
type Rule struct {
	//输出字段，注意：有无该字段与是否输出须保持一致
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
