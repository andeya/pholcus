package spider

import (
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/app/scheduler"
	"github.com/henrylee2cn/pholcus/logs"
)

const (
	USE = " " //注意必须为空格
)

// 蜘蛛规则
type Spider struct {
	Id          int    // 所在SpiderList的下标编号，系统自动分配
	Name        string // 必须保证全局唯一
	Description string
	Pausetime   [2]uint // 暂停区间Pausetime[0]~Pausetime[0]+Pausetime[1]
	*RuleTree
	//以下为可选成员
	MaxPage   int    // UI传参而来，可在涉及采集页数控制时使用
	Keyword   string // 如需使用必须附初始值为常量USE
	UseCookie bool   // 控制下载器运行模式，true:支持登录功能，false:支持大量UserAgent随机轮换
	Proxy     string // 代理服务器 example='localhost:80'
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
	gost.MaxPage = self.MaxPage
	gost.Keyword = self.Keyword
	gost.UseCookie = self.UseCookie
	gost.Proxy = self.Proxy
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

// 添加自身到蜘蛛菜单
func (self *Spider) AddMenu() {
	Menu.Add(self)
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

// 设置暂停时间 pause[0]~(pause[0]+pause[1])
// 当且仅当runtime[0]为true时可覆盖现有值
func (self *Spider) SetPausetime(pause [2]uint, runtime ...bool) {
	if self.Pausetime == [2]uint{} || len(runtime) > 0 && runtime[0] {
		self.Pausetime = pause
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

// 批量url生成请求，并添加至队列
func (self *Spider) BulkAddQueue(urls []string, param map[string]interface{}) {
	for _, url := range urls {
		param["Url"] = url
		self.AddQueue(param)
	}
}

// 生成并添加请求至队列
func (self *Spider) AddQueue(param map[string]interface{}) {
	req := self.newRequest(param)
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

func (self *Spider) newRequest(param map[string]interface{}) *context.Request {
	param["Spider"] = self.GetName()
	req := context.NewRequest(param)
	req.SetSpiderId(self.GetId())
	return req
}

// 根据响应流运行指定解析Rule，仅用于crawl模块，Rule中请使用Parse()代替
func (self *Spider) ExecParse(resp *context.Response) {
	self.RuleTree.Trunk[resp.GetRuleName()].ParseFunc(self, resp)
}
